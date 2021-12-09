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
	dbentities *db.EntityCache
	listeners  map[vhostMapEntry]types.Listener
	logger     *zap.Logger
}

type vhostMapEntry struct {
	vhost string
	port  int
}

func newVhostMapping(d *db.EntityCache, logger *zap.Logger) *vhostMapping {

	return &vhostMapping{
		dbentities: d,
		logger:     logger,
	}
}

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

func (v *vhostMapping) buildVhostMap() map[vhostMapEntry]types.Listener {

	newListeners := make(map[vhostMapEntry]types.Listener)

	for _, listener := range v.dbentities.GetListeners() {
		listener.Attributes = types.NullAttributes

		for _, host := range listener.VirtualHosts {
			newListeners[vhostMapEntry{strings.ToLower(host), listener.Port}] = listener

			v.logger.Info("vhostmap",
				zap.String("host", host),
				zap.Int("port", listener.Port))
		}
	}

	var m sync.Mutex
	m.Lock()
	v.listeners = newListeners
	m.Unlock()

	return newListeners
}

// FIXME this should be lookup in map instead of for loops
// FIXXE map should have a key vhost:port
func (v *vhostMapping) Lookup(hostname, protocol string) (*types.Listener, error) {

	for _, listener := range v.listeners {
		for _, vhost := range listener.VirtualHosts {
			if vhost == hostname {
				if (listener.Port == 80 && protocol == "http") ||
					(listener.Port == 443 && protocol == "https") {
					return &listener, nil
				}
			}
		}
	}

	return nil, errors.New("vhost not found")
}
