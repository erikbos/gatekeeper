package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const developerMetricLabel = "developers"

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

// GetByOrganization retrieves all developer belonging to an organization
func (s *DeveloperStore) GetByOrganization(organizationName string) ([]shared.Developer, error) {

	query := "SELECT * FROM developers WHERE organization_name = ? ALLOW FILTERING"
	developers, err := s.runGetDeveloperQuery(query, organizationName)
	if err != nil {
		return []shared.Developer{}, err
	}

	if len(developers) == 0 {
		s.db.metrics.QueryMiss(developerMetricLabel)
		return developers,
			fmt.Errorf("Could not retrieve developers in organization '%s'", organizationName)
	}

	s.db.metrics.QueryHit(developerMetricLabel)
	return developers, nil
}

// GetCountByOrganization retrieves number of developer belonging to an organization
func (s *DeveloperStore) GetCountByOrganization(organizationName string) int {

	var developerCount int

	query := "SELECT count(*) FROM developers WHERE organization_name = ? ALLOW FILTERING"
	if err := s.db.CassandraSession.Query(query, organizationName).Scan(&developerCount); err != nil {
		s.db.metrics.QueryMiss(developerMetricLabel)
		return -1
	}

	s.db.metrics.QueryHit(developerMetricLabel)
	return developerCount
}

// GetByEmail retrieves a developer from database
func (s *DeveloperStore) GetByEmail(developerOrganization, developerEmail string) (*shared.Developer, error) {

	query := "SELECT * FROM developers WHERE organization_name = ? AND email = ? LIMIT 1 ALLOW FILTERING"
	developers, err := s.runGetDeveloperQuery(query, developerOrganization, developerEmail)
	if err != nil {
		return &shared.Developer{}, err
	}

	if len(developers) == 0 {
		s.db.metrics.QueryMiss(developerMetricLabel)
		return nil, fmt.Errorf("Can not find developer '%s'", developerEmail)
	}

	s.db.metrics.QueryHit(developerMetricLabel)
	return &developers[0], nil
}

// GetByID retrieves a developer from database
func (s *DeveloperStore) GetByID(developerID string) (*shared.Developer, error) {

	query := "SELECT * FROM developers WHERE developer_id = ? LIMIT 1"
	developers, err := s.runGetDeveloperQuery(query, developerID)
	if err != nil {
		return nil, err
	}

	if len(developers) == 0 {
		s.db.metrics.QueryMiss(developerMetricLabel)
		return nil, fmt.Errorf("Can not find developerId '%s'", developerID)
	}

	s.db.metrics.QueryHit(developerMetricLabel)
	return &developers[0], nil
}

// runDeveloperQuery executes CQL query and returns resultset
func (s *DeveloperStore) runGetDeveloperQuery(query string, queryParameters ...interface{}) ([]shared.Developer, error) {

	var developers []shared.Developer

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	// Run query, and transfer in batches of 100 rows
	iterable := s.db.CassandraSession.Query(query, queryParameters...).PageSize(100).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		developers = append(developers, shared.Developer{
			Apps:             shared.Developer{}.Apps.Unmarshal(m["apps"].(string)),
			Attributes:       shared.Developer{}.Attributes.Unmarshal(m["attributes"].(string)),
			DeveloperID:      m["developer_id"].(string),
			CreatedAt:        m["created_at"].(int64),
			CreatedBy:        m["created_by"].(string),
			Email:            m["email"].(string),
			FirstName:        m["first_name"].(string),
			LastName:         m["last_name"].(string),
			LastmodifiedAt:   m["lastmodified_at"].(int64),
			LastmodifiedBy:   m["lastmodified_by"].(string),
			OrganizationName: m["organization_name"].(string),
			Status:           m["status"].(string),
			SuspendedTill:    m["suspended_till"].(int64),
			UserName:         m["user_name"].(string),
		})
		m = map[string]interface{}{}
	}
	if err := iterable.Close(); err != nil {
		log.Error(err)
		return []shared.Developer{}, err
	}
	return developers, nil
}

// UpdateByName UPSERTs a developer in database
func (s *DeveloperStore) UpdateByName(dev *shared.Developer) error {

	dev.Attributes.Tidy()
	dev.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	if err := s.db.CassandraSession.Query(`INSERT INTO developers (
developer_id,
apps,
attributes,
organization_name,
status,
user_name,
email,
first_name,
last_name,
suspended_till,
created_at,
created_by,
lastmodified_at,
lastmodified_by) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,

		dev.DeveloperID,
		dev.Apps.Marshal(),
		dev.Attributes.Marshal(),
		dev.OrganizationName,
		dev.Status,
		dev.UserName,
		dev.Email,
		dev.FirstName,
		dev.LastName,
		dev.SuspendedTill,
		dev.CreatedAt,
		dev.CreatedBy,
		dev.LastmodifiedAt,
		dev.LastmodifiedBy).Exec(); err != nil {

		return fmt.Errorf("Cannot update developer (%v)", err)
	}
	return nil
}

// DeleteByEmail deletes a developer
func (s *DeveloperStore) DeleteByEmail(organizationName, developerEmail string) error {

	developer, err := s.GetByEmail(organizationName, developerEmail)
	if err != nil {
		return err
	}

	query := "DELETE FROM developers WHERE developer_id = ?"
	return s.db.CassandraSession.Query(query, developer.DeveloperID).Exec()
}
