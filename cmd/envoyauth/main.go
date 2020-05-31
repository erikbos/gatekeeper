package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

var (
	version   string
	buildTime string
)

const (
	myName = "envoyauth"
)

type authorizationServer struct {
	config       *APIAuthConfig
	ginEngine    *gin.Engine
	readiness    shared.Readiness
	virtualhosts []shared.VirtualHost
	routes       []shared.Route
	db           *db.Database
	cache        *Cache
	g            *shared.Geoip
	metrics      metricsCollection
}

func main() {
	shared.StartLogging(myName, version, buildTime)

	a := authorizationServer{}
	a.config = loadConfiguration()
	// FIXME we should check if we have all required parameters (use viper package?)

	shared.SetLoggingConfiguration(a.config.LogLevel)
	a.readiness.RegisterMetrics(myName)

	var err error
	a.db, err = db.Connect(a.config.Database, &a.readiness, myName)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	a.cache = newCache(&a.config.Cache)

	if a.config.Geoip.Filename != "" {
		a.g, err = shared.OpenGeoipDatabase(a.config.Geoip.Filename)
		if err != nil {
			log.Fatalf("Geoip db load failed: %v", err)
		}
	}

	a.registerMetrics()
	go StartWebAdminServer(&a)
	go a.GetVirtualHostConfigFromDatabase()
	go a.GetRouteConfigFromDatabase()

	a.startGRPCAuthorizationServer()
}
