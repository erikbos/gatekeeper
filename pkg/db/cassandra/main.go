package cassandra

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// DatabaseConfig holds database connection configuration
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
	QueryRetries    int           `yaml:"queryretries"`
}

// Database holds all our database connection information and performance counters
type Database struct {
	CassandraSession *gocql.Session
	metrics          metricsCollection
}

// New builds new connected database instance
func New(config DatabaseConfig, serviceName string,
	createSchema bool, replicationCount int) (*db.Database, error) {

	cassandraClusterConfig := buildClusterConfig(config)

	if createSchema {
		// Connect to system keyspace first as our keyspace does not exist yet
		cassandraClusterConfig.Keyspace = "system"
		db, err := connect(cassandraClusterConfig, config)
		if err != nil {
			return nil, err
		}

		// Create keyspace with specific replication count
		if err := createKeyspace(db, config.Keyspace, replicationCount); err != nil {
			return nil, err
		}
		// Close session: we cannot use it any further,
		// as we are connected to the wrong keyspace: we need to reconnect.
		db.Close()
	}

	// Connect to Cassandra on correct keyspace
	cassandraClusterConfig.Keyspace = config.Keyspace
	cassandraSession, err := connect(cassandraClusterConfig, config)
	if err != nil {
		return nil, err
	}
	// Create tables within keyspace if requested
	if createSchema {
		if err := createTables(cassandraSession); err != nil {
			return nil, err
		}
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
		Readiness:    NewReadiness(&dbConfig),
	}
	return &database, nil
}

func buildClusterConfig(config DatabaseConfig) *gocql.ClusterConfig {

	clusterConfig := gocql.NewCluster(config.Hostname)

	clusterConfig.Port = config.Port

	if config.TLS.Enable {
		// Empty struct to enable TLS
		clusterConfig.SslOpts = &gocql.SslOptions{}

		if config.TLS.Capath != "" {
			clusterConfig.SslOpts.CaPath = config.TLS.Capath
		}
	}
	clusterConfig.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Username,
		Password: config.Password,
	}
	if config.Timeout != 0 {
		clusterConfig.Timeout = config.Timeout
	}
	if config.QueryRetries != 0 {
		clusterConfig.RetryPolicy = &gocql.SimpleRetryPolicy{
			NumRetries: config.QueryRetries,
		}
	}

	return clusterConfig
}

// connect setups up connectivity to Cassandra
func connect(cluster *gocql.ClusterConfig, config DatabaseConfig) (*gocql.Session, error) {

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
