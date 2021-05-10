package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

const (
	createKeyspaceCQL = "CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': %d };"
)

// createKeyspace creates keyspace if it does not yet exist
func createKeyspace(s *gocql.Session, keyspace string, replicationCount int, logger *zap.Logger) error {

	logger.Info("Creating keyspace if not existing",
		zap.String("keyspace", keyspace),
		zap.Int("replicationCount", replicationCount))
	createKeyspaceQuery := fmt.Sprintf(createKeyspaceCQL, keyspace, replicationCount)
	if err := s.Query(createKeyspaceQuery).Exec(); err != nil {
		return err
	}

	if replicationCount < 3 {
		logger.Error("This database is not suitable for production",
			zap.Int("replicationCount", replicationCount))
	}

	return nil
}

// createTables adds required tables if they do not yet exist
func createTables(s *gocql.Session, logger *zap.Logger) error {

	logger.Info("Creating all tables if not existing")
	for _, query := range createTablesCQL {
		logger.Debug("init database", zap.String("cql", query))
		if err := s.Query(query).Exec(); err != nil {
			logger.Warn("init database statement failed", zap.Error(err))
			return err
		}
	}

	logger.Info("Tables and indices created if not existing")
	return nil
}

// ShowCreateSchemaStatements show CQL statements to create all tables
func ShowCreateSchemaStatements() {

	fmt.Printf(createKeyspaceCQL+"\n\n", "keyspace", 3)

	for _, query := range createTablesCQL {
		fmt.Printf("%s\n\n", query)
	}
}

var createTablesCQL = [...]string{

	`CREATE TABLE IF NOT EXISTS users (
        name text,
        display_name text,
        password text,
        status text,
        roles text,
        created_at bigint,
        created_by text,
        lastmodified_at bigint,
        lastmodified_by text,
        PRIMARY KEY (name)
        )`,

	// Default database user 'admin', password 'passwd', role 'admin'
	`INSERT INTO users (name,password,status,roles,created_by,created_at,lastmodified_at) VALUES('admin','$2a$07$zWlw6WvswAFGZzNpBJg5qelwyg87NM/w4ypXP.NhfpuYmmv.WPyJO','active','["admin"]','initdb',toUnixTimestamp(now()),toUnixTimestamp(now())) IF NOT EXISTS`,

	`CREATE TABLE IF NOT EXISTS roles (
        name text,
        display_name text,
        allows text,
        created_at bigint,
        created_by text,
        lastmodified_at bigint,
        lastmodified_by text,
        PRIMARY KEY (name)
        )`,

	// Default database role 'admin', allowing GET, POST, DELETE on /v1/* path
	`INSERT INTO roles (name,allows,created_by,created_at,lastmodified_at) VALUES('admin','[{"methods":["GET","POST","DELETE"],"paths":["/v1/**"]}]','initdb',toUnixTimestamp(now()),toUnixTimestamp(now())) IF NOT EXISTS`,

	`CREATE TABLE IF NOT EXISTS listeners (
    attributes text,
    created_at bigint,
    created_by text,
    display_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    policies text,
    port int,
    route_group text,
    virtual_hosts text,
    PRIMARY KEY (name)
	)`,

	`CREATE TABLE IF NOT EXISTS routes (
    attributes text,
    created_at bigint,
    created_by text,
    display_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    path text,
    path_type text,
    route_group text,
    PRIMARY KEY (name)
	)`,

	`CREATE TABLE IF NOT EXISTS clusters (
    attributes text,
    created_at bigint,
    created_by text,
    display_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    PRIMARY KEY (name)
	)`,

	`CREATE TABLE IF NOT EXISTS developers (
    apps text,
    attributes text,
    created_at bigint,
    created_by text,
    developer_id text,
    email text,
    first_name text,
    last_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    suspended_till bigint,
    status text,
    user_name text,
    PRIMARY KEY (developer_id)
	)`,

	`CREATE INDEX IF NOT EXISTS ON developers (email)`,
	`CREATE INDEX IF NOT EXISTS ON developers (first_name)`,
	`CREATE INDEX IF NOT EXISTS ON developers (last_name)`,
	`CREATE INDEX IF NOT EXISTS ON developers (lastmodified_at)`,
	`CREATE INDEX IF NOT EXISTS ON developers (status)`,
	`CREATE INDEX IF NOT EXISTS ON developers (user_name)`,

	`CREATE TABLE IF NOT EXISTS developer_apps (
    app_id text,
    attributes text,
    created_at bigint,
    created_by text,
    developer_id text,
    display_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    status text,
    PRIMARY KEY (app_id)
	)`,

	`CREATE INDEX IF NOT EXISTS ON developer_apps (name)`,
	`CREATE INDEX IF NOT EXISTS ON developer_apps (developer_id)`,
	`CREATE INDEX IF NOT EXISTS ON developer_apps (status)`,

	`CREATE TABLE IF NOT EXISTS keys (
    api_products text,
    consumer_key text,
    consumer_secret text,
    app_id text,
    expires_at bigint,
    issued_at bigint,
    status text,
    PRIMARY KEY (consumer_key)
	)`,

	`CREATE INDEX IF NOT EXISTS ON keys (consumer_secret)`,
	`CREATE INDEX IF NOT EXISTS ON keys (app_id)`,
	`CREATE INDEX IF NOT EXISTS ON keys (status)`,

	`CREATE TABLE IF NOT EXISTS oauth_access_token (
        access text,
        access_created_at bigint,
        access_expires_in bigint,
        client_id text,
        code text,
        code_created_at bigint,
        code_expires_in bigint,
        redirect_uri text,
        refresh text,
        refresh_created_at bigint,
        refresh_expires_in bigint,
        scope text,
        user_id text,
        PRIMARY KEY (access)
        )`,

	`CREATE INDEX IF NOT EXISTS ON oauth_access_token (refresh)`,
	`CREATE INDEX IF NOT EXISTS ON oauth_access_token (code)`,

	`CREATE TABLE IF NOT EXISTS api_products (
    attributes text,
    created_at bigint,
    created_by text,
    description text,
    display_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    paths text,
    policies text,
    route_group text,
	PRIMARY KEY (name)
	)`,
}
