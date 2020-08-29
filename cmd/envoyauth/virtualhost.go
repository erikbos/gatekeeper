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
	vhosts     map[vhostMapEntry]shared.VirtualHost
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

		if changedEntity.Resource == db.EntityTypeVirtualhost ||
			changedEntity.Resource == db.EntityTypeRoute {
			log.Printf("%+v", v.buildVhostMap())
		}
	}
}

func (v *vhostMapping) buildVhostMap() map[vhostMapEntry]shared.VirtualHost {

	newVhosts := make(map[vhostMapEntry]shared.VirtualHost)

	for _, virtualhost := range v.dbentities.GetVirtualhosts() {
		virtualhost.Attributes = shared.Attributes{}

		for _, host := range virtualhost.VirtualHosts {
			newVhosts[vhostMapEntry{strings.ToLower(host),
				virtualhost.Port}] = virtualhost
		}
	}

	var m sync.Mutex
	m.Lock()
	v.vhosts = newVhosts
	m.Unlock()

	return newVhosts
}

// FIXME this should be lookup in map instead of for loops
// FIXXE map should have a key vhost:port
func (v *vhostMapping) Lookup(hostname, protocol string) (*shared.VirtualHost, error) {

	for _, vhostEntry := range v.vhosts {
		for _, vhost := range vhostEntry.VirtualHosts {
			if vhost == hostname {
				if (vhostEntry.Port == 80 && protocol == "http") ||
					(vhostEntry.Port == 443 && protocol == "https") {
					return &vhostEntry, nil
				}
			}
		}
	}

	return nil, errors.New("vhost not found")
}
