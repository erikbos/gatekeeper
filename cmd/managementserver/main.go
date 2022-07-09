package main

import (
	"flag"
	"log"
	"os"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/managementserver/audit"
	"github.com/erikbos/gatekeeper/cmd/managementserver/handler"
	"github.com/erikbos/gatekeeper/cmd/managementserver/metrics"
	"github.com/erikbos/gatekeeper/cmd/managementserver/service"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

var (
	version   string // Git version of build, set by Makefile
	buildTime string // Build time, set by Makefile
)

type server struct {
	config   *ManagementServerConfig
	webadmin *webadmin.Webadmin
	db       *db.Database
	handler  *handler.Handler
	metrics  *metrics.Metrics // Metrics store
	logger   *zap.Logger
}

func main() {
	const applicationName = "managementserver"

	filename := flag.String("config", "managementserver-config.yaml", "Configuration filename")
	disableAPIAuthentication := flag.Bool("disableapiauthentication", false, "Disable REST API authentication")
	createSchema := flag.Bool("createschema", false, "Create database schema if it does not exist")
	replicaCount := flag.Int("replicacount", 1, "Replica count to set for database keyspace")
	showCreateSchema := flag.Bool("showcreateschema", false, "Show CQL statements to create database")
	flag.Parse()

	if *showCreateSchema {
		cassandra.ShowCreateSchemaStatements()
		os.Exit(0)
	}

	var s server
	var err error
	if s.config, err = loadConfiguration(*filename); err != nil {
		log.Fatalf("Cannot parse configuration file: (%s)", err)
	}

	logConfig := &shared.Logger{
		Level:    s.config.Logger.Level,
		Filename: s.config.Logger.Filename,
	}
	s.logger = shared.NewLogger(applicationName, logConfig)
	s.logger.Info("Starting",
		zap.String("version", version),
		zap.String("buildtime", buildTime))

	s.metrics = metrics.New(applicationName)

	// Connect to database
	db, err := cassandra.New(s.config.Database, applicationName,
		s.logger, *createSchema, *replicaCount)
	if err != nil {
		s.logger.Fatal("Database connect failed", zap.Error(err))
	}
	s.db = db

	// Wrap database access with cache layer
	// s.db, err = cache.New(&s.config.Cache, db, applicationName, s.logger)
	// if err != nil {
	// 	s.logger.Fatal("Database cache setup failed", zap.Error(err))
	// }

	// Connect to audit database
	auditDb, err := cassandra.New(s.config.Audit.Database, applicationName+"_audit",
		s.logger, *createSchema, *replicaCount)
	if err != nil {
		s.logger.Fatal("Audit database connect failed", zap.Error(err))
	}

	auditLogLogger := shared.NewLogger("audit", &s.config.Audit.Logger)
	auditlog := audit.New(auditDb, auditLogLogger)

	s.webadmin = webadmin.New(s.config.WebAdmin, applicationName)
	s.webadmin.Router.Use(s.metrics.Middleware())

	s.webadmin.Router.GET("/", webadmin.ShowAllRoutes(s.webadmin.Router, applicationName))
	s.webadmin.Router.GET(webadmin.ReadinessCheckPath, webadmin.ReadinessProbe)
	s.webadmin.Router.GET(webadmin.MetricsPath, s.metrics.GinHandler())
	s.webadmin.Router.GET(webadmin.ConfigDumpPath, webadmin.ShowStartupConfiguration(s.config))

	service := service.New(s.db, auditlog)
	s.handler = handler.New(s.webadmin.Router, s.db, service,
		applicationName, *disableAPIAuthentication, s.logger)

	s.webadmin.Start()
}
