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

	"github.com/erikbos/apiauth/pkg/shared"
)

type webAdminConfig struct {
	Listen  string `yaml:"listen"`
	LogFile string `yaml:"logfile"`
}

// StartWebAdminServer starts the admin web UI
func StartWebAdminServer(s *server, c *webAdminConfig) {
	if logFile, err := os.Create(c.LogFile); err == nil {
		gin.DefaultWriter = io.MultiWriter(logFile)
	}

	// disable debuglogging
	gin.SetMode(gin.ReleaseMode)

	// Enable strict checking of posted JSON fields
	gin.EnableJsonDecoderDisallowUnknownFields()

	s.ginEngine = gin.New()

	s.ginEngine.Use(shared.AddRequestID())
	s.ginEngine.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))

	s.registerOrganizationRoutes(s.ginEngine)
	s.registerDeveloperRoutes(s.ginEngine)
	s.registerDeveloperAppRoutes(s.ginEngine)
	s.registerCredentialRoutes(s.ginEngine)
	s.registerAPIProductRoutes(s.ginEngine)
	s.registerClusterRoutes(s.ginEngine)
	s.registerRouteRoutes(s.ginEngine)
	s.registerVirtualHostRoutes(s.ginEngine)

	s.ginEngine.GET("/", s.ShowWebAdminHomePage)
	s.ginEngine.GET("/ready", s.readiness.DisplayStatus)
	s.ginEngine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	s.ginEngine.GET("/config_dump", s.ConfigDump)
	s.ginEngine.Static("/assets", "./assets")

	log.Info("Webadmin listening on ", c.Listen)
	if err := s.ginEngine.Run(c.Listen); err != nil {
		log.Fatal(err)
	}
}

// ShowWebAdminHomePage shows home page
func (s *server) ShowWebAdminHomePage(c *gin.Context) {
	// FIXME feels like hack, is there a better way to pass gin engine context?
	shared.ShowIndexPage(c, s.ginEngine, myName)
}

// configDump pretty prints the active configuration
func (s *server) ConfigDump(c *gin.Context) {
	// We must remove db password from configuration struct before showing
	configToPrint := s.config
	// configToPrint.config.Password = ""

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
