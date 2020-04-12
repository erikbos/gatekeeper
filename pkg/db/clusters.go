package db

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus label for metrics of db interactions
const clusterMetricLabel = "clusters"

// GetClusters retrieves all clusters
func (d *Database) GetClusters() ([]types.Cluster, error) {
	query := "SELECT * FROM clusters"
	clusters, err := d.runGetClusterQuery(query)
	if err != nil {
		return []types.Cluster{}, err
	}
	if len(clusters) == 0 {
		d.metricsQueryMiss(clusterMetricLabel)
		return []types.Cluster{}, errors.New("Can not retrieve list of clusters")
	}
	d.metricsQueryHit(clusterMetricLabel)
	return clusters, nil
}

// GetClusterByName retrieves a cluster from database
func (d *Database) GetClusterByName(clusterName string) (types.Cluster, error) {
	query := "SELECT * FROM clusters WHERE key = ? LIMIT 1"
	clusters, err := d.runGetClusterQuery(query, clusterName)
	if err != nil {
		return types.Cluster{}, err
	}
	if len(clusters) == 0 {
		d.metricsQueryMiss(clusterMetricLabel)
		return types.Cluster{},
			fmt.Errorf("Can not find cluster (%s)", clusterName)
	}
	d.metricsQueryHit(clusterMetricLabel)
	return clusters[0], nil
}

// runGetClusterQuery executes CQL query and returns resultset
func (d *Database) runGetClusterQuery(query string, queryParameters ...interface{}) ([]types.Cluster, error) {
	var clusters []types.Cluster

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iter := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		newCluster := types.Cluster{
			Name:           m["key"].(string),
			HostName:       m["host_name"].(string),
			HostPort:       m["host_port"].(int16),
			CreatedAt:      m["created_at"].(int64),
			CreatedBy:      m["created_by"].(string),
			DisplayName:    m["display_name"].(string),
			LastmodifiedAt: m["lastmodified_at"].(int64),
			LastmodifiedBy: m["lastmodified_by"].(string),
		}
		if m["attributes"] != nil {
			newCluster.Attributes = d.unmarshallJSONArrayOfAttributes(m["attributes"].(string))
		}
		clusters = append(clusters, newCluster)
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		log.Error(err)
		return []types.Cluster{}, err
	}
	return clusters, nil
}

// UpdateClusterByName UPSERTs an cluster in database
func (d *Database) UpdateClusterByName(updatedCluster *types.Cluster) error {
	query := "INSERT INTO clusters (key, display_name, " +
		"host_name, host_port, attributes, " +
		"created_at, created_by, lastmodified_at, lastmodified_by) " +
		"VALUES(?,?,?,?,?,?,?,?,?)"
	updatedCluster.Attributes = types.TidyAttributes(updatedCluster.Attributes)
	attributes := d.marshallArrayOfAttributesToJSON(updatedCluster.Attributes)
	updatedCluster.LastmodifiedAt = types.GetCurrentTimeMilliseconds()
	if err := d.cassandraSession.Query(query,
		updatedCluster.Name, updatedCluster.DisplayName,
		updatedCluster.HostName, updatedCluster.HostPort, attributes,
		updatedCluster.CreatedAt, updatedCluster.CreatedBy,
		updatedCluster.LastmodifiedAt,
		updatedCluster.LastmodifiedBy).Exec(); err != nil {
		return fmt.Errorf("Can not update cluster (%v)", err)
	}
	return nil
}

// DeleteClusterByName deletes a cluster
func (d *Database) DeleteClusterByName(clusterToDelete string) error {
	_, err := d.GetClusterByName(clusterToDelete)
	if err != nil {
		return err
	}
	query := "DELETE FROM clusters WHERE key = ?"
	return d.cassandraSession.Query(query, clusterToDelete).Exec()
}
