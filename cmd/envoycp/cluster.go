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

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	unknownClusterAttributeWarning = "Cluster '%s' has attribute '%s' with unknown value '%s'"
)

// getClusterConfig returns array of all envoy clusters
func (s *server) getEnvoyClusterConfig() ([]cache.Resource, error) {

	envoyClusters := []cache.Resource{}

	for _, cluster := range s.dbentities.GetClusters() {
		if err := cluster.ConfigCheck(); err != nil {
			log.Warningf("Cluster '%s' %s", cluster.Name, err)
		}
		if clusterToAdd := buildEnvoyClusterConfig(cluster); clusterToAdd != nil {
			envoyClusters = append(envoyClusters, clusterToAdd)
		} else {
			log.Warningf("Cluster '%s' was not added", cluster.Name)
		}
	}

	return envoyClusters, nil
}

// buildEnvoyClusterConfig builds one envoy cluster configuration
func buildEnvoyClusterConfig(cluster types.Cluster) *envoyCluster.Cluster {

	envoyCluster := &envoyCluster.Cluster{
		Name:                      cluster.Name,
		ConnectTimeout:            clusterConnectTimeout(cluster),
		ClusterDiscoveryType:      &envoyCluster.Cluster_Type{Type: envoyCluster.Cluster_LOGICAL_DNS},
		DnsLookupFamily:           clusterDNSLookupFamily(cluster),
		DnsResolvers:              clusterDNSResolvers(cluster),
		DnsRefreshRate:            clusterDNSRefreshRate(cluster),
		LbPolicy:                  clusterLbPolicy(cluster),
		HealthChecks:              clusterHealthChecks(cluster),
		CommonHttpProtocolOptions: clusterCommonHTTPProtocolOptions(cluster),
		CircuitBreakers:           clusterCircuitBreakers(cluster),
		TrackClusterStats:         clusterTrackClusterStats(cluster),
	}

	loadAssignment := clusterLoadAssignment(cluster)
	if loadAssignment == nil {
		log.Warnf("Cluster '%s' could not set destination hostname and port", cluster.Name)
		return nil
	}
	envoyCluster.LoadAssignment = loadAssignment

	// Add TLS and HTTP/2 configuration options in case we want to
	value, err := cluster.Attributes.Get(types.AttributeTLS)
	if err == nil && value == types.AttributeValueTrue {
		envoyCluster.TransportSocket = clusterTransportSocket(cluster)
		envoyCluster.Http2ProtocolOptions = clusterHTTP2ProtocolOptions(cluster)
	}

	return envoyCluster
}

func clusterConnectTimeout(cluster types.Cluster) *duration.Duration {

	connectTimeout := cluster.Attributes.GetAsDuration(
		types.AttributeConnectTimeout, types.DefaultClusterConnectTimeout)

	return ptypes.DurationProto(connectTimeout)
}

func clusterLbPolicy(cluster types.Cluster) envoyCluster.Cluster_LbPolicy {

	value, err := cluster.Attributes.Get(types.AttributeLbPolicy)
	if err == nil {
		switch value {
		case types.AttributeValueLBRoundRobin:
			return envoyCluster.Cluster_ROUND_ROBIN

		case types.AttributeValueLBLeastRequest:
			return envoyCluster.Cluster_LEAST_REQUEST

		case types.AttributeValueLBRingHash:
			return envoyCluster.Cluster_RING_HASH

		case types.AttributeValueLBRandom:
			return envoyCluster.Cluster_RANDOM

		case types.AttributeValueLBMaglev:
			return envoyCluster.Cluster_MAGLEV

		default:
			log.Warnf(unknownClusterAttributeWarning,
				cluster.Name, types.AttributeLbPolicy, value)
		}
	}
	return envoyCluster.Cluster_ROUND_ROBIN
}

// clusterLoadAssignment sets cluster loadbalance based upon hostname & port attributes
func clusterLoadAssignment(cluster types.Cluster) *endpoint.ClusterLoadAssignment {

	hostName, err := cluster.Attributes.Get(types.AttributeHost)
	if err != nil {
		return nil
	}
	port := cluster.Attributes.GetAsUInt32(types.AttributePort, 0)
	if port == 0 {
		return nil
	}
	return &endpoint.ClusterLoadAssignment{
		ClusterName: cluster.Name,
		Endpoints: []*endpoint.LocalityLbEndpoints{
			{
				LbEndpoints: []*endpoint.LbEndpoint{
					{
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: buildAddress(hostName, port),
							},
						},
					},
				},
			},
		},
	}
}

func clusterCircuitBreakers(cluster types.Cluster) *envoyCluster.CircuitBreakers {

	maxConnections := cluster.Attributes.GetAsUInt32(types.AttributeMaxConnections, 0)
	maxPendingRequests := cluster.Attributes.GetAsUInt32(types.AttributeMaxPendingRequests, 0)
	maxRequests := cluster.Attributes.GetAsUInt32(types.AttributeMaxRequests, 0)
	maxRetries := cluster.Attributes.GetAsUInt32(types.AttributeMaxRetries, 0)

	return &envoyCluster.CircuitBreakers{
		Thresholds: []*envoyCluster.CircuitBreakers_Thresholds{{
			MaxConnections:     protoUint32orNil(maxConnections),
			MaxPendingRequests: protoUint32orNil(maxPendingRequests),
			MaxRequests:        protoUint32orNil(maxRequests),
			MaxRetries:         protoUint32orNil(maxRetries),
		}},
	}
}

// clusterTrackClusterStats build cluster statistics configuration
func clusterTrackClusterStats(cluster types.Cluster) *envoyCluster.TrackClusterStats {

	return &envoyCluster.TrackClusterStats{
		TimeoutBudgets:       true,
		RequestResponseSizes: true,
	}
}

// clusterHealthCheckConfig builds health configuration for a cluster
func clusterHealthChecks(cluster types.Cluster) []*core.HealthCheck {

	healthCheckProtocol, err := cluster.Attributes.Get(types.AttributeHealthCheckProtocol)
	healthCheckPath, _ := cluster.Attributes.Get(types.AttributeHealthCheckPath)

	if err == nil && healthCheckProtocol == types.AttributeValueHealthCheckProtocolHTTP && healthCheckPath != "" {

		hostName := cluster.Attributes.GetAsString(types.AttributeHost, "")

		healthCheckHostName := cluster.Attributes.GetAsString(types.AttributeHealthHostHeader, hostName)

		interval := cluster.Attributes.GetAsDuration(types.AttributeHealthCheckInterval,
			types.DefaultHealthCheckInterval)

		timeout := cluster.Attributes.GetAsDuration(types.AttributeHealthCheckTimeout,
			types.DefaultHealthCheckTimeout)

		unhealthyThreshold := cluster.Attributes.GetAsUInt32(types.AttributeHealthCheckUnhealthyThreshold,
			types.DefaultHealthCheckUnhealthyThreshold)

		healthyThreshold := cluster.Attributes.GetAsUInt32(types.AttributeHealthCheckHealthyThreshold,
			types.DefaultHealthCheckHealthyThreshold)

		logFile := cluster.Attributes.GetAsString(types.AttributeHealthCheckLogFile, "")

		healthCheck := &core.HealthCheck{
			HealthChecker: &core.HealthCheck_HttpHealthCheck_{
				HttpHealthCheck: &core.HealthCheck_HttpHealthCheck{
					Host:            healthCheckHostName,
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

func clusterHealthCodec(cluster types.Cluster) envoyType.CodecClientType {

	value, err := cluster.Attributes.Get(types.AttributeHTTPProtocol)
	if err == nil {
		switch value {
		case types.AttributeValueHTTPProtocol11:
			return envoyType.CodecClientType_HTTP1

		case types.AttributeValueHTTPProtocol2:
			return envoyType.CodecClientType_HTTP2

		case types.AttributeValueHTTPProtocol3:
			return envoyType.CodecClientType_HTTP3

		default:
			log.Warnf(unknownClusterAttributeWarning,
				cluster.Name, types.AttributeHTTPProtocol, value)
		}
	}
	return envoyType.CodecClientType_HTTP1
}

// clusterCommonHTTPProtocolOptions sets HTTP options applicable to both HTTP/1 and /2
func clusterCommonHTTPProtocolOptions(cluster types.Cluster) *core.HttpProtocolOptions {

	idleTimeout := cluster.Attributes.GetAsDuration(
		types.AttributeIdleTimeout, types.DefaultClusterIdleTimeout)

	return &core.HttpProtocolOptions{
		IdleTimeout: ptypes.DurationProto(idleTimeout),
	}
}

// clusterHTTP2ProtocolOptions returns HTTP/2 parameters
func clusterHTTP2ProtocolOptions(cluster types.Cluster) *core.Http2ProtocolOptions {

	value, err := cluster.Attributes.Get(types.AttributeHTTPProtocol)
	if err == nil {
		switch value {
		case types.AttributeValueHTTPProtocol11:
			return nil
		case types.AttributeValueHTTPProtocol2:
			// according to spec we need to return at least empty struct to enable HTTP/2
			return &core.Http2ProtocolOptions{}
		default:
			log.Warnf(unknownClusterAttributeWarning,
				cluster.Name, types.AttributeHTTPProtocol, value)
			return nil
		}
	}
	// Attribute was not set, we do not have to set HTTP/2 options
	return nil
}

// clusterTransportSocket configures TLS settings
func clusterTransportSocket(cluster types.Cluster) *core.TransportSocket {

	// Set TLS configuration based upon cluster attributes
	TLSContext := &tls.UpstreamTlsContext{
		Sni:              clusterSNIHostname(cluster),
		CommonTlsContext: buildCommonTLSContext(cluster.Name, cluster.Attributes),
	}
	return buildTransportSocket(cluster.Name, TLSContext)
}

// clusterSNIHostname sets SNI hostname used for upstream connections
func clusterSNIHostname(cluster types.Cluster) string {

	// Let's check SNI attribute first
	value, err := cluster.Attributes.Get(types.AttributeSNIHostName)
	if err == nil && value != "" {
		return value
	}
	// If not we will fallback to cluster hostname
	value, err = cluster.Attributes.Get(types.AttributeHost)
	if err == nil && value != "" {
		return value
	}
	return ""
}

func clusterDNSLookupFamily(cluster types.Cluster) envoyCluster.Cluster_DnsLookupFamily {

	value, err := cluster.Attributes.Get(types.AttributeDNSLookupFamiliy)
	if err == nil {
		switch value {
		case types.AttributeValueDNSIPV4Only:
			return envoyCluster.Cluster_V4_ONLY
		case types.AttributeValueDNSIPV6Only:
			return envoyCluster.Cluster_V6_ONLY
		case types.AttributeValueDNSAUTO:
			return envoyCluster.Cluster_AUTO
		default:
			log.Warnf(unknownClusterAttributeWarning,
				cluster.Name, types.AttributeDNSLookupFamiliy, value)
		}
	}
	return envoyCluster.Cluster_AUTO
}

func clusterDNSRefreshRate(cluster types.Cluster) *duration.Duration {

	refreshInterval := cluster.Attributes.GetAsDuration(
		types.AttributeDNSRefreshRate, types.DefaultDNSRefreshRate)

	return ptypes.DurationProto(refreshInterval)
}

func clusterDNSResolvers(cluster types.Cluster) []*core.Address {

	value, err := cluster.Attributes.Get(types.AttributeDNSResolvers)
	if err == nil && value != "" {
		var resolvers []*core.Address

		for _, resolver := range strings.Split(value, ",") {
			resolvers = append(resolvers, buildAddress(resolver, 53))
		}
		return resolvers
	}
	return nil
}
