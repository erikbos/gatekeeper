package main

import (
	"crypto/tls"
	"net"
	"strconv"
	"strings"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	alf "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	accesslog "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erikbos/gatekeeper/cmd/accesslogserver/metrics"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Inspired by https://github.com/apigee/apigee-remote-service-envoy/blob/master/server/accesslog.go
// https://github.com/gridgentoo/kuma/tree/master/pkg/envoy/accesslog

// AccessLogServer receives access logs from the remote Envoy nodes.
type AccessLogServer struct {
	metrics           *metrics.Metrics // Metrics store
	maxStreamDuration time.Duration    // Maximum duration for a access log stream log to live
	logger            *zap.Logger      // Logger for writing access log entries
}

// AccessLogServerConfig holds our configuration
type AccessLogServerConfig struct {
	Listen            string        // Access log server listen port grpc
	Logger            shared.Logger // changelog log configuration
	MaxStreamDuration time.Duration // Max duration of access log session
}

// NewAccessLogServer returns a new AccessLogServer instance
func NewAccessLogServer(streamDuration time.Duration,
	metrics *metrics.Metrics, logger *zap.Logger) *AccessLogServer {

	return &AccessLogServer{
		maxStreamDuration: streamDuration,
		metrics:           metrics,
		logger:            logger,
	}
}

// Start runs a new access log server that listens for incoming log streaming
func (a *AccessLogServer) Start(listen string) {

	a.logger.Info("GRPC ALS listening on " + listen)
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		a.logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	accesslog.RegisterAccessLogServiceServer(grpcServer, a)
	reflection.Register(grpcServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	if err := grpcServer.Serve(lis); err != nil {
		a.logger.Fatal("failed to start GRPC server", zap.Error(err))
	}
}

// StreamAccessLogs implements the access log service.
func (a *AccessLogServer) StreamAccessLogs(
	stream accesslog.AccessLogService_StreamAccessLogsServer) error {

	// Set the time at which we will end stream session
	endTime := time.Now().Add(a.maxStreamDuration)

	var logName string
	for {
		msg, err := stream.Recv()
		if err != nil {
			return nil
		}
		if msg.Identifier != nil {
			logName = msg.Identifier.LogName
		}
		switch entries := msg.LogEntries.(type) {
		case *accesslog.StreamAccessLogsMessage_HttpLogs:
			for _, entry := range entries.HttpLogs.LogEntry {
				a.LogHTTPRequest(msg.Identifier, entry)
			}

		case *accesslog.StreamAccessLogsMessage_TcpLogs:
			a.logger.Error("tcp logging not supported",
				zap.String("logname", logName))
		}

		// close the client stream once the timeout reaches
		// we do this to force envoyproxy to reconnect and
		// rebalance access log clusters targets
		if endTime.Before(time.Now()) {
			return stream.SendAndClose(nil)
		}

	}
}

// LogHTTPRequest logs details of a single HTTP request.
func (a *AccessLogServer) LogHTTPRequest(
	i *accesslog.StreamAccessLogsMessage_Identifier,
	e *alf.HTTPAccessLogEntry) {

	// See https://www.envoyproxy.io/docs/envoy/latest/api-v3/data/accesslog/v3/accesslog.proto#envoy-v3-api-msg-data-accesslog-v3-accesslogcommon
	// for field details

	if e != nil {
		// If nil, populate so we can safely log fields
		if i == nil {
			i = &accesslog.StreamAccessLogsMessage_Identifier{}
		}
		if i.Node == nil {
			i.Node = &core.Node{}
		}
		if i.Node.Locality == nil {
			i.Node.Locality = &core.Locality{}
		}
		c := e.CommonProperties
		if c == nil {
			c = &alf.AccessLogCommon{}
		}
		if c.ResponseFlags == nil {
			c.ResponseFlags = &alf.ResponseFlags{}
		}
		requestTLSProperties := e.CommonProperties.TlsProperties
		if requestTLSProperties == nil {
			requestTLSProperties = &alf.TLSProperties{}
		}
		request := e.Request
		if request == nil {
			request = &alf.HTTPRequestProperties{}
		}
		response := e.Response
		if response == nil {
			response = &alf.HTTPResponseProperties{}
		}

		// Record latency of this entry
		timeNowUTC := time.Now().UTC()
		a.metrics.ObserveAccesLogLatency(timeNowUTC.Sub(e.CommonProperties.StartTime.AsTime()))

		a.metrics.IncAccessLogNodeHits(i.Node.Id, i.Node.Cluster)
		a.metrics.IncAccessLogVHostHits(request.Authority)

		a.logger.Info("http",
			zap.Any("envoy", map[string]string{
				"logname": i.LogName,
				"id":      i.Node.Id,
				"cluster": i.Node.Cluster,
				"region":  i.Node.Locality.Region,
				"zone":    i.Node.Locality.Zone,
			}),

			zap.Any("ts", map[string]int64{
				"downstreamstart":       timestampToUnix(c.StartTime),
				"downstreamend":         timestampAddDurationUnix(c.StartTime, c.TimeToLastRxByte),
				"upstreamstart":         timestampAddDurationUnix(c.StartTime, c.TimeToFirstUpstreamTxByte),
				"upstreamend":           timestampAddDurationUnix(c.StartTime, c.TimeToLastUpstreamTxByte),
				"upstreamreceivedstart": timestampAddDurationUnix(c.StartTime, c.TimeToFirstUpstreamRxByte),
				"upstreamreceivedend":   timestampAddDurationUnix(c.StartTime, c.TimeToLastUpstreamRxByte),
				"downstreamsentstart":   timestampAddDurationUnix(c.StartTime, c.TimeToFirstDownstreamTxByte),
				"downstreamsentend":     timestampAddDurationUnix(c.StartTime, c.TimeToLastDownstreamTxByte),
			}),

			zap.Any("request", map[string]interface{}{
				"remoteaddress": formatAddress(c.DownstreamRemoteAddress),
				"remoteport":    formatPort(c.DownstreamRemoteAddress),
				"localaddress":  formatAddress(c.DownstreamLocalAddress),
				"localport":     formatPort(c.DownstreamLocalAddress),
				"tlsversion":    requestTLSProperties.GetTlsVersion().String(),
				"tlscipher":     tls.CipherSuiteName(uint16(requestTLSProperties.GetTlsCipherSuite().GetValue())),
				"proto":         e.ProtocolVersion.String(),
				"forwardedfor":  request.ForwardedFor,
				"scheme":        request.Scheme,
				"authority":     request.Authority,
				"verb":          request.RequestMethod.String(),
				"path":          request.Path,
				"requestid":     request.RequestId,
				"headersize":    request.RequestHeadersBytes,
				"bodysize":      request.RequestBodyBytes,
			}),

			zap.Any("requestheaders", request.GetRequestHeaders()),

			zap.String("responseflags", formatResponseFlags(c.ResponseFlags)),

			zap.Any("upstream", map[string]string{
				"route":         c.RouteName,
				"cluster":       c.UpstreamCluster,
				"remoteaddress": formatAddress(c.UpstreamRemoteAddress),
				"remoteport":    formatPort(c.UpstreamRemoteAddress),
				"localaddress":  formatAddress(c.UpstreamLocalAddress),
				"localport":     formatPort(c.UpstreamLocalAddress),
				"failure":       c.UpstreamTransportFailureReason,
			}),

			zap.Any("response", map[string]uint64{
				"code":       uint64(response.ResponseCode.GetValue()),
				"headersize": response.ResponseHeadersBytes,
				"bodysize":   response.ResponseBodyBytes,
			}),

			zap.Any("responseheaders", response.GetResponseHeaders()),

			zap.Any("metadata", formatMetadata(e.CommonProperties.Metadata)),
		)
	}
}

// Returns ip address part of core.Address type as string
func formatAddress(address *core.Address) string {
	if address != nil {
		switch t := address.GetAddress().(type) {
		case *core.Address_SocketAddress:
			return t.SocketAddress.GetAddress()
		}
	}
	return ""
}

// Returns ip port part of core.Address type as string
func formatPort(address *core.Address) string {
	if address != nil {
		switch t := address.GetAddress().(type) {
		case *core.Address_SocketAddress:
			return strconv.FormatUint(uint64(t.SocketAddress.GetPortValue()), 10)
		}
	}
	return ""
}

func formatResponseFlags(flags *alf.ResponseFlags) string {

	if flags == nil {
		return ""
	}

	const (
		ResponseFlagDownstreamConnectionTermination = "DC"
		ResponseFlagFailedLocalHealthCheck          = "LH"
		ResponseFlagNoHealthyUpstream               = "UH"
		ResponseFlagUpstreamRequestTimeout          = "UT"
		ResponseFlagLocalReset                      = "LR"
		ResponseFlagUpstreamRemoteReset             = "UR"
		ResponseFlagUpstreamConnectionFailure       = "UF"
		ResponseFlagUpstreamConnectionTermination   = "UC"
		ResponseFlagUpstreamOverflow                = "UO"
		ResponseFlagUpstreamRetryLimitExceeded      = "URX"
		ResponseFlagNoRouteFound                    = "NR"
		ResponseFlagDelayInjected                   = "DI"
		ResponseFlagFaultInjected                   = "FI"
		ResponseFlagRateLimited                     = "RL"
		ResponseFlagUnauthorizedExternalService     = "UAEX"
		ResponseFlagRatelimitServiceError           = "RLSE"
		ResponseFlagStreamIdleTimeout               = "SI"
		ResponseFlagInvalidEnvoyRequestHeaders      = "IH"
		ResponseFlagDownstreamProtocolError         = "DPE"
	)

	values := make([]string, 0)
	if flags.GetFailedLocalHealthcheck() {
		values = append(values, ResponseFlagFailedLocalHealthCheck)
	}
	if flags.GetNoHealthyUpstream() {
		values = append(values, ResponseFlagNoHealthyUpstream)
	}
	if flags.GetUpstreamRequestTimeout() {
		values = append(values, ResponseFlagUpstreamRequestTimeout)
	}
	if flags.GetLocalReset() {
		values = append(values, ResponseFlagLocalReset)
	}
	if flags.GetUpstreamRemoteReset() {
		values = append(values, ResponseFlagUpstreamRemoteReset)
	}
	if flags.GetUpstreamConnectionFailure() {
		values = append(values, ResponseFlagUpstreamConnectionFailure)
	}
	if flags.GetUpstreamConnectionTermination() {
		values = append(values, ResponseFlagUpstreamConnectionTermination)
	}
	if flags.GetUpstreamOverflow() {
		values = append(values, ResponseFlagUpstreamOverflow)
	}
	if flags.GetNoRouteFound() {
		values = append(values, ResponseFlagNoRouteFound)
	}
	if flags.GetDelayInjected() {
		values = append(values, ResponseFlagDelayInjected)
	}
	if flags.GetFaultInjected() {
		values = append(values, ResponseFlagFaultInjected)
	}
	if flags.GetRateLimited() {
		values = append(values, ResponseFlagRateLimited)
	}
	if flags.GetUnauthorizedDetails().GetReason() == alf.ResponseFlags_Unauthorized_EXTERNAL_SERVICE {
		values = append(values, ResponseFlagUnauthorizedExternalService)
	}
	if flags.GetRateLimitServiceError() {
		values = append(values, ResponseFlagRatelimitServiceError)
	}
	if flags.GetDownstreamConnectionTermination() {
		values = append(values, ResponseFlagDownstreamConnectionTermination)
	}
	if flags.GetUpstreamRetryLimitExceeded() {
		values = append(values, ResponseFlagUpstreamRetryLimitExceeded)
	}
	if flags.GetStreamIdleTimeout() {
		values = append(values, ResponseFlagStreamIdleTimeout)
	}
	if flags.GetInvalidEnvoyRequestHeaders() {
		values = append(values, ResponseFlagInvalidEnvoyRequestHeaders)
	}
	if flags.GetDownstreamProtocolError() {
		values = append(values, ResponseFlagDownstreamProtocolError)
	}
	return strings.Join(values, ",")
}

func formatMetadata(m *core.Metadata) interface{} {
	if m != nil {
		return m.FilterMetadata
	}
	return ""
}

// timestampToUnix converts a protobuf time to a UNIX timestamp in milliseconds.
func timestampToUnix(ts *timestamppb.Timestamp) int64 {
	return ts.AsTime().UnixNano() / 1000000
}

func timestampAddDurationUnix(ts *timestamppb.Timestamp, d *durationpb.Duration) int64 {
	return (ts.AsTime().UnixNano() + d.AsDuration().Nanoseconds()) / 1000000
}
