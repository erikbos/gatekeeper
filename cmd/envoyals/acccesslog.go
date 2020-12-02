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
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/erikbos/gatekeeper/cmd/envoyals/metrics"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Inspired by https://github.com/apigee/apigee-remote-service-envoy/blob/master/server/accesslog.go
// https://github.com/gridgentoo/kuma/tree/master/pkg/envoy/accesslog

// AccessLogServer receives access logs from the remote Envoy nodes.
type AccessLogServer struct {
	metrics           *metrics.Metrics
	maxStreamDuration time.Duration // Maximum duration for a access log stream log to live
	logger            *zap.Logger
}

// AccessLogServerConfig holds our configuration
type AccessLogServerConfig struct {
	Listen            string        `yaml:"listen"`            // Access log server listen port grpc
	Logger            shared.Logger `yaml:"logging"`           // changelog log configuration
	maxStreamDuration time.Duration `yaml:"maxstreamduration"` // Max duration of access log session
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
	a.logger.Fatal("failed to start GRPC server", zap.Error(grpcServer.Serve(lis)))
}

// StreamAccessLogs implements the access log service.
func (a *AccessLogServer) StreamAccessLogs(
	stream accesslog.AccessLogService_StreamAccessLogsServer) error {

	node := &core.Node{}
	var logName string
	for {
		msg, err := stream.Recv()
		if err != nil {
			return nil
		}
		if msg.Identifier != nil {
			node = msg.Identifier.Node
			logName = msg.Identifier.LogName
		}
		switch entries := msg.LogEntries.(type) {
		case *accesslog.StreamAccessLogsMessage_HttpLogs:
			for _, entry := range entries.HttpLogs.LogEntry {
				a.LogHTTPRequest(logName, node.Id, node.Cluster, entry)
			}

		case *accesslog.StreamAccessLogsMessage_TcpLogs:
			a.logger.Error("tcp logging not supported", zap.String("logname", logName))
		}
	}
}

// LogHTTPRequest logs details of a single HTTP request.
func (a *AccessLogServer) LogHTTPRequest(logName, nodeID, nodeCluster string, e *alf.HTTPAccessLogEntry) {

	// See https://www.envoyproxy.io/docs/envoy/latest/api-v3/data/accesslog/v3/accesslog.proto#envoy-v3-api-msg-data-accesslog-v3-accesslogcommon
	// for field details

	if e != nil {
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
		a.metrics.ObserveAccesLogLatency(timeNowUTC.Sub(a.pbTimestamp(e.CommonProperties.StartTime)))

		a.metrics.IncAccessLogNodeHits(nodeID, nodeCluster)
		a.metrics.IncAccessLogVHostHits(request.Authority)

		a.logger.Info("http",
			zap.Any("envoy", map[string]string{
				"logname": logName,
				"id":      nodeID,
				"cluster": nodeCluster,
			}),

			zap.Any("ts", map[string]int64{
				"downstreamstart":       a.pbTimestampToUnix(c.StartTime),
				"downstreamend":         a.pbTimestampAddDurationUnix(c.StartTime, c.TimeToLastRxByte),
				"upstreamstart":         a.pbTimestampAddDurationUnix(c.StartTime, c.TimeToFirstUpstreamTxByte),
				"upstreamend":           a.pbTimestampAddDurationUnix(c.StartTime, c.TimeToLastUpstreamTxByte),
				"upstreamreceivedstart": a.pbTimestampAddDurationUnix(c.StartTime, c.TimeToFirstUpstreamRxByte),
				"upstreamreceivedend":   a.pbTimestampAddDurationUnix(c.StartTime, c.TimeToLastUpstreamRxByte),
				"downstreamsentstart":   a.pbTimestampAddDurationUnix(c.StartTime, c.TimeToFirstDownstreamTxByte),
				"downstreamsentend":     a.pbTimestampAddDurationUnix(c.StartTime, c.TimeToLastDownstreamTxByte),
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

func formatAddress(address *core.Address) string {
	if address != nil {
		switch t := address.GetAddress().(type) {
		case *core.Address_SocketAddress:
			return t.SocketAddress.GetAddress()
		}
	}
	return ""
}

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

// pbTimestamp converts a protobuf time to time.Time.
func (a *AccessLogServer) pbTimestamp(ts *timestamp.Timestamp) time.Time {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		a.logger.Error("invalid timestamp", zap.Error(err))
		return time.Time{}
	}
	return t
}

// pbTimestampToUnix converts a protobuf time to a UNIX timestamp in milliseconds.
func (a *AccessLogServer) pbTimestampToUnix(ts *timestamp.Timestamp) int64 {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		a.logger.Error("invalid timestamp", zap.Error(err))
		return 0
	}
	return t.UnixNano() / 1000000
}

func (a *AccessLogServer) pbTimestampAddDurationUnix(ts *timestamp.Timestamp, d *duration.Duration) int64 {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		a.logger.Error("invalid timestamp", zap.Error(err))
		return 0
	}
	du, err := ptypes.Duration(d)
	if err != nil {
		du = 0
	}
	return t.Add(du).UnixNano() / 1000000
}
