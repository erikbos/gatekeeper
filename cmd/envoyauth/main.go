package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/envoyauth/oauth"
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
	applicationName       = "envoyauth"             // Name of application, used in Prometheus metrics
	defaultConfigFileName = "envoyauth-config.yaml" // Default configuration file
	entityRefreshInterval = 3 * time.Second         // interval between database entities refresh loads
)

type authorizationServer struct {
	config     *APIAuthConfig
	webadmin   *webadmin.Webadmin
	db         *db.Database
	dbentities *db.EntityCache
	vhosts     *vhostMapping
	oauth      *oauth.Server
	geoip      *Geoip
	readiness  *shared.Readiness
	metrics    *metrics
	logger     *zap.Logger
}

func main() {
	filename := flag.String("config", defaultConfigFileName, "Configuration filename")
	flag.Parse()

	var a authorizationServer
	var err error
	if a.config, err = loadConfiguration(filename); err != nil {
		fmt.Print("Cannot parse configuration file:")
		panic(err)
	}

	logConfig := &shared.Logger{
		Level:    a.config.Logger.Level,
		Filename: a.config.Logger.Filename,
	}
	a.logger = shared.NewLogger(logConfig)
	a.logger.Info("Starting",
		zap.String("application", applicationName),
		zap.String("version", version),
		zap.String("buildtime", buildTime))

	a.metrics = newMetrics()
	a.metrics.RegisterWithPrometheus()

	database, err := cassandra.New(a.config.Database, applicationName, a.logger, false, 0)
	if err != nil {
		a.logger.Fatal("Database connect failed", zap.Error(err))
	}

	// Wrap database access with cache layer
	a.db, err = cache.New(&a.config.Cache, database, applicationName, a.logger)
	if err != nil {
		a.logger.Fatal("Database cache setup failed", zap.Error(err))
	}

	if a.config.Geoip.Database != "" {
		a.geoip, err = OpenGeoipDatabase(a.config.Geoip.Database)
		if err != nil {
			a.logger.Fatal("Geoip db load failed", zap.Error(err))
		}
	}

	// Start readiness subsystem
	a.readiness = shared.NewReadiness(applicationName, a.logger)
	a.readiness.Start()

	// Start db health check and notify readiness subsystem
	go a.db.RunReadinessCheck(a.readiness.GetChannel())

	go startWebAdmin(&a)

	// Start continously loading of virtual host, routes & cluster data
	entityCacheConf := db.EntityCacheConfig{
		RefreshInterval: entityRefreshInterval,
		Notify:          make(chan db.EntityChangeNotification),
	}
	a.dbentities = db.NewEntityCache(a.db, entityCacheConf, a.logger)
	a.dbentities.Start()

	a.vhosts = newVhostMapping(a.dbentities, a.logger)
	go a.vhosts.WaitFor(entityCacheConf.Notify)

	// // Start service for OAuth2 endpoints
	a.oauth = oauth.New(a.config.OAuth, a.db, a.logger)
	go func() {
		log.Fatal(a.oauth.Start(applicationName))
	}()

	a.StartAuthorizationServer()
}

// startWebAdmin starts the admin web UI
func startWebAdmin(s *authorizationServer) {

	logger := shared.NewLogger(&s.config.WebAdmin.Logger)

	s.webadmin = webadmin.New(s.config.WebAdmin, applicationName, logger)

	// Enable showing indexpage on / that shows all possible routes
	s.webadmin.Router.GET("/", webadmin.ShowAllRoutes(s.webadmin.Router, applicationName))
	s.webadmin.Router.GET(webadmin.LivenessCheckPath, webadmin.LivenessProbe)
	s.webadmin.Router.GET(webadmin.ReadinessCheckPath, s.readiness.ReadinessProbe)
	s.webadmin.Router.GET(webadmin.MetricsPath, gin.WrapH(promhttp.Handler()))
	s.webadmin.Router.GET(webadmin.ConfigDumpPath, webadmin.ShowStartupConfiguration(s.config))

	s.webadmin.Start()
}
