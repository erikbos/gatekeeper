package audit

import (
	"encoding/json"
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

type (
	// Config holds configuration of an auditlog
	Config struct {
		// Database configuration
		Database cassandra.DatabaseConfig `yaml:"database"`
		// audit log configuration
		Logger shared.Logger `yaml:"logging"`
	}

	// Audit is a new auditlogger to be used by service layer to log changes to entities
	Audit struct {
		db     *db.Database
		logger *zap.Logger
	}

	// Requester stores the identity and connection details of an authenticated dbadmin API user
	// who is requesting a change.
	Requester struct {
		RemoteAddr string
		Header     http.Header
		User       string
		Role       string
		RequestID  string
	}
)

// New returns a new Auditlog instance which audit events to logfile and/or audit database.
func New(database *db.Database, logger *zap.Logger) *Audit {

	return &Audit{
		db:     database,
		logger: logger,
	}
}

type eventType string

func (c eventType) String() string {
	return string(c)
}

const (
	eventCreate eventType = "create"
	eventUpdate eventType = "update"
	eventDelete eventType = "delete"
)

// Create logs a created entity to auditlog
func (al *Audit) Create(new interface{}, who Requester) {

	al.log(eventCreate, types.NameOf(new), types.IDOf(new), nil, new, who)
}

// Update logs an updated entity to auditlog
func (al *Audit) Update(old, new interface{}, who Requester) {

	al.log(eventUpdate, types.NameOf(old), types.IDOf(old), old, new, who)
}

// Delete logs a deleted entity to auditlog
func (al *Audit) Delete(old interface{}, who Requester) {

	al.log(eventDelete, types.NameOf(old), types.IDOf(old), old, nil, who)
}

// log logs the changed entity to auditlog
func (al *Audit) log(eType eventType, entityType, entityID string, old, new interface{}, who Requester) {

	auditEntry := genAudit(eType, entityType, entityID, old, new, who)
	err := al.db.Audit.Write(auditEntry)
	if err != nil {
		log.Printf("ERROR %s", err)
	}

	al.logger.Info("audit",
		zap.String("eventType", eType.String()),
		zap.String("entityType", entityType),
		zap.String("entityId", entityID),
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

	var oldValue, newValue []byte
	var err error

	oldValue, err = json.Marshal(old)
	if err != nil {
		oldValue = []byte{}
	}
	newValue, err = json.Marshal(new)
	if err != nil {
		newValue = []byte{}
	}

	return &types.Audit{
		Timestamp:    shared.GetCurrentTimeMilliseconds(),
		AuditType:    eType.String(),
		EntityType:   entityType,
		EntityID:     entityId,
		IPaddress:    who.RemoteAddr,
		RequestID:    who.RequestID,
		User:         who.User,
		Role:         who.Role,
		UserAgent:    who.Header.Get("User-Agent"),
		Organization: "",
		DeveloperID:  "",
		AppID:        "",
		Old:          string(oldValue),
		New:          string(newValue),
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
