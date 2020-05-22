package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// StartXDS brings up XDS system
func (s *server) StartXDS(notifications chan xdsNotifyMesssage) {

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

	for {
		select {
		case n := <-notifications:
			log.Infof("XDS change notify received for resource '%s'", n.resource)

			s.XDSBuildSnapshot()

		case <-time.After(s.config.XDS.ConfigPushInterval):
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

	log.Info("HTTP XDS listening on ", s.config.XDS.HTTPListen)
	err := http.ListenAndServe(s.config.XDS.HTTPListen, &xds.HTTPGateway{Server: s.xds})
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

// registerMetrics registers xds' operational metrics
func (s *server) registerMetrics() {

	metricVirtualHostsCount := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: myName,
			Name:      "xds_virtualhosts_total",
			Help:      "Total number of clusters.",
		}, s.GetVirtualHostCount)

	metricRoutesCount := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: myName,
			Name:      "xds_routes_total",
			Help:      "Total number of routes.",
		}, s.GetRouteCount)

	metricClustersCount := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: myName,
			Name:      "xds_clusters_total",
			Help:      "Total number of clusters.",
		}, s.GetClusterCount)

	s.metrics.xdsDeployments = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "xds_deployments_total",
			Help:      "Total number of xds configuration deployments.",
		}, []string{"resource"})

	prometheus.MustRegister(metricVirtualHostsCount)
	prometheus.MustRegister(metricRoutesCount)
	prometheus.MustRegister(metricClustersCount)
	prometheus.MustRegister(s.metrics.xdsDeployments)
}
