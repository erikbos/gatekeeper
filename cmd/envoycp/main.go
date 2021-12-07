package main

import (
	"flag"
	"fmt"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/envoycp/metrics"
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
	config     *EnvoyCPConfig
	webadmin   *webadmin.Webadmin
	db         *db.Database
	dbentities *db.EntityCache
	readiness  *shared.Readiness
	metrics    *metrics.Metrics
	logger     *zap.Logger
}

func main() {
	const applicationName = "envoycp"

	filename := flag.String("config", "envoycp-config.yaml", "Configuration filename")
	flag.Parse()

	var s server
	var err error
	if s.config, err = loadConfiguration(filename); err != nil {
		fmt.Print("Cannot parse configuration file:\n")
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

	s.metrics = metrics.New(applicationName)
	s.metrics.RegisterWithPrometheus()

	if s.db, err = cassandra.New(s.config.Database, applicationName, s.logger, false, 0); err != nil {
		s.logger.Fatal("Database connect failed", zap.Error(err))
	}

	// Start readiness subsystem
	s.readiness = shared.NewReadiness(applicationName, s.logger)
	s.readiness.Start()

	// Start db health check and notify readiness subsystem
	go s.db.RunReadinessCheck(s.readiness.GetChannel())

	go startWebAdmin(&s, applicationName)

	// Start continously loading of virtual host, routes & cluster data
	entityCacheConf := db.EntityCacheConfig{
		RefreshInterval: s.config.XDS.ConfigCompileInterval,
		Notify:          make(chan db.EntityChangeNotification),
	}
	s.dbentities = db.NewEntityCache(s.db, entityCacheConf, s.logger)
	s.dbentities.Start()

	// Start XDS control plane service
	x := newXDS(s, s.config.XDS, entityCacheConf.Notify)
	x.Start()
}

// startWebAdmin starts the admin web UI
func startWebAdmin(s *server, applicationName string) {

	logger := shared.NewLogger(&s.config.WebAdmin.Logger)

	s.webadmin = webadmin.New(s.config.WebAdmin, applicationName, logger)

	// Enable showing indexpage on / that shows all possible routes
	s.webadmin.Router.GET("/", webadmin.ShowAllRoutes(s.webadmin.Router, applicationName))
	s.webadmin.Router.GET(webadmin.LivenessCheckPath, webadmin.LivenessProbe)
	s.webadmin.Router.GET(webadmin.ReadinessCheckPath, s.readiness.ReadinessProbe)
	s.webadmin.Router.GET(webadmin.MetricsPath, s.metrics.GinHandler())
	s.webadmin.Router.GET(webadmin.ConfigDumpPath, webadmin.ShowStartupConfiguration(s.config))

	s.webadmin.Start()
}
