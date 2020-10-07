package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

var (
	version   string // Git version of build, set by Makefile
	buildTime string // Build time, set by Makefile
)

const (
	applicationName       = "envoycp"             // Name of application, used in Prometheus metrics
	defaultConfigFileName = "envoycp-config.yaml" // Default configuration file
)

type server struct {
	config     *EnvoyCPConfig
	webadmin   *webadmin.Webadmin
	db         *db.Database
	dbentities *db.EntityCache
	readiness  *shared.Readiness
	metrics    *metrics
	logger     *zap.Logger
}

var log string

func main() {

	filename := flag.String("config", defaultConfigFileName, "Configuration filename")
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
	s.logger = shared.NewLogger(logConfig, true)
	s.logger.Info("Starting",
		zap.String("application", applicationName),
		zap.String("version", version),
		zap.String("buildtime", buildTime))

	if s.db, err = cassandra.New(s.config.Database, applicationName, s.logger, false, 0); err != nil {
		s.logger.Fatal("Database connect failed", zap.Error(err))
	}

	s.readiness = shared.StartReadiness(applicationName, s.logger)
	go s.db.RunReadinessCheck(s.readiness.GetChannel())

	s.metrics = newMetrics(&s)
	s.metrics.Start()

	go startWebAdmin(&s)

	// Start continously loading of virtual host, routes & cluster data
	entityCacheConf := db.EntityCacheConfig{
		RefreshInterval: s.config.XDS.ConfigCompileInterval,
		Notify:          make(chan db.EntityChangeNotification),
		Listener:        true,
		Route:           true,
		Cluster:         true,
	}
	s.dbentities = db.NewEntityCache(s.db, entityCacheConf, s.logger)
	s.dbentities.Start()

	// Start XDS control plane service
	x := newXDS(s, s.config.XDS, entityCacheConf.Notify)
	x.Start()
}

// startWebAdmin starts the admin web UI
func startWebAdmin(s *server) {

	logger := shared.NewLogger(&s.config.WebAdmin.Logger, false)

	s.webadmin = webadmin.New(s.config.WebAdmin, applicationName, logger)

	// Enable showing indexpage on / that shows all possible routes
	s.webadmin.Router.GET("/", webadmin.ShowAllRoutes(s.webadmin.Router, applicationName))
	s.webadmin.Router.GET(webadmin.LivenessCheckPath, webadmin.LivenessProbe)
	s.webadmin.Router.GET(webadmin.ReadinessCheckPath, s.readiness.ReadinessProbe)
	s.webadmin.Router.GET(webadmin.MetricsPath, gin.WrapH(promhttp.Handler()))
	s.webadmin.Router.GET(webadmin.ConfigDumpPath, webadmin.ShowStartupConfiguration(s.config))

	s.webadmin.Start()
}
