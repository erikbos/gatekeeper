package db

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// Entityloader contains up to date configuration of listeners, routes and clusters
type Entityloader struct {
	db                    *Database                     // Database handle
	configRefreshInterval time.Duration                 // Interval between database loads
	notify                chan EntityChangeNotification // Notification channel to emit change events
	listeners             types.Listeners               // All listeners loaded from database
	routes                types.Routes                  // All routes loaded from database
	clusters              types.Clusters                // All clusters loaded from database
	listenersLastUpdate   int64                         // Timestamp of most recent load of listeners
	routesLastUpdate      int64                         // Timestamp of most recent load of routes
	clustersLastUpdate    int64                         // Timestamp of most recent load of clusters
	mutex                 sync.Mutex                    // Mutex to use when updating
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

// LoadContinously continuously loads listeners, routes and clusters
// entities from database and updates the in-memory slices.
//
// In case a changed entity has been detect a notification will be sent
//
func (g *Entityloader) LoadContinously() {

	for {
		if newListeners, err := g.db.Listener.GetAll(); err != nil {
			log.Errorf("Could not retrieve listeners from database (%s)", err)
		} else {
			if g.checkForChangeListeners(newListeners) {
				log.Info("Listener configuration reloaded")
				g.notify <- EntityChangeNotification{Resource: EntityTypeListener}
			}
		}

		if newRoutes, err := g.db.Route.GetAll(); err != nil {
			log.Errorf("Could not retrieve routes from database (%s)", err)
		} else {
			if g.checkForChangedRoutes(newRoutes) {
				log.Info("Route configuration reloaded")
				g.notify <- EntityChangeNotification{Resource: EntityTypeRoute}
			}
		}

		if newClusters, err := g.db.Cluster.GetAll(); err != nil {
			log.Errorf("Could not retrieve listeners from database (%s)", err)
		} else {
			if g.checkForChangedClusters(newClusters) {
				log.Info("Cluster configuration reloaded")
				g.notify <- EntityChangeNotification{Resource: EntityTypeCluster}
			}
		}

		time.Sleep(g.configRefreshInterval)
	}
}

// checkForChangedListeners if the loaded list of listeners is shorter
// or one entry has been updated
func (g *Entityloader) checkForChangeListeners(loadedListeners types.Listeners) bool {

	// In case we have less routes one or more was deleted
	if len(loadedListeners) < len(g.listeners) {
		g.updateListeners(loadedListeners)
		return true
	}

	for _, listener := range loadedListeners {
		if g.listenersLastUpdate == 0 || listener.LastmodifiedAt > g.listenersLastUpdate {
			g.updateListeners(loadedListeners)
			return true
		}
	}
	return false
}

func (g *Entityloader) updateListeners(newListeners types.Listeners) {

	g.mutex.Lock()
	g.listeners = newListeners
	g.mutex.Unlock()
	g.listenersLastUpdate = shared.GetCurrentTimeMilliseconds()
}

// checkForChangedRoutes if the loaded list of routes is shorter
// or one entry has been updated
func (g *Entityloader) checkForChangedRoutes(loadedRoutes types.Routes) bool {

	// In case we have less routes one or more was deleted
	if len(loadedRoutes) < len(g.routes) {
		g.updateRoutes(loadedRoutes)
		return true
	}

	for _, route := range loadedRoutes {
		if g.routesLastUpdate == 0 || route.LastmodifiedAt > g.routesLastUpdate {
			g.updateRoutes(loadedRoutes)
			return true
		}
	}
	return false
}

func (g *Entityloader) updateRoutes(newRoutes types.Routes) {

	g.mutex.Lock()
	g.routes = newRoutes
	g.mutex.Unlock()
	g.routesLastUpdate = shared.GetCurrentTimeMilliseconds()
}

// checkForChangeClusters checks if the loaded list of clusters is shorter
// or one entry has been updated
func (g *Entityloader) checkForChangedClusters(loadedClusters types.Clusters) bool {

	// In case we have less routes one or more was deleted
	if len(loadedClusters) < len(g.clusters) {
		g.updateClusters(loadedClusters)
		return true
	}

	for _, cluster := range loadedClusters {
		if g.clustersLastUpdate == 0 || cluster.LastmodifiedAt > g.clustersLastUpdate {
			g.updateClusters(loadedClusters)
			return true
		}
	}
	return false
}

func (g *Entityloader) updateClusters(newClusters types.Clusters) {

	g.mutex.Lock()
	g.clusters = newClusters
	g.mutex.Unlock()
	g.clustersLastUpdate = shared.GetCurrentTimeMilliseconds()
}

// GetListeners returns all listeners
func (g *Entityloader) GetListeners() types.Listeners {

	return g.listeners
}

// GetRoutes returns all listeners
func (g *Entityloader) GetRoutes() types.Routes {

	return g.routes
}

// GetClusters returns number of clusters
func (g *Entityloader) GetClusters() types.Clusters {

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
