package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

type webAdminConfig struct {
	Listen      string `yaml:"listen"`      // Address and port to listen
	IPACL       string `yaml:"ipacl"`       // ip accesslist (e.g. "10.0.0.0/8,192.168.0.0/16")
	LogFileName string `yaml:"logfilename"` // Filename for writing admin access logs
}

// StartWebAdminServer starts the admin web UI
func (s *server) StartWebAdminServer() {

	if logFile, err := os.Create(s.config.WebAdmin.LogFileName); err == nil {
		gin.DefaultWriter = io.MultiWriter(logFile)
	}

	gin.SetMode(gin.ReleaseMode)

	s.ginEngine = gin.New()
	s.ginEngine.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	s.ginEngine.Use(shared.AddRequestID())
	s.ginEngine.Use(shared.WebAdminCheckIPACL(s.config.WebAdmin.IPACL))

	s.ginEngine.GET("/", s.ShowWebAdminHomePage)
	s.ginEngine.GET(shared.LivenessCheckPath, shared.LivenessProbe)
	s.ginEngine.GET(shared.ReadinessCheckPath, s.readiness.ReadinessProbe)
	s.ginEngine.GET(shared.MetricsPath, gin.WrapH(promhttp.Handler()))
	s.ginEngine.GET(shared.ConfigDumpPath, s.ConfigDump)

	log.Info("Webadmin listening on ", s.config.WebAdmin.Listen)
	log.Fatal(s.ginEngine.Run(s.config.WebAdmin.Listen))
}

// ShowWebAdminHomePage shows home page
func (s *server) ShowWebAdminHomePage(c *gin.Context) {
	// FIXME feels like hack, is there a better way to pass gin engine context?
	shared.ShowIndexPage(c, s.ginEngine, applicationName)
}

// configDump pretty prints the active configuration (without password)
func (s *server) ConfigDump(c *gin.Context) {

	c.Header("Content-type", "text/yaml")
	c.String(http.StatusOK, fmt.Sprint(s.config))
}
