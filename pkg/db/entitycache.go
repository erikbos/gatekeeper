package db

import (
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

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
	users               types.Users       // All users loaded from database
	roles               types.Roles       // All roles loaded from database
	listenersLastUpdate int64             // Timestamp of most recent load of listeners
	routesLastUpdate    int64             // Timestamp of most recent load of routes
	clustersLastUpdate  int64             // Timestamp of most recent load of clusters
	usersLastUpdate     int64             // Timestamp of most recent load of users
	rolesLastUpdate     int64             // Timestamp of most recent load of roles
	mutex               sync.Mutex        // Mutex to use when updating
}

// EntityCacheConfig contains configuration on which entities we continously load
type EntityCacheConfig struct {
	RefreshInterval time.Duration                 // Interval between entity loads
	Notify          chan EntityChangeNotification // Notification channel to emit change events
	Listener        bool                          // If true, load listener continously
	Route           bool                          // If true, load route continously
	Cluster         bool                          // If true, load cluster continously
	User            bool                          // If true, load user continously
	Role            bool                          // If true, load role continously
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
	EntityTypeUser     = "user"
	EntityTypeRole     = "role"
)

// NewEntityCache returns a new entity loader
func NewEntityCache(database *Database, config EntityCacheConfig) *EntityCache {

	return &EntityCache{config: config, db: database}
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
		if ec.config.Listener {
			ec.checkForChangedListeners()
		}
		if ec.config.Route {
			ec.checkForChangedRoutes()
		}
		if ec.config.Cluster {
			ec.checkForChangedClusters()
		}
		if ec.config.User {
			ec.checkForChangedUsers()
		}
		if ec.config.Role {
			ec.checkForChangedRoles()
		}
		time.Sleep(ec.config.RefreshInterval)
	}
}

// checkForChangedListeners if the loaded list of listeners is shorter
// or one entry has been updated
func (ec *EntityCache) checkForChangedListeners() {

	loadedListeners, err := ec.db.Listener.GetAll()
	if err != nil {
		log.Errorf("Could not retrieve listeners from database (%s)", err)
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

	log.Info("Listener entities reloaded")
	if ec.config.Notify != nil {
		ec.config.Notify <- EntityChangeNotification{Resource: EntityTypeListener}
	}
}

// checkForChangedRoutes if the loaded list of routes is shorter
// or one entry has been updated
func (ec *EntityCache) checkForChangedRoutes() {

	loadedRoutes, err := ec.db.Route.GetAll()
	if err != nil {
		log.Errorf("Could not retrieve routes from database (%s)", err)
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

	log.Info("Route entities reloaded")
	if ec.config.Notify != nil {
		ec.config.Notify <- EntityChangeNotification{Resource: EntityTypeRoute}
	}
}

// checkForChangedClusters checks if the loaded list of clusters is shorter
// or one entry has been updated
func (ec *EntityCache) checkForChangedClusters() {

	loadedClusters, err := ec.db.Cluster.GetAll()
	if err != nil {
		log.Errorf("Could not retrieve clusters from database (%s)", err)
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

	log.Info("Cluster entities reloaded")
	if ec.config.Notify != nil {
		ec.config.Notify <- EntityChangeNotification{Resource: EntityTypeCluster}
	}
}

// checkForChangedUsers checks if the loaded list of users is shorter
// or one entry has been updated
func (ec *EntityCache) checkForChangedUsers() {

	loadedUsers, err := ec.db.User.GetAll()
	if err != nil {
		log.Errorf("Could not retrieve users from database (%s)", err)
		return
	}
	// In case we have less users one or more was deleted
	if len(loadedUsers) < len(ec.users) {
		ec.updateUsers(loadedUsers)
	}
	for _, user := range loadedUsers {
		if ec.usersLastUpdate == 0 || user.LastmodifiedAt > ec.usersLastUpdate {
			ec.updateUsers(loadedUsers)
		}
	}
}

func (ec *EntityCache) updateUsers(newUsers types.Users) {

	ec.mutex.Lock()
	ec.users = newUsers
	ec.mutex.Unlock()
	ec.usersLastUpdate = shared.GetCurrentTimeMilliseconds()

	log.Info("User entities reloaded")
	if ec.config.Notify != nil {
		ec.config.Notify <- EntityChangeNotification{Resource: EntityTypeUser}
	}
}

// checkForChangedRoles checks if the loaded list of users is shorter
// or one entry has been updated
func (ec *EntityCache) checkForChangedRoles() {

	loadedRoles, err := ec.db.Role.GetAll()
	if err != nil {
		log.Errorf("Could not retrieve roles from database (%s)", err)
		return
	}
	// In case we have less roles one or more was deleted
	if len(loadedRoles) < len(ec.roles) {
		ec.updateRoles(loadedRoles)
	}
	for _, role := range loadedRoles {
		if ec.rolesLastUpdate == 0 || role.LastmodifiedAt > ec.rolesLastUpdate {
			ec.updateRoles(loadedRoles)
		}
	}
}

func (ec *EntityCache) updateRoles(newRoles types.Roles) {

	ec.mutex.Lock()
	ec.roles = newRoles
	ec.mutex.Unlock()
	ec.rolesLastUpdate = shared.GetCurrentTimeMilliseconds()

	log.Info("Role entities reloaded")
	if ec.config.Notify != nil {
		ec.config.Notify <- EntityChangeNotification{Resource: EntityTypeRole}
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

// GetUsers returns all users
func (ec *EntityCache) GetUsers() types.Users {

	return ec.users
}

// GetRoles returns all roles
func (ec *EntityCache) GetRoles() types.Roles {

	return ec.roles
}

// GetUser lookups one user
func (ec *EntityCache) GetUser(username string) (*types.User, error) {

	for _, user := range ec.users {
		if user.Name == username {
			return &user, nil
		}
	}
	return nil, errors.New("User not found")
}

// GetRole lookups one role
func (ec *EntityCache) GetRole(rolename string) (*types.Role, error) {

	for _, role := range ec.roles {
		if role.Name == rolename {
			return &role, nil
		}
	}
	return nil, errors.New("Role not found")
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

// GetUserCount returns number of users
func (ec *EntityCache) GetUserCount() int {

	return len(ec.users)
}

// GetRoleCount returns number of roles
func (ec *EntityCache) GetRoleCount() int {

	return len(ec.roles)
}