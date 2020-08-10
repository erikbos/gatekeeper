package cassandra

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// DatabaseConfig holds configuration configuration
type DatabaseConfig struct {
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	TLS      struct {
		Enable bool   `yaml:"enable"`
		Capath string `yaml:"capath"`
	} `yaml:"tls"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Keyspace        string        `yaml:"keyspace"`
	Timeout         time.Duration `yaml:"timeout"`
	ConnectAttempts int           `yaml:"connectattempts"`
}

// Database holds all our database connection information and performance counters
type Database struct {
	CassandraSession *gocql.Session
	metrics          metricsCollection
}

// New builds new connected database instance
func New(config DatabaseConfig, serviceName string) (*db.Database, error) {

	cassandraSession, err := connect(config, serviceName)
	if err != nil {
		return nil, err
	}

	dbConfig := Database{
		CassandraSession: cassandraSession,
		metrics:          metricsCollection{},
	}

	dbConfig.metrics.register(serviceName, config.Hostname)

	database := db.Database{
		Virtualhost:  NewVirtualhostStore(&dbConfig),
		Route:        NewRouteStore(&dbConfig),
		Cluster:      NewClusterStore(&dbConfig),
		Organization: NewOrganizationStore(&dbConfig),
		Developer:    NewDeveloperStore(&dbConfig),
		DeveloperApp: NewDeveloperAppStore(&dbConfig),
		APIProduct:   NewAPIProductStore(&dbConfig),
		Credential:   NewCredentialStore(&dbConfig),
		OAuth:        NewOAuthStore(&dbConfig),
		Readiness:    NewReadinessCheck(&dbConfig),
	}
	return &database, nil
}

// connect setups up connectivity to Cassandra
func connect(config DatabaseConfig, serviceName string) (*gocql.Session, error) {

	cluster := gocql.NewCluster(config.Hostname)

	cluster.Port = config.Port

	if config.TLS.Enable == true {
		// Empty struct to enable TLS
		cluster.SslOpts = &gocql.SslOptions{}

		if config.TLS.Capath != "" {
			cluster.SslOpts.CaPath = config.TLS.Capath
		}
	}
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Username,
		Password: config.Password,
	}
	cluster.Keyspace = config.Keyspace

	if config.Timeout != 0 {
		cluster.Timeout = config.Timeout
	}
	// cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}

	// In case no number of connect attempts is set we try at least once
	if config.ConnectAttempts == 0 {
		config.ConnectAttempts = 1
	}

	var err error
	var cassandraSession *gocql.Session

	attempt := 0
	for attempt < config.ConnectAttempts {
		attempt++
		log.Debugf("Trying to connect to database, attempt %d of %d", attempt, config.ConnectAttempts)

		cassandraSession, err = cluster.CreateSession()
		if cassandraSession != nil {
			log.Infof("Database connected as '%s' to '%s'", config.Username, config.Hostname)

			return cassandraSession, nil
		}
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("Could not connect to database (%s)", err)
}
