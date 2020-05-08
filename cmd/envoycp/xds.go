package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/erikbos/apiauth/pkg/shared"
	"github.com/prometheus/client_golang/prometheus"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// StartXDS brings up XDS system
func (s *server) StartXDS() {
	s.registerMetrics()

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
	go s.GetRouteConfigFromDatabase()
	go s.GetVirtualHostConfigFromDatabase()

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

var xdsLastUpdate int64

func (s *server) XDSMainloop() {
	var lastConfigurationDeployment int64

	for {
		if xdsLastUpdate > lastConfigurationDeployment {
			log.Info("lastConfigurationDeployment: ", shared.TimeMillisecondsToString(lastConfigurationDeployment))
			log.Info("XdsLastUpdate: ", shared.TimeMillisecondsToString(xdsLastUpdate))
			log.Info("Starting configuration compilation")

			EnvoyClusters, _ := getEnvoyClusterConfig()
			EnvoyRoutes, _ := getEnvoyRouteConfig()
			EnvoyListeners, _ := getEnvoyListenerConfig()

			now := shared.GetCurrentTimeMilliseconds()
			version := fmt.Sprint(now)

			log.Infof("Creating config version %s", version)
			snap := cache.NewSnapshot(version, nil, EnvoyClusters, EnvoyRoutes, EnvoyListeners, nil)
			_ = s.xdsCache.SetSnapshot("jenny", snap)

			lastConfigurationDeployment = now
		}
		time.Sleep(1 * time.Second)
	}
}

// registerMetrics registers xds' operational metrics
func (s *server) registerMetrics() {
	metricVirtualHostsCount := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: myName + "_xds_virtualhosts",
			Help: "Current number of clusters",
		},
		s.GetVirtualHostCount)

	metricRoutesCount := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: myName + "_xds_routes",
			Help: "Current number of routes",
		},
		s.GetRouteCount)

	metricClustersCount := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: myName + "_xds_clusters",
			Help: "Current number of clusters",
		},
		s.GetClusterCount)

	s.metricXdsDeployments = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: myName + "_xds_deployments",
			Help: "Current number of xds configuration deployments",
		},
		[]string{"resource"})

	prometheus.MustRegister(metricVirtualHostsCount)
	prometheus.MustRegister(metricRoutesCount)
	prometheus.MustRegister(metricClustersCount)
	prometheus.MustRegister(s.metricXdsDeployments)
}
