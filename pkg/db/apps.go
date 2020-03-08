package db

import (
	"fmt"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus label for metrics of db interactions
const appsMetricLabel = "apps"

// GetDeveloperAppsByOrganization retrieves all developer apps belonging to an organization
func (d *Database) GetDeveloperAppsByOrganization(organizationName string) ([]types.DeveloperApp, error) {
	query := "SELECT * FROM apps WHERE organization_name = ? ALLOW FILTERING"
	developerapps := d.runGetDeveloperAppQuery(query, organizationName)
	if len(developerapps) == 0 {
		d.metricsQueryMiss(appsMetricLabel)
		return developerapps,
			fmt.Errorf("Can not find developers in organization %s", organizationName)
	}
	d.metricsQueryHit(appsMetricLabel)
	return developerapps, nil
}

// GetDeveloperAppByName returns details of a developer app
func (d *Database) GetDeveloperAppByName(organization, developerAppName string) (types.DeveloperApp, error) {
	query := "SELECT * FROM apps WHERE organization_name = ? AND name = ? LIMIT 1"
	developerapps := d.runGetDeveloperAppQuery(query, organization, developerAppName)
	if len(developerapps) == 0 {
		d.metricsQueryMiss(appsMetricLabel)
		return types.DeveloperApp{},
			fmt.Errorf("Can not find developer app (%s)", developerAppName)
	}
	d.metricsQueryHit(appsMetricLabel)
	return developerapps[0], nil
}

// GetDeveloperAppByID returns details of a developer app
func (d *Database) GetDeveloperAppByID(organization, developerAppID string) (types.DeveloperApp, error) {
	query := "SELECT * FROM apps WHERE key = ? LIMIT 1"
	developerapps := d.runGetDeveloperAppQuery(query, developerAppID)
	if len(developerapps) == 0 {
		d.metricsQueryMiss(appsMetricLabel)
		return types.DeveloperApp{},
			fmt.Errorf("Can not find developer app id (%s)", developerAppID)
	}
	d.metricsQueryHit(appsMetricLabel)
	return developerapps[0], nil
}

// GetDeveloperAppCountByDeveloperID retrieves number of apps belonging to a developer
func (d *Database) GetDeveloperAppCountByDeveloperID(organizationName string) int {
	var developerAppCount int
	query := "SELECT count(*) FROM apps WHERE parent_id = ?"
	if err := d.cassandraSession.Query(query, organizationName).Scan(&developerAppCount); err != nil {
		d.metricsQueryMiss(appsMetricLabel)
		return -1
	}
	d.metricsQueryHit(appsMetricLabel)
	return developerAppCount
}

// runGetDeveloperAppQuery executes CQL query and returns resulset
func (d *Database) runGetDeveloperAppQuery(query string, queryParameters ...interface{}) []types.DeveloperApp {
	var developerapps []types.DeveloperApp

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		developerapps = append(developerapps, types.DeveloperApp{
			AccessType:       m["access_type"].(string),
			AppFamily:        m["app_family"].(string),
			AppID:            m["app_id"].(string),
			AppType:          m["app_type"].(string),
			Attributes:       d.unmarshallJSONArrayOfAttributes(m["attributes"].(string), false),
			CallbackURL:      m["callback_url"].(string),
			CreatedAt:        m["created_at"].(int64),
			CreatedBy:        m["created_by"].(string),
			DeveloperAppID:   m["key"].(string),
			DisplayName:      m["display_name"].(string),
			LastmodifiedAt:   m["lastmodified_at"].(int64),
			LastmodifiedBy:   m["lastmodified_by"].(string),
			Name:             m["name"].(string),
			OrganizationName: m["organization_name"].(string),
			ParentID:         m["parent_id"].(string),
			ParentStatus:     m["parent_status"].(string),
			Status:           m["status"].(string),
		})
		m = map[string]interface{}{}
	}
	return developerapps
}

// UpdateDeveloperAppByName UPSERTs a developer app in database
func (d *Database) UpdateDeveloperAppByName(updatedDeveloperApp types.DeveloperApp) error {
	query := "INSERT INTO apps (key,app_id,attributes," +
		"created_at,created_by, display_name," +
		"lastmodified_at,lastmodified_by,name," +
		"organization_name,parent_id,parent_status," +
		"status) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)"

	Attributes := d.marshallArrayOfAttributesToJSON(updatedDeveloperApp.Attributes, false)
	err := d.cassandraSession.Query(query,
		updatedDeveloperApp.DeveloperAppID, updatedDeveloperApp.AppID, Attributes,
		updatedDeveloperApp.CreatedAt, updatedDeveloperApp.CreatedBy, updatedDeveloperApp.DisplayName,
		updatedDeveloperApp.LastmodifiedAt, updatedDeveloperApp.LastmodifiedBy, updatedDeveloperApp.Name,
		updatedDeveloperApp.OrganizationName, updatedDeveloperApp.ParentID, updatedDeveloperApp.ParentStatus,
		updatedDeveloperApp.Status).Exec()
	if err == nil {
		return nil
	}
	return fmt.Errorf("Can not update developer app (%v)", err)
}

// DeleteDeveloperAppByID deletes an developer app
func (d *Database) DeleteDeveloperAppByID(organizationName, developerAppID string) error {
	_, err := d.GetDeveloperAppByID(organizationName, developerAppID)
	if err != nil {
		return err
	}
	query := "DELETE FROM apps WHERE key = ?"
	return d.cassandraSession.Query(query, developerAppID).Exec()
}
