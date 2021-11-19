package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// List of developer columns we use
	developerColumns = `developer_id,
apps,
attributes,
status,
user_name,
email,
first_name,
last_name,
organization_name,
created_at,
created_by,
lastmodified_at,
lastmodified_by`

	// Prometheus label for metrics of db interactions
	developerMetricLabel = "developers"
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
func (s *DeveloperStore) GetAll(organizationName string) (types.Developers, types.Error) {

	query := "SELECT " + developerColumns + " FROM developers WHERE organization_name = ?"
	developers, err := s.runGetDeveloperQuery(query, organizationName)
	if err != nil {
		s.db.metrics.QueryMiss(developerMetricLabel)
		return types.NullDevelopers, types.NewDatabaseError(err)
	}

	s.db.metrics.QueryHit(developerMetricLabel)
	return developers, nil
}

// GetByEmail retrieves a developer from database
func (s *DeveloperStore) GetByEmail(organizationName, developerEmail string) (*types.Developer, types.Error) {

	query := "SELECT " + developerColumns + " FROM developers WHERE organization_name = ? AND email = ? LIMIT 1 ALLOW FILTERING"
	developers, err := s.runGetDeveloperQuery(query, organizationName, developerEmail)
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
func (s *DeveloperStore) GetByID(organizationName, developerID string) (*types.Developer, types.Error) {

	query := "SELECT " + developerColumns + " FROM developers WHERE organization_name = ? AND developer_id = ? LIMIT 1"
	developers, err := s.runGetDeveloperQuery(query, organizationName, developerID)
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
			Apps:           stringSliceUnmarshal(columnValueString(m, "apps")),
			Attributes:     AttributesUnmarshal(columnValueString(m, "attributes")),
			CreatedAt:      columnValueInt64(m, "created_at"),
			CreatedBy:      columnValueString(m, "created_by"),
			DeveloperID:    columnValueString(m, "developer_id"),
			Email:          columnValueString(m, "email"),
			FirstName:      columnValueString(m, "first_name"),
			LastName:       columnValueString(m, "last_name"),
			LastModifiedAt: columnValueInt64(m, "lastmodified_at"),
			LastModifiedBy: columnValueString(m, "lastmodified_by"),
			Status:         columnValueString(m, "status"),
			UserName:       columnValueString(m, "user_name"),
		})
		m = map[string]interface{}{}
	}
	if err := iterable.Close(); err != nil {
		return types.NullDevelopers, err
	}
	return developers, nil
}

// Update UPSERTs a developer in database
func (s *DeveloperStore) Update(organizationName string, d *types.Developer) types.Error {

	query := "INSERT INTO developers (" + developerColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		d.DeveloperID,
		stringSliceMarshal(d.Apps),
		AttributesMarshal(d.Attributes),
		d.Status,
		d.UserName,
		d.Email,
		d.FirstName,
		d.LastName,
		organizationName,
		d.CreatedAt,
		d.CreatedBy,
		d.LastModifiedAt,
		d.LastModifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(developerMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update developer '%s' (%s)", d.DeveloperID, err))
	}
	return nil
}

// DeleteByID deletes a developer
func (s *DeveloperStore) DeleteByID(organizationName, developerID string) types.Error {

	query := "DELETE FROM developers WHERE WHERE organization_name = ? AND developer_id = ?"
	if err := s.db.CassandraSession.Query(query, organizationName, developerID).Exec(); err != nil {
		s.db.metrics.QueryFailed(developerMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
