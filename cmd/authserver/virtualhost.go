package main

import (
	"errors"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

type vhostMapping struct {
	db                  *db.Database
	vhostListeners      map[vhostMapKey]vhostMapEntry
	defaultOrganization string
	logger              *zap.Logger
}

type vhostMapKey struct {
	vhost string
	port  int
}

type vhostMapEntry struct {
	// Listener associated with this vhost
	listener types.Listener
	// Organization to use for lookups when request needs to be evaluated
	organization types.Organization
}

func newVhostMapping(db *db.Database, defaultOrganization string, logger *zap.Logger) *vhostMapping {
	return &vhostMapping{
		db:                  db,
		defaultOrganization: defaultOrganization,
		logger:              logger,
	}
}

func newVhostMapKey(vhost string, port int) vhostMapKey {
	return vhostMapKey{
		vhost: strings.ToLower(vhost),
		port:  port,
	}
}

// WaitFor updates the vhost to listener map after receiving a trigger that listeners in db have changed
func (v *vhostMapping) WaitFor(entityNotifications chan db.EntityChangeNotification) {

	for changedEntity := range entityNotifications {
		v.logger.Info("Database change notify received",
			zap.String("entity", changedEntity.Resource))

		if changedEntity.Resource == types.TypeListenerName ||
			changedEntity.Resource == types.TypeRouteName {

			v.buildVhostMap()
		}
	}
}

func (v *vhostMapping) buildVhostMap() {

	listeners, err := v.db.Listener.GetAll()
	if err != nil {
		v.logger.Error("cannot retrieve listeners to build vhost map")
		return
	}

	newVHostListeners := make(map[vhostMapKey]vhostMapEntry)
	for _, listener := range listeners {
		for _, host := range listener.VirtualHosts {
			// Retrieve organization details, based upon Listernet attribute Organization, or use
			// default organizationName from startup configuration
			organization, err := v.db.Organization.Get(
				listener.Attributes.GetAsString(types.AttributeOrganization, v.defaultOrganization))
			if err != nil {
				v.logger.Error("cannot retrieve organization details",
					zap.String("listener", listener.Name))
				continue
			}

			key := newVhostMapKey(host, listener.Port)
			entry := vhostMapEntry{
				listener:     listener,
				organization: *organization,
			}
			newVHostListeners[key] = entry

			v.logger.Debug("vhostmap",
				zap.String("host", key.vhost),
				zap.Int("port", key.port),
				zap.String("organization", entry.organization.Name))
		}
	}

	var m sync.Mutex
	m.Lock()
	v.vhostListeners = newVHostListeners
	m.Unlock()
}

// Lookup return listener and organization details based upon vhost and port
func (v *vhostMapping) Lookup(vhost string, port int) (*types.Listener, *types.Organization, error) {

	vhostMapEntry, ok := v.vhostListeners[newVhostMapKey(vhost, port)]
	if ok {
		return &vhostMapEntry.listener, &vhostMapEntry.organization, nil
	}
	return nil, nil, errors.New("vhost not found")
}
