package db

import (
	"errors"
	"fmt"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

//Prometheus label for metrics of db interactions
const developerMetricLabel = "organizations"

//GetDevelopersByOrganization retrieves all developer belonging to an organization
//
func (d *Database) GetDevelopersByOrganization(organizationName string) ([]types.Developer, error) {
	query := "SELECT * FROM developers WHERE organization_name = ? LIMIT 10 ALLOW FILTERING"
	developers := d.runGetDeveloperQuery(query, organizationName)
	if len(developers) > 0 {
		d.metricsQueryHit(developerMetricLabel)
		return developers, nil
	}
	d.metricsQueryMiss(developerMetricLabel)
	return developers, errors.New("Could not retrieve list of developers")
}

//GetDeveloperCountByOrganization retrieves number of developer belonging to an organization
//
func (d *Database) GetDeveloperCountByOrganization(organizationName string) int {
	var developerCount int
	query := "SELECT count(*) FROM developers WHERE organization_name = ? ALLOW FILTERING"
	if err := d.cassandraSession.Query(query, organizationName).Scan(&developerCount); err != nil {
		return -1
	}
	return developerCount
}

//GetDeveloperByEmail retrieves a developer from database
//
func (d *Database) GetDeveloperByEmail(developerOrganization, developerEmail string) (types.Developer, error) {
	query := "SELECT * FROM developers WHERE email = ? LIMIT 1"
	developers := d.runGetDeveloperQuery(query, developerEmail)
	if len(developers) > 0 {
		d.metricsQueryHit(developerMetricLabel)
		// Check of record of developer matches the required org
		if developers[0].OrganizationName != developerOrganization {
			return types.Developer{}, fmt.Errorf("Organization mismatch")
		}
		return developers[0], nil
	}
	d.metricsQueryMiss(developerMetricLabel)
	return types.Developer{}, fmt.Errorf("Could not find developer (%s)", developerEmail)
}

//GetDeveloperByID retrieves a developer from database
//
func (d *Database) GetDeveloperByID(developerID string) (types.Developer, error) {
	query := "SELECT * FROM developers WHERE key = ? LIMIT 1"
	developers := d.runGetDeveloperQuery(query, developerID)
	if len(developers) > 0 {
		d.metricsQueryHit(developerMetricLabel)
		return developers[0], nil
	}
	d.metricsQueryMiss(developerMetricLabel)
	return types.Developer{}, fmt.Errorf("Could not find developerId (%s)", developerID)
}

// runDeveloperQuery executes CQL query and returns resultset
//
func (d *Database) runGetDeveloperQuery(query, queryParameter string) []types.Developer {
	var developers []types.Developer

	//Set timer to record how long this function run
	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameter).Iter()
	m := make(map[string]interface{})

	// from https://github.com/uber/cherami-server/blob/1de31a4ed1d0a9cd33ff64199f7e91f23e99c11e/cmd/tools/cmq/fix.go
	//
	// for iter.Scan(&uuid, &destinationUUID, &name, &status, &lockTimeoutSeconds, &maxDeliveryCount, &skipOlderMessagesSeconds,
	// 	&deadLetterQueueDestinationUUID, &ownerEmail, &startFrom, &isMultiZone, &activeZone, &zoneConfigs, &delaySeconds, &options) {

	for iterable.MapScan(m) {
		developers = append(developers, types.Developer{
			Apps:             d.unmarshallJSONArrayOfStrings(m["apps"].(string)),
			Attributes:       d.unmarshallJSONArrayOfAttributes(m["attributes"].(string), false),
			DeveloperID:      m["key"].(string),
			CreatedAt:        m["created_at"].(int64),
			CreatedBy:        m["created_by"].(string),
			Email:            m["email"].(string),
			FirstName:        m["first_name"].(string),
			LastName:         m["last_name"].(string),
			LastmodifiedAt:   m["lastmodified_at"].(int64),
			LastmodifiedBy:   m["lastmodified_by"].(string),
			OrganizationName: m["organization_name"].(string),
			// password:          m["password"].(string),
			Salt:     m["salt"].(string),
			Status:   m["status"].(string),
			UserName: m["user_name"].(string),
		})
		m = map[string]interface{}{}
	}
	return developers
}

// UpdateDeveloperByName UPSERTs a developer in database
// Upsert is: In case a developer does not exist (primary key not matching) it will create a new row
func (d *Database) UpdateDeveloperByName(updatedDeveloper types.Developer) error {
	query := "INSERT INTO developers (key,apps,attributes, " +
		"created_at, created_by, email, " +
		"first_name, last_name, lastmodified_at, " +
		"lastmodified_by, organization_name, status, user_name)" +
		"VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)"

	Apps := d.marshallArrayOfStringsToJSON(updatedDeveloper.Apps)
	Attributes := d.marshallArrayOfAttributesToJSON(updatedDeveloper.Attributes, false)
	// log.Printf("attributes: %s", updatedDeveloper.Attributes)

	err := d.cassandraSession.Query(query,
		updatedDeveloper.DeveloperID, Apps, Attributes,
		updatedDeveloper.CreatedAt, updatedDeveloper.CreatedBy, updatedDeveloper.Email,
		updatedDeveloper.FirstName, updatedDeveloper.LastName, updatedDeveloper.LastmodifiedAt,
		updatedDeveloper.LastmodifiedBy, updatedDeveloper.OrganizationName, updatedDeveloper.Status,
		updatedDeveloper.UserName).Exec()
	if err == nil {
		return nil
	}
	return fmt.Errorf("Could not update developer (%v)", err)
}

//DeleteDeveloperByEmail deletes a developer
//
func (d *Database) DeleteDeveloperByEmail(organizatioName, developerEmail string) error {
	developer, err := d.GetDeveloperByEmail(organizatioName, developerEmail)
	if err != nil {
		return err
	}
	query := "DELETE FROM developers WHERE key = ?"
	return d.cassandraSession.Query(query, developer.DeveloperID).Exec()
}
