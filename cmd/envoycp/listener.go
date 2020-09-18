package main

import (
	"fmt"
	"time"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	ratelimitconf "github.com/envoyproxy/go-control-plane/envoy/config/ratelimit/v3"
	fileaccesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	grpcaccesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	extauthz "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ratelimit/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoymatcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// getEnvoyListenerConfig returns array of envoy listeners
func (s *server) getEnvoyListenerConfig() ([]cache.Resource, error) {
	envoyListeners := []cache.Resource{}

	uniquePorts := s.getListenerPorts()
	for port := range uniquePorts {
		log.Infof("XDS adding listener & vhosts on port %d", port)
		envoyListeners = append(envoyListeners, s.buildEnvoyListenerConfig(port))
	}
	return envoyListeners, nil
}

// getListenerPorts return unique set of ports from vhost configuration
func (s *server) getListenerPorts() map[int]bool {
	listenerPorts := map[int]bool{}
	for _, listener := range s.dbentities.GetListeners() {
		listenerPorts[listener.Port] = true
	}
	return listenerPorts
}

// getListenerPorts return unique set of ports from vhost configuration
func (s *server) getListenersOfRouteGroup(RouteGroupName string) []string {
	var ListenersInRouteGroup []string

	for _, listener := range s.dbentities.GetListeners() {
		if err := listener.ConfigCheck(); err != nil {
			log.Warningf("Listener '%s' %s", listener.Name, err)
		}

		if listener.RouteGroup == RouteGroupName {
			ListenersInRouteGroup = append(ListenersInRouteGroup, listener.VirtualHosts...)
		}
	}
	return ListenersInRouteGroup
}

func (s *server) buildEnvoyListenerConfig(port int) *listener.Listener {

	newListener := &listener.Listener{
		Name:            fmt.Sprintf("port_%d", port),
		Address:         buildAddress("0.0.0.0", port),
		ListenerFilters: buildListenerFilterHTTP(),
	}

	// add all vhosts belonging to this listener's port
	for _, listener := range s.dbentities.GetListeners() {
		if listener.Port == port {
			newListener.FilterChains = append(newListener.FilterChains, s.buildFilterChainEntry(newListener, listener))
		}
	}

	return newListener
}

func buildListenerFilterHTTP() []*listener.ListenerFilter {

	return []*listener.ListenerFilter{
		{
			Name: wellknown.HttpInspector,
		},
	}
}

func (s *server) buildFilterChainEntry(l *listener.Listener, v types.Listener) *listener.FilterChain {

	httpFilter := s.buildFilter()

	manager := s.buildConnectionManager(httpFilter, v)
	managerProtoBuf, err := ptypes.MarshalAny(manager)
	if err != nil {
		log.Panic(err)
	}

	FilterChainEntry := &listener.FilterChain{
		Filters: []*listener.Filter{{
			Name: wellknown.HTTPConnectionManager,
			ConfigType: &listener.Filter_TypedConfig{
				TypedConfig: managerProtoBuf,
			},
		}},
	}

	// Check if we have a certificate and certificate key
	_, certificateError := v.Attributes.Get(types.AttributeTLSCertificate)
	_, certificateKeyError := v.Attributes.Get(types.AttributeTLSCertificateKey)

	// No certificate details, return and do not try to enable TLS
	if certificateError != nil && certificateKeyError != nil {
		return FilterChainEntry
	}

	// Enable TLS protocol on listener
	l.ListenerFilters = []*listener.ListenerFilter{
		{
			Name: wellknown.TlsInspector,
		},
	}

	// Configure listener to use SNI to match against vhost names
	FilterChainEntry.FilterChainMatch =
		&listener.FilterChainMatch{
			ServerNames: v.VirtualHosts,
		}

	// Set TLS configuration based upon listeners attributes
	downStreamTLSConfig := &tls.DownstreamTlsContext{
		CommonTlsContext: buildCommonTLSContext(v.Name, v.Attributes),
	}
	FilterChainEntry.TransportSocket = buildTransportSocket(v.Name, downStreamTLSConfig)

	return FilterChainEntry
}

func (s *server) buildConnectionManager(httpFilters []*hcm.HttpFilter,
	vhost types.Listener) *hcm.HttpConnectionManager {

	proxyConfig := &s.config.Envoyproxy

	connectionManager := &hcm.HttpConnectionManager{
		CodecType:        hcm.HttpConnectionManager_AUTO,
		StatPrefix:       "ingress_http",
		UseRemoteAddress: protoBool(true),
		HttpFilters:      httpFilters,
		RouteSpecifier:   s.buildRouteSpecifierRDS(vhost.RouteGroup),
		AccessLog:        buildAccessLog(proxyConfig.Logging, vhost),

		CommonHttpProtocolOptions: &core.HttpProtocolOptions{
			IdleTimeout: ptypes.DurationProto(proxyConfig.Misc.HTTPIdleTimeout),
		},
		Http2ProtocolOptions: &core.Http2ProtocolOptions{
			MaxConcurrentStreams:        protoUint32orNil(proxyConfig.Misc.MaxConcurrentStreams),
			InitialConnectionWindowSize: protoUint32orNil(proxyConfig.Misc.InitialConnectionWindowSize),
			InitialStreamWindowSize:     protoUint32orNil(proxyConfig.Misc.InitialStreamWindowSize),
		},
		LocalReplyConfig: buildLocalOverWrite(vhost),
	}
	if proxyConfig.Misc.ServerName != "" {
		connectionManager.ServerName = proxyConfig.Misc.ServerName
	}

	return connectionManager
}

func (s *server) buildFilter() []*hcm.HttpFilter {

	httpFilter := make([]*hcm.HttpFilter, 0, 10)

	if extAuthz := s.buildHTTPFilterExtAuthzConfig(); extAuthz != nil {
		httpFilter = append(httpFilter, &hcm.HttpFilter{
			Name: wellknown.HTTPExternalAuthorization,
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: extAuthz,
			},
		})
	}

	httpFilter = append(httpFilter, &hcm.HttpFilter{
		Name: wellknown.CORS,
	})

	if ratelimiter := s.buildHTTPFilterRateLimiterConfig(); ratelimiter != nil {
		httpFilter = append(httpFilter, &hcm.HttpFilter{
			Name: wellknown.HTTPRateLimit,
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: ratelimiter,
			},
		})
	}

	httpFilter = append(httpFilter, &hcm.HttpFilter{
		Name: wellknown.Router,
	})

	return httpFilter
}

func (s *server) buildHTTPFilterExtAuthzConfig() *anypb.Any {

	if !s.config.Envoyproxy.ExtAuthz.Enable ||
		s.config.Envoyproxy.ExtAuthz.Cluster == "" {
		return nil
	}

	extAuthz := &extauthz.ExtAuthz{
		FailureModeAllow: s.config.Envoyproxy.ExtAuthz.FailureModeAllow,
		Services: &extauthz.ExtAuthz_GrpcService{
			GrpcService: buildGRPCService(s.config.Envoyproxy.ExtAuthz.Cluster,
				s.config.Envoyproxy.ExtAuthz.Timeout),
		},
		TransportApiVersion: core.ApiVersion_V3,
	}
	if s.config.Envoyproxy.ExtAuthz.RequestBodySize > 0 {
		extAuthz.WithRequestBody = s.extAuthzWithRequestBody()
	}

	extAuthzTypedConf, err := ptypes.MarshalAny(extAuthz)
	if err != nil {
		log.Panic(err)
	}
	return extAuthzTypedConf
}

func (s *server) buildHTTPFilterRateLimiterConfig() *anypb.Any {

	if !s.config.Envoyproxy.RateLimiter.Enable ||
		s.config.Envoyproxy.RateLimiter.Cluster == "" {
		return nil
	}

	ratelimit := &ratelimit.RateLimit{
		Domain:          s.config.Envoyproxy.RateLimiter.Domain,
		Stage:           0,
		FailureModeDeny: s.config.Envoyproxy.RateLimiter.FailureModeDeny,
		Timeout:         ptypes.DurationProto(s.config.Envoyproxy.RateLimiter.Timeout),
		RateLimitService: &ratelimitconf.RateLimitServiceConfig{
			GrpcService: buildGRPCService(s.config.Envoyproxy.RateLimiter.Cluster,
				s.config.Envoyproxy.RateLimiter.Timeout),
			TransportApiVersion: core.ApiVersion_V3,
		},
	}

	ratelimitTypedConf, err := ptypes.MarshalAny(ratelimit)
	if err != nil {
		log.Panic(err)
	}
	return ratelimitTypedConf
}

func (s *server) extAuthzWithRequestBody() *extauthz.BufferSettings {
	if s.config.Envoyproxy.ExtAuthz.RequestBodySize > 0 {
		return &extauthz.BufferSettings{
			MaxRequestBytes:     uint32(s.config.Envoyproxy.ExtAuthz.RequestBodySize),
			AllowPartialMessage: false,
		}
	}
	return nil
}

func (s *server) buildRouteSpecifierRDS(routeGroup string) *hcm.HttpConnectionManager_Rds {

	return &hcm.HttpConnectionManager_Rds{
		Rds: &hcm.Rds{
			RouteConfigName: routeGroup,
			ConfigSource:    buildConfigSource(s.config.XDS.Cluster, s.config.XDS.Timeout),
		},
	}
}

func buildAccessLog(config envoyLogConfig, v types.Listener) []*accesslog.AccessLog {

	// Set access log behaviour based upon virtual host attributes
	accessLogFile, error := v.Attributes.Get(types.AttributeAccessLogFile)
	if error == nil && accessLogFile != "" {
		return buildFileAccessLog(config.File.Fields, accessLogFile)
	}
	accessLogCluster, error := v.Attributes.Get(types.AttributeAccessLogCluster)
	if error == nil && accessLogCluster != "" {
		return buildGRPCAccessLog(accessLogCluster, v.Name, types.DefaultClusterConnectTimeout, 0)
	}

	// Fallback is default logging based upon configfile
	if config.File.LogFile != "" {
		return buildFileAccessLog(config.File.Fields, config.File.LogFile)
	}
	if config.GRPC.Cluster != "" {
		return buildGRPCAccessLog(config.GRPC.Cluster, v.Name, config.GRPC.Timeout, config.GRPC.BufferSize)
	}
	return nil
}

func buildFileAccessLog(fields map[string]string, path string) []*accesslog.AccessLog {

	if len(fields) == 0 {
		log.Warnf("To do file access logging a logfield definition in envoycp configuration is required")
		return nil
	}

	jsonFormat := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}
	for field, fieldFormat := range fields {
		if fieldFormat != "" {
			jsonFormat.Fields[field] = &structpb.Value{
				Kind: &structpb.Value_StringValue{StringValue: fieldFormat},
			}
		}
	}
	accessLogConf := &fileaccesslog.FileAccessLog{
		Path: path,
		AccessLogFormat: &fileaccesslog.FileAccessLog_LogFormat{
			LogFormat: &core.SubstitutionFormatString{
				Format: &core.SubstitutionFormatString_JsonFormat{
					JsonFormat: jsonFormat,
				},
			},
		},
	}
	accessLogTypedConf, err := ptypes.MarshalAny(accessLogConf)
	if err != nil {
		log.Panic(err)
	}
	return []*accesslog.AccessLog{{
		Name: wellknown.FileAccessLog,
		ConfigType: &accesslog.AccessLog_TypedConfig{
			TypedConfig: accessLogTypedConf,
		},
	}}
}

func buildGRPCAccessLog(clusterName, LogName string, timeout time.Duration, bufferSize uint32) []*accesslog.AccessLog {

	accessLogConf := &grpcaccesslog.HttpGrpcAccessLogConfig{
		CommonConfig: &grpcaccesslog.CommonGrpcAccessLogConfig{
			LogName:             LogName,
			GrpcService:         buildGRPCService(clusterName, timeout),
			TransportApiVersion: core.ApiVersion_V3,
			BufferSizeBytes:     protoUint32orNil(bufferSize),
		},
	}

	accessLogTypedConf, err := ptypes.MarshalAny(accessLogConf)
	if err != nil {
		log.Panic(err)
	}
	return []*accesslog.AccessLog{{
		Name: wellknown.HTTPGRPCAccessLog,
		ConfigType: &accesslog.AccessLog_TypedConfig{
			TypedConfig: accessLogTypedConf,
		},
	}}
}

// buildLocalOverWrite generates all the local rewrites Envoyproxy should do
func buildLocalOverWrite(vhost types.Listener) *hcm.LocalReplyConfig {

	return &hcm.LocalReplyConfig{
		Mappers: []*hcm.ResponseMapper{
			buildLocalOverWrite429to403(vhost),
		},
	}
}

func buildLocalOverWrite429to403(vhost types.Listener) *hcm.ResponseMapper {

	// This matches on
	// 1) response status code 429
	// 2) metadata path "envoy.filters.http.ext_authz" (wellknown.HTTPExternalAuthorization)
	// 3) checks for presence of key "rl429to403"
	return &hcm.ResponseMapper{
		Filter: &accesslog.AccessLogFilter{
			FilterSpecifier: &accesslog.AccessLogFilter_AndFilter{
				AndFilter: &accesslog.AndFilter{
					Filters: []*accesslog.AccessLogFilter{
						{
							FilterSpecifier: &accesslog.AccessLogFilter_StatusCodeFilter{
								StatusCodeFilter: &accesslog.StatusCodeFilter{
									Comparison: &accesslog.ComparisonFilter{
										Op: accesslog.ComparisonFilter_EQ,
										Value: &core.RuntimeUInt32{
											DefaultValue: 429,
											RuntimeKey:   "rl429to403",
										},
									},
								},
							},
						},
						{
							FilterSpecifier: &accesslog.AccessLogFilter_MetadataFilter{
								MetadataFilter: &accesslog.MetadataFilter{
									Matcher: &envoymatcher.MetadataMatcher{
										Filter: wellknown.HTTPExternalAuthorization,
										Path: []*envoymatcher.MetadataMatcher_PathSegment{
											{
												Segment: &envoymatcher.MetadataMatcher_PathSegment_Key{
													Key: "rl429to403",
												},
											},
										},
										Value: &envoymatcher.ValueMatcher{
											MatchPattern: &envoymatcher.ValueMatcher_PresentMatch{
												PresentMatch: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		StatusCode: protoUint32(403),
	}
}
