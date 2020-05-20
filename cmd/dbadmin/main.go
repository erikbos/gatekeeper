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
	myName = "dbadmin"
)

type server struct {
	config    *DBAdminConfig
	ginEngine *gin.Engine
	db        *db.Database
	readiness shared.Readiness
}

func main() {
	shared.StartLogging(myName, version, buildTime)

	srv := &server{}
	srv.config = loadConfiguration()
	// FIXME we should check if we have all required parameters (use viper package?)

	shared.SetLoggingConfiguration(srv.config.LogLevel)
	srv.readiness.RegisterMetrics(myName)

	var err error
	srv.db, err = db.Connect(srv.config.Database, &srv.readiness, myName)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	StartWebAdminServer(srv, &srv.config.WebAdmin)

}

// boiler plate for later log actual API user
func (s *server) whoAmI() string {
	return "rest-api@test"
}
