package db

import (
	"encoding/json"
	"fmt"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

//Prometheus label for metrics of db interactions
const appCrendetialsMetricLabel = "appcredentials"

//GetAppCredentialByKey returns details of a single apikey
//
func (d *Database) GetAppCredentialByKey(key string) (types.AppCredential, error) {
	var appcredentials []types.AppCredential

	query := "SELECT * FROM app_credentials WHERE key = ? LIMIT 1"
	appcredentials = d.runGetAppCredentialQuery(query, key)
	if len(appcredentials) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
		return appcredentials[0], nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
	return types.AppCredential{}, fmt.Errorf("Could not find apikey '%s'", key)
}

//GetAppCredentialByDeveloperAppID returns an array with apikey details of a developer app
// FIXME contains LIMIT
func (d *Database) GetAppCredentialByDeveloperAppID(organizationAppID string) ([]types.AppCredential, error) {
	var appcredentials []types.AppCredential

	// FIXME hardcoded row limit
	query := "SELECT * FROM app_credentials WHERE organization_app_id = ?"
	appcredentials = d.runGetAppCredentialQuery(query, organizationAppID)
	if len(appcredentials) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
		return appcredentials, nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
	return appcredentials, fmt.Errorf("Could not find apikeys of developer app '%s'", organizationAppID)
}

//GetAppCredentialCountByDeveloperAppID retrieves number of keys beloning to developer app
//
func (d *Database) GetAppCredentialCountByDeveloperAppID(developerAppID string) int {
	var AppCredentialCount int
	query := "SELECT count(*) FROM app_credentials WHERE organization_app_id = ?"
	if err := d.cassandraSession.Query(query, developerAppID).Scan(&AppCredentialCount); err != nil {
		return -1
	}
	return AppCredentialCount
}

//runAppCredentialQuery executes CQL query and returns resulset
//
func (d *Database) runGetAppCredentialQuery(query, queryParameter string) []types.AppCredential {
	var appcredentials []types.AppCredential

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameter).Iter()
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
// Upsert is: In case a developer does not exist (primary key not matching) it will create a new row
func (d *Database) UpdateAppCredentialByKey(updatedAppCredential types.AppCredential) error {
	query := "INSERT INTO app_credentials (key,api_products,attributes," +
		"consumer_secret,expires_at,issued_at," +
		"organization_app_id,organization_name,status)" +
		"VALUES(?,?,?,?,?,?,?,?,?)"

	APIProducts := d.marshallArrayOfProductStatusesToJSON(updatedAppCredential.APIProducts)
	Attributes := d.marshallArrayOfAttributesToJSON(updatedAppCredential.Attributes, false)
	err := d.cassandraSession.Query(query,
		updatedAppCredential.ConsumerKey, APIProducts, Attributes,
		updatedAppCredential.ConsumerSecret, -1, updatedAppCredential.IssuedAt,
		updatedAppCredential.OrganizationAppID, updatedAppCredential.OrganizationName,
		updatedAppCredential.Status).Exec()
	if err == nil {
		return nil
	}
	return fmt.Errorf("Could not update appcredential (%v)", err)
}

//DeleteAppCredentialByKey deletes a developer
//
func (d *Database) DeleteAppCredentialByKey(consumerKey string) error {
	_, err := d.GetAppCredentialByKey(consumerKey)
	if err != nil {
		return err
	}
	query := "DELETE FROM app_credentials WHERE key = ?"
	return d.cassandraSession.Query(query, consumerKey).Exec()
}
