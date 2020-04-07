package main

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/erikbos/apiauth/pkg/db"
)

func getClusterConfig(db *db.Database) ([]cache.Resource, error) {
	// FIXME this should be retrieve asynchronously
	clusters, err := db.GetClusters()
	if err != nil {
		msg := fmt.Sprintf("Could not retrieve clusters from database %v", err)
		log.Error(msg)
		return nil, errors.New(msg)
	}

	EnvoyClusters := []cache.Resource{}
	for i, s := range clusters {
		fmt.Println(i, s)
		EnvoyClusters = append(EnvoyClusters,
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
	return EnvoyClusters, nil
}
