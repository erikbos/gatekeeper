package db

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const developerMetricLabel = "developers"

// GetDevelopersByOrganization retrieves all developer belonging to an organization
func (d *Database) GetDevelopersByOrganization(organizationName string) ([]shared.Developer, error) {

	query := "SELECT * FROM developers WHERE organization_name = ? ALLOW FILTERING"
	developers, err := d.runGetDeveloperQuery(query, organizationName)
	if err != nil {
		return []shared.Developer{}, err
	}

	if len(developers) == 0 {
		d.metricsQueryMiss(developerMetricLabel)
		return developers,
			fmt.Errorf("Could not retrieve developers in organization '%s'", organizationName)
	}

	d.metricsQueryHit(developerMetricLabel)
	return developers, nil
}

// GetDeveloperCountByOrganization retrieves number of developer belonging to an organization
func (d *Database) GetDeveloperCountByOrganization(organizationName string) int {

	var developerCount int

	query := "SELECT count(*) FROM developers WHERE organization_name = ? ALLOW FILTERING"
	if err := d.cassandraSession.Query(query, organizationName).Scan(&developerCount); err != nil {
		d.metricsQueryMiss(developerMetricLabel)
		return -1
	}

	d.metricsQueryHit(developerMetricLabel)
	return developerCount
}

// GetDeveloperByEmail retrieves a developer from database
func (d *Database) GetDeveloperByEmail(developerOrganization, developerEmail string) (shared.Developer, error) {

	query := "SELECT * FROM developers WHERE organization_name = ? AND email = ? LIMIT 1 ALLOW FILTERING"
	developers, err := d.runGetDeveloperQuery(query, developerOrganization, developerEmail)
	if err != nil {
		return shared.Developer{}, err
	}

	if len(developers) == 0 {
		d.metricsQueryMiss(developerMetricLabel)
		return shared.Developer{}, fmt.Errorf("Can not find developer '%s'", developerEmail)
	}

	d.metricsQueryHit(developerMetricLabel)
	return developers[0], nil
}

// GetDeveloperByID retrieves a developer from database
func (d *Database) GetDeveloperByID(developerID string) (shared.Developer, error) {

	query := "SELECT * FROM developers WHERE developer_id = ? LIMIT 1"
	developers, err := d.runGetDeveloperQuery(query, developerID)
	if err != nil {
		return shared.Developer{}, err
	}

	if len(developers) == 0 {
		d.metricsQueryMiss(developerMetricLabel)
		return shared.Developer{}, fmt.Errorf("Can not find developerId '%s'", developerID)
	}

	d.metricsQueryHit(developerMetricLabel)
	return developers[0], nil
}

// runDeveloperQuery executes CQL query and returns resultset
func (d *Database) runGetDeveloperQuery(query string, queryParameters ...interface{}) ([]shared.Developer, error) {

	var developers []shared.Developer

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	// Run query, and transfer in batches of 100 rows
	iterable := d.cassandraSession.Query(query, queryParameters...).PageSize(100).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		developers = append(developers, shared.Developer{
			Apps:             d.unmarshallJSONArrayOfStrings(m["apps"].(string)),
			Attributes:       d.unmarshallJSONArrayOfAttributes(m["attributes"].(string)),
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

// UpdateDeveloperByName UPSERTs a developer in database
func (d *Database) UpdateDeveloperByName(dev *shared.Developer) error {

	dev.Attributes = shared.TidyAttributes(dev.Attributes)
	dev.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	if err := d.cassandraSession.Query(`INSERT INTO developers (
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
lastmodified_by) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,

		dev.DeveloperID,
		d.marshallArrayOfStringsToJSON(dev.Apps),
		d.marshallArrayOfAttributesToJSON(shared.TidyAttributes(dev.Attributes)),
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

// DeleteDeveloperByEmail deletes a developer
func (d *Database) DeleteDeveloperByEmail(organizationName, developerEmail string) error {

	developer, err := d.GetDeveloperByEmail(organizationName, developerEmail)
	if err != nil {
		return err
	}

	query := "DELETE FROM developers WHERE developer_id = ?"
	return d.cassandraSession.Query(query, developer.DeveloperID).Exec()
}
