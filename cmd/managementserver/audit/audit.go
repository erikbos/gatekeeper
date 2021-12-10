package audit

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
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

	// Details of organization, developer and app this audit change applies to
	Environment struct {
		Organization string
		DeveloperID  string
		AppID        string
	}

	// Requester stores the identity and connection details of an authenticated managementserver API user
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

type auditType string

func (c auditType) String() string {
	return string(c)
}

const (
	eventCreate auditType = "create"
	eventUpdate auditType = "update"
	eventDelete auditType = "delete"
)

// Create logs a created entity to auditlog
func (al *Audit) Create(new interface{}, e *Environment, who Requester) {

	al.log(eventCreate, types.NameOf(new), types.IDOf(new), nil, new, e, who)
}

// Update logs an updated entity to auditlog
func (al *Audit) Update(old, new interface{}, e *Environment, who Requester) {

	al.log(eventUpdate, types.NameOf(old), types.IDOf(old), old, new, e, who)
}

// Delete logs a deleted entity to auditlog
func (al *Audit) Delete(old interface{}, e *Environment, who Requester) {

	al.log(eventDelete, types.NameOf(old), types.IDOf(old), old, nil, e, who)
}

// log logs the changed entity to auditlog and database
func (al *Audit) log(aType auditType, entityType, entityID string, oldValue, newValue interface{}, e *Environment, who Requester) {

	if e == nil {
		e = &Environment{}
	}
	auditEntry := &types.Audit{
		ID:           uuid.New().String(),
		Timestamp:    shared.GetCurrentTimeMilliseconds(),
		AuditType:    aType.String(),
		EntityType:   entityType,
		EntityID:     entityID,
		IPaddress:    who.RemoteAddr,
		RequestID:    who.RequestID,
		User:         who.User,
		Role:         who.Role,
		UserAgent:    who.Header.Get("User-Agent"),
		Organization: e.Organization,
		DeveloperID:  e.DeveloperID,
		AppID:        e.AppID,
		OldValue:     al.convertInterfaceMapString(clearSensitiveFields(oldValue)),
		NewValue:     al.convertInterfaceMapString(clearSensitiveFields(newValue)),
	}

	al.logger.Info("m", zap.Any("record", auditEntry))
	if err := al.db.Audit.Write(auditEntry); err != nil {
		al.logger.Error("m", zap.Any("error", err))
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

// convertInterfaceMapString converts interface{} to *map[string]interface{}
func (al *Audit) convertInterfaceMapString(m interface{}) map[string]interface{} {

	var data []byte
	var mapString map[string]interface{}
	var err error

	data, err = json.Marshal(m)
	if err != nil {
		al.logger.Fatal("Cannot marshal", zap.Any("InterfaceStringMap", clearSensitiveFields(m)))
	}

	if err = json.Unmarshal(data, &mapString); err != nil {
		al.logger.Fatal("Cannot unmarshal", zap.Any("InterfaceStringMap", clearSensitiveFields(m)))
	}

	return mapString
}
