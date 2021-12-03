package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// List of audit columns we use
	auditColumns = `id,
timestamp,
change_type,
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
old,
new`

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

// Get retrieves a audit
func (s *AuditStore) Get(auditName string) (types.Audits, types.Error) {

	query := "SELECT " + auditColumns + " FROM audits WHERE name = ? LIMIT 1"
	audits, err := s.runGetAuditQuery(query, auditName)
	if err != nil {
		s.db.metrics.QueryFailed(auditMetricLabel)
		return nil, types.NewDatabaseError(err)

	}

	if len(audits) == 0 {
		s.db.metrics.QueryNotFound(auditMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("can not find audit '%s'", auditName))
	}

	s.db.metrics.QuerySuccessful(auditMetricLabel)
	return audits, nil
}

// runGetAuditQuery executes CQL query and returns resultset
func (s *AuditStore) runGetAuditQuery(query string,
	queryParameters ...interface{}) (types.Audits, error) {

	timer := prometheus.NewTimer(s.db.metrics.queryHistogram)
	defer timer.ObserveDuration()

	var audits types.Audits

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		audits = append(audits, types.Audit{
			Id:           columnToString(m, "id"),
			Timestamp:    columnToInt64(m, "timestamp"),
			ChangeType:   columnToString(m, "change_type"),
			EntityType:   columnToString(m, "entity_type"),
			EntityId:     columnToString(m, "entity_id"),
			IPaddress:    columnToString(m, "ipaddress"),
			RequestID:    columnToString(m, "request_id"),
			User:         columnToString(m, "user"),
			Role:         columnToString(m, "role"),
			UserAgent:    columnToString(m, "user_agent"),
			Organization: columnToString(m, "organization"),
			DeveloperID:  columnToString(m, "developer_id"),
			AppID:        columnToString(m, "app_id"),
			Old:          columnToString(m, "old"),
			New:          columnToString(m, "new"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		return types.Audits{}, err
	}
	return audits, nil
}

// Write an entry in audit log
func (s *AuditStore) Write(a *types.Audit) types.Error {

	query := "INSERT INTO audits (" + auditColumns + ") VALUES(uuid(),?,?,?,?, ?,?,?,?,?, ?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		a.Timestamp,
		a.ChangeType,
		a.EntityType,
		a.EntityId,
		a.IPaddress,
		a.RequestID,
		a.User,
		a.Role,
		a.UserAgent,
		a.Organization,
		a.DeveloperID,
		a.AppID,
		a.Old,
		a.New).Exec(); err != nil {

		s.db.metrics.QueryFailed(auditMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update audit (%s)", err))
	}
	return nil
}
