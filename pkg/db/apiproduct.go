package db

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const apiProductsMetricLabel = "apiproducts"

// GetAPIProductsByOrganization retrieves all api products belonging to an organization
func (d *Database) GetAPIProductsByOrganization(organizationName string) ([]shared.APIProduct, error) {
	query := "SELECT * FROM api_products WHERE organization_name = ? ALLOW FILTERING"
	apiproducts, err := d.runGetAPIProductQuery(query, organizationName)
	if err != nil {
		return []shared.APIProduct{}, err
	}
	if len(apiproducts) == 0 {
		d.metricsQueryMiss(appsMetricLabel)
		return apiproducts,
			fmt.Errorf("Can not find apiproducts in organization %s", organizationName)
	}
	d.metricsQueryHit(appsMetricLabel)
	return apiproducts, nil
}

// GetAPIProductByName returns an apiproduct
func (d *Database) GetAPIProductByName(organizationName, apiproductName string) (shared.APIProduct, error) {
	query := "SELECT * FROM api_products WHERE organization_name = ? AND name = ? LIMIT 1"
	apiproducts, err := d.runGetAPIProductQuery(query, organizationName, apiproductName)
	if err != nil {
		return shared.APIProduct{}, err
	}
	if len(apiproducts) == 0 {
		d.metricsQueryMiss(apiProductsMetricLabel)
		return shared.APIProduct{},
			fmt.Errorf("Could not find apiproduct (%s)", apiproductName)
	}
	d.metricsQueryHit(apiProductsMetricLabel)
	return apiproducts[0], nil
}

// runAPIProductQuery executes CQL query and returns resultset
func (d *Database) runGetAPIProductQuery(query string, queryParameters ...interface{}) ([]shared.APIProduct, error) {
	var apiproducts []shared.APIProduct

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		apiproduct := shared.APIProduct{
			Name:             m["name"].(string),
			DisplayName:      m["display_name"].(string),
			Description:      m["description"].(string),
			RouteSet:         m["route_set"].(string),
			OrganizationName: m["organization_name"].(string),
			Policies:         m["policies"].(string),
			CreatedAt:        m["created_at"].(int64),
			CreatedBy:        m["created_by"].(string),
			LastmodifiedAt:   m["lastmodified_at"].(int64),
			LastmodifiedBy:   m["lastmodified_by"].(string),
		}
		apiproduct.Paths = d.unmarshallJSONArrayOfStrings(m["paths"].(string))
		apiproduct.Attributes = d.unmarshallJSONArrayOfAttributes(m["attributes"].(string))
		apiproducts = append(apiproducts, apiproduct)
		m = map[string]interface{}{}
	}
	if err := iterable.Close(); err != nil {
		log.Error(err)
		return []shared.APIProduct{}, err
	}
	return apiproducts, nil
}

// UpdateAPIProductByName UPSERTs an apiproduct in database
func (d *Database) UpdateAPIProductByName(updatedAPIProduct *shared.APIProduct) error {
	query := "INSERT INTO api_products (name,display_name, attributes," +
		"created_at,created_by, route_set, paths, policies, " +
		"lastmodified_at,lastmodified_by,organization_name) " +
		"VALUES(?,?,?,?,?,?,?,?,?,?,?)"

	updatedAPIProduct.Attributes = shared.TidyAttributes(updatedAPIProduct.Attributes)
	attributes := d.marshallArrayOfAttributesToJSON(updatedAPIProduct.Attributes)

	paths := d.marshallArrayOfStringsToJSON(updatedAPIProduct.Paths)

	updatedAPIProduct.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()
	err := d.cassandraSession.Query(query,
		updatedAPIProduct.Name, updatedAPIProduct.DisplayName, attributes,
		updatedAPIProduct.CreatedAt, updatedAPIProduct.CreatedBy, updatedAPIProduct.RouteSet,
		paths, updatedAPIProduct.Policies,
		updatedAPIProduct.LastmodifiedAt, updatedAPIProduct.LastmodifiedBy,
		updatedAPIProduct.OrganizationName).Exec()
	if err == nil {
		return nil
	}
	return fmt.Errorf("Can not update apiproduct (%v)", err)
}

// DeleteAPIProductByName deletes an apiproduct
func (d *Database) DeleteAPIProductByName(organizationName, apiProduct string) error {
	apiproduct, err := d.GetAPIProductByName(organizationName, apiProduct)
	if err != nil {
		return err
	}
	query := "DELETE FROM api_products WHERE name = ?"
	return d.cassandraSession.Query(query, apiproduct.Name).Exec()
}
