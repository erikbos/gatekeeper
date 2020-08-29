package main

import (
	"strings"

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

// getClusterConfig returns array of all envoy clusters
func (s *server) getEnvoyClusterConfig() ([]cache.Resource, error) {

	envoyClusters := []cache.Resource{}

	for _, cluster := range s.dbentities.GetClusters() {
		warnForUnknownClusterAttributes(cluster)

		envoyClusters = append(envoyClusters, buildEnvoyClusterConfig(cluster))
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
		HealthChecks:              clusterHealthChecks(cluster),
		CommonHttpProtocolOptions: clusterCommonHTTPProtocolOptions(cluster),
		CircuitBreakers:           clusterCircuitBreakers(cluster),
	}

	// Add TLS and HTTP/2 configuration options in case we want to
	value, err := cluster.Attributes.Get(attributeTLSEnable)
	if err == nil && value == attributeValueTrue {
		envoyCluster.TransportSocket = clusterTransportSocket(cluster)
		envoyCluster.Http2ProtocolOptions = clusterHTTP2ProtocolOptions(cluster)
	}

	return envoyCluster
}

func clusterConnectTimeout(cluster shared.Cluster) *duration.Duration {

	connectTimeout := cluster.Attributes.GetAsDuration(
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

func clusterCircuitBreakers(cluster shared.Cluster) *envoyCluster.CircuitBreakers {

	maxConnections := cluster.Attributes.GetAsUInt32(attributeMaxConnections, 0)
	maxPendingRequests := cluster.Attributes.GetAsUInt32(attributeMaxPendingRequests, 0)
	maxRequests := cluster.Attributes.GetAsUInt32(attributeMaxRequests, 0)
	maxRetries := cluster.Attributes.GetAsUInt32(attributeMaxRetries, 0)

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
func clusterHealthChecks(cluster shared.Cluster) []*core.HealthCheck {

	healthCheckProtocol, err := cluster.Attributes.Get(attributeHealthCheckProtocol)
	healthCheckPath, _ := cluster.Attributes.Get(attributeHealthCheckPath)

	if err == nil && healthCheckProtocol == attributeValueHealthCheckProtocolHTTP && healthCheckPath != "" {

		interval := cluster.Attributes.GetAsDuration(attributeHealthCheckInterval, defaultHealthCheckInterval)

		timeout := cluster.Attributes.GetAsDuration(attributeHealthCheckTimeout, defaultHealthCheckTimeout)

		unhealthyThreshold := cluster.Attributes.GetAsUInt32(attributeHealthCheckUnhealthyThreshold, defaultHealthCheckUnhealthyThreshold)

		healthyThreshold := cluster.Attributes.GetAsUInt32(attributeHealthCheckHealthyThreshold, defaultHealthCheckHealthyThreshold)

		logFile := cluster.Attributes.GetAsString(attributeHealthCheckLogFile, "")

		healthCheck := &core.HealthCheck{
			HealthChecker: &core.HealthCheck_HttpHealthCheck_{
				HttpHealthCheck: &core.HealthCheck_HttpHealthCheck{
					Path:            healthCheckPath,
					CodecClientType: clusterHealthCodec(cluster),
				},
			},
			Interval:           ptypes.DurationProto(interval),
			Timeout:            ptypes.DurationProto(timeout),
			UnhealthyThreshold: protoUint32orNil(unhealthyThreshold),
			HealthyThreshold:   protoUint32orNil(healthyThreshold),
		}
		if logFile != "" {
			healthCheck.EventLogPath = logFile
		}

		return append([]*core.HealthCheck{}, healthCheck)
	}
	return nil
}

func clusterHealthCodec(cluster shared.Cluster) envoyType.CodecClientType {

	value, err := cluster.Attributes.Get(attributeHTTPProtocol)
	if err == nil {
		switch value {
		case attributeValueHTTPProtocol2:
			return envoyType.CodecClientType_HTTP2

		case attributeValueHTTPProtocol3:
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

	idleTimeout := cluster.Attributes.GetAsDuration(
		attributeIdleTimeout, defaultClusterIdleTimeout)

	return &core.HttpProtocolOptions{
		IdleTimeout: ptypes.DurationProto(idleTimeout),
	}
}

// clusterHTTP2ProtocolOptions returns HTTP/2 parameters
func clusterHTTP2ProtocolOptions(cluster shared.Cluster) *core.Http2ProtocolOptions {

	value, err := cluster.Attributes.Get(attributeHTTPProtocol)
	if err == nil {
		switch value {
		case attributeValueHTTPProtocol11:
			return nil
		case attributeValueHTTPProtocol2:
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

	// Set TLS configuration based upon cluster attributes
	TLSContext := &tls.UpstreamTlsContext{
		Sni:              clusterSNIHostname(cluster),
		CommonTlsContext: buildCommonTLSContext(cluster.Name, cluster.Attributes),
	}
	return buildTransportSocket(cluster.Name, TLSContext)
}

// clusterSNIHostname sets SNI hostname used for upstream connections
func clusterSNIHostname(cluster shared.Cluster) string {

	value, err := cluster.Attributes.Get(attributeSNIHostName)
	if err == nil && value != "" {
		return value
	}
	return cluster.HostName
}

func clusterDNSRefreshRate(cluster shared.Cluster) *duration.Duration {

	refreshInterval := cluster.Attributes.GetAsDuration(
		attributeDNSRefreshRate, defaultDNSRefreshRate)

	return ptypes.DurationProto(refreshInterval)
}

func clusterDNSResolvers(cluster shared.Cluster) []*core.Address {

	value, err := cluster.Attributes.Get(attributeDNSResolvers)
	if err == nil {
		var resolvers []*core.Address

		for _, resolver := range strings.Split(value, ",") {
			resolvers = append(resolvers, buildAddress(resolver, 53))
		}
		return resolvers
	}
	return nil
}
