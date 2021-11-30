package service

import (
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

type (
	// AuditService logs all create, update and delete events
	AuditService interface {
		// Create logs a created entity
		Create(new interface{}, who Requester)

		// Update logs an updated entity
		Update(old, new interface{}, who Requester)

		// Delete logs a deleted entity
		Delete(old interface{}, who Requester)
	}

	// AuditConfig holds configuration of an auditlog
	AuditConfig struct {
		Logger shared.Logger `yaml:"logging"` // audit log configuration
	}

	// Auditlog is a new auditlogger
	Auditlog struct {
		db     *db.Database
		logger *zap.Logger
	}

	// Requester stores the identity and connection details of an authenticated user
	// who is reqesting a change, it is passed to service layer and auditlog
	Requester struct {
		RemoteAddr string
		Header     http.Header
		User       string
		Role       string
		RequestID  string
	}
)

// NewAuditlog returns a new Auditlog instance
func NewAuditlog(database *db.Database, logger *zap.Logger) *Auditlog {

	return &Auditlog{
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

// Create logs a created entity to auditlog
func (al *Auditlog) Create(new interface{}, who Requester) {

	al.log(createEvent, types.NameOf(new), types.IDOf(new), nil, new, who)
}

// Update logs an updated entity to auditlog
func (al *Auditlog) Update(old, new interface{}, who Requester) {

	al.log(updateEvent, types.NameOf(old), types.IDOf(old), old, new, who)
}

// Delete logs a deleted entity to auditlog
func (al *Auditlog) Delete(old interface{}, who Requester) {

	al.log(deleteEvent, types.NameOf(old), types.IDOf(old), old, nil, who)
}

// log logs the changed entity to auditlog
func (al *Auditlog) log(eType eventType, entityType, entityId string, old, new interface{}, who Requester) {

	auditEntry := genAudit(eType, entityType, entityId, old, new, who)
	err := al.db.Audit.Write(auditEntry)
	if err != nil {
		log.Printf("ERROR %s", err)
	}

	al.logger.Info("changelog",
		zap.String("changeType", eType.String()),
		zap.String("entityType", entityType),
		zap.String("entityId", entityId),
		zap.Any("who", map[string]interface{}{
			"user":      who.User,
			"role":      who.Role,
			"requestId": who.RequestID,
			"address":   who.RemoteAddr,
			"userAgent": who.Header.Get("User-Agent"),
		}),
		zap.Any("old", map[string]interface{}{
			entityType: clearSensitiveFields(old),
		}),
		zap.Any("new", map[string]interface{}{
			entityType: clearSensitiveFields(new),
		}),
	)
}

func genAudit(eType eventType, entityType, entityId string, old, new interface{}, who Requester) *types.Audit {

	return &types.Audit{
		Timestamp:    shared.GetCurrentTimeMilliseconds(),
		ChangeType:   eType.String(),
		EntityType:   entityType,
		EntityId:     entityId,
		IPaddress:    who.RemoteAddr,
		RequestID:    who.RequestID,
		User:         who.User,
		Role:         who.Role,
		UserAgent:    who.Header.Get("User-Agent"),
		Organization: "",
		DeveloperID:  "",
		AppID:        "",
		Old:          "",
		New:          "",
	}
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
