package main

import (
	"flag"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

var (
	version   string // Git version of build, set by Makefile
	buildTime string // Build time, set by Makefile
)

const (
	applicationName       = "envoyauth"             // Name of application, used in Prometheus metrics
	defaultConfigFileName = "envoyauth-config.yaml" // Default configuration file
	entityLoadInterval    = 3 * time.Second         // interval between database entities refreshloads
)

type authorizationServer struct {
	config     *APIAuthConfig
	ginEngine  *gin.Engine
	db         *db.Database
	dbentities *db.Entityloader
	vhosts     *vhostMapping
	cache      *Cache
	oauth      *oauthServer
	geoip      *Geoip
	readiness  *shared.Readiness
	metrics    metricsCollection
}

func main() {
	shared.StartLogging(applicationName, version, buildTime)

	filename := flag.String("config", defaultConfigFileName, "Configuration filename")
	flag.Parse()

	a := authorizationServer{}
	a.config = loadConfiguration(filename)

	shared.SetLoggingConfiguration(a.config.LogLevel)

	var err error
	a.db, err = cassandra.New(a.config.Database, applicationName, false, 0)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}
	a.cache = newCache(&a.config.Cache)

	if a.config.Geoip.Database != "" {
		a.geoip, err = OpenGeoipDatabase(a.config.Geoip.Database)
		if err != nil {
			log.Fatalf("Geoip db load failed: %v", err)
		}
	}

	a.readiness = shared.StartReadiness(applicationName)
	go a.db.RunReadinessCheck(a.readiness.GetChannel())

	a.registerMetrics()
	go StartWebAdminServer(&a)

	// Start continously loading of virtual host, routes & cluster data
	a.dbentities = db.NewEntityLoader(a.db, entityLoadInterval)
	a.dbentities.Start()

	a.vhosts = newVhostMapping(a.dbentities)
	go a.vhosts.WaitFor(a.dbentities.GetChannel())

	// // Start service for OAuth2 endpoints
	a.oauth = newOAuthServer(&a.config.OAuth, a.db, a.cache)
	go a.oauth.Start()

	a.StartAuthorizationServer()
}
