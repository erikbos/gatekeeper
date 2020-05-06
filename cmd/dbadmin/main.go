package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/shared"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	version   string
	buildTime string
)

const (
	myName = "dbadmin"
)

type env struct {
	config    *DBAdminConfig
	db        *db.Database
	ginEngine *gin.Engine
	readiness shared.Readiness
}

func main() {
	shared.StartLogging(myName, version, buildTime)

	e := &env{}
	e.config = loadConfiguration()
	// FIXME we should check if we have all required parameters (use viper package?)

	shared.SetLoggingConfiguration(e.config.LogLevel)

	var err error
	e.db, err = db.Connect(e.config.Database, &e.readiness, myName)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	StartWebAdminServer(e)
}

// boiler plate for later log actual API user
func (e *env) whoAmI() string {
	return "rest-api@test"
}

func setLastModifiedHeader(c *gin.Context, timeStamp int64) {
	c.Header("Last-Modified",
		time.Unix(0, timeStamp*int64(time.Millisecond)).UTC().Format(http.TimeFormat))
}

// returnJSONMessage returns an error message
func returnJSONMessage(c *gin.Context, statusCode int, errorMessage error) {
	c.IndentedJSON(statusCode, gin.H{"message": fmt.Sprintf("%s", errorMessage)})
}
