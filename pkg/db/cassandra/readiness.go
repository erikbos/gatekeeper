package cassandra

import (
	"fmt"
	"time"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

const (
	systemLocalTable    = "system.local"
	healthCheckCQLquery = "SELECT * FROM " + systemLocalTable
	healthCheckInterval = 5 * time.Second
)

// ReadinessCheck holds our database config
type ReadinessCheck struct {
	db *Database
}

// NewReadiness creates readiness instance
func NewReadiness(database *Database) *ReadinessCheck {
	return &ReadinessCheck{
		db: database,
	}
}

// cassandraSystemInfo is used as state of our readiness
type cassandraSystemInfo struct {
	listenAddress string
	clusterName   string
	dataCenter    string
	rack          string
}

// RunReadinessCheck monitors database queryability
func (h *ReadinessCheck) RunReadinessCheck(n chan shared.ReadinessMessage) {

	var connected bool
	var msg string
	for {
		if peers, err := h.db.healthCheckQuery(); err == nil {
			connected = true
			msg = fmt.Sprintf("Database connection established (address: %s, name: %s, dc: %s, rack: %s)",
				peers.listenAddress, peers.clusterName, peers.dataCenter, peers.rack)
		} else {
			connected = false
			msg = fmt.Sprintf("Database connection failed (%s)", err)
		}
		// Send our status to Central Readiness Command
		n <- shared.ReadinessMessage{
			Component: "db",
			Message:   msg,
			Up:        connected,
		}
		time.Sleep(healthCheckInterval)
	}
}

// HealthCheckQuery checks Cassandra connectivity by reading table system.local
func (d *Database) healthCheckQuery() (*cassandraSystemInfo, error) {

	var peers cassandraSystemInfo

	iter := d.CassandraSession.Query(healthCheckCQLquery).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		peers = cassandraSystemInfo{
			listenAddress: m["listen_address"].(string),
			clusterName:   m["cluster_name"].(string),
			dataCenter:    m["data_center"].(string),
			rack:          m["rack"].(string),
		}
	}

	if err := iter.Close(); err != nil {
		d.metrics.QuerySuccessful(systemLocalTable)
		return nil, err
	}

	d.metrics.QuerySuccessful(systemLocalTable)
	return &peers, nil
}
