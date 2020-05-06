package main

import (
	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/shared"

	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	version   string
	buildTime string
)

const (
	myName = "envoycp"
)

type server struct {
	config    *EnvoyCPConfig
	ginEngine *gin.Engine
	db        *db.Database
	xds       xds.Server
	xdsCache  cache.SnapshotCache
	// authLatencyHistogram prometheus.Summary
	readiness shared.Readiness
}

func main() {
	shared.StartLogging(myName, version, buildTime)

	s := server{}
	s.config = loadConfiguration()
	// FIXME we should check if we have all required parameters (use viper package?)

	shared.SetLoggingConfiguration(s.config.LogLevel)

	var err error
	s.db, err = db.Connect(s.config.Database, &s.readiness, myName)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	go s.StartWebAdminServer()

	s.StartXDS()
}
