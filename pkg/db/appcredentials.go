package db

import (
	"encoding/json"
	"errors"

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
	return types.AppCredential{}, errors.New("Could not find apikey")
}

//GetAppCredentialByDeveloperAppID returns an array with apikey details of a developer app
// FIXME contains LIMIT
func (d *Database) GetAppCredentialByDeveloperAppID(organizationAppID string) ([]types.AppCredential, error) {
	var appcredentials []types.AppCredential

	// FIXME hardcoded row limit
	query := "SELECT * FROM app_credentials WHERE organization_app_id = ? LIMIT 1000"
	appcredentials = d.runGetAppCredentialQuery(query, organizationAppID)
	if len(appcredentials) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
		return appcredentials, nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
	return appcredentials, errors.New("Could not find apikeys of developer app")
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
			Attributes:        m["attributes"].(string),
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
