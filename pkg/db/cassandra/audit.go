package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// List of audit columns we use
	auditColumns = `audit_id,
audit_type,
timestamp,
entity_type,
entity_id,
ipaddress,
request_id,
user,
role,
user_agent,
organization,
developer_id,
app_id,
old_value,
new_value`

	// Prometheus label for metrics of db interactions
	auditMetricLabel = "audits"
)

// AuditStore holds our AuditStore config
type AuditStore struct {
	db *Database
}

// NewAuditStore creates audit instance
func NewAuditStore(database *Database) *AuditStore {
	return &AuditStore{
		db: database,
	}
}

// GetOrganization retrieves audit records of an organization
func (s *AuditStore) GetOrganization(organizationName string, params db.AuditFilterParams) (types.Audits, types.Error) {

	query := "SELECT " + auditColumns + " FROM audits WHERE organization = ? AND timestamp >= ? AND timestamp <= ? LIMIT ?"
	return s.runGetAuditQuery(query, organizationName, params.StartTime, params.EndTime, params.Count)
}

// GetAPIProduct retrieves audit records of an apiproduct
func (s *AuditStore) GetAPIProduct(organizationName, apiproductName string, params db.AuditFilterParams) (types.Audits, types.Error) {

	query := "SELECT " + auditColumns + " FROM audits WHERE organization = ? AND entity_type = ? AND entity_id = ? AND timestamp >= ? AND timestamp <= ? LIMIT ?"
	return s.runGetAuditQuery(query, organizationName, types.TypeAPIProductName, apiproductName, params.StartTime, params.EndTime, params.Count)
}

// GetDeveloper retrieves audit records of a developer
func (s *AuditStore) GetDeveloper(organizationName, developerID string, params db.AuditFilterParams) (types.Audits, types.Error) {

	query := "SELECT " + auditColumns + " FROM audits WHERE organization = ? AND developer_id = ? AND timestamp >= ? AND timestamp <= ? LIMIT ?"
	return s.runGetAuditQuery(query, organizationName, developerID, params.StartTime, params.EndTime, params.Count)
}

// GetApplication retrieves audit records of an application
func (s *AuditStore) GetApplication(organizationName, developerID, appID string, params db.AuditFilterParams) (types.Audits, types.Error) {

	query := "SELECT " + auditColumns + " FROM audits WHERE organization = ? AND developer_id = ? AND app_id = ? AND timestamp >= ? AND timestamp <= ? LIMIT ?"
	return s.runGetAuditQuery(query, organizationName, developerID, appID, params.StartTime, params.EndTime, params.Count)
}

// GetUser retrieves audit records of a user
func (s *AuditStore) GetUser(userName string, params db.AuditFilterParams) (types.Audits, types.Error) {

	query := "SELECT " + auditColumns + " FROM audits WHERE user = ? AND timestamp >= ? AND timestamp <= ? LIMIT ?"
	return s.runGetAuditQuery(query, userName, params.StartTime, params.EndTime, params.Count)
}

// runGetAuditQuery executes CQL query and returns resultset
func (s *AuditStore) runGetAuditQuery(query string,
	queryParameters ...interface{}) (types.Audits, types.Error) {

	timer := prometheus.NewTimer(s.db.metrics.queryHistogram)
	defer timer.ObserveDuration()

	var audits types.Audits

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		audits = append(audits, types.Audit{
			ID:           columnToString(m, "audit_id"),
			AuditType:    columnToString(m, "audit_type"),
			Timestamp:    columnToInt64(m, "timestamp"),
			EntityType:   columnToString(m, "entity_type"),
			EntityID:     columnToString(m, "entity_id"),
			IPaddress:    columnToString(m, "ipaddress"),
			RequestID:    columnToString(m, "request_id"),
			User:         columnToString(m, "user"),
			Role:         columnToString(m, "role"),
			UserAgent:    columnToString(m, "user_agent"),
			Organization: columnToString(m, "organization"),
			DeveloperID:  columnToString(m, "developer_id"),
			AppID:        columnToString(m, "app_id"),
			OldValue:     columnToMapString(m, "old_value"),
			NewValue:     columnToMapString(m, "new_value"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		s.db.metrics.QueryFailed(auditMetricLabel)
		return types.Audits{}, types.NewDatabaseError(err)
	}
	s.db.metrics.QuerySuccessful(auditMetricLabel)
	return audits, nil
}

// Write an entry in audit log
func (s *AuditStore) Write(a *types.Audit) types.Error {

	query := "INSERT INTO audits (" + auditColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		a.ID,
		a.AuditType,
		a.Timestamp,
		a.EntityType,
		a.EntityID,
		a.IPaddress,
		a.RequestID,
		a.User,
		a.Role,
		a.UserAgent,
		a.Organization,
		a.DeveloperID,
		a.AppID,
		valueToJSON(a.OldValue),
		valueToJSON(a.NewValue)).Exec(); err != nil {

		s.db.metrics.QueryFailed(auditMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update audit (%s)", err))
	}
	return nil
}
