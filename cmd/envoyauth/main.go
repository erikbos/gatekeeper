package main

import (
	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
)

const (
	myName = "envoyauth"
)

type authorizationServer struct {
	config               *APIAuthConfig
	ginEngine            *gin.Engine
	db                   *db.Database
	c                    *db.Cache
	g                    *types.Geoip
	authLatencyHistogram prometheus.Summary
	readyness            types.Readyness
}

func main() {
	a := authorizationServer{}
	a.config = loadConfiguration()
	// FIXME we should check if we have all required parameters (use viper package?)

	types.SetLoggingConfiguration(a.config.LogLevel)

	var err error
	a.db, err = db.Connect(a.config.Database, myName)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	a.c = db.CacheInit(a.config.Cache.Size, a.config.Cache.TTL, a.config.Cache.NegativeTTL)

	a.g, err = types.OpenGeoipDatabase(a.config.Geoip.Filename)
	if err != nil {
		log.Fatalf("Geoip db load failed: %v", err)
	}

	StartWebAdminServer(&a)
	a.readyness.Up()
	startGRPCAuthenticationServer(a)
}
