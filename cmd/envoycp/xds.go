package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/erikbos/apiauth/pkg/types"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	localhost = "127.0.0.1"
	version   int32
)

// StartXDS brings up XDS system
func (s *server) StartXDS() {
	s.xdsCache = cache.NewSnapshotCache(true, Hasher{}, logger{})
	//config := cache.NewSnapshotCache(false, hash{}, logger{})
	signal := make(chan struct{})
	s.xds = xds.NewServer(context.Background(), s.xdsCache, &callbacks{
		signal:   signal,
		fetches:  0,
		requests: 0,
	})

	go s.GRPCManagementServer()
	go s.HTTPManagementGateway()
	go s.GetClusterConfigFromDatabase()

	s.XDSMainloop()
}

// GRPCManagementServer starts grpc xds listener
func (s *server) GRPCManagementServer() {
	log.Info("Starting GRPC XDS on ", s.config.XDSGRPCListen)
	lis, err := net.Listen("tcp", s.config.XDSGRPCListen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, s.xds)
	v2.RegisterEndpointDiscoveryServiceServer(grpcServer, s.xds)
	v2.RegisterClusterDiscoveryServiceServer(grpcServer, s.xds)
	v2.RegisterRouteDiscoveryServiceServer(grpcServer, s.xds)
	v2.RegisterListenerDiscoveryServiceServer(grpcServer, s.xds)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start GRPC server: %v", err)
	}
}

// HTTPManagementGateway starts http xds listener
func (s *server) HTTPManagementGateway() {
	log.Info("Starting HTTP XDS on ", s.config.XDSHTTPListen)
	err := http.ListenAndServe(s.config.XDSHTTPListen, &xds.HTTPGateway{Server: s.xds})
	if err != nil {
		log.Fatalf("failed to HTTP serve: %v", err)
	}
}

func (s *server) XDSMainloop() {
	var lastConfigurationDeployment int64

	for {
		now := types.GetCurrentTimeMilliseconds()

		if clustersLastUpdate > lastConfigurationDeployment {
			log.Infof("Starting configuration compilation")

			// var clusterName = "service_google"
			// log.Infof(">>>>>>>>>>>>>>>>>>> creating cluster " + clusterName)

			//c := []cache.Resource{resource.MakeCluster(resource.Ads, clusterName)}
			/*h := &core.Address{Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Address:  remoteHost,
					Protocol: core.TCP,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: uint32(443),
					},
				},
			}}*/

			EnvoyClusters, _ := getEnvoyClusterConfig(s.db)

			// var listenerName = "listener_0"
			// var targetHost = "www.google.com"
			// var targetPrefix = "/"
			// var virtualHostName = "local_service"
			// var routeConfigName = "local_route"
			// // log.Infof(">>>>>>>>>>>>>>>>>>> creating listener " + listenerName)
			// //virtual host (inside http filter)
			// v := route.VirtualHost{
			// 	Name:    virtualHostName,
			// 	Domains: []string{"*"},
			// 	Routes: []route.Route{{
			// 		Match: route.RouteMatch{
			// 			PathSpecifier: &route.RouteMatch_Prefix{
			// 				Prefix: targetPrefix,
			// 			},
			// 		},
			// 		Action: &route.Route_Route{
			// 			Route: &route.RouteAction{
			// 				HostRewriteSpecifier: &route.RouteAction_HostRewrite{
			// 					HostRewrite: targetHost,
			// 				},
			// 				ClusterSpecifier: &route.RouteAction_Cluster{
			// 					Cluster: clusterName,
			// 				},
			// 			},
			// 		},
			// 	}},
			// }

			// //http filter (inside listener)
			// manager := &hcm.HttpConnectionManager{
			// 	StatPrefix: "ingress_http",
			// 	RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			// 		RouteConfig: &v2.RouteConfiguration{
			// 			Name:         routeConfigName,
			// 			VirtualHosts: []route.VirtualHost{v},
			// 		},
			// 	},
			// 	HttpFilters: []*hcm.HttpFilter{{
			// 		Name: util.Router,
			// 	}},
			// }
			// pbst, err := proto_types.MarshalAny(manager)
			// if err != nil {
			// 	fmt.Println("yellow")
			// 	panic(err)
			// }
			// fmt.Println(pbst.TypeUrl)

			// //create listener
			// var l = []cache.Resource{
			// 	&v2.Listener{
			// 		Name: listenerName,
			// 		Address: core.Address{
			// 			Address: &core.Address_SocketAddress{
			// 				SocketAddress: &core.SocketAddress{
			// 					Protocol: core.TCP,
			// 					Address:  localhost,
			// 					PortSpecifier: &core.SocketAddress_PortValue{
			// 						PortValue: 10000,
			// 					},
			// 				},
			// 			},
			// 		},
			// 		FilterChains: []listener.FilterChain{{
			// 			Filters: []listener.Filter{{
			// 				Name: util.HTTPConnectionManager,
			// 				ConfigType: &listener.Filter_TypedConfig{
			// 					TypedConfig: pbst,
			// 				},
			// 			}},
			// 		}},
			// 	}}

			// log.Printf("l: %+v", l)
			// log.Printf("v: %+v", v)
			// log.Printf("c: %+v", EnvoyClusters)
			atomic.AddInt32(&version, 1)
			// nodeID := config.GetStatusKeys()[1]

			log.Infof("Creating config version " + fmt.Sprint(version))
			// snap := cache.NewSnapshot(fmt.Sprint(version), nil, EnvoyClusters, nil, l)
			snap := cache.NewSnapshot(fmt.Sprint(version), nil, EnvoyClusters, nil, nil, nil)

			_ = s.xdsCache.SetSnapshot("jenny", snap)

			lastConfigurationDeployment = now
		}
		time.Sleep(1 * time.Second)
	}
}
