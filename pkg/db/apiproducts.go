package db

import (
	"fmt"
	"log"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus label for metrics of db interactions
const apiProductsMetricLabel = "apiproducts"

// GetAPIProductsByOrganization retrieves all api products belonging to an organization
func (d *Database) GetAPIProductsByOrganization(organizationName string) ([]types.APIProduct, error) {
	query := "SELECT * FROM api_products WHERE organization_name = ? ALLOW FILTERING"
	apiproducts := d.runGetAPIProductQuery(query, organizationName)
	if len(apiproducts) == 0 {
		d.metricsQueryMiss(appsMetricLabel)
		return apiproducts,
			fmt.Errorf("Can not find developers in organization %s", organizationName)
	}
	d.metricsQueryHit(appsMetricLabel)
	return apiproducts, nil
}

// GetAPIProductByName returns an apiproduct
func (d *Database) GetAPIProductByName(organizationName, apiproductName string) (types.APIProduct, error) {
	query := "SELECT * FROM api_products WHERE organization_name = ? AND name = ? LIMIT 1"
	log.Printf("%s - %s", organizationName, apiproductName)
	apiproducts := d.runGetAPIProductQuery(query, organizationName, apiproductName)
	if len(apiproducts) == 0 {
		d.metricsQueryMiss(apiProductsMetricLabel)
		return types.APIProduct{},
			fmt.Errorf("Could not find apiproduct (%s)", apiproductName)
	}
	d.metricsQueryHit(apiProductsMetricLabel)
	return apiproducts[0], nil
}

// runAPIProductQuery executes CQL query and returns resultset
func (d *Database) runGetAPIProductQuery(query string, queryParameters ...interface{}) []types.APIProduct {
	var apiproducts []types.APIProduct

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		apiproduct := types.APIProduct{
			Key:              m["key"].(string),
			ApprovalType:     m["approval_type"].(string),
			CreatedAt:        m["created_at"].(int64),
			CreatedBy:        m["created_by"].(string),
			Description:      m["description"].(string),
			DisplayName:      m["display_name"].(string),
			Environments:     m["environments"].(string),
			LastmodifiedAt:   m["lastmodified_at"].(int64),
			LastmodifiedBy:   m["lastmodified_by"].(string),
			Name:             m["name"].(string),
			OrganizationName: m["organization_name"].(string),
			Scopes:           m["scopes"].(string),
		}
		apiproduct.APIResources = d.unmarshallJSONArrayOfStrings(m["api_resources"].(string))
		apiproduct.Proxies = d.unmarshallJSONArrayOfStrings(m["proxies"].(string))
		apiproduct.Attributes = d.unmarshallJSONArrayOfAttributes(m["attributes"].(string), true)
		apiproducts = append(apiproducts, apiproduct)
		m = map[string]interface{}{}
	}
	return apiproducts
}

// UpdateAPIProductByName UPSERTs an apiproduct in database
func (d *Database) UpdateAPIProductByName(updatedAPIProduct types.APIProduct) error {
	query := "INSERT INTO api_products (key,name,display_name, attributes," +
		"created_at,created_by, api_resources," +
		"lastmodified_at,lastmodified_by,organization_name) " +
		"VALUES(?,?,?,?,?, ?,?,?,?,?)"

	Attributes := d.marshallArrayOfAttributesToJSON(updatedAPIProduct.Attributes, false)
	err := d.cassandraSession.Query(query,
		updatedAPIProduct.Key, updatedAPIProduct.Name, updatedAPIProduct.DisplayName, Attributes,
		updatedAPIProduct.CreatedAt, updatedAPIProduct.CreatedBy, updatedAPIProduct.APIResources,
		updatedAPIProduct.LastmodifiedAt, updatedAPIProduct.LastmodifiedBy,
		updatedAPIProduct.OrganizationName).Exec()
	if err == nil {
		return nil
	}
	return fmt.Errorf("Can not update api product (%v)", err)
}

// DeleteAPIProductByName deletes an apiproduct
func (d *Database) DeleteAPIProductByName(organizationName, developerEmail string) error {
	apiproduct, err := d.GetAPIProductByName(organizationName, developerEmail)
	if err != nil {
		return err
	}
	query := "DELETE FROM api_products WHERE key = ?"
	return d.cassandraSession.Query(query, apiproduct.Key).Exec()
}
