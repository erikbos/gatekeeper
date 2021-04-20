package main

import (
	"fmt"
	"strings"
	"time"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_ratelimit "github.com/envoyproxy/go-control-plane/envoy/config/ratelimit/v3"
	envoy_extention_fileaccesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	envoy_extention_grpcaccesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	envoy_filter_extauthz "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoy_filter_ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ratelimit/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// Default HttpProtocolOptions idle timeout, as the period in which there are no active requests
	listenerIdleTimeout = 5 * time.Minute

	// ExtAuthZ Authentication timeout
	defaultAuthenticationTimeout = 10 * time.Millisecond

	// Ratelimiting timeout
	defaultRateLimitingTimeout = 10 * time.Millisecond

	// Default buffer size for accesslogging via grpc
	// (see https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#extensions-access-loggers-grpc-v3-httpgrpcaccesslogconfig)
	accessLogBufferSizeDefault = 16384
)

// getEnvoyListenerConfig returns array of envoy listeners
func (s *server) getEnvoyListenerConfig() ([]cache.Resource, error) {
	envoyListeners := []cache.Resource{}

	uniquePorts := s.getListenerPorts()
	for port := range uniquePorts {
		s.logger.Info("Compiling configuration", zap.Uint32("port", port))
		envoyListeners = append(envoyListeners, s.buildEnvoyListenerConfig(port))
	}
	return envoyListeners, nil
}

// getListenerPorts return unique set of ports from vhost configuration
func (s *server) getListenerPorts() map[uint32]bool {

	listenerPorts := map[uint32]bool{}
	for _, listener := range s.dbentities.GetListeners() {
		listenerPorts[uint32(listener.Port)] = true
	}
	return listenerPorts
}

// getVhostsInRouteGroup return all vhosts in a route group
func (s *server) getVhostsInRouteGroup(routeGroupName string) []string {
	var VhostsInRouteGroup []string

	for _, listener := range s.dbentities.GetListeners() {
		if listener.RouteGroup == routeGroupName {
			VhostsInRouteGroup = append(VhostsInRouteGroup, listener.VirtualHosts...)
		}
	}
	return VhostsInRouteGroup
}

func (s *server) buildEnvoyListenerConfig(port uint32) *envoy_listener.Listener {

	envoyListener := &envoy_listener.Listener{
		Name:            fmt.Sprintf("port_%d", port),
		Address:         buildAddress("0.0.0.0", port),
		ListenerFilters: buildListenerFilterHTTP(),
	}

	// add all vhosts belonging to this listener's port
	for _, configuredListener := range s.dbentities.GetListeners() {
		if configuredListener.Port == int(port) {
			envoyListener.FilterChains = append(envoyListener.FilterChains,
				s.buildFilterChainEntry(configuredListener, envoyListener))

			TLSEnabled := configuredListener.Attributes.GetAsString(types.AttributeTLS, "")
			if TLSEnabled == types.AttributeValueTrue {
				// Enable TLS protocol on listener
				envoyListener.ListenerFilters = []*envoy_listener.ListenerFilter{
					{
						Name: wellknown.TlsInspector,
					},
				}
			}
			// In case of a second non TLS listener on the same port we stop as
			// we can add only one entry in the filter chain match for non-TLS
			//
			// as a result the setting attributes of the second (or third) listener on same port
			// will be ignored: all share the settings of the first configured listener
			if TLSEnabled != types.AttributeValueTrue {
				s.logger.Error("Cannot add listener, already one active on port",
					zap.Uint32("port", port),
					zap.String("listener", configuredListener.Name))
				break
			}
		}
	}

	return envoyListener
}

func buildListenerFilterHTTP() []*envoy_listener.ListenerFilter {

	return []*envoy_listener.ListenerFilter{
		{
			Name: wellknown.HttpInspector,
		},
	}
}

func (s *server) buildFilterChainEntry(v types.Listener, configuredListener *envoy_listener.Listener) *envoy_listener.FilterChain {

	manager := s.buildConnectionManager(v)
	managerProtoBuf, err := anypb.New(manager)
	if err != nil {
		s.logger.Panic("buildFilterChainEntry", zap.Error(err))
	}

	FilterChainEntry := &envoy_listener.FilterChain{
		Filters: []*envoy_listener.Filter{{
			Name: wellknown.HTTPConnectionManager,
			ConfigType: &envoy_listener.Filter_TypedConfig{
				TypedConfig: managerProtoBuf,
			},
		}},
	}

	// Is TLS-enabled set to true?
	value := v.Attributes.GetAsString(types.AttributeTLS, "")
	if value != types.AttributeValueTrue {
		return FilterChainEntry
	}

	// Check if we have a certificate and certificate key
	_, certificateError := v.Attributes.Get(types.AttributeTLSCertificate)
	_, certificateKeyError := v.Attributes.Get(types.AttributeTLSCertificateKey)

	// No certificate details, return and do not try to enable TLS
	if certificateError != nil && certificateKeyError != nil {
		return FilterChainEntry
	}

	// Configure listener to use SNI to match against vhost names
	FilterChainEntry.FilterChainMatch =
		&envoy_listener.FilterChainMatch{
			ServerNames: v.VirtualHosts,
		}

	// Set TLS configuration based upon listeners attributes
	downStreamTLSConfig := &envoy_tls.DownstreamTlsContext{
		CommonTlsContext: buildCommonTLSContext(v.Name, v.Attributes),
	}
	FilterChainEntry.TransportSocket = buildTransportSocket(v.Name, downStreamTLSConfig)

	return FilterChainEntry
}

func (s *server) buildConnectionManager(listener types.Listener) *envoy_hcm.HttpConnectionManager {

	connectionManager := &envoy_hcm.HttpConnectionManager{
		CodecType:                 envoy_hcm.HttpConnectionManager_AUTO,
		StatPrefix:                "ingress_http",
		UseRemoteAddress:          protoBool(true),
		HttpFilters:               s.buildFilter(listener),
		RouteSpecifier:            s.buildRouteSpecifierRDS(listener.RouteGroup),
		AccessLog:                 s.buildAccessLog(listener),
		CommonHttpProtocolOptions: listenerCommonHTTPProtocolOptions(listener),
		Http2ProtocolOptions:      buildHTTP2ProtocolOptions(listener),
		// LocalReplyConfig:          buildLocalOverWrite(listener),
	}

	// Override Server response header
	if serverName, err := listener.Attributes.Get(types.AttributeServerName); err == nil {
		connectionManager.ServerName = serverName
	}

	return connectionManager
}

func (s *server) buildFilter(listener types.Listener) []*envoy_hcm.HttpFilter {

	httpFilter := make([]*envoy_hcm.HttpFilter, 0, 10)

	if filters, err := listener.Attributes.Get(types.AttributeListenerFilters); err == nil {
		for _, filter := range strings.Split(filters, ",") {
			switch filter {
			case wellknown.HTTPExternalAuthorization:
				if extAuthz := s.buildHTTPFilterExtAuthzConfig(listener); extAuthz != nil {
					httpFilter = append(httpFilter, &envoy_hcm.HttpFilter{
						Name: wellknown.HTTPExternalAuthorization,
						ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
							TypedConfig: extAuthz,
						},
					})
				}

			case wellknown.CORS:
				httpFilter = append(httpFilter, &envoy_hcm.HttpFilter{
					Name: wellknown.CORS,
				})

			case wellknown.HTTPRateLimit:
				if ratelimiter := s.buildHTTPFilterRateLimiterConfig(listener); ratelimiter != nil {
					httpFilter = append(httpFilter, &envoy_hcm.HttpFilter{
						Name: wellknown.HTTPRateLimit,
						ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
							TypedConfig: ratelimiter,
						},
					})
				}
			}
		}
	}

	// Always add http router filter as last, as we want to forward traffic
	httpFilter = append(httpFilter, &envoy_hcm.HttpFilter{
		Name: wellknown.Router,
	})

	return httpFilter
}

func (s *server) buildHTTPFilterExtAuthzConfig(listener types.Listener) *anypb.Any {

	cluster, err := listener.Attributes.Get(types.AttributeExtAuthzCluster)
	if err != nil || cluster == "" {
		return nil
	}
	timeout := listener.Attributes.GetAsDuration(types.AttributeExtAuthzTimeout,
		defaultAuthenticationTimeout)

	var failureModeAllow bool
	val, err := listener.Attributes.Get(types.AttributeExtAuthzFailureModeAllow)
	if err == nil && val == types.AttributeValueTrue {
		failureModeAllow = true
	}

	extAuthz := &envoy_filter_extauthz.ExtAuthz{
		FailureModeAllow: failureModeAllow,
		Services: &envoy_filter_extauthz.ExtAuthz_GrpcService{
			GrpcService: buildGRPCService(cluster, timeout),
		},
		TransportApiVersion: envoy_core.ApiVersion_V3,
	}

	requestBodySize := listener.Attributes.GetAsUInt32(types.AttributeExtAuthzRequestBodySize, 0)
	if requestBodySize > 0 {
		extAuthz.WithRequestBody = &envoy_filter_extauthz.BufferSettings{
			MaxRequestBytes:     uint32(requestBodySize),
			AllowPartialMessage: false,
		}
	}

	extAuthzTypedConf, e := anypb.New(extAuthz)
	if e != nil {
		s.logger.Panic("buildHTTPFilterExtAuthzConfig", zap.Error(err))
	}
	return extAuthzTypedConf
}

func (s *server) buildHTTPFilterRateLimiterConfig(listener types.Listener) *anypb.Any {

	cluster, err := listener.Attributes.Get(types.AttributeRateLimitingCluster)
	if err != nil || cluster == "" {
		return nil
	}
	timeout := listener.Attributes.GetAsDuration(types.AttributeRateLimitingTimeout,
		defaultRateLimitingTimeout)

	domain := listener.Attributes.GetAsString(types.AttributeRateLimitingDomain, "")

	var failureModeAllow bool
	val, err := listener.Attributes.Get(types.AttributeRateLimitingFailureModeAllow)
	if err == nil && val == types.AttributeValueTrue {
		failureModeAllow = true
	}

	ratelimit := &envoy_filter_ratelimit.RateLimit{
		Domain:          domain,
		Stage:           0,
		FailureModeDeny: failureModeAllow,
		Timeout:         durationpb.New(timeout),
		RateLimitService: &envoy_ratelimit.RateLimitServiceConfig{
			GrpcService:         buildGRPCService(cluster, timeout),
			TransportApiVersion: envoy_core.ApiVersion_V3,
		},
	}

	ratelimitTypedConf, e := anypb.New(ratelimit)
	if e != nil {
		s.logger.Panic("buildHTTPFilterExtAuthzConfig", zap.Error(err))
	}
	return ratelimitTypedConf
}

func (s *server) buildRouteSpecifierRDS(routeGroup string) *envoy_hcm.HttpConnectionManager_Rds {

	if routeGroup == "" || s.config.XDS.Cluster == "" {
		return nil
	}

	return &envoy_hcm.HttpConnectionManager_Rds{
		Rds: &envoy_hcm.Rds{
			RouteConfigName: routeGroup,
			ConfigSource:    buildConfigSource(s.config.XDS.Cluster, s.config.XDS.Timeout),
		},
	}
}

func (s *server) buildAccessLog(listener types.Listener) []*envoy_accesslog.AccessLog {

	accessLog := make([]*envoy_accesslog.AccessLog, 0, 10)

	// Set up access logging to file, in case we have a filename
	accessLogFile, error := listener.Attributes.Get(types.AttributeAccessLogFile)
	accessLogFileFields, error2 := listener.Attributes.Get(types.AttributeAccessLogFileFields)
	if error == nil && accessLogFile != "" &&
		error2 == nil && accessLogFileFields != "" {
		accessLog = append(accessLog, s.buildFileAccessLog(accessLogFile, accessLogFileFields))
	}

	// Set up access logging to cluster, in case we have a cluster name
	accessLogCluster, error := listener.Attributes.Get(types.AttributeAccessLogCluster)
	if error == nil && accessLogCluster != "" {
		// Get the buffer size so envoy can cache in memory
		accessLogClusterBuffer := listener.Attributes.GetAsUInt32(
			types.AttributeAccessLogClusterBufferSize, accessLogBufferSizeDefault)

		accessLog = append(accessLog, s.buildGRPCAccessLog(accessLogCluster, listener.Name,
			types.DefaultClusterConnectTimeout, accessLogClusterBuffer))
	}
	return accessLog
}

func (s *server) buildFileAccessLog(path, fields string) *envoy_accesslog.AccessLog {

	jsonFormat := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}
	for _, fieldFormat := range strings.Split(fields, ",") {
		fieldConfig := strings.Split(fieldFormat, "=")
		if len(fieldConfig) == 2 {
			key := strings.TrimSpace(fieldConfig[0])
			value := strings.TrimSpace(fieldConfig[1])

			jsonFormat.Fields[key] = &structpb.Value{
				Kind: &structpb.Value_StringValue{
					StringValue: value,
				},
			}
		}
	}
	accessLogConf := &envoy_extention_fileaccesslog.FileAccessLog{
		Path: path,
		AccessLogFormat: &envoy_extention_fileaccesslog.FileAccessLog_LogFormat{
			LogFormat: &envoy_core.SubstitutionFormatString{
				Format: &envoy_core.SubstitutionFormatString_JsonFormat{
					JsonFormat: jsonFormat,
				},
			},
		},
	}
	accessLogTypedConf, err := anypb.New(accessLogConf)
	if err != nil {
		s.logger.Panic("buildFileAccessLog", zap.Error(err))
	}
	return &envoy_accesslog.AccessLog{
		Name: wellknown.FileAccessLog,
		ConfigType: &envoy_accesslog.AccessLog_TypedConfig{
			TypedConfig: accessLogTypedConf,
		},
	}
}

func (s *server) buildGRPCAccessLog(clusterName, logName string,
	timeout time.Duration, bufferSize uint32) *envoy_accesslog.AccessLog {

	accessLogConf := &envoy_extention_grpcaccesslog.HttpGrpcAccessLogConfig{
		CommonConfig: &envoy_extention_grpcaccesslog.CommonGrpcAccessLogConfig{
			LogName:             logName,
			GrpcService:         buildGRPCService(clusterName, timeout),
			TransportApiVersion: envoy_core.ApiVersion_V3,
			BufferSizeBytes:     protoUint32orNil(bufferSize),
		},
	}

	accessLogTypedConf, err := anypb.New(accessLogConf)
	if err != nil {
		s.logger.Panic("buildGRPCAccessLog", zap.Error(err))
	}
	return &envoy_accesslog.AccessLog{
		Name: wellknown.HTTPGRPCAccessLog,
		ConfigType: &envoy_accesslog.AccessLog_TypedConfig{
			TypedConfig: accessLogTypedConf,
		},
	}
}

func listenerCommonHTTPProtocolOptions(listener types.Listener) *envoy_core.HttpProtocolOptions {

	idleTimeout := listener.Attributes.GetAsDuration(
		types.AttributeIdleTimeout, listenerIdleTimeout)

	return &envoy_core.HttpProtocolOptions{
		IdleTimeout: durationpb.New(idleTimeout),
	}
}

func buildHTTP2ProtocolOptions(listener types.Listener) *envoy_core.Http2ProtocolOptions {

	maxConcurrentStreams := listener.Attributes.GetAsUInt32(types.AttributeMaxConcurrentStreams, 0)
	initialConnectionWindowSize := listener.Attributes.GetAsUInt32(types.AttributeInitialConnectionWindowSize, 0)
	initialStreamWindowSize := listener.Attributes.GetAsUInt32(types.AttributeInitialStreamWindowSize, 0)

	return &envoy_core.Http2ProtocolOptions{
		MaxConcurrentStreams:        protoUint32orNil(maxConcurrentStreams),
		InitialConnectionWindowSize: protoUint32orNil(initialConnectionWindowSize),
		InitialStreamWindowSize:     protoUint32orNil(initialStreamWindowSize),
	}
}

// buildLocalOverWrite generates all the local rewrites Envoyproxy should do
// func buildLocalOverWrite(vhost types.Listener) *hcm.LocalReplyConfig {

// 	return &hcm.LocalReplyConfig{
// 		Mappers: []*hcm.ResponseMapper{
// 			buildLocalOverWrite429to403(vhost),
// 		},
// 	}
// }

// func buildLocalOverWrite429to403(vhost types.Listener) *hcm.ResponseMapper {

// 	// This matches on
// 	// 1) response status code 429
// 	// 2) metadata path "envoy.filters.http.ext_authz" (wellknown.HTTPExternalAuthorization)
// 	// 3) checks for presence of key "rl429to403"
// 	// 4) set respons status code to 403
// 	return &hcm.ResponseMapper{
// 		Filter: &accesslog.AccessLogFilter{
// 			FilterSpecifier: &accesslog.AccessLogFilter_AndFilter{
// 				AndFilter: &accesslog.AndFilter{
// 					Filters: []*accesslog.AccessLogFilter{
// 						{
// 							FilterSpecifier: &accesslog.AccessLogFilter_StatusCodeFilter{
// 								StatusCodeFilter: &accesslog.StatusCodeFilter{
// 									Comparison: &accesslog.ComparisonFilter{
// 										Op: accesslog.ComparisonFilter_EQ,
// 										Value: &core.RuntimeUInt32{
// 											DefaultValue: 429,
// 											RuntimeKey:   "rl429to403",
// 										},
// 									},
// 								},
// 							},
// 						},
// 						{
// 							FilterSpecifier: &accesslog.AccessLogFilter_MetadataFilter{
// 								MetadataFilter: &accesslog.MetadataFilter{
// 									Matcher: &envoymatcher.MetadataMatcher{
// 										Filter: wellknown.HTTPExternalAuthorization,
// 										Path: []*envoymatcher.MetadataMatcher_PathSegment{
// 											{
// 												Segment: &envoymatcher.MetadataMatcher_PathSegment_Key{
// 													Key: "rl429to403",
// 												},
// 											},
// 										},
// 										Value: &envoymatcher.ValueMatcher{
// 											MatchPattern: &envoymatcher.ValueMatcher_PresentMatch{
// 												PresentMatch: true,
// 											},
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		StatusCode: protoUint32(403),
// 	}
// }
