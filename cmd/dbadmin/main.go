package main

import (
	"flag"

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
	applicationName = "dbadmin"
)

type server struct {
	config    *DBAdminConfig
	ginEngine *gin.Engine
	db        *db.Database
	readiness *shared.Readiness
}

func main() {
	shared.StartLogging(applicationName, version, buildTime)

	filename := flag.String("config", defaultConfigFileName, "Configuration filename")
	flag.Parse()

	s := server{
		config: loadConfiguration(filename),
	}

	shared.SetLoggingConfiguration(s.config.LogLevel)
	// s.readiness.RegisterMetrics(applicationName)

	var err error
	s.db, err = cassandra.New(s.config.Database, applicationName)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	s.readiness = shared.StartReadiness(applicationName)
	go s.db.RunReadinessCheck(s.readiness.GetChannel())

	StartWebAdminServer(&s)

}
