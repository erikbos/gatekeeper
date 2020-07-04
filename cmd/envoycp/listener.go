package main

import (
	"fmt"
	"sync"
	"time"

	accessLog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	fileAccessLog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	extAuthz "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	spb "github.com/golang/protobuf/ptypes/struct"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"

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

		newVirtualHosts, err := s.db.GetVirtualHosts()
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
			Name: "envoy.filters.listener.http_inspector",
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

	// Configure TLS in case when we have a certificate + key
	certificate, error1 := shared.GetAttribute(v.Attributes, attributeTLSCertificate)
	certificateKey, error2 := shared.GetAttribute(v.Attributes, attributeTLSCertificateKey)

	// No certificate details, return and do not enable TLS
	if error1 != nil && error2 != nil {
		return FilterChainEntry
	}

	// Enable TLS protocol on listener
	l.ListenerFilters = []*listener.ListenerFilter{
		{
			Name: "envoy.filters.listener.tls_inspector",
		},
	}

	// Configure listener to be able to match SNI
	FilterChainEntry.FilterChainMatch =
		&listener.FilterChainMatch{
			ServerNames: v.VirtualHosts,
		}

	// Add TLS certifcate configuration
	downStreamTLSConfig := &tls.DownstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			TlsCertificates: []*tls.TlsCertificate{
				{
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
							InlineString: certificate,
						},
					},
					PrivateKey: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
							InlineString: certificateKey,
						},
					},
				},
			},
			AlpnProtocols: s.listenerALPNOptions(v),
		},
	}

	tlsContext, err := ptypes.MarshalAny(downStreamTLSConfig)
	if err != nil {
		log.Panic(err)
	}
	FilterChainEntry.TransportSocket =
		&core.TransportSocket{
			Name: "tls",
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: tlsContext,
			},
		}
	return FilterChainEntry
}

// ALPNlistenerALPNOptionsOptions sets TLS's ALPN supported protocols
func (s *server) listenerALPNOptions(v shared.VirtualHost) []string {

	value, err := shared.GetAttribute(v.Attributes, attributeHTTPProtocol)
	if err == nil {
		switch value {
		case attributeHTTPProtocolHTTP11:
			return []string{"http/1.1"}

		case attributeHTTPProtocolHTTP2:
			return []string{"h2", "http/1.1"}

		default:
			log.Warnf("Virtual host '%s' has attribute '%s' with unsupported value '%s'",
				v.Name, attributeHTTPProtocol, value)
		}
	}

	return []string{"http/1.1"}
}

func (s *server) buildConnectionManager(httpFilters []*hcm.HttpFilter, v shared.VirtualHost) *hcm.HttpConnectionManager {

	return &hcm.HttpConnectionManager{
		CodecType:        hcm.HttpConnectionManager_AUTO,
		StatPrefix:       "ingress_http",
		UseRemoteAddress: protoBool(true),
		RouteSpecifier:   s.buildRouteSpecifier(v.RouteGroup),
		HttpFilters:      httpFilters,
		AccessLog:        buildAccessLog(s.config.XDS.Envoy.LogFilename, s.config.XDS.Envoy.LogFields),
	}
}

func (s *server) buildFilter() []*hcm.HttpFilter {
	return []*hcm.HttpFilter{
		{
			Name: "envoy.ext_authz",
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: s.buildExtAuthzFilterConfig(),
			},
		},
		{
			Name: "envoy.filters.http.cors",
		},
		{
			Name: "envoy.filters.http.router",
		},
	}
}
func (s *server) buildExtAuthzFilterConfig() *anypb.Any {
	if s.config.XDS.ExtAuthz.Cluster == "" {
		return nil
	}

	extAuthz := &extAuthz.ExtAuthz{
		FailureModeAllow: s.config.XDS.ExtAuthz.FailureModeAllow,
		Services: &extAuthz.ExtAuthz_GrpcService{
			GrpcService: s.buildGRPCService(1),
		},
	}
	if s.config.XDS.ExtAuthz.RequestBodySize > 0 {
		extAuthz.WithRequestBody = s.extAuthzWithRequestBody()
	}

	extAuthzTypedConf, err := ptypes.MarshalAny(extAuthz)
	if err != nil {
		log.Panic(err)
	}
	return extAuthzTypedConf
}

func (s *server) extAuthzWithRequestBody() *extAuthz.BufferSettings {
	if s.config.XDS.ExtAuthz.RequestBodySize > 0 {
		return &extAuthz.BufferSettings{
			MaxRequestBytes:     uint32(s.config.XDS.ExtAuthz.RequestBodySize),
			AllowPartialMessage: false,
		}
	}
	return nil
}

func (s *server) buildGRPCService(timeout time.Duration) *core.GrpcService {
	return &core.GrpcService{
		Timeout: ptypes.DurationProto(s.config.XDS.ExtAuthz.Timeout),
		TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
			EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
				ClusterName: s.config.XDS.ExtAuthz.Cluster,
			},
		},
	}
}

func (s *server) buildRouteSpecifier(RouteGroup string) *hcm.HttpConnectionManager_RouteConfig {
	return &hcm.HttpConnectionManager_RouteConfig{
		RouteConfig: s.buildEnvoyVirtualHostRouteConfig(RouteGroup, s.routes),
	}
}

// func buildRouteSpecifierRDS(RouteGroup string) *hcm.HttpConnectionManager_Rds {
// 	// TODO
// 	XdsCluster := "xds_cluster"

// 	return &hcm.HttpConnectionManager_Rds{
// 		Rds: &hcm.Rds{
// 			RouteConfigName: RouteGroup,
// 			ConfigSource: &core.ConfigSource{
// 				ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
// 					ApiConfigSource: &core.ApiConfigSource{
// 						ApiType: core.ApiConfigSource_GRPC,
// 						GrpcServices: []*core.GrpcService{{
// 							TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
// 								EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: XdsCluster},
// 							},
// 						}},
// 					},
// 				},
// 			},
// 		},
// 	}
// }

func buildAccessLog(logFilename string, fieldFormats map[string]string) []*accessLog.AccessLog {
	jsonFormat := &spb.Struct{
		Fields: map[string]*spb.Value{},
	}
	for field, fieldFormat := range fieldFormats {
		if fieldFormat != "" {
			jsonFormat.Fields[field] = &spb.Value{
				Kind: &spb.Value_StringValue{StringValue: fieldFormat},
			}
		}
	}
	accessLogConf := &fileAccessLog.FileAccessLog{
		Path: logFilename,
		AccessLogFormat: &fileAccessLog.FileAccessLog_JsonFormat{
			JsonFormat: jsonFormat,
		},
	}
	accessLogTypedConf, err := ptypes.MarshalAny(accessLogConf)
	if err != nil {
		log.Panic(err)
	}
	return []*accessLog.AccessLog{{
		// TODO: this should be configurable
		Name: "envoy.file_access_log",
		ConfigType: &accessLog.AccessLog_TypedConfig{
			TypedConfig: accessLogTypedConf,
		},
	}}
}
