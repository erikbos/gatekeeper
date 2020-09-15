package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// Prometheus label for metrics of db interactions
	developerAppsMetricLabel = "developerapps"

	// List of developer app columns we use
	developerAppColumns = `app_id,
developer_id,
name,
display_name,
attributes,
organization_name,
status,
created_at,
created_by,
lastmodified_at,
lastmodified_by`
)

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
func (s *DeveloperAppStore) GetByOrganization(organizationName string) (types.DeveloperApps, error) {

	query := "SELECT " + developerAppColumns + " FROM developer_apps WHERE organization_name = ? ALLOW FILTERING"
	developerapps, err := s.runGetDeveloperAppQuery(query, organizationName)
	if err != nil {
		s.db.metrics.QueryMiss(developerAppsMetricLabel)
		return types.DeveloperApps{}, err
	}

	// if len(developerapps) == 0 {
	// 	s.db.metrics.QueryMiss(developerAppsMetricLabel)
	// 	return developerapps,
	// 		fmt.Errorf("Can not find developer apps in organization '%s'", organizationName)
	// }

	s.db.metrics.QueryHit(developerAppsMetricLabel)
	return developerapps, nil
}

// GetByName returns a developer app
func (s *DeveloperAppStore) GetByName(organization, developerAppName string) (*types.DeveloperApp, error) {

	query := "SELECT " + developerAppColumns + " FROM developer_apps WHERE organization_name = ? AND name = ? LIMIT 1"
	developerapps, err := s.runGetDeveloperAppQuery(query, organization, developerAppName)
	if err != nil {
		s.db.metrics.QueryMiss(developerAppsMetricLabel)
		return nil, err
	}

	if len(developerapps) == 0 {
		s.db.metrics.QueryMiss(developerAppsMetricLabel)
		return nil, fmt.Errorf("Can not find developer app '%s'", developerAppName)
	}

	s.db.metrics.QueryHit(developerAppsMetricLabel)
	return &developerapps[0], nil
}

// GetByID returns a developer app
func (s *DeveloperAppStore) GetByID(organization, developerAppID string) (*types.DeveloperApp, error) {

	query := "SELECT " + developerAppColumns + " FROM developer_apps WHERE app_id = ? LIMIT 1"
	developerapps, err := s.runGetDeveloperAppQuery(query, developerAppID)
	if err != nil {
		return nil, err
	}

	if len(developerapps) == 0 {
		s.db.metrics.QueryMiss(developerAppsMetricLabel)
		return nil, fmt.Errorf("Can not find developer app id '%s'", developerAppID)
	}

	s.db.metrics.QueryHit(developerAppsMetricLabel)
	return &developerapps[0], nil
}

// GetCountByDeveloperID retrieves number of apps belonging to a developer
func (s *DeveloperAppStore) GetCountByDeveloperID(developerID string) int {

	var developerAppCount int
	query := "SELECT count(*) FROM developer_apps WHERE developer_id = ?"
	if err := s.db.CassandraSession.Query(query, developerID).Scan(&developerAppCount); err != nil {
		s.db.metrics.QueryMiss(developerAppsMetricLabel)
		return -1
	}

	s.db.metrics.QueryHit(developerAppsMetricLabel)
	return developerAppCount
}

// runGetDeveloperAppQuery executes CQL query and returns resulset
func (s *DeveloperAppStore) runGetDeveloperAppQuery(query string, queryParameters ...interface{}) (types.DeveloperApps, error) {
	var developerapps types.DeveloperApps

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	iterable := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		developerapps = append(developerapps, types.DeveloperApp{
			AppID:            columnValueString(m, "app_id"),
			DeveloperID:      columnValueString(m, "developer_id"),
			Name:             columnValueString(m, "name"),
			DisplayName:      columnValueString(m, "display_name"),
			Attributes:       types.DeveloperApp{}.Attributes.Unmarshal(m["attributes"].(string)),
			OrganizationName: columnValueString(m, "organization_name"),
			Status:           columnValueString(m, "status"),
			CreatedAt:        columnValueInt64(m, "created_at"),
			CreatedBy:        columnValueString(m, "created_by"),
			LastmodifiedAt:   columnValueInt64(m, "lastmodified_at"),
			LastmodifiedBy:   columnValueString(m, "lastmodified_by"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iterable.Close(); err != nil {
		log.Error(err)
		return types.DeveloperApps{}, err
	}
	return developerapps, nil
}

// UpdateByName UPSERTs a developer app
func (s *DeveloperAppStore) UpdateByName(app *types.DeveloperApp) error {

	app.Attributes.Tidy()
	app.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	query := "INSERT INTO developer_apps (" + developerAppColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		app.AppID,
		app.DeveloperID,
		app.Name,
		app.DisplayName,
		app.Attributes.Marshal(),
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
