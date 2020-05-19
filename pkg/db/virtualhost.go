package db

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/apiauth/pkg/shared"
)

// Prometheus label for metrics of db interactions
const virtualHostMetricLabel = "virtualhosts"

// GetVirtualHosts retrieves all virtualhosts
func (d *Database) GetVirtualHosts() ([]shared.VirtualHost, error) {

	query := "SELECT * FROM virtual_hosts"
	virtualhosts, err := d.runGetVirtualHostQuery(query)
	if err != nil {
		return []shared.VirtualHost{}, err
	}

	if len(virtualhosts) == 0 {
		d.metricsQueryMiss(virtualHostMetricLabel)
		return []shared.VirtualHost{}, errors.New("Can not retrieve list of virtualhosts")
	}

	d.metricsQueryHit(virtualHostMetricLabel)
	return virtualhosts, nil
}

// GetVirtualHostByName retrieves a virtualhost from database
func (d *Database) GetVirtualHostByName(virtualHost string) (shared.VirtualHost, error) {

	query := "SELECT * FROM virtual_hosts WHERE name = ? LIMIT 1"
	virtualhosts, err := d.runGetVirtualHostQuery(query, virtualHost)
	if err != nil {
		return shared.VirtualHost{}, err
	}

	if len(virtualhosts) == 0 {
		d.metricsQueryMiss(virtualHostMetricLabel)
		return shared.VirtualHost{},
			fmt.Errorf("Can not find route (%s)", virtualHost)
	}

	d.metricsQueryHit(virtualHostMetricLabel)
	return virtualhosts[0], nil
}

// runGetVirtualHostQuery executes CQL query and returns resultset
func (d *Database) runGetVirtualHostQuery(query string,
	queryParameters ...interface{}) ([]shared.VirtualHost, error) {

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	var virtualhosts []shared.VirtualHost

	iter := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		newVirtualHost := shared.VirtualHost{
			Name:             m["name"].(string),
			DisplayName:      m["display_name"].(string),
			Port:             m["port"].(int),
			RouteSet:         m["route_set"].(string),
			Policies:         m["policies"].(string),
			OrganizationName: m["organization_name"].(string),
			CreatedAt:        m["created_at"].(int64),
			CreatedBy:        m["created_by"].(string),
			LastmodifiedAt:   m["lastmodified_at"].(int64),
			LastmodifiedBy:   m["lastmodified_by"].(string),
		}
		if m["virtual_hosts"] != nil {
			newVirtualHost.VirtualHosts = d.unmarshallJSONArrayOfStrings(m["virtual_hosts"].(string))
		}
		if m["attributes"] != nil {
			newVirtualHost.Attributes = d.unmarshallJSONArrayOfAttributes(m["attributes"].(string))
		}
		virtualhosts = append(virtualhosts, newVirtualHost)
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		log.Error(err)
		return []shared.VirtualHost{}, err
	}
	return virtualhosts, nil
}

// UpdateVirtualHostByName updates a virtualhost in database
func (d *Database) UpdateVirtualHostByName(updatedVirtualHost *shared.VirtualHost) error {
	query := "INSERT INTO virtual_hosts " +
		"(name, display_name, virtual_hosts, port, route_set, policies, attributes, " +
		"organization_name, created_at, created_by, lastmodified_at, lastmodified_by) " +
		"VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"

	virtualhosts := d.marshallArrayOfStringsToJSON(updatedVirtualHost.VirtualHosts)

	updatedVirtualHost.Attributes = shared.TidyAttributes(updatedVirtualHost.Attributes)
	attributes := d.marshallArrayOfAttributesToJSON(updatedVirtualHost.Attributes)

	updatedVirtualHost.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()
	err := d.cassandraSession.Query(query,
		updatedVirtualHost.Name, updatedVirtualHost.DisplayName,
		virtualhosts, updatedVirtualHost.Port, updatedVirtualHost.RouteSet,
		updatedVirtualHost.Policies, attributes,
		updatedVirtualHost.OrganizationName,
		updatedVirtualHost.CreatedAt, updatedVirtualHost.CreatedBy,
		updatedVirtualHost.LastmodifiedAt,
		updatedVirtualHost.LastmodifiedBy).Exec()
	if err != nil {
		return fmt.Errorf("Can not update virtualhost (%v)", err)
	}
	return nil
}

// DeleteVirtualHostByName deletes a virtualhost
func (d *Database) DeleteVirtualHostByName(virtualHostToDelete string) error {

	_, err := d.GetVirtualHostByName(virtualHostToDelete)
	if err != nil {
		return err
	}

	query := "DELETE FROM virtual_hosts WHERE name = ?"
	return d.cassandraSession.Query(query, virtualHostToDelete).Exec()
}
