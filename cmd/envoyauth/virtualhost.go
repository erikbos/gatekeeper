package main

import (
	"errors"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type vhostMapping struct {
	dbentities *db.Entityloader
	listeners  map[vhostMapEntry]shared.Listener
}

type vhostMapEntry struct {
	vhost string
	port  int
}

func newVhostMapping(d *db.Entityloader) *vhostMapping {

	return &vhostMapping{
		dbentities: d,
	}
}

func (v *vhostMapping) WaitFor(entityNotifications chan db.EntityChangeNotification) {

	for changedEntity := range entityNotifications {
		log.Infof("Database change notify received for entity group '%s'",
			changedEntity.Resource)

		if changedEntity.Resource == db.EntityTypeListener ||
			changedEntity.Resource == db.EntityTypeRoute {
			log.Printf("%+v", v.buildVhostMap())
		}
	}
}

func (v *vhostMapping) buildVhostMap() map[vhostMapEntry]shared.Listener {

	newListeners := make(map[vhostMapEntry]shared.Listener)

	for _, listener := range v.dbentities.GetListeners() {
		listener.Attributes = shared.Attributes{}

		for _, host := range listener.VirtualHosts {
			newListeners[vhostMapEntry{strings.ToLower(host),
				listener.Port}] = listener
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
func (v *vhostMapping) Lookup(hostname, protocol string) (*shared.Listener, error) {

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
