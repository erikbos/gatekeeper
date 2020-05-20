package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

type webAdminConfig struct {
	Listen  string `yaml:"listen"`
	IPACL   string `yaml:"ipacl"`
	LogFile string `yaml:"logfile"`
}

// StartWebAdminServer starts the admin web UI
func StartWebAdminServer(a *authorizationServer) {
	if logFile, err := os.Create(a.config.WebAdmin.LogFile); err == nil {
		gin.DefaultWriter = io.MultiWriter(logFile)
	}

	gin.SetMode(gin.ReleaseMode)

	a.ginEngine = gin.New()
	a.ginEngine.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	a.ginEngine.Use(shared.AddRequestID())
	a.ginEngine.Use(shared.WebAdminCheckIPACL(a.config.WebAdmin.IPACL))

	a.ginEngine.GET("/", a.ShowWebAdminHomePage)
	a.ginEngine.GET("/liveness", shared.LivenessProbe)
	a.ginEngine.GET("/readiness", a.readiness.ReadinessProbe)
	a.ginEngine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	a.ginEngine.GET("/config_dump", a.ConfigDump)

	log.Info("Webadmin listening on ", a.config.WebAdmin.Listen)
	if err := a.ginEngine.Run(a.config.WebAdmin.Listen); err != nil {
		log.Fatal(err)
	}
}

// ShowWebAdminHomePage shows home page
func (a *authorizationServer) ShowWebAdminHomePage(c *gin.Context) {
	// FIXME feels like hack, is there a better way to pass gin engine context?
	shared.ShowIndexPage(c, a.ginEngine, myName)
}

// ConfigDump pretty prints the active configuration
func (a *authorizationServer) ConfigDump(c *gin.Context) {
	// We must remove db password from configuration struct before showing
	configToPrint := a.config
	configToPrint.Database.Password = "[redacted]"

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
