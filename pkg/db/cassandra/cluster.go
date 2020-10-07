package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// Prometheus label for metrics of db interactions
	clusterMetricLabel = "clusters"

	// List of cluster columns we use
	clusterColumns = `name,
display_name,
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
func (s *ClusterStore) GetAll() (types.Clusters, types.Error) {

	query := "SELECT " + clusterColumns + " FROM clusters"
	clusters, err := s.runGetClusterQuery(query)
	if err != nil {
		s.db.metrics.QueryFailed(clusterMetricLabel)
		return types.NullClusters, types.NewDatabaseError(err)
	}

	s.db.metrics.QueryHit(clusterMetricLabel)
	return clusters, nil
}

// Get retrieves a cluster from database
func (s *ClusterStore) Get(clusterName string) (*types.Cluster, types.Error) {

	query := "SELECT " + clusterColumns + " FROM clusters WHERE name = ? LIMIT 1"
	clusters, err := s.runGetClusterQuery(query, clusterName)
	if err != nil {
		s.db.metrics.QueryFailed(clusterMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(clusters) == 0 {
		s.db.metrics.QueryMiss(clusterMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("Can not find cluster (%s)", clusterName))
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
		return types.Clusters{}, err
	}
	return clusters, nil
}

// Update UPSERTs an cluster in database
func (s *ClusterStore) Update(c *types.Cluster) types.Error {

	query := "INSERT INTO clusters (" + clusterColumns + ") VALUES(?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		c.Name,
		c.DisplayName,
		c.Attributes.Marshal(),
		c.CreatedAt,
		c.CreatedBy,
		c.LastmodifiedAt,
		c.LastmodifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(clusterMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("Cannot update cluster '%s'", c.Name))
	}
	return nil
}

// Delete deletes a cluster
func (s *ClusterStore) Delete(clusterToDelete string) types.Error {

	query := "DELETE FROM clusters WHERE name = ?"
	if err := s.db.CassandraSession.Query(query, clusterToDelete).Exec(); err != nil {
		s.db.metrics.QueryFailed(clusterMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
