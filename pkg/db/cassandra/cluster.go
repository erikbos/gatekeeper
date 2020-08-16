package cassandra

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const clusterMetricLabel = "clusters"

// ClusterStore holds our database config
type ClusterStore struct {
	db *Database
}

// NewClusterStore creates cluster instance
func NewClusterStore(database *Database) *ClusterStore {
	return &ClusterStore{
		db: database,
	}
}

// GetAll retrieves all clusters
func (s *ClusterStore) GetAll() ([]shared.Cluster, error) {

	query := "SELECT * FROM clusters"
	clusters, err := s.runGetClusterQuery(query)
	if err != nil {
		return []shared.Cluster{}, err
	}

	if len(clusters) == 0 {
		s.db.metrics.QueryMiss(clusterMetricLabel)
		return []shared.Cluster{}, errors.New("Can not retrieve list of clusters")
	}

	s.db.metrics.QueryHit(clusterMetricLabel)
	return clusters, nil
}

// GetByName retrieves a cluster from database
func (s *ClusterStore) GetByName(clusterName string) (*shared.Cluster, error) {

	query := "SELECT * FROM clusters WHERE name = ? LIMIT 1"
	clusters, err := s.runGetClusterQuery(query, clusterName)

	if err != nil {
		return nil, err
	}

	if len(clusters) == 0 {
		s.db.metrics.QueryMiss(clusterMetricLabel)
		return nil, fmt.Errorf("Can not find cluster (%s)", clusterName)
	}

	s.db.metrics.QueryHit(clusterMetricLabel)
	return &clusters[0], nil
}

// runGetClusterQuery executes CQL query and returns resultset
func (s *ClusterStore) runGetClusterQuery(query string, queryParameters ...interface{}) ([]shared.Cluster, error) {
	var clusters []shared.Cluster

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		clusters = append(clusters, shared.Cluster{
			Name:           m["name"].(string),
			HostName:       m["host_name"].(string),
			Port:           m["port"].(int),
			Attributes:     shared.Cluster{}.Attributes.Unmarshal(m["attributes"].(string)),
			CreatedAt:      m["created_at"].(int64),
			CreatedBy:      m["created_by"].(string),
			DisplayName:    m["display_name"].(string),
			LastmodifiedAt: m["lastmodified_at"].(int64),
			LastmodifiedBy: m["lastmodified_by"].(string),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		log.Error(err)
		return []shared.Cluster{}, err
	}
	return clusters, nil
}

// UpdateByName UPSERTs an cluster in database
func (s *ClusterStore) UpdateByName(c *shared.Cluster) error {

	c.Attributes.Tidy()
	c.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	if err := s.db.CassandraSession.Query(`INSERT INTO clusters (
name,
display_name,
host_name,
port,
attributes,
created_at,
created_by,
lastmodified_at,
lastmodified_by) VALUES(?,?,?,?,?,?,?,?,?)`,

		c.Name,
		c.DisplayName,
		c.HostName,
		c.Port,
		c.Attributes.Marshal(),
		c.CreatedAt,
		c.CreatedBy,
		c.LastmodifiedAt,
		c.LastmodifiedBy).Exec(); err != nil {

		return fmt.Errorf("Can not update cluster '%s' (%v)", c.Name, err)
	}

	return nil
}

// DeleteByName deletes a cluster
func (s *ClusterStore) DeleteByName(clusterToDelete string) error {

	_, err := s.GetByName(clusterToDelete)
	if err != nil {
		return err
	}

	query := "DELETE FROM clusters WHERE name = ?"
	return s.db.CassandraSession.Query(query, clusterToDelete).Exec()
}
