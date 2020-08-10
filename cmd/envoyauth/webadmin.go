package main

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

type webAdminConfig struct {
	Listen      string `yaml:"listen"`
	IPACL       string `yaml:"ipacl"`
	LogFileName string `yaml:"logfilename"`
}

// StartWebAdminServer starts the admin web UI
func StartWebAdminServer(a *authorizationServer) {

	if logFile, err := os.Create(a.config.WebAdmin.LogFileName); err == nil {
		gin.DefaultWriter = io.MultiWriter(logFile)
	}

	gin.SetMode(gin.ReleaseMode)

	a.ginEngine = gin.New()
	a.ginEngine.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	a.ginEngine.Use(shared.AddRequestID())
	a.ginEngine.Use(shared.WebAdminCheckIPACL(a.config.WebAdmin.IPACL))

	a.ginEngine.GET("/", a.ShowWebAdminHomePage)
	a.ginEngine.GET(shared.LivenessCheckPath, shared.LivenessProbe)
	a.ginEngine.GET(shared.ReadinessCheckPath, a.readiness.ReadinessProbe)
	a.ginEngine.GET(shared.MetricsPath, gin.WrapH(promhttp.Handler()))
	a.ginEngine.GET(shared.ConfigDumpPath, a.ConfigDump)

	log.Info("Webadmin listening on ", a.config.WebAdmin.Listen)
	if err := a.ginEngine.Run(a.config.WebAdmin.Listen); err != nil {
		log.Fatal(err)
	}
}

// ShowWebAdminHomePage shows home page
func (a *authorizationServer) ShowWebAdminHomePage(c *gin.Context) {
	// FIXME feels like hack, is there a better way to pass gin engine context?
	shared.ShowIndexPage(c, a.ginEngine, applicationName)
}

// ConfigDump pretty prints the active configuration
func (a *authorizationServer) ConfigDump(c *gin.Context) {
	// We must remove db password from configuration struct before showing
	configToPrint := a.config
	configToPrint.Database.Password = "[redacted]"

	configDump, err := yaml.Marshal(configToPrint)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Content-type", "text/yaml")
	c.String(http.StatusOK, string(configDump))
}
