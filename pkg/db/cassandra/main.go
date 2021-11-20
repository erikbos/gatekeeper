package cassandra

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

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
func New(config DatabaseConfig, serviceName string, logger *zap.Logger,
	createSchema bool, replicationCount int) (*db.Database, error) {

	logger = logger.With(zap.String("system", "db"))

	cassandraClusterConfig := buildClusterConfig(config)

	if createSchema {
		// Connect to system keyspace first as our keyspace does not exist yet
		cassandraClusterConfig.Keyspace = "system"
		db, err := connect(cassandraClusterConfig, config, logger)
		if err != nil {
			return nil, err
		}

		// Create keyspace with specific replication count
		if err := createKeyspace(db, config.Keyspace, replicationCount, logger); err != nil {
			return nil, err
		}
		// Close session: we cannot use it any further,
		// as we are connected to the wrong keyspace: we need to reconnect.
		db.Close()
	}

	// Connect to Cassandra on correct keyspace
	cassandraClusterConfig.Keyspace = config.Keyspace
	cassandraSession, err := connect(cassandraClusterConfig, config, logger)
	if err != nil {
		return nil, err
	}
	// Create tables within keyspace if requested
	if createSchema {
		if err := createTables(cassandraSession, logger); err != nil {
			return nil, err
		}
	}

	dbConfig := Database{
		CassandraSession: cassandraSession,
		metrics:          metricsCollection{},
	}

	dbConfig.metrics.register(serviceName, config.Hostname)

	database := db.Database{
		Listener:     NewListenerStore(&dbConfig),
		Route:        NewRouteStore(&dbConfig),
		Cluster:      NewClusterStore(&dbConfig),
		Organization: NewOrganizationStore(&dbConfig),
		Developer:    NewDeveloperStore(&dbConfig),
		DeveloperApp: NewDeveloperAppStore(&dbConfig),
		APIProduct:   NewAPIProductStore(&dbConfig),
		Key:          NewKeyStore(&dbConfig),
		OAuth:        NewOAuthStore(&dbConfig),
		User:         NewUserStore(&dbConfig),
		Role:         NewRoleStore(&dbConfig),
		Readiness:    NewReadiness(&dbConfig),
	}
	return &database, nil
}

func buildClusterConfig(config DatabaseConfig) *gocql.ClusterConfig {

	clusterConfig := gocql.NewCluster(config.Hostname)

	clusterConfig.Port = config.Port

	if config.TLS.Enable {
		// Enforece TLS, minimum TLS1.2, set hostname to match with presented certificate
		clusterConfig.SslOpts = &gocql.SslOptions{
			Config: &tls.Config{
				MinVersion: tls.VersionTLS12,
				ServerName: config.Hostname,
			},
		}

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
func connect(cluster *gocql.ClusterConfig, config DatabaseConfig,
	logger *zap.Logger) (*gocql.Session, error) {

	// In case no number of connect attempts is set we try at least once
	if config.ConnectAttempts == 0 {
		config.ConnectAttempts = 1
	}

	var err error

	attempt := 0
	for attempt < config.ConnectAttempts {
		attempt++
		logger.Debug(fmt.Sprintf("Attemping to connect to database, attempt %d of %d",
			attempt, config.ConnectAttempts))

		var cassandraSession *gocql.Session
		cassandraSession, err = cluster.CreateSession()
		if cassandraSession != nil {
			logger.Info(fmt.Sprintf("Connected to database as %s to %s",
				config.Username, config.Hostname))

			return cassandraSession, nil
		}
		time.Sleep(3 * time.Second)
	}
	return nil, err
}
