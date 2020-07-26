package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const appsMetricLabel = "developerapps"

// DeveloperAppStore holds our database config
type DeveloperAppStore struct {
	db *Database
}

// NewDeveloperAppStore creates developer app instance
func NewDeveloperAppStore(database *Database) *DeveloperAppStore {
	return &DeveloperAppStore{
		db: database,
	}
}

// GetByOrganization retrieves all developer apps belonging to an organization
func (s *DeveloperAppStore) GetByOrganization(organizationName string) ([]shared.DeveloperApp, error) {

	query := "SELECT * FROM developer_apps WHERE organization_name = ? ALLOW FILTERING"
	developerapps, err := s.runGetDeveloperAppQuery(query, organizationName)
	if err != nil {
		return []shared.DeveloperApp{}, err
	}

	if len(developerapps) == 0 {
		s.db.metrics.QueryMiss(appsMetricLabel)
		return developerapps,
			fmt.Errorf("Can not find developer apps in organization '%s'", organizationName)
	}

	s.db.metrics.QueryHit(appsMetricLabel)
	return developerapps, nil
}

// GetByName returns a developer app
func (s *DeveloperAppStore) GetByName(organization, developerAppName string) (*shared.DeveloperApp, error) {

	query := "SELECT * FROM developer_apps WHERE organization_name = ? AND name = ? LIMIT 1"
	developerapps, err := s.runGetDeveloperAppQuery(query, organization, developerAppName)
	if err != nil {
		return nil, err
	}

	if len(developerapps) == 0 {
		s.db.metrics.QueryMiss(appsMetricLabel)
		return nil, fmt.Errorf("Can not find developer app '%s'", developerAppName)
	}

	s.db.metrics.QueryHit(appsMetricLabel)
	return &developerapps[0], nil
}

// GetByID returns a developer app
func (s *DeveloperAppStore) GetByID(organization, developerAppID string) (*shared.DeveloperApp, error) {

	query := "SELECT * FROM developer_apps WHERE app_id = ? LIMIT 1"
	developerapps, err := s.runGetDeveloperAppQuery(query, developerAppID)
	if err != nil {
		return nil, err
	}

	if len(developerapps) == 0 {
		s.db.metrics.QueryMiss(appsMetricLabel)
		return nil, fmt.Errorf("Can not find developer app id '%s'", developerAppID)
	}

	s.db.metrics.QueryHit(appsMetricLabel)
	return &developerapps[0], nil
}

// GetCountByDeveloperID retrieves number of apps belonging to a developer
func (s *DeveloperAppStore) GetCountByDeveloperID(developerID string) int {

	var developerAppCount int
	query := "SELECT count(*) FROM developer_apps WHERE developer_id = ?"
	if err := s.db.CassandraSession.Query(query, developerID).Scan(&developerAppCount); err != nil {
		s.db.metrics.QueryMiss(appsMetricLabel)
		return -1
	}

	s.db.metrics.QueryHit(appsMetricLabel)
	return developerAppCount
}

// runGetDeveloperAppQuery executes CQL query and returns resulset
func (s *DeveloperAppStore) runGetDeveloperAppQuery(query string, queryParameters ...interface{}) ([]shared.DeveloperApp, error) {
	var developerapps []shared.DeveloperApp

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	iterable := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		developerapps = append(developerapps, shared.DeveloperApp{
			AppID:            m["app_id"].(string),
			Attributes:       s.db.UnmarshallJSONArrayOfAttributes(m["attributes"].(string)),
			CreatedAt:        m["created_at"].(int64),
			CreatedBy:        m["created_by"].(string),
			DisplayName:      m["display_name"].(string),
			LastmodifiedAt:   m["lastmodified_at"].(int64),
			LastmodifiedBy:   m["lastmodified_by"].(string),
			Name:             m["name"].(string),
			OrganizationName: m["organization_name"].(string),
			DeveloperID:      m["developer_id"].(string),
			Status:           m["status"].(string),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iterable.Close(); err != nil {
		log.Error(err)
		return []shared.DeveloperApp{}, err
	}
	return developerapps, nil
}

// UpdateByName UPSERTs a developer app
func (s *DeveloperAppStore) UpdateByName(app *shared.DeveloperApp) error {

	app.Attributes = shared.TidyAttributes(app.Attributes)
	app.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	if err := s.db.CassandraSession.Query(`INSERT INTO developer_apps (
app_id,
developer_id,
name,
display_name,
attributes,
organization_name,
status,
created_at,
created_by,
lastmodified_at,
lastmodified_by) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,

		app.AppID,
		app.DeveloperID,
		app.Name,
		app.DisplayName,
		s.db.MarshallArrayOfAttributesToJSON(app.Attributes),
		app.OrganizationName,
		app.Status,
		app.CreatedAt,
		app.CreatedBy,
		app.LastmodifiedAt,
		app.LastmodifiedBy).Exec(); err != nil {

		return fmt.Errorf("Cannot update developer app '%s' (%v)",
			app.AppID, err)
	}
	return nil
}

// DeleteByID deletes a developer app
func (s *DeveloperAppStore) DeleteByID(organizationName, developerAppID string) error {

	_, err := s.GetByID(organizationName, developerAppID)
	if err != nil {
		return err
	}

	query := "DELETE FROM developer_apps WHERE app_id = ?"
	return s.db.CassandraSession.Query(query, developerAppID).Exec()
}
