package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const appCredentialsMetricLabel = "credentials"

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
func (s *CredentialStore) GetByKey(organizationName, key *string) (*shared.DeveloperAppKey, error) {

	var appcredentials []shared.DeveloperAppKey
	var err error

	if organizationName == nil {
		query := "SELECT * FROM credentials WHERE consumer_key = ? LIMIT 1"
		appcredentials, err = s.runGetAppCredentialQuery(query, key)
	} else {
		query := "SELECT * FROM credentials WHERE consumer_key = ? AND organization_name = ? LIMIT 1 ALLOW FILTERING"
		appcredentials, err = s.runGetAppCredentialQuery(query, key, organizationName)
	}
	if err != nil {
		return nil, err
	}

	if len(appcredentials) == 0 {
		s.db.metrics.QueryMiss(appCredentialsMetricLabel)
		return nil, fmt.Errorf("Can not find apikey '%s'", *key)
	}

	s.db.metrics.QueryHit(appCredentialsMetricLabel)
	return &appcredentials[0], nil
}

// GetByDeveloperAppID returns an array with apikey details of a developer app
func (s *CredentialStore) GetByDeveloperAppID(developerAppID string) ([]shared.DeveloperAppKey, error) {

	query := "SELECT * FROM credentials WHERE app_id = ?"
	appcredentials, err := s.runGetAppCredentialQuery(query, developerAppID)
	if err != nil {
		return nil, err
	}

	if len(appcredentials) == 0 {
		s.db.metrics.QueryMiss(appCredentialsMetricLabel)
		// Not being able to find a developer is not an error
		return appcredentials, nil
	}

	s.db.metrics.QueryHit(appCredentialsMetricLabel)
	return appcredentials, nil
}

// GetCountByDeveloperAppID retrieves number of keys beloning to developer app
func (s *CredentialStore) GetCountByDeveloperAppID(developerAppID string) int {

	var AppCredentialCount int

	query := "SELECT count(*) FROM credentials WHERE app_id = ?"
	if err := s.db.CassandraSession.Query(query, developerAppID).Scan(&AppCredentialCount); err != nil {
		s.db.metrics.QueryMiss(appCredentialsMetricLabel)
		return -1
	}

	s.db.metrics.QueryHit(appCredentialsMetricLabel)
	return AppCredentialCount
}

// runAppCredentialQuery executes CQL query and returns resulset
func (s *CredentialStore) runGetAppCredentialQuery(query string, queryParameters ...interface{}) ([]shared.DeveloperAppKey, error) {

	var appcredentials []shared.DeveloperAppKey

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	iterable := s.db.CassandraSession.Query(query, queryParameters...).Iter()
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
			appcredential.Attributes = s.db.UnmarshallJSONArrayOfAttributes(m["attributes"].(string))
		}
		if m["api_products"] != nil {
			appcredential.APIProducts = s.db.UnmarshallJSONArrayOfProductStatuses(m["api_products"].(string))
		}

		appcredentials = append(appcredentials, appcredential)
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iterable.Close(); err != nil {
		log.Error(err)
		return nil, err
	}
	return appcredentials, nil
}

// UpdateByKey UPSERTs credentials in database
func (s *CredentialStore) UpdateByKey(c *shared.DeveloperAppKey) error {

	c.Attributes = shared.TidyAttributes(c.Attributes)

	if err := s.db.CassandraSession.Query(`INSERT INTO credentials (
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
		s.db.MarshallArrayOfProductStatusesToJSON(c.APIProducts),
		s.db.MarshallArrayOfAttributesToJSON(c.Attributes),
		c.AppID,
		c.OrganizationName,
		c.Status,
		c.IssuedAt,
		c.ExpiresAt).Exec(); err != nil {

		return fmt.Errorf("Can not update credential '%s', (%v)", c.ConsumerKey, err)
	}
	return nil
}

// DeleteByKey deletes credentials
func (s *CredentialStore) DeleteByKey(organizationName, consumerKey string) error {

	_, err := s.GetByKey(&organizationName, &consumerKey)
	if err != nil {
		return err
	}

	query := "DELETE FROM credentials WHERE consumer_key = ?"
	return s.db.CassandraSession.Query(query, consumerKey).Exec()
}