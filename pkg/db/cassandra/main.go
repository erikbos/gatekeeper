package cassandra

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// DatabaseConfig holds configuration configuration
type DatabaseConfig struct {
	Hostname string        `yaml:"hostname" json:"hostname"`
	Port     int           `yaml:"port"     json:"port"`
	Username string        `yaml:"username" json:"username" `
	Password string        `yaml:"password" json:"password"`
	Keyspace string        `yaml:"keyspace" json:"keyspace"`
	Timeout  time.Duration `yaml:"timeout"  json:"timeout"`
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
	cluster.SslOpts = &gocql.SslOptions{
		Config: &tls.Config{
			InsecureSkipVerify: true,
		},
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

	cassandraSession, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("Could not connect to database (%s)", err)
	}

	log.Infof("Database connected as '%s' to '%s'",
		config.Username, config.Hostname)

	return cassandraSession, nil
}
