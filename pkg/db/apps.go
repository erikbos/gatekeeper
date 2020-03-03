package db

import (
	"errors"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

//GetDeveloperAppsByOrganization retrieves all developer belonging to an organization
// FIXME
func (d *Database) GetDeveloperAppsByOrganization(organizationName string) ([]types.DeveloperApp, error) {
	query := "SELECT * FROM apps WHERE organization_name = ? LIMIT 10 ALLOW FILTERING"
	developerapps := d.runGetDeveloperAppQuery(query, organizationName)
	if len(developerapps) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "apps").Inc()
		return developerapps, nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "apps").Inc()
	return developerapps, errors.New("Could not find developer by name")
}

//GetDeveloperAppByName returns details of a DeveloperApplication looked up by Name
//
func (d *Database) GetDeveloperAppByName(organization, developerAppName string) (types.DeveloperApp, error) {
	query := "SELECT * FROM apps WHERE name = ? LIMIT 1"
	developerapps := d.runGetDeveloperAppQuery(query, developerAppName)
	if len(developerapps) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "apps").Inc()
		return developerapps[0], nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "apps").Inc()
	return types.DeveloperApp{}, errors.New("Could not find developer by name")
}

//GetDeveloperAppByID returns details of a DeveloperApplication looked up by ID
//
func (d *Database) GetDeveloperAppByID(organization, developerAppID string) (types.DeveloperApp, error) {
	query := "SELECT * FROM apps WHERE key = ? LIMIT 1"
	developerapps := d.runGetDeveloperAppQuery(query, developerAppID)
	if len(developerapps) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "apps").Inc()
		return developerapps[0], nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "apps").Inc()
	return types.DeveloperApp{}, errors.New("Could not find developer by app id")
}

func (d *Database) runGetDeveloperAppQuery(query, queryParameter string) []types.DeveloperApp {
	var developerapps []types.DeveloperApp

	//Set timer to record how long this function run
	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameter).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		developerapps = append(developerapps, types.DeveloperApp{
			AccessType:  m["access_type"].(string),
			AppFamily:   m["app_family"].(string),
			AppID:       m["app_id"].(string),
			AppType:     m["app_type"].(string),
			Attributes:  d.unmarshallJSONArrayOfAttributes(m["attributes"].(string), false),
			CallbackURL: m["callback_url"].(string),
			CreatedAt:   m["created_at"].(int64),
			CreatedBy:   m["created_by"].(string),
			// DeveloperAppID:   developerAppID,
			DisplayName:      m["display_name"].(string),
			Key:              m["key"].(string),
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
func (d *Database) UpdateDeveloperAppByName(updatedDeveloper types.DeveloperApp) error {
	// query := "INSERT INTO apps (key,apps,attributes, " +
	// 	"created_at, created_by, email, " +
	// 	"first_name, last_name, lastmodified_at, " +
	// 	"lastmodified_by, organization_name, status, user_name)" +
	// 	"VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)"

	// Apps := d.marshallArrayOfStringsToJSON(updatedDeveloper.Apps)
	// Attributes := d.marshallArrayOfAttributesToJSON(updatedDeveloper.Attributes, false)
	// log.Printf("attributes: %s", updatedDeveloper.Attributes)

	// err := d.cassandraSession.Query(query,
	// 	updatedDeveloper.DeveloperID, Apps, Attributes,
	// 	updatedDeveloper.CreatedAt, updatedDeveloper.CreatedBy, updatedDeveloper.Email,
	// 	updatedDeveloper.FirstName, updatedDeveloper.LastName, updatedDeveloper.LastmodifiedAt,
	// 	updatedDeveloper.LastmodifiedBy, updatedDeveloper.OrganizationName, updatedDeveloper.Status,
	// 	updatedDeveloper.UserName).Exec()
	// if err == nil {
	// 	return nil
	// }
	// return fmt.Errorf("Could not update developer app (%v)", err)
	return errors.New("Nope")
}

//DeleteDeveloperAppByName deletes an developer app
func (d *Database) DeleteDeveloperAppByName(organizationName, developerAppName string) error {
	_, err := d.GetDeveloperAppByName(organizationName, developerAppName)
	if err != nil {
		return err
	}
	query := "DELETE FROM apps WHERE key = ?"
	return d.cassandraSession.Query(query, developerAppName).Exec()
}
