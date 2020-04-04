package db

import (
	"errors"
	"fmt"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus label for metrics of db interactions
const clusterMetricLabel = "clusters"

// GetClusters retrieves all clusters
func (d *Database) GetClusters() ([]types.Cluster, error) {
	query := "SELECT * FROM clusters"
	clusters := d.runGetClusterQuery(query)
	if len(clusters) == 0 {
		d.metricsQueryMiss(clusterMetricLabel)
		return clusters, errors.New("Can not retrieve list of clusters")
	}
	d.metricsQueryHit(clusterMetricLabel)
	return clusters, nil
}

// GetClusterByName retrieves a cluster from database
func (d *Database) GetClusterByName(clusterName string) (types.Cluster, error) {
	query := "SELECT * FROM clusters WHERE key = ? LIMIT 1"
	clusters := d.runGetClusterQuery(query, clusterName)
	if len(clusters) == 0 {
		d.metricsQueryMiss(clusterMetricLabel)
		return types.Cluster{},
			fmt.Errorf("Can not find cluster (%s)", clusterName)
	}
	d.metricsQueryHit(clusterMetricLabel)
	return clusters[0], nil
}

// runGetClusterQuery executes CQL query and returns resultset
func (d *Database) runGetClusterQuery(query string, queryParameters ...interface{}) []types.Cluster {
	var clusters []types.Cluster

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		clusters = append(clusters, types.Cluster{
			Name:           m["key"].(string),
			HostName:       m["host_name"].(string),
			HostPort:       m["host_port"].(int16),
			CreatedAt:      m["created_at"].(int64),
			CreatedBy:      m["created_by"].(string),
			DisplayName:    m["display_name"].(string),
			LastmodifiedAt: m["lastmodified_at"].(int64),
			LastmodifiedBy: m["lastmodified_by"].(string),
		})
		m = map[string]interface{}{}
	}
	return clusters
}

// UpdateClusterByName UPSERTs an cluster in database
func (d *Database) UpdateClusterByName(updatedCluster types.Cluster) error {
	query := "INSERT INTO clusters (key, display_name, " +
		"host_name, host_port, " +
		"created_at, created_by, lastmodified_at, lastmodified_by) " +
		"VALUES(?,?,?,?,?,?,?,?)"

	if err := d.cassandraSession.Query(query,
		updatedCluster.Name, updatedCluster.DisplayName,
		updatedCluster.HostName, updatedCluster.HostPort,
		updatedCluster.CreatedAt, updatedCluster.CreatedBy,
		updatedCluster.LastmodifiedAt, updatedCluster.LastmodifiedBy).Exec(); err != nil {
		return fmt.Errorf("Can not update cluster (%v)", err)
	}
	return nil
}

// DeleteClusterByName deletes an cluster
func (d *Database) DeleteClusterByName(clusterToDelete string) error {
	_, err := d.GetClusterByName(clusterToDelete)
	if err != nil {
		return err
	}
	query := "DELETE FROM clusters WHERE key = ?"
	return d.cassandraSession.Query(query, clusterToDelete).Exec()
}
