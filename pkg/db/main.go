package db

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/erikbos/apiauth/pkg/shared"

	"github.com/gocql/gocql"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Database holds all our database connection information and performance counters
type Database struct {
	ServiceName           string
	Config                DatabaseConfig
	cassandraSession      *gocql.Session
	readiness             *shared.Readiness
	dbLookupHitsCounter   *prometheus.CounterVec
	dbLookupMissesCounter *prometheus.CounterVec
	dbLookupHistogram     prometheus.Summary
}

// DatabaseConfig holds configuration configuration
type DatabaseConfig struct {
	Hostname            string `yaml:"hostname"`
	Port                int    `yaml:"port"`
	Username            string `yaml:"username"`
	Password            string `yaml:"password"`
	Keyspace            string `yaml:"keyspace"`
	Timeout             string `yaml:"timeout"`
	HealthcheckInterval string `yaml:"healthcheckinterval"`
}

// Connect setups up connectivity to Cassandra
func Connect(config DatabaseConfig, r *shared.Readiness, serviceName string) (*Database, error) {
	var err error
	d := Database{
		ServiceName: serviceName,
		Config:      config,
		readiness:   r,
	}
	cluster := gocql.NewCluster(d.Config.Hostname)
	cluster.Port = d.Config.Port
	cluster.SslOpts = &gocql.SslOptions{
		CertPath:               "selfsigned.crt",
		KeyPath:                "selfsigned.key",
		EnableHostVerification: false,
	}
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: d.Config.Username,
		Password: d.Config.Password,
	}
	cluster.Keyspace = d.Config.Keyspace

	if d.Config.Timeout != "" {
		timeout, err := time.ParseDuration(d.Config.Timeout)
		if err != nil {
			log.Fatalf("Cannot parse database timeout %v", err)
		}
		cluster.Timeout = timeout
	}
	// cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}

	d.cassandraSession, err = cluster.CreateSession()
	if err != nil {
		return nil, errors.New("Could not connect to database")
	}

	d.registerMetrics()

	go d.runHealthCheck(d.Config.HealthcheckInterval)

	return &d, nil
}

// registerMetrics registers database performance metrics
func (d *Database) registerMetrics() {
	d.dbLookupHitsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: d.ServiceName + "_database_lookup_hits_total",
			Help: "Number of successful database lookups.",
		}, []string{"hostname", "table"})
	prometheus.MustRegister(d.dbLookupHitsCounter)

	d.dbLookupMissesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: d.ServiceName + "database_lookup_misses_total",
			Help: "Number of unsuccesful database lookups.",
		}, []string{"hostname", "table"})
	prometheus.MustRegister(d.dbLookupMissesCounter)

	d.dbLookupHistogram = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: d.ServiceName + "_database_lookup_latency",
			Help: "Database lookup latency in seconds.",
			Objectives: map[float64]float64{
				0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001,
			},
		})
	prometheus.MustRegister(d.dbLookupHistogram)
}

func (d *Database) metricsQueryHit(metricLabel string) {
	d.dbLookupHitsCounter.WithLabelValues(d.Config.Hostname, metricLabel).Inc()
}

func (d *Database) metricsQueryMiss(metricLabel string) {
	d.dbLookupMissesCounter.WithLabelValues(d.Config.Hostname, metricLabel).Inc()
}

// unmarshallJSONArrayOfStrings unpacks JSON array of strings
// e.g. [\"PetStore5\",\"PizzaShop1\"] to []string
//
func (d *Database) unmarshallJSONArrayOfStrings(jsonArrayOfStrings string) []string {
	if jsonArrayOfStrings != "" {
		var StringValues []string
		err := json.Unmarshal([]byte(jsonArrayOfStrings), &StringValues)
		if err == nil {
			return StringValues
		}
	}
	return nil
}

// MarshallArrayOfStringsToJSON packs array of string into JSON
// e.g. []string to [\"PetStore5\",\"PizzaShop1\"]
//
func (d *Database) marshallArrayOfStringsToJSON(ArrayOfStrings []string) string {
	if len(ArrayOfStrings) > 0 {
		ArrayOfStringsInJSON, err := json.Marshal(ArrayOfStrings)
		if err == nil {
			return string(ArrayOfStringsInJSON)
		}
	}
	return "[]"
}

// unmarshallJSONArrayOfAttributes unpacks JSON array of attribute bags
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (d *Database) unmarshallJSONArrayOfAttributes(jsonArrayOfAttributes string) []shared.AttributeKeyValues {
	if jsonArrayOfAttributes != "" {
		var ResponseAttributes = make([]shared.AttributeKeyValues, 0)
		if err := json.Unmarshal([]byte(jsonArrayOfAttributes), &ResponseAttributes); err == nil {
			return ResponseAttributes
		}
	}
	return nil
}

// marshallArrayOfAttributesToJSON packs array of attributes into JSON
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (d *Database) marshallArrayOfAttributesToJSON(ArrayOfAttributes []shared.AttributeKeyValues) string {

	if len(ArrayOfAttributes) > 0 {
		ArrayOfAttributesInJSON, err := json.Marshal(ArrayOfAttributes)
		if err == nil {
			return string(ArrayOfAttributesInJSON)
		}
	}
	return "[]"
}

// unmarshallJSONArrayOfAttributes unpacks JSON array of attribute bags
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (d *Database) unmarshallJSONArrayOfProductStatuses(jsonArrayOfAttributes string) []shared.APIProductStatus {
	if jsonArrayOfAttributes != "" {
		var ResponseAttributes = make([]shared.APIProductStatus, 0)
		if err := json.Unmarshal([]byte(jsonArrayOfAttributes), &ResponseAttributes); err == nil {
			return ResponseAttributes
		}
	}
	return nil
}

// marshallArrayOfAttributesToJSON packs array of attributes into JSON
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (d *Database) marshallArrayOfProductStatusesToJSON(ArrayOfAttributes []shared.APIProductStatus) string {

	if len(ArrayOfAttributes) > 0 {
		ArrayOfAttributesInJSON, err := json.Marshal(ArrayOfAttributes)
		if err == nil {
			return string(ArrayOfAttributesInJSON)
		}
	}
	return "[]"
}
