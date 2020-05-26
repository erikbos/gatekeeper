package db

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const clusterMetricLabel = "clusters"

// GetClusters retrieves all clusters
func (d *Database) GetClusters() ([]shared.Cluster, error) {

	query := "SELECT * FROM clusters"
	clusters, err := d.runGetClusterQuery(query)
	if err != nil {
		return []shared.Cluster{}, err
	}

	if len(clusters) == 0 {
		d.metricsQueryMiss(clusterMetricLabel)
		return []shared.Cluster{}, errors.New("Can not retrieve list of clusters")
	}

	d.metricsQueryHit(clusterMetricLabel)
	return clusters, nil
}

// GetClusterByName retrieves a cluster from database
func (d *Database) GetClusterByName(clusterName string) (shared.Cluster, error) {

	query := "SELECT * FROM clusters WHERE name = ? LIMIT 1"
	clusters, err := d.runGetClusterQuery(query, clusterName)

	if err != nil {
		return shared.Cluster{}, err
	}

	if len(clusters) == 0 {
		d.metricsQueryMiss(clusterMetricLabel)
		return shared.Cluster{},
			fmt.Errorf("Can not find cluster (%s)", clusterName)
	}

	d.metricsQueryHit(clusterMetricLabel)
	return clusters[0], nil
}

// runGetClusterQuery executes CQL query and returns resultset
func (d *Database) runGetClusterQuery(query string, queryParameters ...interface{}) ([]shared.Cluster, error) {
	var clusters []shared.Cluster

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iter := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		newCluster := shared.Cluster{
			Name:           m["name"].(string),
			HostName:       m["host_name"].(string),
			Port:           m["port"].(int),
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
		return []shared.Cluster{}, err
	}
	return clusters, nil
}

// UpdateClusterByName UPSERTs an cluster in database
func (d *Database) UpdateClusterByName(c *shared.Cluster) error {

	c.Attributes = shared.TidyAttributes(c.Attributes)
	c.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	if err := d.cassandraSession.Query(`INSERT INTO clusters (
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
		d.marshallArrayOfAttributesToJSON(c.Attributes),
		c.CreatedAt,
		c.CreatedBy,
		c.LastmodifiedAt,
		c.LastmodifiedBy).Exec(); err != nil {

		return fmt.Errorf("Can not update cluster '%s' (%v)", c.Name, err)
	}

	return nil
}

// DeleteClusterByName deletes a cluster
func (d *Database) DeleteClusterByName(clusterToDelete string) error {

	_, err := d.GetClusterByName(clusterToDelete)
	if err != nil {
		return err
	}

	query := "DELETE FROM clusters WHERE name = ?"
	return d.cassandraSession.Query(query, clusterToDelete).Exec()
}
