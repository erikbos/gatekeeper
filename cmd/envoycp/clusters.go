package main

import (
	"sync"
	"time"

	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/types"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/sirupsen/logrus"
)

var clusters []types.Cluster
var clustersLastUpdate int64
var mux sync.Mutex

// FIXME this should be implemented using channels
// FIXME this does not detect removed records

// getClusterConfigFromDatabase continously gets the current configuration
func getClusterConfigFromDatabase(db *db.Database) {
	for {
		newClusterList, err := db.GetClusters()
		if err != nil {
			log.Errorf("Could not retrieve clusters from database (%s)", err)
		} else {
			for _, s := range newClusterList {
				// Is a cluster updated since last time we stored it?
				if s.LastmodifiedAt > clustersLastUpdate {
					now := types.GetCurrentTimeMilliseconds()
					mux.Lock()
					clusters = newClusterList
					clustersLastUpdate = now
					mux.Unlock()
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// getClusterConfig returns array of envoyclusters
func getEnvoyClusterConfig(db *db.Database) ([]cache.Resource, error) {
	envoyClusters := []cache.Resource{}
	for _, s := range clusters {
		envoyClusters = append(envoyClusters, buildEnvoyClusterConfigV2(s))
	}
	return envoyClusters, nil
}

func buildEnvoyClusterConfigV2(clusterConfig types.Cluster) *api.Cluster {
	address := returnCoreAddress(clusterConfig.HostName, clusterConfig.Port)
	envoyCluster := &api.Cluster{
		Name:           clusterConfig.Name,
		ConnectTimeout: ptypes.DurationProto(1 * time.Second),
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
	envoyCluster.TransportSocket = returnTransportSocket(clusterConfig.HostName)
	envoyCluster.HealthChecks = buildHealthCheckConfig(clusterConfig)

	value, err := types.GetAttribute(clusterConfig.Attributes, "HTTP2Enabled")
	if err == nil && value == "True" {
		envoyCluster.Http2ProtocolOptions = &core.Http2ProtocolOptions{}
	}

	log.Debugf("buildEnvoyClusterConfigV2: %v", envoyCluster)
	return envoyCluster
}

func buildHealthCheckConfig(clusterConfig types.Cluster) []*core.HealthCheck {
	healthChecks := []*core.HealthCheck{}

	healthCheckPath, err := types.GetAttribute(clusterConfig.Attributes, "HealthCheckPath")
	if err == nil && healthCheckPath != "" {
		healthCheckInterval, err := types.GetAttribute(clusterConfig.Attributes, "HealthCheckInterval")
		healthcheckIntervalAsDuration, err := time.ParseDuration(healthCheckInterval)
		if err != nil {
			healthcheckIntervalAsDuration = 30 * time.Second
		}

		healthCheckTimeout, err := types.GetAttribute(clusterConfig.Attributes, "HealthCheckTimeout")
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

func returnCoreAddress(hostname string, port int) *core.Address {
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

func returnTransportSocket(sniHostname string) *core.TransportSocket {
	tlsContext, err := ptypes.MarshalAny(&auth.UpstreamTlsContext{
		Sni: sniHostname,
	})
	if err != nil {
		panic(err)
	}
	return &core.TransportSocket{
		Name: "tls",
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsContext,
		},
	}
}
