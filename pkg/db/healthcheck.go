package db

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const (
	healthCheckMetricLabel = "system.local"
	healthCheckCQLquery    = "select * from system.local"

	minimumHealthCheckInterval = 2 * time.Second
)

// HealthCheckStatus state of our healthiness
type HealthCheckStatus struct {
	listenAddress string
	clusterName   string
	dataCenter    string
	rack          string
}

// runHealthCheck runs continous query against databse to confirm connectivity
func (d *Database) runHealthCheck(healthcheckInterval string) {

	interval := d.getHealthCheckInterval(healthcheckInterval)

	var connected bool
	for {
		if peers, err := d.HealthCheckQuery(); err == nil {
			if !connected {
				log.Infof("Database connected (%s, %s, %s, %s)",
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
		time.Sleep(interval)
	}
}

// getHealthCheckInterval parses config option to set db healthcheck interval
func (d *Database) getHealthCheckInterval(healthcheckInterval string) time.Duration {

	interval, err := time.ParseDuration(healthcheckInterval)
	if err != nil {
		log.Fatalf("Cannot parse '%s' as db healthCheckInterval (%s)", healthcheckInterval, err)
	}

	if interval < minimumHealthCheckInterval {
		log.Fatalf("Db healthcheck interval set to low, should be >= '%s'", minimumHealthCheckInterval)
	}

	return interval
}

// HealthCheckQuery checks Cassandra connectivity by reading table system.local
func (d *Database) HealthCheckQuery() (HealthCheckStatus, error) {
	var peers HealthCheckStatus

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

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
