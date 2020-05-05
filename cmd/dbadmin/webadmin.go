package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/erikbos/apiauth/pkg/shared"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type webAdminConfig struct {
	Listen  string `yaml:"listen"`
	LogFile string `yaml:"logfile"`
}

// StartWebAdminServer starts the admin web UI
func StartWebAdminServer(e *env) {
	if logFile, err := os.Create(e.config.WebAdmin.LogFile); err == nil {
		gin.DefaultWriter = io.MultiWriter(logFile)
	}

	if e.config.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	e.ginEngine = gin.New()
	e.ginEngine.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))

	e.registerOrganizationRoutes(e.ginEngine)
	e.registerDeveloperRoutes(e.ginEngine)
	e.registerDeveloperAppRoutes(e.ginEngine)
	e.registerCredentialRoutes(e.ginEngine)
	e.registerAPIProductRoutes(e.ginEngine)
	e.registerClusterRoutes(e.ginEngine)
	e.registerRouteRoutes(e.ginEngine)
	e.registerVirtualHostRoutes(e.ginEngine)

	e.ginEngine.GET("/", e.ShowWebAdminHomePage)
	e.ginEngine.GET("/ready", e.readiness.DisplayStatus)
	e.ginEngine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	e.ginEngine.GET("/config_dump", e.ConfigDump)
	e.ginEngine.Static("/assets", "./assets")

	log.Info("Webadmin listening on ", e.config.WebAdmin.Listen)
	if err := e.ginEngine.Run(e.config.WebAdmin.Listen); err != nil {
		log.Fatal(err)
	}
}

// ShowWebAdminHomePage shows home page
func (e *env) ShowWebAdminHomePage(c *gin.Context) {
	// FIXME feels like hack, is there a better way to pass gin engine context?
	shared.ShowIndexPage(c, e.ginEngine, myName)
}

// configDump pretty prints the active configuration
func (e *env) ConfigDump(c *gin.Context) {
	// We must remove db password from configuration struct before showing
	configToPrint := e.config
	configToPrint.Database.Password = ""

	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(configToPrint)
	if err != nil {
		return
	}

	c.Header("Content-type", "text/json")
	c.String(http.StatusOK, buffer.String())
}
