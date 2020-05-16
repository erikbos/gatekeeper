package main

import (
	"fmt"
	"sync"
	"time"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	accessLog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	filterAccessLog "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	extAuthz "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/ext_authz/v2"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	spb "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/erikbos/apiauth/pkg/shared"
)

const (
	virtualHostDataRefreshInterval = 2 * time.Second

	extAuthzTimeout = 100 * time.Millisecond
)

// FIXME this does not detect removed records
// GetVirtualHostConfigFromDatabase continously gets the current configuration
func (s *server) GetVirtualHostConfigFromDatabase() {
	var virtualHostsLastUpdate int64
	var virtualHostsMutex sync.Mutex

	for {
		var xdsPushNeeded bool

		newVirtualHosts, err := s.db.GetVirtualHosts()
		if err != nil {
			log.Errorf("Could not retrieve virtualhosts from database (%s)", err)
		} else {
			if virtualHostsLastUpdate == 0 {
				log.Info("Initial load of virtualhosts done")
			}
			for _, virtualhost := range newVirtualHosts {
				// Is a virtualhosts updated since last time we stored it?
				if virtualhost.LastmodifiedAt > virtualHostsLastUpdate {
					virtualHostsMutex.Lock()
					s.virtualhosts = newVirtualHosts
					virtualHostsMutex.Unlock()

					virtualHostsLastUpdate = shared.GetCurrentTimeMilliseconds()
					xdsPushNeeded = true
				}
			}
		}
		if xdsPushNeeded {
			// FIXME this should be notification via channel
			xdsLastUpdate = shared.GetCurrentTimeMilliseconds()
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
func (s *server) getVirtualHostsOfRouteSet(routeSetName string) []string {
	var virtualHostsInRouteSet []string

	for _, virtualhost := range s.virtualhosts {
		if virtualhost.RouteSet == routeSetName {
			virtualHostsInRouteSet = append(virtualHostsInRouteSet, virtualhost.VirtualHosts...)
		}
	}
	return virtualHostsInRouteSet
}

func (s *server) buildEnvoyListenerConfig(port int) *api.Listener {

	newListener := &api.Listener{
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

// buildAddress builds an Envoy address to connect to
func buildAddress(hostname string, port int) *core.Address {
	return &core.Address{Address: &core.Address_SocketAddress{
		SocketAddress: &core.SocketAddress{
			Address:  hostname,
			Protocol: core.SocketAddress_TCP,
			PortSpecifier: &core.SocketAddress_PortValue{
				PortValue: uint32(port),
			},
		},
	}}
}

func buildListenerFilterHTTP() []*listener.ListenerFilter {
	return []*listener.ListenerFilter{
		{
			Name: "envoy.filters.listener.http_inspector",
		},
	}
}

func (s *server) buildFilterChainEntry(l *api.Listener, v shared.VirtualHost) *listener.FilterChain {
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
	certificate, error1 := shared.GetAttribute(v.Attributes, "TLSCertificate")
	certificateKey, error2 := shared.GetAttribute(v.Attributes, "TLSCertificateKey")

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
	downStreamTLSConfig := &auth.DownstreamTlsContext{
		CommonTlsContext: &auth.CommonTlsContext{
			TlsCertificates: []*auth.TlsCertificate{
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
		case attributeHTTPProtocolHTTP1:
			return []string{"http/1.1"}

		case attributeHTTPProtocolHTTP2:
			return []string{"h2", "http/1.1"}

		default:
			log.Warnf("listenerALPNOptions: vhost %s has attribute %s with unknown value %s",
				v.Name, attributeHTTPProtocol, value)
		}
	}

	return []string{"http/1.1"}
}

func (s *server) buildConnectionManager(httpFilters []*hcm.HttpFilter, v shared.VirtualHost) *hcm.HttpConnectionManager {

	return &hcm.HttpConnectionManager{
		CodecType:        hcm.HttpConnectionManager_AUTO,
		StatPrefix:       "ingress_http",
		UseRemoteAddress: &wrappers.BoolValue{Value: true},
		RouteSpecifier:   s.buildRouteSpecifier(v.RouteSet),
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
		Timeout: ptypes.DurationProto(s.getAuthzTimeout()),
		TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
			EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
				ClusterName: s.config.XDS.ExtAuthz.Cluster,
			},
		},
	}
}

// getAuthTimeout gets extauthz timeout
func (s *server) getAuthzTimeout() time.Duration {
	interval, err := time.ParseDuration(s.config.XDS.ExtAuthz.Timeout)
	if err != nil {
		log.Warnf("Cannot parse '%s' as AuthzTimeout (%s)", s.config.XDS.ExtAuthz.Timeout, err)
		return extAuthzTimeout
	}
	return interval
}

func (s *server) buildRouteSpecifier(routeSet string) *hcm.HttpConnectionManager_RouteConfig {
	return &hcm.HttpConnectionManager_RouteConfig{
		RouteConfig: s.buildEnvoyVirtualHostRouteConfig(routeSet, s.routes),
	}
}

// func buildRouteSpecifierRDS(routeSet string) *hcm.HttpConnectionManager_Rds {
// 	// FIXME
// 	XdsCluster := "xds_cluster"

// 	return &hcm.HttpConnectionManager_Rds{
// 		Rds: &hcm.Rds{
// 			RouteConfigName: routeSet,
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

func buildAccessLog(logFilename string, fieldFormats map[string]string) []*filterAccessLog.AccessLog {
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
	accessLogConf := &accessLog.FileAccessLog{
		Path:            logFilename,
		AccessLogFormat: &accessLog.FileAccessLog_JsonFormat{JsonFormat: jsonFormat},
	}
	accessLogTypedConf, err := ptypes.MarshalAny(accessLogConf)
	if err != nil {
		log.Panic(err)
	}
	return []*filterAccessLog.AccessLog{{
		Name: "envoy.file_access_log",
		ConfigType: &filterAccessLog.AccessLog_TypedConfig{
			TypedConfig: accessLogTypedConf,
		},
	}}
}
