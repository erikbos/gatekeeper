package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// List of organization columns we use
	organizationColumns = `name,
display_name,
attributes,
created_at,
created_by,
lastmodified_at,
lastmodified_by`
	// Prometheus label for metrics of db interactions
	organizationMetricLabel = "organizations"
)

// OrganizationStore holds our database config
type OrganizationStore struct {
	db *Database
}

// NewOrganizationStore creates organization instance
func NewOrganizationStore(database *Database) *OrganizationStore {
	return &OrganizationStore{
		db: database,
	}
}

// GetAll retrieves all organizations
func (s *OrganizationStore) GetAll() (types.Organizations, types.Error) {

	query := "SELECT " + organizationColumns + " FROM organizations"
	organizations, err := s.runGetOrganizationQuery(query)
	if err != nil {
		s.db.metrics.QueryFailed(organizationMetricLabel)
		return types.NullOrganizations, types.NewDatabaseError(err)
	}

	s.db.metrics.QueryHit(organizationMetricLabel)
	return organizations, nil
}

// Get retrieves a organization from database
func (s *OrganizationStore) Get(organizationName string) (*types.Organization, types.Error) {

	query := "SELECT " + organizationColumns + " FROM organizations WHERE name = ? LIMIT 1"
	organizations, err := s.runGetOrganizationQuery(query, organizationName)
	if err != nil {
		s.db.metrics.QueryFailed(organizationMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(organizations) == 0 {
		s.db.metrics.QueryMiss(organizationMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("can not find organization '%s'", organizationName))
	}

	s.db.metrics.QueryHit(organizationMetricLabel)
	return &organizations[0], nil
}

// runGetOrganizationQuery executes CQL query and returns resultset
func (s *OrganizationStore) runGetOrganizationQuery(query string, queryParameters ...interface{}) (types.Organizations, error) {
	var organizations types.Organizations

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		organizations = append(organizations, types.Organization{
			Name:           columnValueString(m, "name"),
			DisplayName:    columnValueString(m, "display_name"),
			Attributes:     AttributesUnmarshal(columnValueString(m, "attributes")),
			CreatedAt:      columnValueInt64(m, "created_at"),
			CreatedBy:      columnValueString(m, "created_by"),
			LastModifiedAt: columnValueInt64(m, "lastmodified_at"),
			LastModifiedBy: columnValueString(m, "lastmodified_by"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		return types.Organizations{}, err
	}
	return organizations, nil
}

// Update UPSERTs an organization in database
func (s *OrganizationStore) Update(o *types.Organization) types.Error {

	query := "INSERT INTO organizations (" + organizationColumns + ") VALUES(?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		o.Name,
		o.DisplayName,
		AttributesMarshal(o.Attributes),
		o.CreatedAt,
		o.CreatedBy,
		o.LastModifiedAt,
		o.LastModifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(organizationMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update organization '%s' (%s)", o.Name, err))
	}
	return nil
}

// Delete deletes a organization
func (s *OrganizationStore) Delete(organizationToDelete string) types.Error {

	query := "DELETE FROM organizations WHERE name = ?"
	if err := s.db.CassandraSession.Query(query, organizationToDelete).Exec(); err != nil {
		s.db.metrics.QueryFailed(organizationMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
