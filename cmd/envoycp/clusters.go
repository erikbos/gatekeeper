package main

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/types"
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

func getClusterConfig(db *db.Database) ([]cache.Resource, error) {
	envoyClusters := []cache.Resource{}
	for _, s := range clusters {
		envoyClusters = append(envoyClusters,
			&v2.Cluster{
				Name:           s.Name,
				ConnectTimeout: 1 * time.Second,
				ClusterDiscoveryType: &v2.Cluster_Type{
					Type: v2.Cluster_LOGICAL_DNS,
				},
				DnsLookupFamily: v2.Cluster_V4_ONLY,
				LbPolicy:        v2.Cluster_ROUND_ROBIN,
				LoadAssignment: &v2.ClusterLoadAssignment{
					ClusterName: s.Name,
					Endpoints: []endpoint.LocalityLbEndpoints{
						{
							LbEndpoints: []endpoint.LbEndpoint{
								{
									HostIdentifier: &endpoint.LbEndpoint_Endpoint{
										Endpoint: &endpoint.Endpoint{
											Address: &core.Address{
												Address: &core.Address_SocketAddress{
													SocketAddress: &core.SocketAddress{
														Address:  s.HostName,
														Protocol: core.TCP,
														PortSpecifier: &core.SocketAddress_PortValue{
															PortValue: uint32(s.HostPort),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				TlsContext: &auth.UpstreamTlsContext{
					Sni: s.HostName,
				},
			},
		)
	}
	return envoyClusters, nil
}
