package db

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	minimumHealthCheckInterval = 2 * time.Second
)

// HealthCheckStatus state of our healthiness
type HealthCheckStatus struct {
	listenAddress string
	clusterName   string
	dataCenter    string
	rack          string
}

func (d *Database) runHealthCheck(interval time.Duration) {
	var connected bool

	if interval < minimumHealthCheckInterval {
		log.Fatalf("Db healthcheck interval lower than %s", minimumHealthCheckInterval)
	}

	for {
		if peers, err := d.HealthCheckQuery(); err == nil {
			d.readiness.Up()

			if !connected {
				log.Infof("Database connected (%s, %s, %s, %s)",
					peers.listenAddress, peers.clusterName, peers.dataCenter, peers.rack)
				log.Infof("Database healthcheck ok")
				connected = true
			}
		} else {
			d.readiness.Down()

			log.Infof("Database healthcheck failed (%s)", err)
			connected = false
		}
		time.Sleep(interval)
	}
}

// HealthCheckQuery checks Cassandra connectivity by reading table system.local
func (d *Database) HealthCheckQuery() (HealthCheckStatus, error) {
	var peers HealthCheckStatus

	query := "select * from system.local"
	iter := d.cassandraSession.Query(query).Iter()
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
		return HealthCheckStatus{}, err
	}
	return peers, nil
}
