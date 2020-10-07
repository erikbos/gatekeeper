package service

import (
	"net/http"

	"go.uber.org/zap"

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
		db     *db.Database
		logger *zap.Logger
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
func NewChangelogService(database *db.Database, logger *zap.Logger) *Changelog {

	return &Changelog{
		db:     database,
		logger: logger,
	}
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

	cl.logger.Info("changelog",
		zap.String("username", who.Username),
		zap.String("ip", who.RemoteAddr),
		zap.String("requestid", who.RequestID),
		zap.String("requestid", who.RequestID),
		zap.String("changetype", eventType),
		zap.Reflect("headers", who.Header),
		zap.Reflect("old", old),
		zap.Reflect("new", new))
}
