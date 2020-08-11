package main

import (
	"flag"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

var (
	version   string
	buildTime string
)

const (
	applicationName       = "envoycp"
	defaultConfigFileName = "envoycp-config.yaml"
)

type server struct {
	config       *EnvoyCPConfig
	ginEngine    *gin.Engine
	db           *db.Database
	readiness    *shared.Readiness
	virtualhosts []shared.VirtualHost
	routes       []shared.Route
	clusters     []shared.Cluster
	xds          xds.Server
	xdsCache     cache.SnapshotCache
	metrics      metricsCollection
}

func main() {
	shared.StartLogging(applicationName, version, buildTime)

	filename := flag.String("config", defaultConfigFileName, "Configuration filename")
	flag.Parse()

	s := server{
		config: loadConfiguration(filename),
	}

	shared.SetLoggingConfiguration(s.config.LogLevel)

	var err error
	s.db, err = cassandra.New(s.config.Database, applicationName, false, 0)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	s.readiness = shared.StartReadiness(applicationName)
	go s.db.RunReadinessCheck(s.readiness.GetChannel())

	s.registerMetrics()
	go s.StartWebAdminServer()

	xdsNotify := make(chan xdsNotifyMesssage)
	go s.GetVirtualHostConfigFromDatabase(xdsNotify)
	go s.GetRouteConfigFromDatabase(xdsNotify)
	go s.GetClusterConfigFromDatabase(xdsNotify)
	s.StartXDS(xdsNotify)
}
