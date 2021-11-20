package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (

	// Prometheus label for metrics of db interactions
	keysMetricLabel = "keys"

	// List of key columns we use
	keysColumn = `consumer_key,
consumer_secret,
scopes,
api_products,
attributes,
app_id,
status,
issued_at,
expires_at`
)

// KeyStore holds our database config
type KeyStore struct {
	db *Database
}

// NewKeyStore creates key instance
func NewKeyStore(database *Database) *KeyStore {
	return &KeyStore{
		db: database,
	}
}

// GetByKey returns details of a single apikey
func (s *KeyStore) GetByKey(organization, key *string) (*types.Key, types.Error) {

	var keys types.Keys
	var err error

	query := "SELECT " + keysColumn + " FROM keys WHERE consumer_key = ? LIMIT 1"
	keys, err = s.runGetKeyQuery(query, key)
	if err != nil {
		s.db.metrics.QueryFailed(keysMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(keys) == 0 {
		s.db.metrics.QueryNotFound(keysMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("can not find apikey '%s'", *key))
	}

	s.db.metrics.QuerySuccessful(keysMetricLabel)
	return &keys[0], nil
}

// GetByDeveloperAppID returns an array with apikey details of a developer app
func (s *KeyStore) GetByDeveloperAppID(organization, developerAppID string) (types.Keys, types.Error) {

	query := "SELECT " + keysColumn + " FROM keys WHERE app_id = ?"
	keys, err := s.runGetKeyQuery(query, developerAppID)
	if err != nil {
		s.db.metrics.QueryFailed(keysMetricLabel)
		return nil, types.NewDatabaseError(err)
	}
	s.db.metrics.QuerySuccessful(keysMetricLabel)
	return keys, nil
}

// GetCountByAPIProductName counts the number of times an apiproduct has been assigned to keys
func (s *KeyStore) GetCountByAPIProductName(organization, apiProductName string) (int, types.Error) {

	query := "SELECT api_products FROM keys"
	keys, err := s.runGetKeyQuery(query)
	if err != nil {
		s.db.metrics.QueryFailed(keysMetricLabel)
		return 0, types.NewDatabaseError(err)
	}
	var count int
	for _, key := range keys {
		for _, product := range key.APIProducts {
			if product.Apiproduct == apiProductName {
				count++
			}
		}
	}
	return count, nil
}

// runGetKeyQuery executes CQL query and returns resultset
func (s *KeyStore) runGetKeyQuery(query string, queryParameters ...interface{}) (types.Keys, error) {

	var keys types.Keys

	timer := prometheus.NewTimer(s.db.metrics.queryHistogram)
	defer timer.ObserveDuration()

	iterable := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		keys = append(keys, types.Key{
			ConsumerKey:    columnToString(m, "consumer_key"),
			ConsumerSecret: columnToString(m, "consumer_secret"),
			Scopes:         columnToStringSlice(m, "scopes"),
			APIProducts:    KeyAPIProductStatusesUnmarshal(m["api_products"].(string)),
			Attributes:     columnToAttributes(m, "attributes"),
			AppID:          columnToString(m, "app_id"),
			Status:         columnToString(m, "status"),
			IssuedAt:       columnToInt64(m, "issued_at"),
			ExpiresAt:      columnToInt64(m, "expires_at"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iterable.Close(); err != nil {
		return nil, err
	}
	return keys, nil
}

// UpdateByKey UPSERTs keys in database
func (s *KeyStore) UpdateByKey(organization string, k *types.Key) types.Error {

	query := "INSERT INTO keys (" + keysColumn + ") VALUES(?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		k.ConsumerKey,
		k.ConsumerSecret,
		k.Scopes,
		KeyAPIProductStatusesMarshal(k.APIProducts),
		attributesToColumn(k.Attributes),
		k.AppID,
		k.Status,
		k.IssuedAt,
		k.ExpiresAt).Exec(); err != nil {

		s.db.metrics.QueryFailed(keysMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update key '%s' (%s)", k.ConsumerKey, err))
	}
	return nil
}

// DeleteByKey deletes keys
func (s *KeyStore) DeleteByKey(organization, consumerKey string) types.Error {

	query := "DELETE FROM keys WHERE consumer_key = ?"
	if err := s.db.CassandraSession.Query(query, consumerKey).Exec(); err != nil {
		s.db.metrics.QueryFailed(keysMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}

func (s *KeyStore) Metric(what string) {
	s.db.metrics.QueryFailed(keysMetricLabel)
}
