package main

import (
	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/shared"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	version   string
	buildTime string
)

const (
	myName = "envoyauth"
)

type authorizationServer struct {
	config    *APIAuthConfig
	ginEngine *gin.Engine
	readiness shared.Readiness
	db        *db.Database
	c         *db.Cache
	g         *shared.Geoip
	metrics   struct {
		authLatencyHistogram prometheus.Summary
	}
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

	a.c = db.CacheInit(a.config.Cache.Size, a.config.Cache.TTL, a.config.Cache.NegativeTTL)

	a.g, err = shared.OpenGeoipDatabase(a.config.Geoip.Filename)
	if err != nil {
		log.Fatalf("Geoip db load failed: %v", err)
	}

	go StartWebAdminServer(&a)
	startGRPCAuthenticationServer(a)
}
