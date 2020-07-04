package main

import (
	"strings"
	"sync"
	"time"

	envoyCluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoyType "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

const (
	clusterDataRefreshInterval = 2 * time.Second
)

// FIXME this does not detect removed records
// getClusterConfigFromDatabase continuously gets the current configuration
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

					warnForUnknownClusterAttributes(cluster)
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
func buildEnvoyClusterConfig(cluster shared.Cluster) *envoyCluster.Cluster {

	envoyCluster := &envoyCluster.Cluster{
		Name:                      cluster.Name,
		ConnectTimeout:            clusterConnectTimeout(cluster),
		ClusterDiscoveryType:      &envoyCluster.Cluster_Type{Type: envoyCluster.Cluster_LOGICAL_DNS},
		DnsLookupFamily:           envoyCluster.Cluster_V4_ONLY,
		DnsResolvers:              clusterDNSResolvers(cluster),
		DnsRefreshRate:            clusterDNSRefreshRate(cluster),
		LbPolicy:                  envoyCluster.Cluster_ROUND_ROBIN,
		LoadAssignment:            clusterLoadAssignment(cluster),
		HealthChecks:              clusterHealthCheckConfig(cluster),
		CommonHttpProtocolOptions: clusterCommonHTTPProtocolOptions(cluster),
		CircuitBreakers:           clusterCircuitBreaker(cluster),
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

	connectTimeout := shared.GetAttributeAsDuration(cluster.Attributes,
		attributeConnectTimeout, defaultClusterConnectTimeout)

	return ptypes.DurationProto(connectTimeout)
}

func clusterLoadAssignment(cluster shared.Cluster) *endpoint.ClusterLoadAssignment {

	return &endpoint.ClusterLoadAssignment{
		ClusterName: cluster.Name,
		Endpoints:   buildEndpoint(cluster.HostName, cluster.Port),
	}
}

func buildEndpoint(hostname string, port int) []*endpoint.LocalityLbEndpoints {

	return []*endpoint.LocalityLbEndpoints{
		{
			LbEndpoints: []*endpoint.LbEndpoint{
				{
					HostIdentifier: &endpoint.LbEndpoint_Endpoint{
						Endpoint: &endpoint.Endpoint{
							Address: buildAddress(hostname, port),
						},
					},
				},
			},
		},
	}
}

func clusterCircuitBreaker(cluster shared.Cluster) *envoyCluster.CircuitBreakers {

	maxConnections := shared.GetAttributeAsInt(cluster.Attributes, attributeMaxConnections, 0)
	maxPendingRequests := shared.GetAttributeAsInt(cluster.Attributes, attributeMaxPendingRequests, 0)
	maxRequests := shared.GetAttributeAsInt(cluster.Attributes, attributeMaxRequests, 0)
	maxRetries := shared.GetAttributeAsInt(cluster.Attributes, attributeMaxRetries, 0)

	return &envoyCluster.CircuitBreakers{
		Thresholds: []*envoyCluster.CircuitBreakers_Thresholds{{
			MaxConnections:     protoUint32orNil(maxConnections),
			MaxPendingRequests: protoUint32orNil(maxPendingRequests),
			MaxRequests:        protoUint32orNil(maxRequests),
			MaxRetries:         protoUint32orNil(maxRetries),
		}},
	}
}

// clusterHealthCheckConfig builds health configuration for a cluster
func clusterHealthCheckConfig(cluster shared.Cluster) []*core.HealthCheck {

	protocol, err := shared.GetAttribute(cluster.Attributes, attributeHealthCheckProtocol)
	path, _ := shared.GetAttribute(cluster.Attributes, attributeHealthCheckPath)

	if err == nil && protocol == attributeValueHTTP && path != "" {

		healthCheckInterval := shared.GetAttributeAsDuration(cluster.Attributes,
			attributeHealthCheckInterval, defaultHealthCheckInterval)

		healthCheckTimeout := shared.GetAttributeAsDuration(cluster.Attributes,
			attributeHealthCheckTimeout, defaultHealthCheckTimeout)

		healthCheckUnhealthyThreshold := shared.GetAttributeAsInt(cluster.Attributes,
			attributeHealthCheckUnhealthyThreshold, defaultHealthCheckUnhealthyThreshold)

		healthCheckHealthyThreshold := shared.GetAttributeAsInt(cluster.Attributes,
			attributeHealthCheckHealthyThreshold, defaultHealthCheckHealthyThreshold)

		healthCheckLogFile := shared.GetAttributeAsString(cluster.Attributes,
			attributeHealthCheckLogFile, "")

		healthCheck := &core.HealthCheck{
			HealthChecker: &core.HealthCheck_HttpHealthCheck_{
				HttpHealthCheck: &core.HealthCheck_HttpHealthCheck{
					Path:            path,
					CodecClientType: clusterHealthCodec(cluster),
				},
			},
			Interval:           ptypes.DurationProto(healthCheckInterval),
			Timeout:            ptypes.DurationProto(healthCheckTimeout),
			UnhealthyThreshold: protoUint32orNil(healthCheckUnhealthyThreshold),
			HealthyThreshold:   protoUint32orNil(healthCheckHealthyThreshold),
		}
		if healthCheckLogFile != "" {
			healthCheck.EventLogPath = healthCheckLogFile
		}

		return append([]*core.HealthCheck{}, healthCheck)
	}
	return nil
}

func clusterHealthCodec(cluster shared.Cluster) envoyType.CodecClientType {

	value, err := shared.GetAttribute(cluster.Attributes, attributeHTTPProtocol)
	if err == nil {
		switch value {
		case attributeHTTPProtocolHTTP2:
			return envoyType.CodecClientType_HTTP2

		case attributeHTTPProtocolHTTP3:
			return envoyType.CodecClientType_HTTP3

		default:
			log.Warnf("Cluster '%s' has attribute '%s' with unknown value '%s'",
				cluster.Name, attributeHTTPProtocol, value)
		}
	}
	return envoyType.CodecClientType_HTTP1
}

// clusterCommonHTTPProtocolOptions sets HTTP options applicable to both HTTP/1 and /2
func clusterCommonHTTPProtocolOptions(cluster shared.Cluster) *core.HttpProtocolOptions {

	idleTimeout := shared.GetAttributeAsDuration(cluster.Attributes,
		attributeIdleTimeout, defaultClusterIdleTimeout)

	return &core.HttpProtocolOptions{
		IdleTimeout: ptypes.DurationProto(idleTimeout),
	}
}

// clusterHTTP2ProtocolOptions returns HTTP/2 parameters
func clusterHTTP2ProtocolOptions(cluster shared.Cluster) *core.Http2ProtocolOptions {

	value, err := shared.GetAttribute(cluster.Attributes, attributeHTTPProtocol)
	if err == nil {
		switch value {
		case attributeHTTPProtocolHTTP11:
			return nil
		case attributeHTTPProtocolHTTP2:
			// according to spec we need to return at least empty struct to enable HTTP/2
			return &core.Http2ProtocolOptions{}
		}
	}

	log.Warnf("ClusterProtocol: '%s' has attribute '%s' with unknown value '%s'",
		cluster.Name, attributeHTTPProtocol, value)
	return nil
}

// clusterTransportSocket configures TLS settings
func clusterTransportSocket(cluster shared.Cluster) *core.TransportSocket {

	TLSContext := &tls.UpstreamTlsContext{
		Sni: clusterSNIHostname(cluster),
		CommonTlsContext: &tls.CommonTlsContext{
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
		case attributeHTTPProtocolHTTP11:
			return []string{"http/1.1"}
		case attributeHTTPProtocolHTTP2:
			return []string{"h2", "http/1.1"}
		}
	}

	log.Warnf("Cluster '%s' has attribute '%s' with unsupported value '%s'",
		cluster.Name, attributeHTTPProtocol, value)

	return []string{"http/1.1"}
}

// clusterALPNOptions sets TLS minimum and max cipher options
func clusterTLSOptions(cluster shared.Cluster) *tls.TlsParameters {

	tlsParameters := &tls.TlsParameters{}
	if minVersion, err := shared.GetAttribute(cluster.Attributes, attributeTLSMinimumVersion); err == nil {
		tlsParameters.TlsMinimumProtocolVersion = tlsVersion(minVersion)
	}

	if maxVersion, err := shared.GetAttribute(cluster.Attributes, attributeTLSMaximumVersion); err == nil {
		tlsParameters.TlsMaximumProtocolVersion = tlsVersion(maxVersion)
	}
	tlsParameters.CipherSuites = tlsCipherSuites(cluster)
	return tlsParameters
}

func tlsVersion(version string) tls.TlsParameters_TlsProtocol {

	switch version {
	case attributeValueTLS10:
		return tls.TlsParameters_TLSv1_0
	case attributeValueTLS11:
		return tls.TlsParameters_TLSv1_1
	case attributeValueTLS12:
		return tls.TlsParameters_TLSv1_2
	case attributeValueTLS13:
		return tls.TlsParameters_TLSv1_3
	}
	return tls.TlsParameters_TLS_AUTO
}

func tlsCipherSuites(cluster shared.Cluster) []string {

	value, err := shared.GetAttribute(cluster.Attributes, attributeTLSCipherSuites)
	if err == nil {
		var ciphers []string

		for _, cipher := range strings.Split(value, ",") {
			ciphers = append(ciphers, strings.TrimSpace(cipher))
		}
		return ciphers
	}
	return nil
}

func clusterDNSRefreshRate(cluster shared.Cluster) *duration.Duration {

	refreshInterval := shared.GetAttributeAsDuration(cluster.Attributes,
		attributeDNSRefreshRate, defaultDNSRefreshRate)

	return ptypes.DurationProto(refreshInterval)
}

func clusterDNSResolvers(cluster shared.Cluster) []*core.Address {

	value, err := shared.GetAttribute(cluster.Attributes, attributeDNSResolvers)
	if err == nil {
		var resolvers []*core.Address

		for _, resolver := range strings.Split(value, ",") {
			resolvers = append(resolvers, buildAddress(resolver, 53))
		}
		return resolvers
	}
	return nil
}
