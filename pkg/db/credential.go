package db

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const appCredentialsMetricLabel = "credentials"

// GetAppCredentialByKey returns details of a single apikey
func (d *Database) GetAppCredentialByKey(organizationName string, key *string) (*shared.DeveloperAppKey, error) {

	query := "SELECT * FROM credentials WHERE consumer_key = ? AND organization_name = ? LIMIT 1 ALLOW FILTERING"
	appcredentials, err := d.runGetAppCredentialQuery(query, key, organizationName)
	if err != nil {
		return &shared.DeveloperAppKey{}, err
	}

	if len(appcredentials) == 0 {
		d.metricsQueryMiss(appCredentialsMetricLabel)
		return nil, fmt.Errorf("Can not find apikey '%s'", *key)
	}

	d.metricsQueryHit(appCredentialsMetricLabel)
	return &appcredentials[0], nil
}

// GetAppCredentialByDeveloperAppID returns an array with apikey details of a developer app
func (d *Database) GetAppCredentialByDeveloperAppID(developerAppID string) ([]shared.DeveloperAppKey, error) {

	query := "SELECT * FROM credentials WHERE app_id = ?"
	appcredentials, err := d.runGetAppCredentialQuery(query, developerAppID)
	if err != nil {
		return []shared.DeveloperAppKey{}, err
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

	query := "SELECT count(*) FROM credentials WHERE app_id = ?"
	if err := d.cassandraSession.Query(query, developerAppID).Scan(&AppCredentialCount); err != nil {
		d.metricsQueryMiss(appCredentialsMetricLabel)
		return -1
	}

	d.metricsQueryHit(appCredentialsMetricLabel)
	return AppCredentialCount
}

// runAppCredentialQuery executes CQL query and returns resulset
func (d *Database) runGetAppCredentialQuery(query string, queryParameters ...interface{}) ([]shared.DeveloperAppKey, error) {

	var appcredentials []shared.DeveloperAppKey

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		appcredential := shared.DeveloperAppKey{
			ConsumerKey:      m["consumer_key"].(string),
			ConsumerSecret:   m["consumer_secret"].(string),
			ExpiresAt:        m["expires_at"].(int64),
			IssuedAt:         m["issued_at"].(int64),
			AppID:            m["app_id"].(string),
			OrganizationName: m["organization_name"].(string),
			Status:           m["status"].(string),
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
		return []shared.DeveloperAppKey{}, err
	}
	return appcredentials, nil
}

// UpdateAppCredentialByKey UPSERTs appcredentials in database
func (d *Database) UpdateAppCredentialByKey(c *shared.DeveloperAppKey) error {

	c.Attributes = shared.TidyAttributes(c.Attributes)

	if err := d.cassandraSession.Query(`INSERT INTO credentials (
consumer_key,
consumer_secret,
api_products,
attributes,
app_id,
organization_name,
status,
issued_at,
expires_at) VALUES(?,?,?,?,?,?,?,?,?)`,

		c.ConsumerKey,
		c.ConsumerSecret,
		d.marshallArrayOfProductStatusesToJSON(c.APIProducts),
		d.marshallArrayOfAttributesToJSON(c.Attributes),
		c.AppID,
		c.OrganizationName,
		c.Status,
		c.IssuedAt,
		c.ExpiresAt).Exec(); err != nil {

		return fmt.Errorf("Can not update credential '%s', (%v)", c.ConsumerKey, err)
	}
	return nil
}

// DeleteAppCredentialByKey deletes a developer
func (d *Database) DeleteAppCredentialByKey(organizationName, consumerKey string) error {

	_, err := d.GetAppCredentialByKey(organizationName, &consumerKey)
	if err != nil {
		return err
	}

	query := "DELETE FROM credentials WHERE consumer_key = ?"
	return d.cassandraSession.Query(query, consumerKey).Exec()
}
