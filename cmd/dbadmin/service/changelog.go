package service

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
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

type eventType string

func (c eventType) String() string {
	return string(c)
}

const (
	createEvent eventType = "create"
	updateEvent eventType = "update"
	deleteEvent eventType = "delete"
)

// Create logs a created entity to changelog
func (cl *Changelog) Create(new interface{}, who Requester) {

	cl.log(createEvent, types.NameOf(new), nil, new, who)
}

// Update logs an updated entity to changelog
func (cl *Changelog) Update(old, new interface{}, who Requester) {

	cl.log(updateEvent, types.NameOf(old), old, new, who)
}

// Delete logs a deleted entity to changelog
func (cl *Changelog) Delete(old interface{}, who Requester) {

	cl.log(deleteEvent, types.NameOf(old), old, nil, who)
}

// log logs the changed entity to changelog
func (cl *Changelog) log(eType eventType, entityType string, old, new interface{}, who Requester) {

	cl.logger.Info("changelog",
		zap.String("changetype", eType.String()),
		zap.String("entity", entityType),
		zap.Any("who", map[string]interface{}{
			"ip":        who.RemoteAddr,
			"user":      who.User,
			"role":      who.Role,
			"requestid": who.RequestID,
			// "headers", who.Header,
		}),
		zap.Any("old", map[string]interface{}{
			entityType: clearSensitiveFields(old),
		}),
		zap.Any("new", map[string]interface{}{
			entityType: clearSensitiveFields(new),
		}),
	)
}

// Clear sensitive fields such as User.Password that we do not want end up in the changelog.
// Returns new struct in case fields have been cleared to not modify original.
func clearSensitiveFields(m interface{}) interface{} {

	switch t := m.(type) {
	case types.User:
		// Copy & empty password of User
		scrubbedUser := t
		scrubbedUser.Password = ""

		return scrubbedUser
	default:
		// Do not modify other types, return as is
		return m
	}
}
