package db

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const appsMetricLabel = "developerapps"

// GetDeveloperAppsByOrganization retrieves all developer apps belonging to an organization
func (d *Database) GetDeveloperAppsByOrganization(organizationName string) ([]shared.DeveloperApp, error) {

	query := "SELECT * FROM developer_apps WHERE organization_name = ? ALLOW FILTERING"
	developerapps, err := d.runGetDeveloperAppQuery(query, organizationName)
	if err != nil {
		return []shared.DeveloperApp{}, err
	}

	if len(developerapps) == 0 {
		d.metricsQueryMiss(appsMetricLabel)
		return developerapps,
			fmt.Errorf("Can not find developer apps in organization '%s'", organizationName)
	}

	d.metricsQueryHit(appsMetricLabel)
	return developerapps, nil
}

// GetDeveloperAppByName returns details of a developer app
func (d *Database) GetDeveloperAppByName(organization, developerAppName string) (*shared.DeveloperApp, error) {

	query := "SELECT * FROM developer_apps WHERE organization_name = ? AND name = ? LIMIT 1"
	developerapps, err := d.runGetDeveloperAppQuery(query, organization, developerAppName)
	if err != nil {
		return nil, err
	}

	if len(developerapps) == 0 {
		d.metricsQueryMiss(appsMetricLabel)
		return nil, fmt.Errorf("Can not find developer app '%s'", developerAppName)
	}

	d.metricsQueryHit(appsMetricLabel)
	return &developerapps[0], nil
}

// GetDeveloperAppByID returns details of a developer app
func (d *Database) GetDeveloperAppByID(organization, developerAppID string) (*shared.DeveloperApp, error) {

	query := "SELECT * FROM developer_apps WHERE app_id = ? LIMIT 1"
	developerapps, err := d.runGetDeveloperAppQuery(query, developerAppID)
	if err != nil {
		return nil, err
	}

	if len(developerapps) == 0 {
		d.metricsQueryMiss(appsMetricLabel)
		return nil, fmt.Errorf("Can not find developer app id '%s'", developerAppID)
	}

	d.metricsQueryHit(appsMetricLabel)
	return &developerapps[0], nil
}

// GetDeveloperAppCountByDeveloperID retrieves number of apps belonging to a developer
func (d *Database) GetDeveloperAppCountByDeveloperID(developerID string) int {

	var developerAppCount int
	query := "SELECT count(*) FROM developer_apps WHERE developer_id = ?"
	if err := d.cassandraSession.Query(query, developerID).Scan(&developerAppCount); err != nil {
		d.metricsQueryMiss(appsMetricLabel)
		return -1
	}

	d.metricsQueryHit(appsMetricLabel)
	return developerAppCount
}

// runGetDeveloperAppQuery executes CQL query and returns resulset
func (d *Database) runGetDeveloperAppQuery(query string, queryParameters ...interface{}) ([]shared.DeveloperApp, error) {
	var developerapps []shared.DeveloperApp

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		developerapps = append(developerapps, shared.DeveloperApp{
			AppID:            m["app_id"].(string),
			Attributes:       d.unmarshallJSONArrayOfAttributes(m["attributes"].(string)),
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

// UpdateDeveloperAppByName UPSERTs a developer app in database
func (d *Database) UpdateDeveloperAppByName(app *shared.DeveloperApp) error {

	app.Attributes = shared.TidyAttributes(app.Attributes)
	app.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	if err := d.cassandraSession.Query(`INSERT INTO developer_apps (
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
		d.marshallArrayOfAttributesToJSON(app.Attributes),
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

// DeleteDeveloperAppByID deletes an developer app
func (d *Database) DeleteDeveloperAppByID(organizationName, developerAppID string) error {

	_, err := d.GetDeveloperAppByID(organizationName, developerAppID)
	if err != nil {
		return err
	}

	query := "DELETE FROM developer_apps WHERE app_id = ?"
	return d.cassandraSession.Query(query, developerAppID).Exec()
}
