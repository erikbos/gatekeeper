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

var virtualhosts []shared.VirtualHost

// FIXME this does not detect removed records

// GetVirtualHostConfigFromDatabase continously gets the current configuration
func (s *server) GetVirtualHostConfigFromDatabase() {
	var virtualHostsLastUpdate int64
	var virtualHostsMutex sync.Mutex

	for {
		newVirtualHosts, err := s.db.GetVirtualHosts()
		if err != nil {
			log.Errorf("Could not retrieve virtualhosts from database (%s)", err)
		} else {
			for _, s := range newVirtualHosts {
				// Is a virtualhosts updated since last time we stored it?
				if s.LastmodifiedAt > virtualHostsLastUpdate {
					log.Info("Virtual hosts config changed!")
					now := shared.GetCurrentTimeMilliseconds()

					virtualHostsMutex.Lock()
					virtualhosts = newVirtualHosts
					virtualHostsLastUpdate = now
					virtualHostsMutex.Unlock()

					// FIXME this should be notification via channel
					xdsLastUpdate = now
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// getEnvoyListenerConfig returns array of envoy listeners
func getEnvoyListenerConfig() ([]cache.Resource, error) {
	envoyListeners := []cache.Resource{}

	uniquePorts := getVirtualHostPorts(virtualhosts)
	for port := range uniquePorts {
		log.Infof("adding listener for port %d", port)
		envoyListeners = append(envoyListeners, buildEnvoyListenerConfig(port, virtualhosts))
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

func buildEnvoyListenerConfig(port int, vhosts []shared.VirtualHost) *api.Listener {

	listenerName := fmt.Sprintf("port_%d", port)
	newListener := &api.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: uint32(port),
					},
				},
			},
		},
		ListenerFilters: []*listener.ListenerFilter{
			{
				Name: "envoy.filters.listener.http_inspector",
			},
		},
	}

	// add all vhosts belonging to this listener's port
	for _, virtualhost := range vhosts {
		if virtualhost.Port == port {
			newListener.FilterChains = append(newListener.FilterChains, buildFilterChainEntry(newListener, virtualhost))
		}
	}

	return newListener
}

func buildFilterChainEntry(l *api.Listener, v shared.VirtualHost) *listener.FilterChain {
	httpFilter := buildFilter()

	manager := buildConnectionManager(httpFilter, v)
	pbst, err := ptypes.MarshalAny(manager)
	if err != nil {
		panic(err)
	}

	FilterChainEntry := &listener.FilterChain{
		// FilterChainMatch: &listener.FilterChainMatch{
		// 	ServerNames: []string{v.VirtualHosts[0]},
		// },
		Filters: []*listener.Filter{{
			Name: wellknown.HTTPConnectionManager,
			ConfigType: &listener.Filter_TypedConfig{
				TypedConfig: pbst,
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
			AlpnProtocols: []string{"http/1.1"},
		},
	}

	// Enable HTTP/2
	if value, err := shared.GetAttribute(v.Attributes, "HTTP2Enabled"); err == nil {
		if value == "true" {
			downStreamTLSConfig.CommonTlsContext.AlpnProtocols = []string{"h2", "http/1.1"}
		}
	}

	tlsContext, err = ptypes.MarshalAny(downStreamTLSConfig)
	if err != nil {
		panic(err)
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

func buildConnectionManager(httpFilters []*hcm.HttpFilter, v shared.VirtualHost) *hcm.HttpConnectionManager {

	// FIXME
	// XdsCluster := "xds_cluster"

	routeConfig := buildEnvoyRouteConfig(v.RouteSet, routes)

	httpConnectionManager := &hcm.HttpConnectionManager{
		CodecType:        hcm.HttpConnectionManager_AUTO,
		StatPrefix:       "ingress_http",
		UseRemoteAddress: &wrappers.BoolValue{Value: true},

		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig,
		},

		// RouteSpecifier: &hcm.HttpConnectionManager_Rds{
		// 	Rds: &hcm.Rds{
		// 		RouteConfigName: v.RouteSet,
		// 		ConfigSource: &core.ConfigSource{
		// 			ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
		// 				ApiConfigSource: &core.ApiConfigSource{
		// 					ApiType: core.ApiConfigSource_GRPC,
		// 					GrpcServices: []*core.GrpcService{{
		// 						TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
		// 							EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: XdsCluster},
		// 						},
		// 					}},
		// 				},
		// 			},
		// 		},
		// 	},
		// },
		HttpFilters: httpFilters,
		AccessLog:   buildAccessLog("/dev/stdout", accessLogFields),
	}
	return httpConnectionManager
}

func buildFilter() []*hcm.HttpFilter {
	return []*hcm.HttpFilter{
		{
			Name: "envoy.filters.http.router",
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
		panic(err)
	}
	return []*filterAccessLog.AccessLog{{
		Name:       "envoy.file_access_log",
		ConfigType: &filterAccessLog.AccessLog_TypedConfig{TypedConfig: accessLogTypedConf},
	}}
}

var accessLogFields map[string]string = map[string]string{
	"start_time": "%START_TIME%",
	"request_id": "%REQ(REQUEST-ID)%",
	"caller":     "%REQ(CALLER)%",
	// "request_method":                    "%REQ(:METHOD)%",
	// "request_path":                      "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%",
	// "content_type":                      "%REQ(CONTENT-TYPE)%",
	"protocol":       "%PROTOCOL%",
	"response_code":  "%RESPONSE_CODE%",
	"response_flags": "%RESPONSE_FLAGS%",
	"bytes_sent":     "%BYTES_SENT%",
	"bytes_received": "%BYTES_RECEIVED%",
	// "request_duration":                  "%DURATION%",
	// "response_duration":                 "%RESPONSE_DURATION%",
	"upstream_response_time": "%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%",
	"client_address":         "%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%",
	"x_forwarded_for":        "%REQ(X-FORWARDED-FOR)%",
	"user_agent":             "%REQ(USER-AGENT)%",
	// "http2_authority":                   "%REQ(:AUTHORITY)%",
	"upstream_cluster":                  "%UPSTREAM_CLUSTER%",
	"upstream_host":                     "%UPSTREAM_HOST%",
	"route_name":                        "%ROUTE_NAME%",
	"upstream_transport_failure_reason": "%UPSTREAM_TRANSPORT_FAILURE_REASON%",
	"downstream_remote_address":         "%DOWNSTREAM_REMOTE_ADDRESS%",
	"method":                            "%REQ(:METHOD)%",
	"path":                              "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%",
	// "x-forwarded-for":                   "%REQ(X-FORWARDED-FOR)%",
	"user-agent": "%REQ(USER-AGENT)%",
}