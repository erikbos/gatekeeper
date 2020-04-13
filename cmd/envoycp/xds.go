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
	var version int32
	var lastConfigurationDeployment int64

	for {
		now := types.GetCurrentTimeMilliseconds()

		if clustersLastUpdate > lastConfigurationDeployment {
			log.Infof("Starting configuration compilation")

			EnvoyClusters, _ := getEnvoyClusterConfig()
			EnvoyRoutes, _ := getEnvoyRouteConfig()

			atomic.AddInt32(&version, 1)
			log.Infof("Creating config version " + fmt.Sprint(version))
			snap := cache.NewSnapshot(fmt.Sprint(version), nil, EnvoyClusters, EnvoyRoutes, nil, nil)
			_ = s.xdsCache.SetSnapshot("jenny", snap)

			lastConfigurationDeployment = now
		}
		time.Sleep(1 * time.Second)
	}
}
