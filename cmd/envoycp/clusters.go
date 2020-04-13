package main

import (
	"sync"
	"time"

	"github.com/erikbos/apiauth/pkg/shared"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/sirupsen/logrus"
)

var clusters []shared.Cluster
var clustersLastUpdate int64
var clusterMutex sync.Mutex

// FIXME this should be implemented using channels
// FIXME this does not detect removed records

// getClusterConfigFromDatabase continously gets the current configuration
func (s *server) GetClusterConfigFromDatabase() {
	for {
		newClusterList, err := s.db.GetClusters()
		if err != nil {
			log.Errorf("Could not retrieve clusters from database (%s)", err)
		} else {
			for _, s := range newClusterList {
				// Is a cluster updated since last time we stored it?
				if s.LastmodifiedAt > clustersLastUpdate {
					now := shared.GetCurrentTimeMilliseconds()
					clusterMutex.Lock()
					clusters = newClusterList
					clustersLastUpdate = now
					clusterMutex.Unlock()
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// getClusterConfig returns array of all envoy clusters
func getEnvoyClusterConfig() ([]cache.Resource, error) {
	envoyClusters := []cache.Resource{}
	for _, s := range clusters {
		envoyClusters = append(envoyClusters, buildEnvoyClusterConfig(s))
	}
	return envoyClusters, nil
}

// buildEnvoyClusterConfig buils one envoy cluster configuration
func buildEnvoyClusterConfig(clusterConfig shared.Cluster) *api.Cluster {
	address := coreAddress(clusterConfig.HostName, clusterConfig.Port)
	cluster := &api.Cluster{
		Name:           clusterConfig.Name,
		ConnectTimeout: ptypes.DurationProto(2 * time.Second),
		ClusterDiscoveryType: &api.Cluster_Type{
			Type: api.Cluster_LOGICAL_DNS,
		},
		DnsLookupFamily: api.Cluster_V4_ONLY,
		LbPolicy:        api.Cluster_ROUND_ROBIN,
		LoadAssignment: &api.ClusterLoadAssignment{
			ClusterName: clusterConfig.Name,
			Endpoints: []*endpoint.LocalityLbEndpoints{
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
			},
		},
	}

	value, err := shared.GetAttribute(clusterConfig.Attributes, "HTTP2Enabled")
	if err == nil && value == "True" {
		cluster.Http2ProtocolOptions = &core.Http2ProtocolOptions{}
		cluster.TransportSocket = transportSocket(clusterConfig.HostName, true)
	} else {
		cluster.TransportSocket = transportSocket(clusterConfig.HostName, false)
	}

	cluster.HealthChecks = buildHealthCheckConfig(clusterConfig)

	// FIXME add circuit breaker support
	// cluster.CircuitBreakers = ...

	log.Debugf("buildEnvoyClusterConfig: %+v", cluster)
	return cluster
}

// buildHealthCheckConfig builds health configuration for a cluster
func buildHealthCheckConfig(clusterConfig shared.Cluster) []*core.HealthCheck {
	healthChecks := []*core.HealthCheck{}

	healthCheckPath, err := shared.GetAttribute(clusterConfig.Attributes, "HealthCheckPath")
	if err == nil && healthCheckPath != "" {
		healthCheckInterval, err := shared.GetAttribute(clusterConfig.Attributes, "HealthCheckInterval")
		healthcheckIntervalAsDuration, err := time.ParseDuration(healthCheckInterval)
		if err != nil {
			healthcheckIntervalAsDuration = 30 * time.Second
		}

		healthCheckTimeout, err := shared.GetAttribute(clusterConfig.Attributes, "HealthCheckTimeout")
		healthcheckTimeoutAsDuration, err := time.ParseDuration(healthCheckTimeout)
		if err != nil {
			healthcheckTimeoutAsDuration = 30 * time.Second
		}

		healthCheck := &core.HealthCheck{
			Interval:           ptypes.DurationProto(healthcheckIntervalAsDuration),
			Timeout:            ptypes.DurationProto(healthcheckTimeoutAsDuration),
			UnhealthyThreshold: &wrappers.UInt32Value{Value: 2},
			HealthyThreshold:   &wrappers.UInt32Value{Value: 1},
			HealthChecker: &core.HealthCheck_HttpHealthCheck_{
				HttpHealthCheck: &core.HealthCheck_HttpHealthCheck{
					Path: healthCheckPath,
				},
			},
			EventLogPath: "/tmp/healthcheck",
		}
		healthChecks = append(healthChecks, healthCheck)
	}
	return healthChecks
}

// coreAddress builds an Envoy Address to connect to
func coreAddress(hostname string, port int) *core.Address {
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

// transportSocket builds an Envoy TransportSocket sets HTTP protocol(s) to be used
func transportSocket(sniHostname string, http2Enabled bool) *core.TransportSocket {
	TLSContext := &auth.UpstreamTlsContext{
		Sni: sniHostname,
	}
	if http2Enabled {
		TLSContext.CommonTlsContext = &auth.CommonTlsContext{
			AlpnProtocols: []string{"h2", "http/1.1"},
		}
	} else {
		TLSContext.CommonTlsContext = &auth.CommonTlsContext{
			AlpnProtocols: []string{"http/1.1"},
		}
	}
	tlsContext, err := ptypes.MarshalAny(TLSContext)
	if err != nil {
		return nil
	}
	return &core.TransportSocket{
		Name: "tls",
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsContext,
		},
	}
}
