package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type env struct {
	config    *DBAdminConfig
	db        *db.Database
	ginEngine *gin.Engine
	readyness types.Readyness
}

func main() {
	e := &env{}
	e.config = loadConfiguration()
	// FIXME we should check if we have all required parameters (use viper package?)

	types.SetLoggingConfiguration(e.config.LogLevel)

	var err error
	e.db, err = db.Connect(e.config.Database.Hostname, e.config.Database.Port,
		e.config.Database.Username, e.config.Database.Password, e.config.Database.Keyspace)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	if e.config.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	e.ginEngine = gin.New()
	e.ginEngine.Use(gin.LoggerWithFormatter(logRequstparam))

	e.registerOrganizationRoutes(e.ginEngine)
	e.registerDeveloperRoutes(e.ginEngine)
	e.registerDeveloperAppRoutes(e.ginEngine)
	e.registerCredentialRoutes(e.ginEngine)
	e.registerAPIProductRoutes(e.ginEngine)
	e.registerClusterRoutes(e.ginEngine)

	e.ginEngine.Static("/assets", "./assets")
	e.ginEngine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	e.ginEngine.GET("/", e.ShowWebAdminHomePage)
	e.ginEngine.GET("/ready", e.readyness.DisplayReadyness)

	e.readyness.Up()

	log.Info("Start listening on ", e.config.WebAdminListen)
	e.ginEngine.Run(e.config.WebAdminListen)
}

func logRequstparam(param gin.LogFormatterParams) string {
	return fmt.Sprintf("%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\"\n",
		param.ClientIP,
		param.TimeStamp.Format(time.RFC3339),
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency/time.Millisecond,
		param.Request.UserAgent(),
		param.ErrorMessage,
	)
}

// boiler plate for later log actual API user
func (e *env) whoAmI() string {
	return "rest-api@test"
}

// ShowWebAdminHomePage shows home page
func (e *env) ShowWebAdminHomePage(c *gin.Context) {
	// FIXME feels like hack, is there a better way to pass gin engine context?
	types.ShowIndexPage(c, e.ginEngine)
}

func setLastModifiedHeader(c *gin.Context, timeStamp int64) {
	c.Header("Last-Modified",
		time.Unix(0, timeStamp*int64(time.Millisecond)).UTC().Format(http.TimeFormat))
}

// returnJSONMessage returns an error message
func returnJSONMessage(c *gin.Context, statusCode int, errorMessage error) {
	c.IndentedJSON(statusCode, gin.H{"message": fmt.Sprintf("%s", errorMessage)})
}
