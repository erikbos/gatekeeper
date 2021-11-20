package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// List of developer app columns we use
	developerAppColumns = `app_id,
developer_id,
name,
display_name,
attributes,
status,
callback_url,
scopes,
created_at,
created_by,
lastmodified_at,
lastmodified_by`

	// Prometheus label for metrics of db interactions
	developerAppsMetricLabel = "developerapps"
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

// GetAll retrieves all developer apps
func (s *DeveloperAppStore) GetAll(organizationName string) (types.DeveloperApps, types.Error) {

	query := "SELECT " + developerAppColumns + " FROM developer_apps"
	developerapps, err := s.runGetDeveloperAppQuery(query)
	if err != nil {
		s.db.metrics.QueryNotFound(developerAppsMetricLabel)
		return types.NullDeveloperApps, types.NewDatabaseError(err)
	}

	s.db.metrics.QuerySuccessful(developerAppsMetricLabel)
	return developerapps, nil
}

// GetAllByDeveloperID retrieves all developer apps from a developer
func (s *DeveloperAppStore) GetAllByDeveloperID(organizationName, developerID string) (types.DeveloperApps, types.Error) {

	query := "SELECT " + developerAppColumns + " FROM developer_apps WHERE developer_id = ?"
	developerapps, err := s.runGetDeveloperAppQuery(query, developerID)
	if err != nil {
		s.db.metrics.QueryNotFound(developerAppsMetricLabel)
		return types.NullDeveloperApps, types.NewDatabaseError(err)
	}

	s.db.metrics.QuerySuccessful(developerAppsMetricLabel)
	return developerapps, nil
}

// GetByName returns a developer app
func (s *DeveloperAppStore) GetByName(organizationName, developerEmail, developerAppName string) (*types.DeveloperApp, types.Error) {

	query := "SELECT " + developerAppColumns + " FROM developer_apps WHERE name = ? LIMIT 1"
	developerapps, err := s.runGetDeveloperAppQuery(query, developerAppName)
	if err != nil {
		s.db.metrics.QueryNotFound(developerAppsMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(developerapps) == 0 {
		s.db.metrics.QueryNotFound(developerAppsMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("can not find developer app '%s'", developerAppName))
	}

	s.db.metrics.QuerySuccessful(developerAppsMetricLabel)
	return &developerapps[0], nil
}

// GetByID returns a developer app
func (s *DeveloperAppStore) GetByID(organizationName, developerAppID string) (*types.DeveloperApp, types.Error) {

	query := "SELECT " + developerAppColumns + " FROM developer_apps WHERE app_id = ? LIMIT 1"
	developerapps, err := s.runGetDeveloperAppQuery(query, developerAppID)
	if err != nil {
		return nil, types.NewDatabaseError(err)
	}

	if len(developerapps) == 0 {
		s.db.metrics.QueryNotFound(developerAppsMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("can not find developer app id '%s'", developerAppID))
	}

	s.db.metrics.QuerySuccessful(developerAppsMetricLabel)
	return &developerapps[0], nil
}

// GetCountByDeveloperID retrieves number of apps belonging to a developer
func (s *DeveloperAppStore) GetCountByDeveloperID(organizationName, developerID string) (int, types.Error) {

	var developerAppCount int
	query := "SELECT count(*) FROM developer_apps WHERE developer_id = ?"
	if err := s.db.CassandraSession.Query(query, developerID).Scan(&developerAppCount); err != nil {
		s.db.metrics.QueryNotFound(developerAppsMetricLabel)
		return -1, types.NewDatabaseError(err)
	}

	s.db.metrics.QuerySuccessful(developerAppsMetricLabel)
	return developerAppCount, nil
}

// runGetDeveloperAppQuery executes CQL query and returns resultset
func (s *DeveloperAppStore) runGetDeveloperAppQuery(query string, queryParameters ...interface{}) (types.DeveloperApps, error) {
	var developerapps types.DeveloperApps

	timer := prometheus.NewTimer(s.db.metrics.queryHistogram)
	defer timer.ObserveDuration()

	iterable := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		developerapps = append(developerapps, types.DeveloperApp{
			AppID:          columnToString(m, "app_id"),
			Attributes:     columnToAttributes(m, "attributes"),
			CallbackUrl:    columnToString(m, "callback_url"),
			CreatedAt:      columnToInt64(m, "created_at"),
			CreatedBy:      columnToString(m, "created_by"),
			DeveloperID:    columnToString(m, "developer_id"),
			DisplayName:    columnToString(m, "display_name"),
			LastModifiedAt: columnToInt64(m, "lastmodified_at"),
			LastModifiedBy: columnToString(m, "lastmodified_by"),
			Scopes:         columnToStringSlice(m, "scopes"),
			Status:         columnToString(m, "status"),
			Name:           columnToString(m, "name"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iterable.Close(); err != nil {
		return types.NullDeveloperApps, err
	}
	return developerapps, nil
}

// Update UPSERTs a developer app
func (s *DeveloperAppStore) Update(organizationName string, app *types.DeveloperApp) types.Error {

	query := "INSERT INTO developer_apps (" + developerAppColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		app.AppID,
		app.DeveloperID,
		app.Name,
		app.DisplayName,
		attributesToColumn(app.Attributes),
		app.Status,
		app.CallbackUrl,
		app.Scopes,
		app.CreatedAt,
		app.CreatedBy,
		app.LastModifiedAt,
		app.LastModifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(developerAppsMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update developer app '%s' (%s)", app.AppID, err))
	}
	return nil
}

// DeleteByID deletes a developer app
func (s *DeveloperAppStore) DeleteByID(organizationName, developerAppID string) types.Error {

	query := "DELETE FROM developer_apps WHERE app_id = ?"
	if err := s.db.CassandraSession.Query(query, developerAppID).Exec(); err != nil {
		s.db.metrics.QueryFailed(developerAppsMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
