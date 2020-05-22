package db

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	healthCheckMetricLabel = "system.local"
	healthCheckCQLquery    = "select * from system.local"
	healthCheckInterval    = 5 * time.Second
)

// HealthCheckStatus state of our healthiness
type HealthCheckStatus struct {
	listenAddress string
	clusterName   string
	dataCenter    string
	rack          string
}

// runContinousHealthCheck monitors database queryability
func (d *Database) runContinousHealthCheck() {

	var connected bool
	for {
		if peers, err := d.HealthCheckQuery(); err == nil {
			if !connected {
				log.Infof("Database connected (address: %s, name: %s, dc: %s, rack: %s)",
					peers.listenAddress, peers.clusterName, peers.dataCenter, peers.rack)
				log.Infof("Database healthcheck ok")
				connected = true
			}
			d.readiness.Up()
		} else {
			log.Warnf("Database healthcheck failed (%s)", err)
			connected = false

			d.readiness.Down()
		}
		time.Sleep(healthCheckInterval)
	}
}

// HealthCheckQuery checks Cassandra connectivity by reading table system.local
func (d *Database) HealthCheckQuery() (HealthCheckStatus, error) {
	var peers HealthCheckStatus

	// timer := prometheus.NewTimer(d.dbLookupHistogram)
	// defer timer.ObserveDuration()

	iter := d.cassandraSession.Query(healthCheckCQLquery).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		peers = HealthCheckStatus{
			listenAddress: m["listen_address"].(string),
			clusterName:   m["cluster_name"].(string),
			dataCenter:    m["data_center"].(string),
			rack:          m["rack"].(string),
		}
	}

	if err := iter.Close(); err != nil {
		d.metricsQueryMiss(healthCheckMetricLabel)
		return HealthCheckStatus{}, err
	}

	d.metricsQueryHit(healthCheckMetricLabel)
	return peers, nil
}
