package db

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const oauthMetricLabel = "oauth"

// OAuthAccessTokenGetByAccess retrieves an access token
func (d *Database) OAuthAccessTokenGetByAccess(accessToken string) (*shared.OAuthAccessToken, error) {

	query := "SELECT * FROM oauth_access_token WHERE access = ? LIMIT 1"

	return d.runGetOAuthAccessTokenQuery(query, accessToken)
}

// OAuthAccessTokenGetByCode retrieves token by code
func (d *Database) OAuthAccessTokenGetByCode(code string) (*shared.OAuthAccessToken, error) {

	query := "SELECT * FROM oauth_access_token WHERE code = ? LIMIT 1"

	return d.runGetOAuthAccessTokenQuery(query, code)
}

// OAuthAccessTokenGetByRefresh retrieves token by refreshcode
func (d *Database) OAuthAccessTokenGetByRefresh(refresh string) (*shared.OAuthAccessToken, error) {

	query := "SELECT * FROM oauth_access_token WHERE refresh = ? LIMIT 1"

	return d.runGetOAuthAccessTokenQuery(query, refresh)
}

// runGetOAuthAccessTokenQuery executes CQL query and returns resultset
func (d *Database) runGetOAuthAccessTokenQuery(query, queryParameter string) (*shared.OAuthAccessToken, error) {

	var accessToken shared.OAuthAccessToken

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameter).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		accessToken = shared.OAuthAccessToken{
			ClientID:         m["client_id"].(string),
			UserID:           m["user_id"].(string),
			RedirectURI:      m["redirect_uri"].(string),
			Scope:            m["scope"].(string),
			Code:             m["code"].(string),
			CodeCreatedAt:    m["code_created_at"].(int64),
			CodeExpiresIn:    m["code_expires_in"].(int64),
			Access:           m["access"].(string),
			AccessCreatedAt:  m["access_created_at"].(int64),
			AccessExpiresIn:  m["access_expires_in"].(int64),
			Refresh:          m["refresh"].(string),
			RefreshCreatedAt: m["refresh_created_at"].(int64),
			RefreshExpiresIn: m["refresh_expires_in"].(int64),
		}
	}
	if err := iterable.Close(); err != nil {
		log.Error(err)
		d.metricsQueryMiss(organizationMetricLabel)
		return nil, err
	}
	d.metricsQueryHit(oauthMetricLabel)

	log.Printf("runGetOAuthAccessTokenQuery: %+v", accessToken)
	return &accessToken, nil
}

// OAuthAccessTokenCreate UPSERTs an organization in database
func (d *Database) OAuthAccessTokenCreate(t *shared.OAuthAccessToken) error {

	// log.Printf("OAuthAccessTokenCreate: %+v", t)
	if err := d.cassandraSession.Query(`INSERT INTO oauth_access_token (
client_id,
user_id,
redirect_uri,
scope,
code,
code_created_at,
code_expires_in,
access,
access_created_at,
access_expires_in,
refresh,
refresh_created_at,
refresh_expires_in) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,

		t.ClientID,
		t.UserID,
		t.RedirectURI,
		t.Scope,
		t.Code,
		t.CodeCreatedAt,
		t.CodeExpiresIn,
		t.Access,
		t.AccessCreatedAt,
		t.AccessExpiresIn,
		t.Refresh,
		t.RefreshCreatedAt,
		t.RefreshExpiresIn).Exec(); err != nil {

		return fmt.Errorf("Cannot update access token '%s' (%v)", t.Access, err)
	}
	return nil
}

// OAuthAccessTokenRemoveByAccess deletes an access token
func (d *Database) OAuthAccessTokenRemoveByAccess(accessTokenToDelete *string) error {

	query := "DELETE FROM oauth_access_token WHERE access = ?"

	return d.cassandraSession.Query(query, accessTokenToDelete).Exec()
}

// OAuthAccessTokenRemoveByCode deletes an access token
func (d *Database) OAuthAccessTokenRemoveByCode(codeToDelete *string) error {

	query := "DELETE FROM oauth_access_token WHERE code = ?"

	return d.cassandraSession.Query(query, codeToDelete).Exec()
}

// OAuthAccessTokenRemoveByRefresh deletes an access token
func (d *Database) OAuthAccessTokenRemoveByRefresh(refreshToDelete *string) error {

	query := "DELETE FROM oauth_access_token WHERE refresh = ?"

	return d.cassandraSession.Query(query, refreshToDelete).Exec()
}
