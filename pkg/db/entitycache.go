package db

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// EntityCache contains up to date entities like listeners, routes, clusters, users and roles
type EntityCache struct {
	db                  *Database         // Database handle
	config              EntityCacheConfig // Loader configuration
	listeners           types.Listeners   // All listeners loaded from database
	routes              types.Routes      // All routes loaded from database
	clusters            types.Clusters    // All clusters loaded from database
	listenersLastUpdate int64             // Timestamp of most recent load of listeners
	routesLastUpdate    int64             // Timestamp of most recent load of routes
	clustersLastUpdate  int64             // Timestamp of most recent load of clusters
	mutex               sync.Mutex        // Mutex to use when updating
	logger              *zap.Logger       // Logger
}

// EntityCacheConfig contains configuration on which entities we continously load
type EntityCacheConfig struct {
	RefreshInterval time.Duration                 // Interval between entity loads
	Notify          chan EntityChangeNotification // Notification channel to emit change events
}

// EntityChangeNotification is the msg send when we noticed a change in an entity
type EntityChangeNotification struct {
	Resource string // Name of resource type that has been changed
}

// NewEntityCache returns a new entity loader
func NewEntityCache(database *Database, config EntityCacheConfig, logger *zap.Logger) *EntityCache {

	return &EntityCache{config: config,
		db:     database,
		logger: logger.With(zap.String("system", "entitycache")),
	}
}

// Start kicks off continously refreshing entities from database at interval
func (ec *EntityCache) Start() {

	go ec.loadContinously()
}

// loadContinously continuously loads listeners, routes and clusters
// entities from database and updates the in-memory slices.
//
// In case a changed entity has been detect a notification will be sent
func (ec *EntityCache) loadContinously() {

	for {
		ec.checkForChangedListeners()
		ec.checkForChangedRoutes()
		ec.checkForChangedClusters()
		time.Sleep(ec.config.RefreshInterval)
	}
}

// checkForChangedListeners if the loaded list of listeners is shorter
// or one entry has been updated
func (ec *EntityCache) checkForChangedListeners() {

	loadedListeners, err := ec.db.Listener.GetAll()
	if err != nil {
		ec.logger.Error("Cannot retrieve listeners from database", zap.Error(err))
		return
	}
	// In case we have less routes one or more was deleted
	if len(loadedListeners) < len(ec.listeners) {
		ec.updateListeners(loadedListeners)
	}
	for _, listener := range loadedListeners {
		if ec.listenersLastUpdate == 0 || listener.LastmodifiedAt > ec.listenersLastUpdate {
			ec.updateListeners(loadedListeners)
		}
	}
}

func (ec *EntityCache) updateListeners(newListeners types.Listeners) {

	ec.mutex.Lock()
	ec.listeners = newListeners
	ec.mutex.Unlock()
	ec.listenersLastUpdate = shared.GetCurrentTimeMilliseconds()

	ec.logger.Info("Listener entities reloaded")
	if ec.config.Notify != nil {
		ec.config.Notify <- EntityChangeNotification{Resource: EntityTypeListener}
	}
}

// checkForChangedRoutes if the loaded list of routes is shorter
// or one entry has been updated
func (ec *EntityCache) checkForChangedRoutes() {

	loadedRoutes, err := ec.db.Route.GetAll()
	if err != nil {
		ec.logger.Error("Cannot retrieve routes from database", zap.Error(err))
		return
	}
	// In case we have less routes one or more was deleted
	if len(loadedRoutes) < len(ec.routes) {
		ec.updateRoutes(loadedRoutes)
	}
	for _, route := range loadedRoutes {
		if ec.routesLastUpdate == 0 || route.LastmodifiedAt > ec.routesLastUpdate {
			ec.updateRoutes(loadedRoutes)
		}
	}
}

func (ec *EntityCache) updateRoutes(newRoutes types.Routes) {

	ec.mutex.Lock()
	ec.routes = newRoutes
	ec.mutex.Unlock()
	ec.routesLastUpdate = shared.GetCurrentTimeMilliseconds()

	ec.logger.Info("Route entities reloaded")
	if ec.config.Notify != nil {
		ec.config.Notify <- EntityChangeNotification{Resource: EntityTypeRoute}
	}
}

// checkForChangedClusters checks if the loaded list of clusters is shorter
// or one entry has been updated
func (ec *EntityCache) checkForChangedClusters() {

	loadedClusters, err := ec.db.Cluster.GetAll()
	if err != nil {
		ec.logger.Error("Cannot retrieve clusters from database", zap.Error(err))
		return
	}
	// In case we have less routes one or more was deleted
	if len(loadedClusters) < len(ec.clusters) {
		ec.updateClusters(loadedClusters)
	}
	for _, cluster := range loadedClusters {
		if ec.clustersLastUpdate == 0 || cluster.LastmodifiedAt > ec.clustersLastUpdate {
			ec.updateClusters(loadedClusters)
		}
	}
}

func (ec *EntityCache) updateClusters(newClusters types.Clusters) {

	ec.mutex.Lock()
	ec.clusters = newClusters
	ec.mutex.Unlock()
	ec.clustersLastUpdate = shared.GetCurrentTimeMilliseconds()

	ec.logger.Info("Cluster entities reloaded")
	if ec.config.Notify != nil {
		ec.config.Notify <- EntityChangeNotification{Resource: EntityTypeCluster}
	}
}

// GetListeners returns all listeners
func (ec *EntityCache) GetListeners() types.Listeners {

	return ec.listeners
}

// GetRoutes returns all routes
func (ec *EntityCache) GetRoutes() types.Routes {

	return ec.routes
}

// GetClusters returns all clusters
func (ec *EntityCache) GetClusters() types.Clusters {

	return ec.clusters
}

// GetListenerCount returns number of listeners
func (ec *EntityCache) GetListenerCount() int {

	return len(ec.listeners)
}

// GetRouteCount returns number of routes
func (ec *EntityCache) GetRouteCount() int {

	return len(ec.routes)
}

// GetClusterCount returns number of clusters
func (ec *EntityCache) GetClusterCount() int {

	return len(ec.clusters)
}
