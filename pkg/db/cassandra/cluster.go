package cassandra

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// Prometheus label for metrics of db interactions
	clusterMetricLabel = "clusters"

	// List of cluster columns we use
	clusterColumns = `name,
display_name,
host_name,
port,
attributes,
created_at,
created_by,
lastmodified_at,
lastmodified_by`
)

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
func (s *ClusterStore) GetAll() (types.Clusters, error) {

	query := "SELECT " + clusterColumns + " FROM clusters"
	clusters, err := s.runGetClusterQuery(query)
	if err != nil {
		return types.Clusters{}, err
	}

	if len(clusters) == 0 {
		s.db.metrics.QueryMiss(clusterMetricLabel)
		return types.Clusters{}, errors.New("Can not retrieve list of clusters")
	}

	s.db.metrics.QueryHit(clusterMetricLabel)
	return clusters, nil
}

// GetByName retrieves a cluster from database
func (s *ClusterStore) GetByName(clusterName string) (*types.Cluster, error) {

	query := "SELECT " + clusterColumns + " FROM clusters WHERE name = ? LIMIT 1"
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
func (s *ClusterStore) runGetClusterQuery(query string, queryParameters ...interface{}) (types.Clusters, error) {
	var clusters types.Clusters

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		clusters = append(clusters, types.Cluster{
			Name:           columnValueString(m, "name"),
			DisplayName:    columnValueString(m, "display_name"),
			HostName:       columnValueString(m, "host_name"),
			Port:           columnValueInt(m, "port"),
			Attributes:     types.Cluster{}.Attributes.Unmarshal(columnValueString(m, "attributes")),
			CreatedAt:      columnValueInt64(m, "created_at"),
			CreatedBy:      columnValueString(m, "created_by"),
			LastmodifiedAt: columnValueInt64(m, "lastmodified_at"),
			LastmodifiedBy: columnValueString(m, "lastmodified_by"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		log.Error(err)
		return types.Clusters{}, err
	}
	return clusters, nil
}

// UpdateByName UPSERTs an cluster in database
func (s *ClusterStore) UpdateByName(c *types.Cluster) error {

	c.Attributes.Tidy()
	c.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	query := "INSERT INTO clusters (" + clusterColumns + ") VALUES(?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
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
