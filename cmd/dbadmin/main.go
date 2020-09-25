package main

import (
	"flag"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/handler"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

var (
	version   string // Git version of build, set by Makefile
	buildTime string // Build time, set by Makefile
)

const (
	applicationName       = "dbadmin"             // Name of application, used in Prometheus metrics
	defaultConfigFileName = "dbadmin-config.yaml" // Default configuration file
)

type server struct {
	config    *DBAdminConfig
	router    *gin.Engine
	db        *db.Database
	handler   *handler.Handler
	readiness *shared.Readiness
}

func main() {
	shared.StartLogging(applicationName, version, buildTime)

	filename := flag.String("config", defaultConfigFileName, "Configuration filename")
	createSchema := flag.Bool("createschema", false, "Create database schema if it does not exist")
	replicaCount := flag.Int("replicacount", 3, "Replica count to set for database schema")
	flag.Parse()

	s := server{
		config: loadConfiguration(filename),
	}

	shared.SetLoggingConfiguration(s.config.LogLevel)
	// s.readiness.RegisterMetrics(applicationName)

	var err error
	s.db, err = cassandra.New(s.config.Database, applicationName, *createSchema, *replicaCount)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	s.readiness = shared.StartReadiness(applicationName)
	go s.db.RunReadinessCheck(s.readiness.GetChannel())

	StartWebAdminServer(&s)

}
