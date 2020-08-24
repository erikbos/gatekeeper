package main

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
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

	"github.com/erikbos/gatekeeper/pkg/db"
)

// XDS holds configuration of XDS server
type XDS struct {
	server               server                             // Main server configuration
	xdsConfig            xdsConfig                          // Configuration of our XDS server
	xds                  xds.Server                         // Handlers of various services supported by XDS
	notify               <-chan db.EntityChangeNotification // Channel to receive notifications that configurations has changed
	snapshotCacheVersion int64                              // Unique version id of our cache
	snapshotCache        cache.SnapshotCache                // Cache of all snapshots for all Envoy nodes we are serving
	snapshotLatest       cache.Snapshot                     // Latest compiled configuration snapshot
}

type xdsConfig struct {
	GRPCListen            string        `yaml:"grpclisten"`            // GRPC listen port
	ConfigCompileInterval time.Duration `yaml:"configcompileinterval"` // Interval between configuration compilations
	Cluster               string        `yaml:"cluster"`               // Name of cluster providing XDS service
	Timeout               time.Duration `yaml:"timeout"`               // Max duration of request to XDS cluster
}

// xdsNotifyMesssage tells us configuration has changed
type xdsNotifyMesssage struct {
	resource string // Resource type that has been changed ("virtualhost", "route", "cluster")
}

func newXDS(s server, config xdsConfig, signal <-chan db.EntityChangeNotification) *XDS {

	return &XDS{
		server:    s,
		xdsConfig: config,
		notify:    signal,
	}
}

// Start brings up XDS system
func (x *XDS) Start() {

	x.snapshotCache = cache.NewSnapshotCache(false, cache.IDHash{}, logger{})
	streamCallbacks := newCallback(&x.server)
	x.xds = xds.NewServer(context.Background(), x.snapshotCache, streamCallbacks)

	go x.CompileSnapshotsForNewNodes(streamCallbacks.signal)
	go x.GRPCManagementServer()
	for {
		select {
		case n := <-x.notify:
			log.Infof("XDS change notify received for resource group '%s'", n.Resource)

			x.server.metrics.xdsDeployments.WithLabelValues(n.Resource).Inc()

			x.CreateNewSnapshot(streamCallbacks)

		case <-time.After(x.xdsConfig.ConfigCompileInterval):
			// Nothing, just wait configcompileinterval seconds
		}
	}
}

// GRPCManagementServer starts grpc xds listener
func (x *XDS) GRPCManagementServer() {

	log.Info("GRPC XDS listening on ", x.xdsConfig.GRPCListen)
	lis, err := net.Listen("tcp", x.xdsConfig.GRPCListen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	discoveryservice.RegisterAggregatedDiscoveryServiceServer(grpcServer, x.xds)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, x.xds)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, x.xds)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, x.xds)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, x.xds)

	log.Fatalf("failed to start GRPC server: %v", grpcServer.Serve(lis))
}

// CreateNewSnapshot compiles configuration into snapshot
func (x *XDS) CreateNewSnapshot(streamCallbacks *callback) {

	atomic.AddInt64(&x.snapshotCacheVersion, 1)
	ts := time.Now().UTC().Format(time.RFC3339)
	version := fmt.Sprintf(ts+"-V%d", x.snapshotCacheVersion)

	log.Infof("Creating configuration snapshot version '%s'", version)

	EnvoyClusters, _ := x.server.getEnvoyClusterConfig()
	EnvoyRoutes, _ := x.server.getEnvoyRouteConfig()
	EnvoyListeners, _ := x.server.getEnvoyListenerConfig()

	x.snapshotLatest = cache.NewSnapshot(version, nil, EnvoyClusters, EnvoyRoutes, EnvoyListeners, nil, nil)

	// Update snapshot cache for each connected Envoy we are aware of
	for _, node := range streamCallbacks.connections {
		x.snapshotCache.SetSnapshot(node.Id, x.snapshotLatest)
	}
}

// CompileSnapshotsForNewNodes waits for messages of new Envoys coming online and
// initiates compilation of configuration snapshots
func (x *XDS) CompileSnapshotsForNewNodes(signal <-chan newNode) {

	for {
		newNode := <-signal

		fields := log.Fields{
			"id": newNode.nodeID,
		}
		log.WithFields(fields).Info("NewNodeDiscovery")

		// In case our snapshot version is still zero it means we have not yet done
		// our first configuration compilation: we skip. In this case XDSCreateNewSnapshot()
		// provide this Envoy a configuration as its connection was registered by OnStreamRequest().
		if x.snapshotCacheVersion != 0 {
			// Update cache for this newly connect Envoy we have not seen before
			x.snapshotCache.SetSnapshot(newNode.nodeID, x.snapshotLatest)
		}
	}
}
