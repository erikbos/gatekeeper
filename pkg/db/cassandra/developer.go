package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// Prometheus label for metrics of db interactions
	developerMetricLabel = "developers"

	// List of developer columns we use
	developerColumns = `developer_id,
apps,
attributes,
status,
user_name,
email,
first_name,
last_name,
suspended_till,
created_at,
created_by,
lastmodified_at,
lastmodified_by`
)

// DeveloperStore holds our database config
type DeveloperStore struct {
	db *Database
}

// NewDeveloperStore creates developer instance
func NewDeveloperStore(database *Database) *DeveloperStore {
	return &DeveloperStore{
		db: database,
	}
}

// GetAll retrieves all developer
func (s *DeveloperStore) GetAll() (types.Developers, types.Error) {

	query := "SELECT " + developerColumns + " FROM developers"
	developers, err := s.runGetDeveloperQuery(query)
	if err != nil {
		s.db.metrics.QueryMiss(developerMetricLabel)
		return types.NullDevelopers, types.NewDatabaseError(err)
	}

	s.db.metrics.QueryHit(developerMetricLabel)
	return developers, nil
}

// GetByEmail retrieves a developer from database
func (s *DeveloperStore) GetByEmail(developerEmail string) (*types.Developer, types.Error) {

	query := "SELECT " + developerColumns + " FROM developers WHERE email = ? LIMIT 1 ALLOW FILTERING"
	developers, err := s.runGetDeveloperQuery(query, developerEmail)
	if err != nil {
		s.db.metrics.QueryFailed(developerMetricLabel)
		return &types.NullDeveloper, types.NewDatabaseError(err)
	}

	if len(developers) == 0 {
		s.db.metrics.QueryMiss(developerMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("can not find developer '%s'", developerEmail))
	}

	s.db.metrics.QueryHit(developerMetricLabel)
	return &developers[0], nil
}

// GetByID retrieves a developer from database
func (s *DeveloperStore) GetByID(developerID string) (*types.Developer, types.Error) {

	query := "SELECT " + developerColumns + " FROM developers WHERE developer_id = ? LIMIT 1"
	developers, err := s.runGetDeveloperQuery(query, developerID)
	if err != nil {
		s.db.metrics.QueryFailed(developerMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(developers) == 0 {
		s.db.metrics.QueryMiss(developerMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("can not find developerId '%s'", developerID))
	}

	s.db.metrics.QueryHit(developerMetricLabel)
	return &developers[0], nil
}

// runDeveloperQuery executes CQL query and returns resultset
func (s *DeveloperStore) runGetDeveloperQuery(query string, queryParameters ...interface{}) (types.Developers, error) {

	var developers types.Developers

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	// Run query, and transfer in batches of 100 rows
	iterable := s.db.CassandraSession.Query(query, queryParameters...).PageSize(100).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		developers = append(developers, types.Developer{
			DeveloperID:    columnValueString(m, "developer_id"),
			Apps:           types.Developer{}.Apps.Unmarshal(columnValueString(m, "apps")),
			Attributes:     types.Developer{}.Attributes.Unmarshal(columnValueString(m, "attributes")),
			Status:         columnValueString(m, "status"),
			UserName:       columnValueString(m, "user_name"),
			Email:          columnValueString(m, "email"),
			FirstName:      columnValueString(m, "first_name"),
			LastName:       columnValueString(m, "last_name"),
			SuspendedTill:  columnValueInt64(m, "suspended_till"),
			CreatedAt:      columnValueInt64(m, "created_at"),
			CreatedBy:      columnValueString(m, "created_by"),
			LastmodifiedAt: columnValueInt64(m, "lastmodified_at"),
			LastmodifiedBy: columnValueString(m, "lastmodified_by"),
		})
		m = map[string]interface{}{}
	}
	if err := iterable.Close(); err != nil {
		return types.Developers{}, err
	}
	return developers, nil
}

// Update UPSERTs a developer in database
func (s *DeveloperStore) Update(d *types.Developer) types.Error {

	query := "INSERT INTO developers (" + developerColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		d.DeveloperID,
		d.Apps.Marshal(),
		d.Attributes.Marshal(),
		d.Status,
		d.UserName,
		d.Email,
		d.FirstName,
		d.LastName,
		d.SuspendedTill,
		d.CreatedAt,
		d.CreatedBy,
		d.LastmodifiedAt,
		d.LastmodifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(developerMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update developer '%s'", d.DeveloperID))
	}
	return nil
}

// DeleteByID deletes a developer
func (s *DeveloperStore) DeleteByID(developerID string) types.Error {

	query := "DELETE FROM developers WHERE developer_id = ?"
	if err := s.db.CassandraSession.Query(query, developerID).Exec(); err != nil {
		s.db.metrics.QueryFailed(developerMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
