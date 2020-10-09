package service

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
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

	// ChangelogConfig holds configuration of a changelog
	ChangelogConfig struct {
		Logger shared.Logger `yaml:"logging"` // changelog log configuration
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
		User       string
		Role       string
		RequestID  string
	}
)

// NewChangelog returns a new Changelog instance
func NewChangelog(database *db.Database, logger *zap.Logger) *Changelog {

	return &Changelog{
		db:     database,
		logger: logger,
	}
}

const (
	createEvent = "create"
	updateEvent = "update"
	deleteEvent = "delete"
)

// Create logs a created entity
func (cl *Changelog) Create(new interface{}, who Requester) {

	cl.log(createEvent, db.Typeof(new), nil, new, who)
}

// Update logs an updated entity
func (cl *Changelog) Update(old, new interface{}, who Requester) {

	cl.log(updateEvent, db.Typeof(old), old, new, who)
}

// Delete logs a deleted entity
func (cl *Changelog) Delete(old interface{}, who Requester) {

	cl.log(deleteEvent, db.Typeof(old), old, nil, who)
}

// log logs a changed entity
func (cl *Changelog) log(eventType, entityType string, old, new interface{}, who Requester) {

	cl.logger.Info("changelog",
		zap.String("changetype", eventType),
		zap.String("entity", entityType),
		zap.Any("who", map[string]interface{}{
			"ip":        who.RemoteAddr,
			"user":      who.User,
			"role":      who.Role,
			"requestid": who.RequestID,
			// "headers", who.Header,
		}),
		zap.Any("old", map[string]interface{}{
			entityType: old,
		}),
		zap.Any("new", map[string]interface{}{
			entityType: new,
		}))
}
