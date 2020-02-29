package db

import (
	"errors"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

//GetAPIProductByName returns an array with apiproduct details of a developer app
//
func (d *Database) GetAPIProductByName(apiproductname string) (types.APIProduct, error) {
	query := "SELECT * FROM api_products WHERE name = ? LIMIT 1"
	apiproducts := d.runGetAPIProductQuery(query, apiproductname)
	if len(apiproducts) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
		return apiproducts[0], nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
	return types.APIProduct{}, errors.New("Could not find apikeys of developer app")
}

// runAPIProductQuery executes CQL query and returns resultset
//
func (d *Database) runGetAPIProductQuery(query, queryParameter string) []types.APIProduct {
	var apiproducts []types.APIProduct

	//Set timer to record how long this function run
	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameter).Iter()
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
