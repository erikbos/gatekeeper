package db

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/apiauth/pkg/shared"
)

// Prometheus label for metrics of db interactions
const appCredentialsMetricLabel = "appcredentials"

// GetAppCredentialByKey returns details of a single apikey
func (d *Database) GetAppCredentialByKey(organizationName, key string) (shared.AppCredential, error) {
	query := "SELECT * FROM app_credentials WHERE key = ? AND organization_name = ? LIMIT 1"
	appcredentials, err := d.runGetAppCredentialQuery(query, key, organizationName)
	if err != nil {
		return shared.AppCredential{}, err
	}
	if len(appcredentials) == 0 {
		d.metricsQueryMiss(appCredentialsMetricLabel)
		return shared.AppCredential{}, fmt.Errorf("Can not find apikey '%s'", key)
	}
	d.metricsQueryHit(appCredentialsMetricLabel)
	return appcredentials[0], nil
}

// GetAppCredentialByDeveloperAppID returns an array with apikey details of a developer app
func (d *Database) GetAppCredentialByDeveloperAppID(organizationAppID string) ([]shared.AppCredential, error) {
	query := "SELECT * FROM app_credentials WHERE organization_app_id = ?"
	appcredentials, err := d.runGetAppCredentialQuery(query, organizationAppID)
	if err != nil {
		return []shared.AppCredential{}, err
	}
	if len(appcredentials) == 0 {
		d.metricsQueryMiss(appCredentialsMetricLabel)
		// Not being able to find a developer is not an error
		return appcredentials, nil
	}
	d.metricsQueryHit(appCredentialsMetricLabel)
	return appcredentials, nil
}

// GetAppCredentialCountByDeveloperAppID retrieves number of keys beloning to developer app
func (d *Database) GetAppCredentialCountByDeveloperAppID(developerAppID string) int {
	var AppCredentialCount int
	query := "SELECT count(*) FROM app_credentials WHERE organization_app_id = ?"
	if err := d.cassandraSession.Query(query, developerAppID).Scan(&AppCredentialCount); err != nil {
		d.metricsQueryMiss(appCredentialsMetricLabel)
		return -1
	}
	d.metricsQueryHit(appCredentialsMetricLabel)
	return AppCredentialCount
}

// runAppCredentialQuery executes CQL query and returns resulset
func (d *Database) runGetAppCredentialQuery(query string, queryParameters ...interface{}) ([]shared.AppCredential, error) {
	var appcredentials []shared.AppCredential

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		appcredential := shared.AppCredential{
			ConsumerKey:       m["key"].(string),
			AppStatus:         m["app_status"].(string),
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
		if m["attributes"] != nil {
			appcredential.Attributes = d.unmarshallJSONArrayOfAttributes(m["attributes"].(string))
		}
		if m["api_products"] != nil {
			appcredential.APIProducts = d.unmarshallJSONArrayOfProductStatuses(m["api_products"].(string))
		}

		appcredentials = append(appcredentials, appcredential)
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iterable.Close(); err != nil {
		log.Error(err)
		return []shared.AppCredential{}, err
	}
	return appcredentials, nil
}

// UpdateAppCredentialByKey UPSERTs appcredentials in database
func (d *Database) UpdateAppCredentialByKey(updatedAppCredential *shared.AppCredential) error {

	APIProducts := d.marshallArrayOfProductStatusesToJSON(updatedAppCredential.APIProducts)

	updatedAppCredential.Attributes = shared.TidyAttributes(updatedAppCredential.Attributes)
	Attributes := d.marshallArrayOfAttributesToJSON(updatedAppCredential.Attributes)

	if err := d.cassandraSession.Query(
		"INSERT INTO app_credentials (key, api_products, attributes, "+
			"consumer_secret, expires_at, issued_at,"+
			"organization_app_id, organization_name, status)"+
			"VALUES(?,?,?,?,?,?,?,?,?)",
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
