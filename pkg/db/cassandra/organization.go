package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const organizationMetricLabel = "organizations"

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
	query := "SELECT * FROM organizations ALLOW FILTERING"
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

	query := "SELECT * FROM organizations WHERE name = ? LIMIT 1"
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
			Attributes:     s.db.UnmarshallJSONArrayOfAttributes(m["attributes"].(string)),
			CreatedAt:      m["created_at"].(int64),
			CreatedBy:      m["created_by"].(string),
			DisplayName:    m["display_name"].(string),
			LastmodifiedAt: m["lastmodified_at"].(int64),
			LastmodifiedBy: m["lastmodified_by"].(string),
			Name:           m["name"].(string),
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

	o.Attributes = shared.TidyAttributes(o.Attributes)
	o.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	if err := s.db.CassandraSession.Query(`INSERT INTO organizations (
name,
display_name,
attributes,
created_at,
created_by,
lastmodified_at,
lastmodified_by) VALUES(?,?,?,?,?,?,?)`,

		o.Name,
		o.DisplayName,
		s.db.MarshallArrayOfAttributesToJSON(o.Attributes),
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
