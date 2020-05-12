package main

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/shared"
)

var (
	version   string
	buildTime string
)

const (
	myName = "envoycp"
)

type server struct {
	config       *EnvoyCPConfig
	ginEngine    *gin.Engine
	db           *db.Database
	readiness    shared.Readiness
	virtualhosts []shared.VirtualHost
	routes       []shared.Route
	clusters     []shared.Cluster
	xds          xds.Server
	xdsCache     cache.SnapshotCache
	metrics      struct {
		xdsDeployments *prometheus.CounterVec
	}
}

func main() {
	shared.StartLogging(myName, version, buildTime)

	s := server{}
	s.config = loadConfiguration()
	// FIXME we should check if we have all required parameters (use viper package?)

	shared.SetLoggingConfiguration(s.config.LogLevel)
	s.readiness.RegisterMetrics(myName)

	var err error
	s.db, err = db.Connect(s.config.Database, &s.readiness, myName)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	s.registerMetrics()
	go s.StartWebAdminServer()
	go s.GetVirtualHostConfigFromDatabase()
	go s.GetRouteConfigFromDatabase()
	go s.GetClusterConfigFromDatabase()
	s.StartXDS()
}
