package main

import (
	"fmt"
	"sync"
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

	"github.com/erikbos/gatekeeper/pkg/shared"
)

const (
	virtualHostDataRefreshInterval = 2 * time.Second
	extAuthzTimeout                = 100 * time.Millisecond
)

// FIXME this does not detect removed records
// GetVirtualHostConfigFromDatabase continuously gets the current configuration
func (s *server) GetVirtualHostConfigFromDatabase(n chan xdsNotifyMesssage) {
	var virtualHostsLastUpdate int64
	var virtualHostsMutex sync.Mutex

	for {
		var xdsPushNeeded bool

		newVirtualHosts, err := s.db.Virtualhost.GetAll()
		if err != nil {
			log.Errorf("Could not retrieve virtualhosts from database (%s)", err)
		} else {
			if virtualHostsLastUpdate == 0 {
				log.Info("Initial load of virtualhosts")
			}
			for _, virtualhost := range newVirtualHosts {
				// Is a virtualhosts updated since last time we stored it?
				if virtualhost.LastmodifiedAt > virtualHostsLastUpdate {
					virtualHostsMutex.Lock()
					s.virtualhosts = newVirtualHosts
					virtualHostsMutex.Unlock()

					virtualHostsLastUpdate = shared.GetCurrentTimeMilliseconds()
					xdsPushNeeded = true

					warnForUnknownVirtualHostAttributes(virtualhost)
				}
			}
		}
		if xdsPushNeeded {
			n <- xdsNotifyMesssage{
				resource: "virtualhost",
			}
			// Increase xds deployment metric
			s.metrics.xdsDeployments.WithLabelValues("virtualhosts").Inc()
		}
		time.Sleep(virtualHostDataRefreshInterval)
	}
}

// GetVirtualHostCount returns number of virtualhosts
func (s *server) GetVirtualHostCount() float64 {
	return float64(len(s.virtualhosts))
}

// getEnvoyListenerConfig returns array of envoy listeners
func (s *server) getEnvoyListenerConfig() ([]cache.Resource, error) {
	envoyListeners := []cache.Resource{}

	uniquePorts := s.getVirtualHostPorts()
	for port := range uniquePorts {
		log.Infof("Adding listener & vhosts on port %d", port)
		envoyListeners = append(envoyListeners, s.buildEnvoyListenerConfig(port))
	}
	return envoyListeners, nil
}

// getVirtualHostPorts return unique set of ports from vhost configuration
func (s *server) getVirtualHostPorts() map[int]bool {
	listenerPorts := map[int]bool{}
	for _, virtualhost := range s.virtualhosts {
		listenerPorts[virtualhost.Port] = true
	}
	return listenerPorts
}

// getVirtualHostPorts return unique set of ports from vhost configuration
func (s *server) getVirtualHostsOfRouteGroup(RouteGroupName string) []string {
	var virtualHostsInRouteGroup []string

	for _, virtualhost := range s.virtualhosts {
		if virtualhost.RouteGroup == RouteGroupName {
			virtualHostsInRouteGroup = append(virtualHostsInRouteGroup, virtualhost.VirtualHosts...)
		}
	}
	return virtualHostsInRouteGroup
}

func (s *server) buildEnvoyListenerConfig(port int) *listener.Listener {

	newListener := &listener.Listener{
		Name:            fmt.Sprintf("port_%d", port),
		Address:         buildAddress("0.0.0.0", port),
		ListenerFilters: buildListenerFilterHTTP(),
	}

	// add all vhosts belonging to this listener's port
	for _, virtualhost := range s.virtualhosts {
		if virtualhost.Port == port {
			newListener.FilterChains = append(newListener.FilterChains, s.buildFilterChainEntry(newListener, virtualhost))
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

func (s *server) buildFilterChainEntry(l *listener.Listener, v shared.VirtualHost) *listener.FilterChain {

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
	_, certificateError := v.Attributes.Get(attributeTLSCertificate)
	_, certificateKeyError := v.Attributes.Get(attributeTLSCertificateKey)

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

	// Set TLS configuration based upon virtualhosts attributes
	downStreamTLSConfig := &tls.DownstreamTlsContext{
		CommonTlsContext: buildCommonTLSContext(v.Name, v.Attributes),
	}
	FilterChainEntry.TransportSocket = buildTransportSocket(v.Name, downStreamTLSConfig)

	return FilterChainEntry
}

func (s *server) buildConnectionManager(httpFilters []*hcm.HttpFilter,
	vhost shared.VirtualHost) *hcm.HttpConnectionManager {

	proxyConfig := &s.config.Envoyproxy

	return &hcm.HttpConnectionManager{
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
			MaxConcurrentStreams:        protoUint32(proxyConfig.Misc.MaxConcurrentStreams),
			InitialConnectionWindowSize: protoUint32(proxyConfig.Misc.InitialConnectionWindowSize),
			InitialStreamWindowSize:     protoUint32(proxyConfig.Misc.InitialStreamWindowSize),
		},
		ServerName:       proxyConfig.Misc.ServerName,
		LocalReplyConfig: buildLocalOverWrite(vhost),
	}
}

func (s *server) buildFilter() []*hcm.HttpFilter {

	return []*hcm.HttpFilter{
		{
			Name: wellknown.HTTPExternalAuthorization,
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: s.buildExtAuthzFilterConfig(),
			},
		},
		{
			Name: wellknown.CORS,
		},
		{
			Name: wellknown.HTTPRateLimit,
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: s.buildRateLimiterFilterConfig(),
			},
		},
		{
			Name: wellknown.Router,
		},
	}
}

func (s *server) buildExtAuthzFilterConfig() *anypb.Any {

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

func (s *server) buildRateLimiterFilterConfig() *anypb.Any {

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

// Not necessary anymore as we receive routes via RDS (buildRouteSpecifierRDS)
//
// func (s *server) buildRouteSpecifier(RouteGroup string) *hcm.HttpConnectionManager_RouteConfig {

// 	return &hcm.HttpConnectionManager_RouteConfig{
// 		RouteConfig: s.buildEnvoyVirtualHostRouteConfig(RouteGroup, s.routes),
// 	}
// }

func (s *server) buildRouteSpecifierRDS(routeGroup string) *hcm.HttpConnectionManager_Rds {

	return &hcm.HttpConnectionManager_Rds{
		Rds: &hcm.Rds{
			RouteConfigName: routeGroup,
			ConfigSource:    buildConfigSource(s.config.XDS.Cluster, s.config.XDS.Timeout),
		},
	}
}

func buildAccessLog(config envoyLogConfig, v shared.VirtualHost) []*accesslog.AccessLog {

	// Set access log behaviour based upon virtual host attributes
	accessLogFileName, error := v.Attributes.Get(attributeAccessLogFileName)
	if error == nil && accessLogFileName != "" {
		return buildFileAccessLog(config.File.Fields, accessLogFileName)
	}
	accessLogClusterName, error := v.Attributes.Get(attributeAccessLogClusterName)
	if error == nil && accessLogClusterName != "" {
		return buildGRPCAccessLog(accessLogClusterName, v.Name, defaultClusterConnectTimeout, 0)
	}

	// Fallback is default logging based upon configfile
	if config.File.LogFileName != "" {
		return buildFileAccessLog(config.File.Fields, config.File.LogFileName)
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
func buildLocalOverWrite(vhost shared.VirtualHost) *hcm.LocalReplyConfig {

	return &hcm.LocalReplyConfig{
		Mappers: []*hcm.ResponseMapper{
			buildLocalOverWrite429to403(vhost),
		},
	}
}

func buildLocalOverWrite429to403(vhost shared.VirtualHost) *hcm.ResponseMapper {

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
