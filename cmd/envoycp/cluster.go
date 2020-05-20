package main

import (
	"sync"
	"time"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

const (
	clusterDataRefreshInterval   = 2 * time.Second
	defaultClusterConnectTimeout = 2 * time.Second

	attributeConnectTimeout      = "ConnectTimeout"
	attributeIdleTimeout         = "IdleTimeout"
	attributeTLSEnabled          = "TLSEnabled"
	attributeTLSMinimumVersion   = "TLSMinimumVersion"
	attributeTLSMaximumVersion   = "TLSMaximumVersion"
	attributeHTTPProtocol        = "HTTPProtocol"
	attributeHTTPProtocolHTTP1   = "HTTP/1.1"
	attributeHTTPProtocolHTTP2   = "HTTP/2"
	attributeHTTPProtocolHTTP3   = "HTTP/3"
	attributeSNIHostName         = "SNIHostName"
	attributeHealthCheckProtocol = "HealthCheckProtocol"
	attributeHealthCheckPath     = "HealthCheckPath"
	attributeHealthCheckInterval = "HealthCheckInterval"
	attributeHealthCheckTimeout  = "HealthCheckTimeout"

	attributeValueTrue  = "true"
	attributeValueHTTP  = "HTTP"
	attributeValueTLS10 = "TLSv10"
	attributeValueTLS11 = "TLSv11"
	attributeValueTLS12 = "TLSv12"
	attributeValueTLS13 = "TLSv13"
)

// FIXME this does not detect removed records
// getClusterConfigFromDatabase continously gets the current configuration
func (s *server) GetClusterConfigFromDatabase(n chan xdsNotifyMesssage) {
	var clustersLastUpdate int64
	var clusterMutex sync.Mutex

	for {
		var xdsPushNeeded bool

		newClusterList, err := s.db.GetClusters()
		if err != nil {
			log.Errorf("Could not retrieve clusters from database (%s)", err)
		} else {
			// Is one of the cluster updated since last time pushed config to Envoy?
			if clustersLastUpdate == 0 {
				log.Info("Initial load of clusters")
			}
			for _, cluster := range newClusterList {
				if cluster.LastmodifiedAt > clustersLastUpdate {
					clusterMutex.Lock()
					s.clusters = newClusterList
					clusterMutex.Unlock()

					clustersLastUpdate = shared.GetCurrentTimeMilliseconds()
					xdsPushNeeded = true
				}
			}
		}
		if xdsPushNeeded {
			n <- xdsNotifyMesssage{
				resource: "cluster",
			}
			// Increase xds deployment metric
			s.metrics.xdsDeployments.WithLabelValues("clusters").Inc()
		}
		time.Sleep(clusterDataRefreshInterval)
	}
}

// GetClusterCount returns number of clusters
func (s *server) GetClusterCount() float64 {
	return float64(len(s.clusters))
}

// getClusterConfig returns array of all envoy clusters
func (s *server) getEnvoyClusterConfig() ([]cache.Resource, error) {
	envoyClusters := []cache.Resource{}
	for _, s := range s.clusters {
		envoyClusters = append(envoyClusters, buildEnvoyClusterConfig(s))
	}
	return envoyClusters, nil
}

// buildEnvoyClusterConfig builds one envoy cluster configuration
func buildEnvoyClusterConfig(cluster shared.Cluster) *api.Cluster {

	envoyCluster := &api.Cluster{
		Name:                      cluster.Name,
		ConnectTimeout:            clusterConnectTimeout(cluster),
		ClusterDiscoveryType:      &api.Cluster_Type{Type: api.Cluster_LOGICAL_DNS},
		DnsLookupFamily:           api.Cluster_V4_ONLY,
		LbPolicy:                  api.Cluster_ROUND_ROBIN,
		LoadAssignment:            clusterLoadAssignment(cluster),
		HealthChecks:              clusterHealthCheckConfig(cluster),
		CommonHttpProtocolOptions: clusterCommonHTTPProtocolOptions(cluster),
		// CircuitBreakers:      clusterCircuitBreaker(cluster),
	}

	// Add TLS and HTTP/2 configuration options in case we want to
	value, err := shared.GetAttribute(cluster.Attributes, attributeTLSEnabled)
	if err == nil && value == attributeValueTrue {
		envoyCluster.TransportSocket = clusterTransportSocket(cluster)
		envoyCluster.Http2ProtocolOptions = clusterHTTP2ProtocolOptions(cluster)
	}

	return envoyCluster
}

func clusterConnectTimeout(cluster shared.Cluster) *duration.Duration {
	connectTimeout, _ := shared.GetAttribute(cluster.Attributes, attributeConnectTimeout)
	connectTimeoutAsDuration, err := time.ParseDuration(connectTimeout)
	if err != nil {
		connectTimeoutAsDuration = defaultClusterConnectTimeout
	}
	return ptypes.DurationProto(connectTimeoutAsDuration)
}

func clusterLoadAssignment(cluster shared.Cluster) *api.ClusterLoadAssignment {
	return &api.ClusterLoadAssignment{
		ClusterName: cluster.Name,
		Endpoints:   buildEndpoint(cluster.HostName, cluster.Port),
	}
}

func buildEndpoint(hostname string, port int) []*endpoint.LocalityLbEndpoints {
	address := &core.Address{Address: &core.Address_SocketAddress{
		SocketAddress: &core.SocketAddress{
			Address:  hostname,
			Protocol: core.SocketAddress_TCP,
			PortSpecifier: &core.SocketAddress_PortValue{
				PortValue: uint32(port),
			},
		},
	}}

	return []*endpoint.LocalityLbEndpoints{
		{
			LbEndpoints: []*endpoint.LbEndpoint{
				{
					HostIdentifier: &endpoint.LbEndpoint_Endpoint{
						Endpoint: &endpoint.Endpoint{
							Address: address,
						},
					},
				},
			},
		},
	}

}

// func clusterCircuitBreaker(cluster shared.Cluster) *envoy_cluster.CircuitBreakers {
// 	return &envoy_cluster.CircuitBreakers{
// 		Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{{
// 			MaxConnections:     u32nil(service.MaxConnections),
// 			MaxPendingRequests: u32nil(service.MaxPendingRequests),
// 			MaxRequests:        u32nil(service.MaxRequests),
// 			MaxRetries:         u32nil(service.MaxRetries),
// 		}},
// 	}
// }

// clusterHealthCheckConfig builds health configuration for a cluster
func clusterHealthCheckConfig(cluster shared.Cluster) []*core.HealthCheck {

	value, err := shared.GetAttribute(cluster.Attributes, attributeHealthCheckProtocol)
	if err == nil && value == attributeValueHTTP {
		healthCheckPath, _ := shared.GetAttribute(cluster.Attributes, attributeHealthCheckPath)

		healthCheckInterval, _ := shared.GetAttribute(cluster.Attributes, attributeHealthCheckInterval)
		healthcheckIntervalAsDuration, err := time.ParseDuration(healthCheckInterval)
		if err != nil {
			healthcheckIntervalAsDuration = 10 * time.Second
		}

		healthCheckTimeout, _ := shared.GetAttribute(cluster.Attributes, attributeHealthCheckTimeout)
		healthcheckTimeoutAsDuration, err := time.ParseDuration(healthCheckTimeout)
		if err != nil {
			healthcheckTimeoutAsDuration = 10 * time.Second
		}

		healthCheck := &core.HealthCheck{
			HealthChecker: &core.HealthCheck_HttpHealthCheck_{
				HttpHealthCheck: &core.HealthCheck_HttpHealthCheck{
					Path:            healthCheckPath,
					CodecClientType: clusterHealthCodec(cluster),
				},
			},
			Interval:           ptypes.DurationProto(healthcheckIntervalAsDuration),
			Timeout:            ptypes.DurationProto(healthcheckTimeoutAsDuration),
			UnhealthyThreshold: &wrappers.UInt32Value{Value: 2},
			HealthyThreshold:   &wrappers.UInt32Value{Value: 1},
			EventLogPath:       "/tmp/healthcheck",
		}
		return append([]*core.HealthCheck{}, healthCheck)
	}
	return nil
}

func clusterHealthCodec(cluster shared.Cluster) envoy_type.CodecClientType {

	value, err := shared.GetAttribute(cluster.Attributes, attributeHTTPProtocol)
	if err == nil {
		switch value {
		case attributeHTTPProtocolHTTP2:
			return envoy_type.CodecClientType_HTTP2

		case attributeHTTPProtocolHTTP3:
			return envoy_type.CodecClientType_HTTP3

		default:
			log.Warnf("Cluster %s has attribute %s with unknown value %s",
				cluster.Name, attributeHTTPProtocol, value)
		}
	}
	return envoy_type.CodecClientType_HTTP1
}

// clusterCommonHTTPProtocolOptions sets HTTP options applicable to both HTTP/1 and /2
func clusterCommonHTTPProtocolOptions(cluster shared.Cluster) *core.HttpProtocolOptions {

	idleTimeout, _ := shared.GetAttribute(cluster.Attributes, attributeIdleTimeout)
	idleTimeoutAsDuration, err := time.ParseDuration(idleTimeout)
	if err != nil {
		return nil
	}
	return &core.HttpProtocolOptions{
		IdleTimeout: ptypes.DurationProto(idleTimeoutAsDuration),
	}
}

// clusterHTTP2ProtocolOptions returns HTTP/2 parameters
func clusterHTTP2ProtocolOptions(cluster shared.Cluster) *core.Http2ProtocolOptions {

	value, err := shared.GetAttribute(cluster.Attributes, attributeHTTPProtocol)
	if err == nil {
		switch value {
		case attributeHTTPProtocolHTTP1:
			return nil
		case attributeHTTPProtocolHTTP2:
			// according to spec we need to return at least empty struct to enable HTTP/2
			return &core.Http2ProtocolOptions{}
		}
	}

	log.Warnf("ClusterProtocol: %s has attribute %s with unknown value %s",
		cluster.Name, attributeHTTPProtocol, value)
	return nil
}

// clusterTransportSocket configures TLS settings
func clusterTransportSocket(cluster shared.Cluster) *core.TransportSocket {
	TLSContext := &auth.UpstreamTlsContext{
		Sni: clusterSNIHostname(cluster),
		CommonTlsContext: &auth.CommonTlsContext{
			AlpnProtocols: clusterALPNOptions(cluster),
			TlsParams:     clusterTLSOptions(cluster),
		},
	}
	tlsContextProtoBuf, err := ptypes.MarshalAny(TLSContext)
	if err != nil {
		return nil
	}
	return &core.TransportSocket{
		Name: "tls",
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsContextProtoBuf,
		},
	}
}

// clusterSNIHostname sets SNI hostname used by TLS
func clusterSNIHostname(cluster shared.Cluster) string {
	value, err := shared.GetAttribute(cluster.Attributes, attributeSNIHostName)
	if err == nil && value != "" {
		return value
	}
	return cluster.HostName
}

// clusterALPNOptions sets TLS's ALPN supported protocols
func clusterALPNOptions(cluster shared.Cluster) []string {

	value, err := shared.GetAttribute(cluster.Attributes, attributeHTTPProtocol)
	if err == nil {
		switch value {
		case attributeHTTPProtocolHTTP1:
			return []string{"http/1.1"}
		case attributeHTTPProtocolHTTP2:
			return []string{"h2", "http/1.1"}
		}
	}

	log.Warnf("clusterALPNOptions: cluster %s has attribute %s with unknown value %s",
		cluster.Name, attributeHTTPProtocol, value)

	return []string{"http/1.1"}
}

// clusterALPNOptions sets TLS minimum and max cipher options
func clusterTLSOptions(cluster shared.Cluster) *auth.TlsParameters {
	tlsParameters := &auth.TlsParameters{}
	if minVersion, err := shared.GetAttribute(cluster.Attributes, attributeTLSMinimumVersion); err == nil {
		tlsParameters.TlsMinimumProtocolVersion = tlsVersion(minVersion)
	}
	if maxVersion, err := shared.GetAttribute(cluster.Attributes, attributeTLSMaximumVersion); err == nil {
		tlsParameters.TlsMaximumProtocolVersion = tlsVersion(maxVersion)
	}
	return tlsParameters
}

func tlsVersion(version string) auth.TlsParameters_TlsProtocol {
	switch version {
	case attributeValueTLS10:
		return auth.TlsParameters_TLSv1_0
	case attributeValueTLS11:
		return auth.TlsParameters_TLSv1_1
	case attributeValueTLS12:
		return auth.TlsParameters_TLSv1_2
	case attributeValueTLS13:
		return auth.TlsParameters_TLSv1_3
	}
	return auth.TlsParameters_TLS_AUTO
}

// func u32nil(val uint32) *wrappers.UInt32Value {
// 	switch val {
// 	case 0:
// 		return nil
// 	default:
// 		return &wrappers.UInt32Value{Value: val}
// 	}
// }
