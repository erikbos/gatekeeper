package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/erikbos/apiauth/pkg/shared"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// StartWebAdminServer starts the admin web UI
func (s *server) StartWebAdminServer() {
	if s.config.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	s.ginEngine = gin.New()
	s.ginEngine.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))

	s.ginEngine.GET("/", s.ShowWebAdminHomePage)
	s.ginEngine.GET("/ready", s.readiness.DisplayStatus)
	s.ginEngine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	s.ginEngine.GET("/config_dump", s.ConfigDump)

	log.Info("Webadmin listening on ", s.config.WebAdminListen)
	if err := s.ginEngine.Run(s.config.WebAdminListen); err != nil {
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