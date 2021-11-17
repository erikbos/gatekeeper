package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// Prometheus label for metrics of db interactions
	oauthMetricLabel = "oauth"
	// Default TTL for OAuth token row, Cassandra will expire row afer this period
	defaultOAuthtokenTTL = 86400

	// List of columns we use
	oauthColumns = `client_id,
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
refresh_expires_in`
)

// OAuthStore holds our database config
type OAuthStore struct {
	db *Database
}

// NewOAuthStore creates oauth instance
func NewOAuthStore(database *Database) *OAuthStore {
	return &OAuthStore{
		db: database,
	}
}

// OAuthAccessTokenGetByAccess retrieves an access token
func (s *OAuthStore) OAuthAccessTokenGetByAccess(accessToken string) (*types.OAuthAccessToken, error) {

	query := "SELECT " + oauthColumns + " FROM oauth_access_token WHERE access = ? LIMIT 1"
	return s.runGetOAuthAccessTokenQuery(query, accessToken)
}

// OAuthAccessTokenGetByCode retrieves token by code
func (s *OAuthStore) OAuthAccessTokenGetByCode(code string) (*types.OAuthAccessToken, error) {

	query := "SELECT " + oauthColumns + " FROM oauth_access_token WHERE code = ? LIMIT 1"
	return s.runGetOAuthAccessTokenQuery(query, code)
}

// OAuthAccessTokenGetByRefresh retrieves token by refreshcode
func (s *OAuthStore) OAuthAccessTokenGetByRefresh(refresh string) (*types.OAuthAccessToken, error) {

	query := "SELECT " + oauthColumns + " FROM oauth_access_token WHERE refresh = ? LIMIT 1"
	return s.runGetOAuthAccessTokenQuery(query, refresh)
}

// runGetOAuthAccessTokenQuery executes CQL query and returns resultset
func (s *OAuthStore) runGetOAuthAccessTokenQuery(query, queryParameter string) (*types.OAuthAccessToken, error) {

	var accessToken types.OAuthAccessToken

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	iterable := s.db.CassandraSession.Query(query, queryParameter).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		accessToken = types.OAuthAccessToken{
			ClientID:         columnValueString(m, "client_id"),
			UserID:           columnValueString(m, "user_id"),
			RedirectURI:      columnValueString(m, "redirect_uri"),
			Scope:            columnValueString(m, "scope"),
			Code:             columnValueString(m, "code"),
			CodeCreatedAt:    columnValueInt64(m, "code_created_at"),
			CodeExpiresIn:    columnValueInt64(m, "code_expires_in"),
			Access:           columnValueString(m, "access"),
			AccessCreatedAt:  columnValueInt64(m, "access_created_at"),
			AccessExpiresIn:  columnValueInt64(m, "access_expires_in"),
			Refresh:          columnValueString(m, "refresh"),
			RefreshCreatedAt: columnValueInt64(m, "refresh_created_at"),
			RefreshExpiresIn: columnValueInt64(m, "refresh_expires_in"),
		}
	}
	if err := iterable.Close(); err != nil {
		s.db.metrics.QueryMiss(oauthMetricLabel)
		return nil, err
	}
	s.db.metrics.QueryHit(oauthMetricLabel)

	// log.Printf("runGetOAuthAccessTokenQuery: %+v", accessToken)
	return &accessToken, nil
}

// OAuthAccessTokenCreate UPSERTs a token in database
func (s *OAuthStore) OAuthAccessTokenCreate(t *types.OAuthAccessToken) error {

	// "USING TTL %d" is used to give each inserted token row a time-to-live,
	// this will force the database to expire the row.
	//
	// OAuth packages will check CreatedAt + ExpiresIn to check validity of a retrieved token,
	/// but does not actively delete from database.
	query := fmt.Sprintf("INSERT INTO oauth_access_token ("+oauthColumns+
		") VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?) USING TTL %d", defaultOAuthtokenTTL)
	if err := s.db.CassandraSession.Query(query,
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

		return fmt.Errorf("cannot update access token '%s' (%s)", t.Access, err)
	}
	return nil
}

// OAuthAccessTokenRemoveByAccess deletes an access token
func (s *OAuthStore) OAuthAccessTokenRemoveByAccess(accessTokenToDelete string) error {

	query := "DELETE FROM oauth_access_token WHERE access = ?"

	return s.db.CassandraSession.Query(query, accessTokenToDelete).Exec()
}

// OAuthAccessTokenRemoveByCode deletes an access token
func (s *OAuthStore) OAuthAccessTokenRemoveByCode(codeToDelete string) error {

	query := "DELETE FROM oauth_access_token WHERE code = ?"

	return s.db.CassandraSession.Query(query, codeToDelete).Exec()
}

// OAuthAccessTokenRemoveByRefresh deletes an access token
func (s *OAuthStore) OAuthAccessTokenRemoveByRefresh(refreshToDelete string) error {

	query := "DELETE FROM oauth_access_token WHERE refresh = ?"

	return s.db.CassandraSession.Query(query, refreshToDelete).Exec()
}
