package db

import (
	"encoding/json"
	"fmt"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus label for metrics of db interactions
const appCrendetialsMetricLabel = "appcredentials"

// GetAppCredentialByKey returns details of a single apikey
func (d *Database) GetAppCredentialByKey(organizationName, key string) (types.AppCredential, error) {
	var appcredentials []types.AppCredential

	query := "SELECT * FROM app_credentials WHERE key = ? AND organization_name = ? LIMIT 1"
	appcredentials = d.runGetAppCredentialQuery(query, key, organizationName)
	if len(appcredentials) == 0 {
		d.metricsQueryMiss(appCrendetialsMetricLabel)
		return types.AppCredential{}, fmt.Errorf("Can not find apikey '%s'", key)
	}
	d.metricsQueryHit(appCrendetialsMetricLabel)
	return appcredentials[0], nil
}

// GetAppCredentialByDeveloperAppID returns an array with apikey details of a developer app
func (d *Database) GetAppCredentialByDeveloperAppID(organizationAppID string) ([]types.AppCredential, error) {
	var appcredentials []types.AppCredential

	query := "SELECT * FROM app_credentials WHERE organization_app_id = ?"
	appcredentials = d.runGetAppCredentialQuery(query, organizationAppID)
	if len(appcredentials) == 0 {
		d.metricsQueryMiss(appCrendetialsMetricLabel)
		// Not being able to find a developer is not an error
		return appcredentials, nil
	}
	d.metricsQueryHit(appCrendetialsMetricLabel)
	return appcredentials, nil
}

// GetAppCredentialCountByDeveloperAppID retrieves number of keys beloning to developer app
func (d *Database) GetAppCredentialCountByDeveloperAppID(developerAppID string) int {
	var AppCredentialCount int
	query := "SELECT count(*) FROM app_credentials WHERE organization_app_id = ?"
	if err := d.cassandraSession.Query(query, developerAppID).Scan(&AppCredentialCount); err != nil {
		d.metricsQueryMiss(appCrendetialsMetricLabel)
		return -1
	}
	d.metricsQueryHit(appCrendetialsMetricLabel)
	return AppCredentialCount
}

// runAppCredentialQuery executes CQL query and returns resulset
func (d *Database) runGetAppCredentialQuery(query string, queryParameters ...interface{}) []types.AppCredential {
	var appcredentials []types.AppCredential

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		appcredential := types.AppCredential{
			ConsumerKey:       m["key"].(string),
			AppStatus:         m["app_status"].(string),
			Attributes:        d.unmarshallJSONArrayOfAttributes(m["attributes"].(string), false),
			CompanyStatus:     m["company_status"].(string),
			ConsumerSecret:    m["consumer_secret"].(string),
			CredentialMethod:  m["credential_method"].(string),
			DeveloperStatus:   m["developer_status"].(string),
			ExpiresAt:         m["expires_at"].(int64),
			IssuedAt:          m["issued_at"].(int64),
			OrganizationAppID: m["organization_app_id"].(string),
			OrganizationName:  m["organization_name"].(string),
			Scopes:            m["scopes"].(string),
			Status:            m["status"].(string),
		}
		if m["api_products"].(string) != "" {
			appcredential.APIProducts = make([]types.APIProductStatus, 0)
			json.Unmarshal([]byte(m["api_products"].(string)), &appcredential.APIProducts)
		}
		appcredentials = append(appcredentials, appcredential)
		m = map[string]interface{}{}
	}
	return appcredentials
}

// UpdateAppCredentialByKey UPSERTs appcredentials in database
func (d *Database) UpdateAppCredentialByKey(updatedAppCredential types.AppCredential) error {
	query := "INSERT INTO app_credentials (key,api_products,attributes," +
		"consumer_secret,expires_at,issued_at," +
		"organization_app_id,organization_name,status)" +
		"VALUES(?,?,?,?,?,?,?,?,?)"

	APIProducts := d.marshallArrayOfProductStatusesToJSON(updatedAppCredential.APIProducts)
	Attributes := d.marshallArrayOfAttributesToJSON(updatedAppCredential.Attributes, false)
	if err := d.cassandraSession.Query(query,
		updatedAppCredential.ConsumerKey, APIProducts, Attributes,
		updatedAppCredential.ConsumerSecret, -1, updatedAppCredential.IssuedAt,
		updatedAppCredential.OrganizationAppID, updatedAppCredential.OrganizationName,
		updatedAppCredential.Status).Exec(); err != nil {
		return fmt.Errorf("Can not update appcredential (%v)", err)
	}
	return nil
}

// DeleteAppCredentialByKey deletes a developer
func (d *Database) DeleteAppCredentialByKey(organizationName, consumerKey string) error {
	_, err := d.GetAppCredentialByKey(organizationName, consumerKey)
	if err != nil {
		return err
	}
	query := "DELETE FROM app_credentials WHERE key = ?"
	return d.cassandraSession.Query(query, consumerKey).Exec()
}