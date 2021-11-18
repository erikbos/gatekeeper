package main

import (
	"strings"
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_upstreams "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/golang/protobuf/ptypes/duration"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	unknownClusterAttributeValueWarning = "Unsupported attribute value"
)

// getClusterConfig returns array of all envoy clusters
func (s *server) getEnvoyClusterConfig() ([]cache.Resource, error) {

	envoyClusters := []cache.Resource{}

	for _, cluster := range s.dbentities.GetClusters() {
		if err := cluster.Validate(); err != nil {
			s.logger.Warn("Cluster has unsupported configuration",
				zap.String("cluster", cluster.Name), zap.Error(err))
		}
		if clusterToAdd := s.buildEnvoyClusterConfig(cluster); clusterToAdd != nil {
			envoyClusters = append(envoyClusters, clusterToAdd)
		} else {
			s.logger.Warn("Cluster not added", zap.String("cluster", cluster.Name))
		}
	}

	return envoyClusters, nil
}

// buildEnvoyClusterConfig builds one envoy cluster configuration
func (s *server) buildEnvoyClusterConfig(cluster types.Cluster) *envoy_cluster.Cluster {

	envoyCluster := &envoy_cluster.Cluster{
		Name:                          cluster.Name,
		ConnectTimeout:                s.clusterConnectTimeout(cluster),
		ClusterDiscoveryType:          &envoy_cluster.Cluster_Type{Type: envoy_cluster.Cluster_LOGICAL_DNS},
		DnsLookupFamily:               s.clusterDNSLookupFamily(cluster),
		DnsResolvers:                  s.clusterDNSResolvers(cluster),
		DnsRefreshRate:                s.clusterDNSRefreshRate(cluster),
		LbPolicy:                      s.clusterLbPolicy(cluster),
		HealthChecks:                  s.clusterHealthChecks(cluster),
		CircuitBreakers:               s.clusterCircuitBreakers(cluster),
		TrackClusterStats:             s.clusterTrackClusterStats(cluster),
		TypedExtensionProtocolOptions: s.clusterTypedExtensionProtocolOptions(cluster),
	}

	loadAssignment := s.clusterLoadAssignment(cluster)
	if loadAssignment == nil {
		s.logger.Warn("Cannot set destination host or port", zap.String("cluster", cluster.Name))
		return nil
	}
	envoyCluster.LoadAssignment = loadAssignment

	// Add TLS and HTTP/2 configuration options in case we want to
	value, err := cluster.Attributes.Get(types.AttributeTLS)
	if err == nil && value == types.AttributeValueTrue {
		envoyCluster.TransportSocket = s.clusterTransportSocket(cluster)
	}

	return envoyCluster
}

func (s *server) clusterConnectTimeout(cluster types.Cluster) *duration.Duration {

	connectTimeout := cluster.Attributes.GetAsDuration(
		types.AttributeConnectTimeout, types.DefaultClusterConnectTimeout)

	return durationpb.New(connectTimeout)
}

func (s *server) clusterLbPolicy(cluster types.Cluster) envoy_cluster.Cluster_LbPolicy {

	value, err := cluster.Attributes.Get(types.AttributeLbPolicy)
	if err == nil {
		switch value {
		case types.AttributeValueLBRoundRobin:
			return envoy_cluster.Cluster_ROUND_ROBIN

		case types.AttributeValueLBLeastRequest:
			return envoy_cluster.Cluster_LEAST_REQUEST

		case types.AttributeValueLBRingHash:
			return envoy_cluster.Cluster_RING_HASH

		case types.AttributeValueLBRandom:
			return envoy_cluster.Cluster_RANDOM

		case types.AttributeValueLBMaglev:
			return envoy_cluster.Cluster_MAGLEV

		default:
			s.logger.Warn(unknownClusterAttributeValueWarning,
				zap.String("cluster", cluster.Name),
				zap.String("attribute", types.AttributeLbPolicy))
		}
	}
	return envoy_cluster.Cluster_ROUND_ROBIN
}

// clusterLoadAssignment sets cluster loadbalance based upon hostname & port attributes
func (s *server) clusterLoadAssignment(cluster types.Cluster) *envoy_endpoint.ClusterLoadAssignment {

	hostName, err := cluster.Attributes.Get(types.AttributeHost)
	if err != nil {
		return nil
	}
	port := cluster.Attributes.GetAsUInt32(types.AttributePort, 0)
	if port == 0 {
		return nil
	}
	return &envoy_endpoint.ClusterLoadAssignment{
		ClusterName: cluster.Name,
		Endpoints: []*envoy_endpoint.LocalityLbEndpoints{
			{
				LbEndpoints: []*envoy_endpoint.LbEndpoint{
					{
						HostIdentifier: &envoy_endpoint.LbEndpoint_Endpoint{
							Endpoint: &envoy_endpoint.Endpoint{
								Address: buildAddress(hostName, port),
							},
						},
					},
				},
			},
		},
	}
}

func (s *server) clusterCircuitBreakers(cluster types.Cluster) *envoy_cluster.CircuitBreakers {

	maxConnections := cluster.Attributes.GetAsUInt32(types.AttributeMaxConnections, 0)
	maxPendingRequests := cluster.Attributes.GetAsUInt32(types.AttributeMaxPendingRequests, 0)
	maxRequests := cluster.Attributes.GetAsUInt32(types.AttributeMaxRequests, 0)
	maxRetries := cluster.Attributes.GetAsUInt32(types.AttributeMaxRetries, 0)

	return &envoy_cluster.CircuitBreakers{
		Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{{
			MaxConnections:     protoUint32orNil(maxConnections),
			MaxPendingRequests: protoUint32orNil(maxPendingRequests),
			MaxRequests:        protoUint32orNil(maxRequests),
			MaxRetries:         protoUint32orNil(maxRetries),
		}},
	}
}

// clusterTrackClusterStats build cluster statistics configuration
func (s *server) clusterTrackClusterStats(cluster types.Cluster) *envoy_cluster.TrackClusterStats {

	return &envoy_cluster.TrackClusterStats{
		TimeoutBudgets:       true,
		RequestResponseSizes: true,
	}
}

// clusterHealthCheckConfig builds health configuration for a cluster
func (s *server) clusterHealthChecks(cluster types.Cluster) []*envoy_core.HealthCheck {

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

		healthCheck := &envoy_core.HealthCheck{
			HealthChecker: &envoy_core.HealthCheck_HttpHealthCheck_{
				HttpHealthCheck: &envoy_core.HealthCheck_HttpHealthCheck{
					Host:            healthCheckHostName,
					Path:            healthCheckPath,
					CodecClientType: s.clusterHealthCodec(cluster),
				},
			},
			Interval:           durationpb.New(interval),
			Timeout:            durationpb.New(timeout),
			UnhealthyThreshold: protoUint32orNil(unhealthyThreshold),
			HealthyThreshold:   protoUint32orNil(healthyThreshold),
		}
		if logFile != "" {
			healthCheck.EventLogPath = logFile
		}

		return append([]*envoy_core.HealthCheck{}, healthCheck)
	}
	return nil
}

func (s *server) clusterHealthCodec(cluster types.Cluster) envoy_type.CodecClientType {

	value, err := cluster.Attributes.Get(types.AttributeHTTPProtocol)
	if err == nil {
		switch value {
		case types.AttributeValueHTTPProtocol11:
			return envoy_type.CodecClientType_HTTP1

		case types.AttributeValueHTTPProtocol2:
			return envoy_type.CodecClientType_HTTP2

		case types.AttributeValueHTTPProtocol3:
			return envoy_type.CodecClientType_HTTP3

		default:
			s.logger.Warn(unknownClusterAttributeValueWarning,
				zap.String("cluster", cluster.Name),
				zap.String("attribute", types.AttributeHTTPProtocol))
		}
	}
	return envoy_type.CodecClientType_HTTP1
}

// clusterTransportSocket configures TLS settings
func (s *server) clusterTransportSocket(cluster types.Cluster) *envoy_core.TransportSocket {

	// Set TLS configuration based upon cluster attributes
	TLSContext := &envoy_tls.UpstreamTlsContext{
		Sni:              s.clusterSNIHostname(cluster),
		CommonTlsContext: buildCommonTLSContext(cluster.Name, cluster.Attributes),
	}
	return buildTransportSocket(cluster.Name, TLSContext)
}

// clusterSNIHostname sets SNI hostname used for upstream connections
func (s *server) clusterSNIHostname(cluster types.Cluster) string {

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

func (s *server) clusterDNSLookupFamily(cluster types.Cluster) envoy_cluster.Cluster_DnsLookupFamily {

	value, err := cluster.Attributes.Get(types.AttributeDNSLookupFamily)
	if err == nil {
		switch value {
		case types.AttributeValueDNSIPV4Only:
			return envoy_cluster.Cluster_V4_ONLY
		case types.AttributeValueDNSIPV6Only:
			return envoy_cluster.Cluster_V6_ONLY
		case types.AttributeValueDNSAUTO:
			return envoy_cluster.Cluster_AUTO
		default:
			s.logger.Warn(unknownClusterAttributeValueWarning,
				zap.String("cluster", cluster.Name),
				zap.String("attribute", types.AttributeDNSLookupFamily))
		}
	}
	return envoy_cluster.Cluster_AUTO
}

func (s *server) clusterDNSRefreshRate(cluster types.Cluster) *duration.Duration {

	refreshInterval := cluster.Attributes.GetAsDuration(
		types.AttributeDNSRefreshRate, types.DefaultDNSRefreshRate)

	return durationpb.New(refreshInterval)
}

func (s *server) clusterDNSResolvers(cluster types.Cluster) []*envoy_core.Address {

	value, err := cluster.Attributes.Get(types.AttributeDNSResolvers)
	if err == nil && value != "" {
		var resolvers []*envoy_core.Address

		for _, resolver := range strings.Split(value, ",") {
			resolvers = append(resolvers, buildAddress(resolver, 53))
		}
		return resolvers
	}
	return nil
}

func (s *server) clusterTypedExtensionProtocolOptions(cluster types.Cluster) map[string]*anypb.Any {

	idleTimeout, _ := cluster.Attributes.Get(types.AttributeIdleTimeout)
	clusterHTTPProtocol, _ := cluster.Attributes.Get(types.AttributeHTTPProtocol)

	if idleTimeout == "" && clusterHTTPProtocol == "" {
		return nil
	}

	httpProtocolOptions := &envoy_upstreams.HttpProtocolOptions{
		UpstreamProtocolOptions: &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{},
			},
		},
	}
	if idleTimeoutDuration, err := time.ParseDuration(idleTimeout); err == nil {
		httpProtocolOptions.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{
			IdleTimeout: durationpb.New(idleTimeoutDuration),
		}
	}

	if clusterHTTPProtocol != "" {
		switch clusterHTTPProtocol {
		case types.AttributeValueHTTPProtocol11:
			httpProtocolOptions.UpstreamProtocolOptions = &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{},
				},
			}
		case types.AttributeValueHTTPProtocol2:
			httpProtocolOptions.UpstreamProtocolOptions = &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{},
				},
			}
		case types.AttributeValueHTTPProtocol3:
			httpProtocolOptions.UpstreamProtocolOptions = &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &envoy_upstreams.HttpProtocolOptions_ExplicitHttpConfig_Http3ProtocolOptions{},
				},
			}
		default:
			s.logger.Warn(unknownClusterAttributeValueWarning,
				zap.String("cluster", cluster.Name),
				zap.String("attribute", types.AttributeHTTPProtocol))
			return nil
		}
	}

	httpProtocolOptionsTypedConf, e := anypb.New(httpProtocolOptions)
	if e != nil {
		s.logger.Panic("clusterTypedExtensionProtocolOptions", zap.Error(e))
	}
	return map[string]*anypb.Any{
		"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": httpProtocolOptionsTypedConf,
	}
}
