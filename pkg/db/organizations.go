package db

import (
	"errors"
	"fmt"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

//Prometheus label for metrics of db interactions
const organizationMetricLabel = "organizations"

//GetOrganizations retrieves all organizations
//
func (d *Database) GetOrganizations() ([]types.Organization, error) {
	query := "SELECT * FROM organizations WHERE key != ? ALLOW FILTERING"
	organizations := d.runGetOrganizationQuery(query, "")
	if len(organizations) == 0 {
		d.metricsQueryMiss(organizationMetricLabel)
		return organizations, errors.New("Can not retrieve list of organizations")
	}
	d.metricsQueryHit(organizationMetricLabel)
	return organizations, nil
}

//GetOrganizationByName retrieves an organization from database
//
func (d *Database) GetOrganizationByName(organizationName string) (types.Organization, error) {
	query := "SELECT * FROM organizations WHERE name = ? LIMIT 1"
	organizations := d.runGetOrganizationQuery(query, organizationName)
	if len(organizations) == 0 {
		d.metricsQueryMiss(organizationMetricLabel)
		return types.Organization{},
			fmt.Errorf("Can not find organization (%s)", organizationName)
	}
	d.metricsQueryHit(organizationMetricLabel)
	return organizations[0], nil
}

// runGetOrganizationQuery executes CQL query and returns resultset
//
func (d *Database) runGetOrganizationQuery(query, queryParameter string) []types.Organization {
	var organizations []types.Organization

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameter).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		organizations = append(organizations, types.Organization{
			Attributes:     d.unmarshallJSONArrayOfAttributes(m["attributes"].(string), false),
			CreatedAt:      m["created_at"].(int64),
			CreatedBy:      m["created_by"].(string),
			DisplayName:    m["display_name"].(string),
			LastmodifiedAt: m["lastmodified_at"].(int64),
			LastmodifiedBy: m["lastmodified_by"].(string),
			Name:           m["name"].(string),
		})
		m = map[string]interface{}{}
	}
	return organizations
}

// UpdateOrganizationByName UPSERTs an organization in database
// Upsert is: In case an organization does not exist (primary key not matching) it will create a new row
func (d *Database) UpdateOrganizationByName(updatedOrganization types.Organization) error {
	query := "INSERT INTO organizations (key, name, display_name, attributes, " +
		"created_at, created_by, lastmodified_at, lastmodified_by) " +
		"VALUES(?,?,?,?,?,?,?,?)"

	Attributes := d.marshallArrayOfAttributesToJSON(updatedOrganization.Attributes, false)
	if err := d.cassandraSession.Query(query,
		updatedOrganization.Name, updatedOrganization.Name,
		updatedOrganization.DisplayName, Attributes, updatedOrganization.CreatedAt,
		updatedOrganization.CreatedBy, updatedOrganization.LastmodifiedAt,
		updatedOrganization.LastmodifiedBy).Exec(); err != nil {
		return fmt.Errorf("Can not update organization (%v)", err)
	}
	return nil
}

//DeleteOrganizationByName deletes an organization
//
func (d *Database) DeleteOrganizationByName(organizationToDelete string) error {
	_, err := d.GetOrganizationByName(organizationToDelete)
	if err != nil {
		return err
	}
	query := "DELETE FROM organizations WHERE key = ?"
	return d.cassandraSession.Query(query, organizationToDelete).Exec()
}
