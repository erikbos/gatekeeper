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
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
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
	Listen                string        `yaml:"listen"`                // Extauthz listen port grpc
	ConfigCompileInterval time.Duration `yaml:"configcompileinterval"` // Interval between configuration compilations
	Cluster               string        `yaml:"cluster"`               // Name of cluster providing XDS service
	Timeout               time.Duration `yaml:"timeout"`               // Max duration of request to XDS cluster
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

	// snapshotCache will contain a configuration for each connected Envoy
	x.snapshotCache = cache.NewSnapshotCache(false, cache.IDHash{},
		newCacheLogger(x.server.logger))

	streamCallbacks := newCallback(&x.server)
	x.xds = xds.NewServer(context.Background(), x.snapshotCache, streamCallbacks)

	go x.CompileSnapshotsForNewNodes(streamCallbacks.signal)
	go x.GRPCManagementServer()
	for {
		select {
		case n := <-x.notify:
			x.server.logger.Info("Database change notify received",
				zap.String("entity", n.Resource))

			// Create snapshot given we got a notification a entity was updated
			x.CreateNewSnapshot(streamCallbacks)

			// Update our stats
			m := x.server.metrics
			m.SetEntityCount(types.TypeListenerName, len(x.server.dbentities.GetListeners()))
			m.SetEntityCount(types.TypeRouteName, len(x.server.dbentities.GetRoutes()))
			m.SetEntityCount(types.TypeClusterName, len(x.server.dbentities.GetClusters()))
			m.IncXDSSnapshotCreateCount(n.Resource)

		case <-time.After(x.xdsConfig.ConfigCompileInterval):
			// Nothing, just wait configcompileinterval seconds
		}
	}
}

// GRPCManagementServer starts grpc xds listener
func (x *XDS) GRPCManagementServer() {

	x.server.logger.Info("GRPC XDS listening on " + x.xdsConfig.Listen)
	lis, err := net.Listen("tcp", x.xdsConfig.Listen)
	if err != nil {
		x.server.logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	discoveryservice.RegisterAggregatedDiscoveryServiceServer(grpcServer, x.xds)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, x.xds)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, x.xds)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, x.xds)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, x.xds)

	reflection.Register(grpcServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	if err := grpcServer.Serve(lis); err != nil {
		x.server.logger.Fatal("failed to start GRPC server", zap.Error(err))
	}
}

// CreateNewSnapshot compiles configuration into snapshot
func (x *XDS) CreateNewSnapshot(streamCallbacks *callback) {

	atomic.AddInt64(&x.snapshotCacheVersion, 1)
	ts := time.Now().UTC().Format(time.RFC3339)
	version := fmt.Sprintf(ts+"-V%d", x.snapshotCacheVersion)

	x.server.logger.Info("Creating configuration snapshot", zap.String("version", version))

	EnvoyClusters, _ := x.server.getEnvoyClusterConfig()
	EnvoyRoutes, _ := x.server.getEnvoyRouteConfig()
	EnvoyListeners, _ := x.server.getEnvoyListenerConfig()

	x.snapshotLatest = cache.NewSnapshot(version, nil, EnvoyClusters, EnvoyRoutes, EnvoyListeners, nil, nil)

	// Update snapshot cache for each connected Envoy we are aware of
	for _, node := range streamCallbacks.connections {
		if err := x.snapshotCache.SetSnapshot(node.Id, x.snapshotLatest); err != nil {
			x.server.logger.Info("Cannot set snapshot for node", zap.String("id", node.Id))
		}
	}
}

// CompileSnapshotsForNewNodes waits for messages of new Envoys coming online and
// initiates compilation of configuration snapshots
func (x *XDS) CompileSnapshotsForNewNodes(signal <-chan newNode) {

	for {
		newNode := <-signal

		x.server.logger.Info("NewNodeDiscovery", zap.String("id", newNode.nodeID))

		// In case our snapshot version is still zero it means we have not yet done
		// our first configuration compilation: we skip. In this case XDSCreateNewSnapshot()
		// provide this Envoy a configuration as its connection was registered by OnStreamRequest().
		if x.snapshotCacheVersion != 0 {
			// Update cache for this newly connect Envoy we have not seen before
			if err := x.snapshotCache.SetSnapshot(newNode.nodeID, x.snapshotLatest); err != nil {
				x.server.logger.Warn("Cannot set snapshot for node", zap.String("id", newNode.nodeID))
			}
		}
	}
}
