package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/handler"
	"github.com/erikbos/gatekeeper/cmd/dbadmin/service"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/db/cache"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

var (
	version   string // Git version of build, set by Makefile
	buildTime string // Build time, set by Makefile
)

const (
	applicationName       = "dbadmin"             // Name of application, used in Prometheus metrics
	defaultConfigFileName = "dbadmin-config.yaml" // Default configuration file
)

type server struct {
	config    *DBAdminConfig
	webadmin  *webadmin.Webadmin
	db        *db.Database
	handler   *handler.Handler
	readiness *shared.Readiness
	logger    *zap.Logger
}

func main() {
	filename := flag.String("config", defaultConfigFileName, "Configuration filename")
	createSchema := flag.Bool("createschema", false, "Create database schema if it does not exist")
	replicaCount := flag.Int("replicacount", 3, "Replica count to set for database schema")
	enableAPIAuthentication := flag.Bool("enableapiauthentication", false, "Enable REST API authentication")
	flag.Parse()

	var s server
	var err error
	if s.config, err = loadConfiguration(filename); err != nil {
		fmt.Print(err)
		panic(err)
	}

	logConfig := &shared.Logger{
		Level:    s.config.Logger.Level,
		Filename: s.config.Logger.Filename,
	}
	s.logger = shared.NewLogger(logConfig)
	s.logger.Info("Starting",
		zap.String("application", applicationName),
		zap.String("version", version),
		zap.String("buildtime", buildTime))

	// s.readiness.RegisterMetrics(applicationName)

	// Connect to db
	db, err := cassandra.New(s.config.Database, applicationName,
		s.logger, *createSchema, *replicaCount)
	if err != nil {
		s.logger.Fatal("Database connect failed", zap.Error(err))
	}

	// Wrap database access with cache layer
	s.db, err = cache.New(&s.config.Cache, db, applicationName, s.logger)
	if err != nil {
		s.logger.Fatal("Database cache setup failed", zap.Error(err))
	}

	// Start readiness subsystem
	s.readiness = shared.NewReadiness(applicationName, s.logger)
	s.readiness.Start()

	// Start db health check and notify readiness subsystem
	go s.db.RunReadinessCheck(s.readiness.GetChannel())

	startWebAdmin(&s, *enableAPIAuthentication)
}

// startWebAdmin starts the admin web UI
func startWebAdmin(s *server, enableAPIAuthentication bool) {

	webAdminLogger := shared.NewLogger(&s.config.WebAdmin.Logger)
	s.webadmin = webadmin.New(s.config.WebAdmin, applicationName, webAdminLogger)

	// Enable showing indexpage on / that shows all possible routes
	s.webadmin.Router.GET("/", webadmin.ShowAllRoutes(s.webadmin.Router, applicationName))
	s.webadmin.Router.GET(webadmin.LivenessCheckPath, webadmin.LivenessProbe)
	s.webadmin.Router.GET(webadmin.ReadinessCheckPath, s.readiness.ReadinessProbe)
	s.webadmin.Router.GET(webadmin.MetricsPath, gin.WrapH(promhttp.Handler()))
	s.webadmin.Router.GET(webadmin.ConfigDumpPath, webadmin.ShowStartupConfiguration(s.config))

	changeLogLogger := shared.NewLogger(&s.config.Changelog.Logger)

	service := service.New(s.db, changeLogLogger)
	s.handler = handler.NewHandler(s.webadmin.Router, s.db, service, webAdminLogger, enableAPIAuthentication)

	s.webadmin.Start()
}
