package main

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
)

const (
	myName = "envoycp"
)

type server struct {
	config               *EnvoyCPConfig
	ginEngine            *gin.Engine
	db                   *db.Database
	xds                  xds.Server
	xdsCache             cache.SnapshotCache
	authLatencyHistogram prometheus.Summary
	readyness            types.Readyness
}

func main() {
	s := server{}
	s.config = loadConfiguration()
	// FIXME we should check if we have all required parameters (use viper package?)

	types.SetLoggingConfiguration(s.config.LogLevel)

	var err error
	s.db, err = db.Connect(s.config.Database, myName)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	go s.StartWebAdminServer()
	s.readyness.Up()

	s.StartXDS()
}
