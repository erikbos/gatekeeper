package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// Prometheus label for metrics of db interactions
	organizationMetricLabel = "organizations"

	// List of organization columns we use
	organizationColumns = `name,
display_name,
attributes,
created_at,
created_by,
lastmodified_at,
lastmodified_by`
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

	// FIXME this ugly workaround to have to pass an argument
	query := "SELECT " + organizationColumns + " FROM organizations ALLOW FILTERING"
	organizations, err := s.runGetOrganizationQuery(query, "")
	if err != nil {
		s.db.metrics.QueryFailed(organizationMetricLabel)
		return types.NullOrganizations, types.NewDatabaseError(err)
	}

	s.db.metrics.QueryHit(organizationMetricLabel)
	return organizations, nil
}

// Get retrieves an organization
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
			fmt.Errorf("Cannot find organization '%s'", organizationName))
	}

	s.db.metrics.QueryHit(organizationMetricLabel)
	return &organizations[0], nil
}

// runGetOrganizationQuery executes CQL query and returns resultset
func (s *OrganizationStore) runGetOrganizationQuery(query, queryParameter string) (types.Organizations, error) {
	var organizations types.Organizations

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	var iter *gocql.Iter
	if queryParameter == "" {
		iter = s.db.CassandraSession.Query(query).Iter()
	} else {
		iter = s.db.CassandraSession.Query(query, queryParameter).Iter()
	}

	m := make(map[string]interface{})
	for iter.MapScan(m) {
		organizations = append(organizations, types.Organization{
			Name:           columnValueString(m, "name"),
			DisplayName:    columnValueString(m, "display_name"),
			Attributes:     types.Organization{}.Attributes.Unmarshal(columnValueString(m, "attributes")),
			CreatedAt:      columnValueInt64(m, "created_at"),
			CreatedBy:      columnValueString(m, "created_by"),
			LastmodifiedAt: columnValueInt64(m, "lastmodified_at"),
			LastmodifiedBy: columnValueString(m, "lastmodified_by"),
		})
		m = map[string]interface{}{}
	}
	if err := iter.Close(); err != nil {
		return types.NullOrganizations, err
	}
	return organizations, nil
}

// Update UPSERTs an organization
func (s *OrganizationStore) Update(o *types.Organization) types.Error {

	query := "INSERT INTO organizations (" + organizationColumns + ") VALUES(?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		o.Name,
		o.DisplayName,
		o.Attributes.Marshal(),
		o.CreatedAt,
		o.CreatedBy,
		o.LastmodifiedAt,
		o.LastmodifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(organizationMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("Cannot update organization '%s'", o.Name))
	}
	return nil
}

// Delete deletes an organization
func (s *OrganizationStore) Delete(organizationToDelete string) types.Error {

	query := "DELETE FROM organizations WHERE name = ?"
	if err := s.db.CassandraSession.Query(query, organizationToDelete).Exec(); err != nil {
		s.db.metrics.QueryFailed(organizationMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
