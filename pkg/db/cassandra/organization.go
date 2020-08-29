package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
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
func (s *OrganizationStore) GetAll() ([]shared.Organization, error) {

	// FIXME this ugly workaround to have to pass an argument
	query := "SELECT " + organizationColumns + " FROM organizations ALLOW FILTERING"
	organizations, err := s.runGetOrganizationQuery(query, "")
	if err != nil {
		return []shared.Organization{}, fmt.Errorf("Cannot retrieve list of organizations (%s)", err)
	}

	if len(organizations) == 0 {
		s.db.metrics.QueryMiss(organizationMetricLabel)
	} else {
		s.db.metrics.QueryHit(organizationMetricLabel)
	}

	return organizations, nil
}

// GetByName retrieves an organization
func (s *OrganizationStore) GetByName(organizationName string) (*shared.Organization, error) {

	query := "SELECT " + organizationColumns + " FROM organizations WHERE name = ? LIMIT 1"
	organizations, err := s.runGetOrganizationQuery(query, organizationName)
	if err != nil {
		return nil, err
	}

	if len(organizations) == 0 {
		s.db.metrics.QueryMiss(organizationMetricLabel)
		return nil, fmt.Errorf("Can not find organization (%s)", organizationName)
	}

	s.db.metrics.QueryHit(organizationMetricLabel)
	return &organizations[0], nil
}

// runGetOrganizationQuery executes CQL query and returns resultset
func (s *OrganizationStore) runGetOrganizationQuery(query, queryParameter string) ([]shared.Organization, error) {
	var organizations []shared.Organization

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	var iter *gocql.Iter
	if queryParameter == "" {
		iter = s.db.CassandraSession.Query(query).Iter()
	} else {
		iter = s.db.CassandraSession.Query(query, queryParameter).Iter()
	}

	if iter.NumRows() == 0 {
		_ = iter.Close()
		return []shared.Organization{}, nil
	}

	m := make(map[string]interface{})
	for iter.MapScan(m) {
		organizations = append(organizations, shared.Organization{
			Name:           columnValueString(m, "name"),
			DisplayName:    columnValueString(m, "display_name"),
			Attributes:     shared.Organization{}.Attributes.Unmarshal(columnValueString(m, "attributes")),
			CreatedAt:      columnValueInt64(m, "created_at"),
			CreatedBy:      columnValueString(m, "created_by"),
			LastmodifiedAt: columnValueInt64(m, "lastmodified_at"),
			LastmodifiedBy: columnValueString(m, "lastmodified_by"),
		})
		m = map[string]interface{}{}
	}
	if err := iter.Close(); err != nil {
		log.Error(err)
		return []shared.Organization{}, err
	}
	return organizations, nil
}

// UpdateByName UPSERTs an organization
func (s *OrganizationStore) UpdateByName(o *shared.Organization) error {

	o.Attributes.Tidy()
	o.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	query := "INSERT INTO organizations (" + organizationColumns + ") VALUES(?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		o.Name,
		o.DisplayName,
		o.Attributes.Marshal(),
		o.CreatedAt,
		o.CreatedBy,
		o.LastmodifiedAt,
		o.LastmodifiedBy).Exec(); err != nil {

		return fmt.Errorf("Cannot update organization '%s' (%v)", o.Name, err)
	}
	return nil
}

// DeleteByName deletes an organization
func (s *OrganizationStore) DeleteByName(organizationToDelete string) error {

	_, err := s.GetByName(organizationToDelete)
	if err != nil {
		return err
	}

	query := "DELETE FROM organizations WHERE name = ?"
	return s.db.CassandraSession.Query(query, organizationToDelete).Exec()
}
