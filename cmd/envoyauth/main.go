package main

import (
	"flag"
	"fmt"
	"time"

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
	applicationName       = "envoyauth"             // Name of application, used in Prometheus metrics
	defaultConfigFileName = "envoyauth-config.yaml" // Default configuration file
	entityRefreshInterval = 3 * time.Second         // interval between database entities refresh loads
)

type authorizationServer struct {
	config   *APIAuthConfig
	webadmin *webadmin.Webadmin
	// router     *gin.Engine
	db         *db.Database
	dbentities *db.EntityCache
	vhosts     *vhostMapping
	cache      *Cache
	oauth      *oauthServer
	geoip      *Geoip
	readiness  *shared.Readiness
	metrics    metricsCollection
	logger     *zap.Logger
}

func main() {
	filename := flag.String("config", defaultConfigFileName, "Configuration filename")
	flag.Parse()

	var a authorizationServer
	var err error
	if a.config, err = loadConfiguration(filename); err != nil {
		fmt.Print(err)
		panic(err)
	}

	logConfig := &shared.Logger{
		Level:    a.config.Logger.Level,
		Filename: a.config.Logger.Filename,
	}
	a.logger = shared.NewLogger(logConfig, true)
	a.logger.Info("Starting",
		zap.String("application", applicationName),
		zap.String("version", version),
		zap.String("buildtime", buildTime))

	a.db, err = cassandra.New(a.config.Database, applicationName, a.logger, false, 0)
	if err != nil {
		a.logger.Fatal("Database connect failed", zap.Error(err))
	}
	a.cache = newCache(&a.config.Cache)

	if a.config.Geoip.Database != "" {
		a.geoip, err = OpenGeoipDatabase(a.config.Geoip.Database)
		if err != nil {
			a.logger.Fatal("Geoip db load failed", zap.Error(err))
		}
	}

	a.readiness = shared.StartReadiness(applicationName, a.logger)
	go a.db.RunReadinessCheck(a.readiness.GetChannel())

	a.registerMetrics()
	go startWebAdmin(&a)

	// Start continously loading of virtual host, routes & cluster data
	entityCacheConf := db.EntityCacheConfig{
		RefreshInterval: entityRefreshInterval,
		Notify:          make(chan db.EntityChangeNotification),
		Listener:        true,
		Route:           true,
		Cluster:         true,
	}
	a.dbentities = db.NewEntityCache(a.db, entityCacheConf, a.logger)
	a.dbentities.Start()

	a.vhosts = newVhostMapping(a.dbentities, a.logger)
	go a.vhosts.WaitFor(entityCacheConf.Notify)

	// // Start service for OAuth2 endpoints
	a.oauth = newOAuthServer(&a.config.OAuth, a.db, a.cache)
	go a.oauth.Start()

	a.StartAuthorizationServer()
}

// startWebAdmin starts the admin web UI
func startWebAdmin(s *authorizationServer) {

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
