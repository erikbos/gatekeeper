package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/erikbos/apiauth/pkg/shared"
	log "github.com/sirupsen/logrus"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	accessLog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	filterAccessLog "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	spb "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"
)

const (
	virtualHostRefreshInterval = 2 * time.Second
)

var virtualhosts []shared.VirtualHost

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
			for _, s := range newVirtualHosts {
				// Is a virtualhosts updated since last time we stored it?
				if s.LastmodifiedAt > virtualHostsLastUpdate {
					virtualHostsMutex.Lock()
					virtualhosts = newVirtualHosts
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
			s.metricXdsDeployments.WithLabelValues("virtualhosts").Inc()
		}
		time.Sleep(virtualHostRefreshInterval)
	}
}

// GetVirtualHostCount returns number of virtualhosts
func (s *server) GetVirtualHostCount() float64 {
	return float64(len(virtualhosts))
}

// getEnvoyListenerConfig returns array of envoy listeners
func (s *server) getEnvoyListenerConfig() ([]cache.Resource, error) {
	envoyListeners := []cache.Resource{}

	uniquePorts := getVirtualHostPorts(virtualhosts)
	for port := range uniquePorts {
		log.Infof("adding listener for port %d", port)
		envoyListeners = append(envoyListeners, s.buildEnvoyListenerConfig(port, virtualhosts))
	}
	return envoyListeners, nil
}

// getVirtualHostPorts return unique set of ports from vhost configuration
func getVirtualHostPorts(vhosts []shared.VirtualHost) map[int]bool {
	listenerPorts := map[int]bool{}
	for _, virtualhost := range vhosts {
		listenerPorts[virtualhost.Port] = true
	}
	return listenerPorts
}

// getVirtualHostPorts return unique set of ports from vhost configuration
func getVirtualHostsOfRouteSet(routeSetName string) []string {
	var virtualHostsInRouteSet []string

	for _, virtualhost := range virtualhosts {
		if virtualhost.RouteSet == routeSetName {
			virtualHostsInRouteSet = append(virtualHostsInRouteSet, virtualhost.VirtualHosts...)
		}
	}
	return virtualHostsInRouteSet
}

func (s *server) buildEnvoyListenerConfig(port int, vhosts []shared.VirtualHost) *api.Listener {

	newListener := &api.Listener{
		Name:            fmt.Sprintf("port_%d", port),
		Address:         buildAddress("0.0.0.0", port),
		ListenerFilters: buildListenerFilterHTTP(),
	}

	// add all vhosts belonging to this listener's port
	for _, virtualhost := range vhosts {
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
	httpFilter := buildFilter()

	manager := s.buildConnectionManager(httpFilter, v)
	managerProtoBuf, err := ptypes.MarshalAny(manager)
	if err != nil {
		log.Panic(err)
	}

	FilterChainEntry := &listener.FilterChain{
		// FilterChainMatch: &listener.FilterChainMatch{
		// 	ServerNames: []string{v.VirtualHosts[0]},
		// },
		Filters: []*listener.Filter{{
			Name: wellknown.HTTPConnectionManager,
			ConfigType: &listener.Filter_TypedConfig{
				TypedConfig: managerProtoBuf,
			},
		}},
	}

	// Configure TLS in case when we have a certificate + key
	var tlsContext *any.Any

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

	// Configure listener to do match SNI
	FilterChainEntry.FilterChainMatch =
		&listener.FilterChainMatch{
			ServerNames: []string{v.VirtualHosts[0]},
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
		},
	}

	// Enable HTTP/2
	if value, err := shared.GetAttribute(v.Attributes, "HTTP2Enabled"); err == nil {
		if value == "true" {
			downStreamTLSConfig.CommonTlsContext.AlpnProtocols = []string{"h2", "http/1.1"}
		}
	} else {
		// by default we will do http/1.1
		downStreamTLSConfig.CommonTlsContext.AlpnProtocols = []string{"http/1.1"}
	}

	tlsContext, err = ptypes.MarshalAny(downStreamTLSConfig)
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

func (s *server) buildConnectionManager(httpFilters []*hcm.HttpFilter, v shared.VirtualHost) *hcm.HttpConnectionManager {

	return &hcm.HttpConnectionManager{
		CodecType:        hcm.HttpConnectionManager_AUTO,
		StatPrefix:       "ingress_http",
		UseRemoteAddress: &wrappers.BoolValue{Value: true},
		RouteSpecifier:   buildRouteSpecifier(v.RouteSet),
		HttpFilters:      httpFilters,
		AccessLog:        buildAccessLog(s.config.XDS.EnvoyLogFilename, s.config.XDS.EnvoyLogFields),
	}
}

func buildFilter() []*hcm.HttpFilter {
	return []*hcm.HttpFilter{
		{
			Name: "envoy.filters.http.cors",
		},
		{
			Name: "envoy.filters.http.router",
		},
	}
}

func buildRouteSpecifier(routeSet string) *hcm.HttpConnectionManager_RouteConfig {
	return &hcm.HttpConnectionManager_RouteConfig{
		RouteConfig: buildEnvoyVirtualHostRouteConfig(routeSet, routes),
	}
}

func buildRouteSpecifierRDS(routeSet string) *hcm.HttpConnectionManager_Rds {
	// FIXME
	XdsCluster := "xds_cluster"

	return &hcm.HttpConnectionManager_Rds{
		Rds: &hcm.Rds{
			RouteConfigName: routeSet,
			ConfigSource: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
					ApiConfigSource: &core.ApiConfigSource{
						ApiType: core.ApiConfigSource_GRPC,
						GrpcServices: []*core.GrpcService{{
							TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
								EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: XdsCluster},
							},
						}},
					},
				},
			},
		},
	}
}

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
		Name:       "envoy.file_access_log",
		ConfigType: &filterAccessLog.AccessLog_TypedConfig{TypedConfig: accessLogTypedConf},
	}}
}
