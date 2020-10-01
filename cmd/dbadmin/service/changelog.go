package service

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/db"
)

type (
	// ChangelogService logs all create, update and delete events
	ChangelogService interface {
		// Create logs a created entity
		Create(new interface{}, who Requester)

		// Update logs an updated entity
		Update(old, new interface{}, who Requester)

		// Delete logs a deleted entity
		Delete(old interface{}, who Requester)
	}

	// Changelog is a new changelogger
	Changelog struct {
		db *db.Database
	}

	// Requester stores the identity and connection details of an authenticated user
	// who is reqesting a change, it is passed to service layer and changelog
	Requester struct {
		RemoteAddr string
		Header     http.Header
		Username   string
		RequestID  string
	}
)

// NewChangelogService returns a new Changelog instance
func NewChangelogService(database *db.Database) *Changelog {

	return &Changelog{db: database}
}

const (
	createEvent = "CREATE"
	updateEvent = "UPDATE"
	deleteEvent = "DELETE"
)

// Create logs a created entity
func (cl *Changelog) Create(new interface{}, who Requester) {

	cl.log(createEvent, nil, new, who)
}

// Update logs an updated entity
func (cl *Changelog) Update(old, new interface{}, who Requester) {

	cl.log(updateEvent, old, new, who)
}

// Delete logs a deleted entity
func (cl *Changelog) Delete(old interface{}, who Requester) {

	cl.log(deleteEvent, old, nil, who)
}

// log logs a changed entity
func (cl *Changelog) log(eventType string, old, new interface{}, who Requester) {

	log.Infof("who: %s", who)
	log.Infof("eventType: %s", eventType)
	log.Infof("OLD %+v\n", old)
	log.Infof("NEW %+v\n", new)
}
