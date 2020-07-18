package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// xdsNotifyMesssage is used to notify
type xdsNotifyMesssage struct {
	resource string
}

// StartXDS brings up XDS system
func (s *server) StartXDS(notifications chan xdsNotifyMesssage) {

	s.xdsCache = cache.NewSnapshotCache(true, cache.IDHash{}, logger{})
	streamCallbacks := &callbacks{
		signal:   make(chan struct{}),
		fetches:  0,
		requests: 0,
		srv:      s,
	}
	s.xds = xds.NewServer(context.Background(), s.xdsCache, streamCallbacks)

	go s.GRPCManagementServer()
	go s.HTTPManagementGateway()

	for {
		select {
		case n := <-notifications:
			log.Infof("XDS change notify received for resource group '%s'", n.resource)

			s.XDSBuildSnapshot()

		case <-time.After(s.config.XDS.ConfigPushInterval):
			// Nothing, just wait ConfigPushInterval seconds everytime
		}
	}
}

// GRPCManagementServer starts grpc xds listener
func (s *server) GRPCManagementServer() {

	log.Info("GRPC XDS listening on ", s.config.XDS.GRPCListen)
	lis, err := net.Listen("tcp", s.config.XDS.GRPCListen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	discoveryservice.RegisterAggregatedDiscoveryServiceServer(grpcServer, s.xds)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, s.xds)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, s.xds)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, s.xds)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, s.xds)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start GRPC server: %v", err)
	}
}

type xdsHTTPGatewayWrapper struct {
	Server xds.HTTPGateway
}

// this wrapper is required:
// - xds.HTTPGateway function proto mismatches http.Handler because it returns error
// - we can pass xdsHTTPGatewayWrapper to http.ListenAndServe
func (h xdsHTTPGatewayWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Server.ServeHTTP(r)
}

// HTTPManagementGateway starts http xds listener
func (s *server) HTTPManagementGateway() {

	log.Info("HTTP XDS listening on ", s.config.XDS.HTTPListen)
	err := http.ListenAndServe(s.config.XDS.HTTPListen,
		xdsHTTPGatewayWrapper{
			Server: xds.HTTPGateway{},
		})
	if err != nil {
		log.Fatalf("failed to HTTP serve: %v", err)
	}
}

func (s *server) XDSBuildSnapshot() {

	version := fmt.Sprint(shared.GetCurrentTimeMilliseconds())
	log.Infof("Creating configuration snapshot version %s", version)

	EnvoyClusters, _ := s.getEnvoyClusterConfig()
	EnvoyRoutes, _ := s.getEnvoyRouteConfig()
	EnvoyListeners, _ := s.getEnvoyListenerConfig()

	snapshot := cache.NewSnapshot(version, nil, EnvoyClusters, EnvoyRoutes, EnvoyListeners, nil)
	_ = s.xdsCache.SetSnapshot("jenny", snapshot)
}
