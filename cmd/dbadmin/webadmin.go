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
func StartWebAdminServer(s *server) {

	if logFile, err := os.Create(s.config.WebAdmin.LogFileName); err == nil {
		gin.DefaultWriter = io.MultiWriter(logFile)
	}

	// disable debuglogging
	gin.SetMode(gin.ReleaseMode)

	// Enable strict checking of posted JSON fields
	gin.EnableJsonDecoderDisallowUnknownFields()

	s.ginEngine = gin.New()
	s.ginEngine.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	s.ginEngine.Use(shared.AddRequestID())
	s.ginEngine.Use(shared.WebAdminCheckIPACL(s.config.WebAdmin.IPACL))

	s.registerOrganizationRoutes(s.ginEngine)
	s.registerDeveloperRoutes(s.ginEngine)
	s.registerDeveloperAppRoutes(s.ginEngine)
	s.registerCredentialRoutes(s.ginEngine)
	s.registerAPIProductRoutes(s.ginEngine)
	s.registerClusterRoutes(s.ginEngine)
	s.registerRouteRoutes(s.ginEngine)
	s.registerVirtualHostRoutes(s.ginEngine)

	s.ginEngine.GET("/", s.ShowWebAdminHomePage)
	s.ginEngine.GET(shared.LivenessCheckPath, shared.LivenessProbe)
	s.ginEngine.GET(shared.ReadinessCheckPath, s.readiness.ReadinessProbe)
	s.ginEngine.GET(shared.MetricsPath, gin.WrapH(promhttp.Handler()))
	s.ginEngine.GET(shared.ConfigDumpPath, s.ConfigDump)

	log.Info("Webadmin listening on ", s.config.WebAdmin.Listen)
	if err := s.ginEngine.Run(s.config.WebAdmin.Listen); err != nil {
		log.Fatal(err)
	}
}

// ShowWebAdminHomePage shows home page
func (s *server) ShowWebAdminHomePage(c *gin.Context) {
	// FIXME feels like hack, is there a better way to pass gin engine context?
	shared.ShowIndexPage(c, s.ginEngine, applicationName)
}

// configDump pretty prints the active configuration
func (s *server) ConfigDump(c *gin.Context) {
	// We must remove db password from configuration struct before showing
	configToPrint := s.config
	configToPrint.Database.Password = "[redacted]"

	configDump, err := yaml.Marshal(configToPrint)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Content-type", "text/yaml")
	c.String(http.StatusOK, string(configDump))
}
