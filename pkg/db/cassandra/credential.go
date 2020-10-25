package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (

	// Prometheus label for metrics of db interactions
	appCredentialsMetricLabel = "credentials"

	// List of credential columns we use
	appCredentialsColumn = `consumer_key,
consumer_secret,
api_products,
attributes,
app_id,
status,
issued_at,
expires_at`
)

// CredentialStore holds our database config
type CredentialStore struct {
	db *Database
}

// NewCredentialStore creates credential instance
func NewCredentialStore(database *Database) *CredentialStore {
	return &CredentialStore{
		db: database,
	}
}

// GetByKey returns details of a single apikey
func (s *CredentialStore) GetByKey(key *string) (*types.DeveloperAppKey, types.Error) {

	var appcredentials types.DeveloperAppKeys
	var err error

	query := "SELECT " + appCredentialsColumn + " FROM credentials WHERE consumer_key = ? LIMIT 1"
	appcredentials, err = s.runGetAppCredentialQuery(query, key)
	if err != nil {
		s.db.metrics.QueryFailed(appCredentialsMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(appcredentials) == 0 {
		s.db.metrics.QueryMiss(appCredentialsMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("Can not find apikey '%s'", *key))
	}

	s.db.metrics.QueryHit(appCredentialsMetricLabel)
	return &appcredentials[0], nil
}

// GetByDeveloperAppID returns an array with apikey details of a developer app
func (s *CredentialStore) GetByDeveloperAppID(developerAppID string) (types.DeveloperAppKeys, types.Error) {

	query := "SELECT " + appCredentialsColumn + " FROM credentials WHERE app_id = ?"
	appcredentials, err := s.runGetAppCredentialQuery(query, developerAppID)
	if err != nil {
		s.db.metrics.QueryFailed(appCredentialsMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(appcredentials) == 0 {
		s.db.metrics.QueryMiss(appCredentialsMetricLabel)
		// Not being able to find a developer is not an error
		return appcredentials, types.NewItemNotFoundError(
			fmt.Errorf("Can not find api keys of developer app id '%s'", developerAppID))
	}

	s.db.metrics.QueryHit(appCredentialsMetricLabel)
	return appcredentials, nil
}

// runAppCredentialQuery executes CQL query and returns resulset
func (s *CredentialStore) runGetAppCredentialQuery(query string, queryParameters ...interface{}) (types.DeveloperAppKeys, error) {

	var appcredentials types.DeveloperAppKeys

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	iterable := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		appcredentials = append(appcredentials, types.DeveloperAppKey{
			ConsumerKey:    columnValueString(m, "consumer_key"),
			ConsumerSecret: columnValueString(m, "consumer_secret"),
			APIProducts:    types.DeveloperAppKey{}.APIProducts.Unmarshal(m["api_products"].(string)),
			Attributes:     types.DeveloperAppKey{}.Attributes.Unmarshal(m["attributes"].(string)),
			AppID:          columnValueString(m, "app_id"),
			Status:         columnValueString(m, "status"),
			IssuedAt:       columnValueInt64(m, "issued_at"),
			ExpiresAt:      columnValueInt64(m, "expires_at"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iterable.Close(); err != nil {
		return nil, err
	}
	return appcredentials, nil
}

// UpdateByKey UPSERTs credentials in database
func (s *CredentialStore) UpdateByKey(c *types.DeveloperAppKey) types.Error {

	query := "INSERT INTO credentials (" + appCredentialsColumn + ") VALUES(?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		c.ConsumerKey,
		c.ConsumerSecret,
		c.APIProducts.Marshal(),
		c.Attributes.Marshal(),
		c.AppID,
		c.Status,
		c.IssuedAt,
		c.ExpiresAt).Exec(); err != nil {

		s.db.metrics.QueryFailed(appCredentialsMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("Cannot update credential '%s'", c.ConsumerKey))
	}
	return nil
}

// DeleteByKey deletes credentials
func (s *CredentialStore) DeleteByKey(consumerKey string) types.Error {

	query := "DELETE FROM credentials WHERE consumer_key = ?"
	if err := s.db.CassandraSession.Query(query, consumerKey).Exec(); err != nil {
		s.db.metrics.QueryFailed(appCredentialsMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
