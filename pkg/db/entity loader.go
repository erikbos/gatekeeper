package db

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Entityloader bla
type Entityloader struct {
	db                    *Database                     // Database handle
	configRefreshInterval time.Duration                 // Interval between database loads
	notify                chan EntityChangeNotification // Notification channel to emit change events
	listeners             []shared.Listener             // All listeners loaded from database
	routes                []shared.Route                // All routes loaded from database
	clusters              []shared.Cluster              // All clusters loaded from database
	listenersLastUpdate   int64                         // Timestamp of most recent load of listeners
	routesLastUpdate      int64                         // Timestamp of most recent load of routes
	clustersLastUpdate    int64                         // Timestamp of most recent load of clusters
}

// EntityChangeNotification is the msg send when we noticed a change in an entity
type EntityChangeNotification struct {
	Resource string // Name of resource type that has been changed
}

// Entity types that we load and send via notification channel
const (
	EntityTypeListener = "listener"
	EntityTypeRoute    = "route"
	EntityTypeCluster  = "cluster"
)

// NewEntityLoader returns a new entity loader
func NewEntityLoader(database *Database, interval time.Duration) *Entityloader {

	return &Entityloader{
		configRefreshInterval: interval,
		notify:                make(chan EntityChangeNotification),
		db:                    database,
	}
}

// Start continously monitors the database for configuration entities
func (g *Entityloader) Start() {

	go g.LoadContinously()
}

// LoadContinously continuously loads listeners, routes
// and clusters entities from database and updates the in-memory list
// in case a changed entity has been detect a notification will be send
//
// FIXME this code does not detect removed records
//
func (g *Entityloader) LoadContinously() {

	for {
		if newListeners, err := g.db.Listener.GetAll(); err != nil {
			log.Errorf("Could not retrieve listeners from database (%s)", err)
		} else {
			if g.listenerConfigChanged(newListeners) {
				log.Info("Listener configuration loaded")
				g.notify <- EntityChangeNotification{Resource: EntityTypeListener}
			}
		}

		if newRoutes, err := g.db.Route.GetAll(); err != nil {
			log.Errorf("Could not retrieve listeners from database (%s)", err)
		} else {
			if g.routeConfigChanged(newRoutes) {
				log.Info("Route configuration loaded")
				g.notify <- EntityChangeNotification{Resource: EntityTypeRoute}
			}
		}

		if newClusters, err := g.db.Cluster.GetAll(); err != nil {
			log.Errorf("Could not retrieve listeners from database (%s)", err)
		} else {
			if g.clusterConfigChanged(newClusters) {
				log.Info("Cluster configuration loaded")
				g.notify <- EntityChangeNotification{Resource: EntityTypeCluster}
			}
		}

		time.Sleep(g.configRefreshInterval)
	}
}

func (g *Entityloader) listenerConfigChanged(newConfig []shared.Listener) bool {

	for _, listener := range newConfig {
		if g.listenersLastUpdate == 0 ||
			listener.LastmodifiedAt > g.listenersLastUpdate {

			var m sync.Mutex
			m.Lock()
			g.listeners = newConfig
			m.Unlock()

			g.listenersLastUpdate = shared.GetCurrentTimeMilliseconds()
			return true
		}
	}
	return false
}

func (g *Entityloader) routeConfigChanged(newConfig []shared.Route) bool {

	for _, route := range newConfig {
		if g.routesLastUpdate == 0 || route.LastmodifiedAt > g.routesLastUpdate {

			var m sync.Mutex
			m.Lock()
			g.routes = newConfig
			m.Unlock()

			g.routesLastUpdate = shared.GetCurrentTimeMilliseconds()
			return true
		}
	}
	return false
}

func (g *Entityloader) clusterConfigChanged(newConfig []shared.Cluster) bool {

	for _, cluster := range newConfig {
		if g.clustersLastUpdate == 0 || cluster.LastmodifiedAt > g.clustersLastUpdate {

			var m sync.Mutex
			m.Lock()
			g.clusters = newConfig
			m.Unlock()

			g.clustersLastUpdate = shared.GetCurrentTimeMilliseconds()
			return true
		}
	}
	return false
}

// GetListeners returns all listeners
func (g *Entityloader) GetListeners() []shared.Listener {

	return g.listeners
}

// GetRoutes returns all listeners
func (g *Entityloader) GetRoutes() []shared.Route {

	return g.routes
}

// GetClusters returns number of clusters
func (g *Entityloader) GetClusters() []shared.Cluster {

	return g.clusters
}

// GetListenerCount returns number of listeners
func (g *Entityloader) GetListenerCount() int {

	return len(g.listeners)
}

// GetRouteCount returns number of routes
func (g *Entityloader) GetRouteCount() int {

	return len(g.routes)
}

// GetClusterCount returns number of clusters
func (g *Entityloader) GetClusterCount() int {

	return len(g.clusters)
}

// GetChannel returns notification channel
func (g *Entityloader) GetChannel() chan EntityChangeNotification {

	return g.notify
}
