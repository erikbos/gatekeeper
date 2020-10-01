package main

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/handler"
	"github.com/erikbos/gatekeeper/cmd/dbadmin/service"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type webAdminConfig struct {
	Listen  string `yaml:"listen"`  // Address and port to listen
	IPACL   string `yaml:"ipacl"`   // ip accesslist (e.g. "10.0.0.0/8,192.168.0.0/16")
	LogFile string `yaml:"logfile"` // File for writing admin access logs
	TLS     struct {
		certFile string `yaml:"certfile"` // TLS certifcate file
		keyFile  string `yaml:"keyfile"`  // TLS certifcate key file
	} `yaml:"tls"`
}

// StartWebAdminServer starts the admin web UI
func StartWebAdminServer(s *server, enableAPIAuthentication bool) {

	if logFile, err := os.Create(s.config.WebAdmin.LogFile); err == nil {
		gin.DefaultWriter = io.MultiWriter(logFile)
	}

	// disable debuglogging
	gin.SetMode(gin.ReleaseMode)

	// Enable strict field checking of POSTed JSON
	gin.EnableJsonDecoderDisallowUnknownFields()

	s.router = gin.New()
	s.router.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	s.router.Use(shared.AddRequestID())
	s.router.Use(shared.WebAdminCheckIPACL(s.config.WebAdmin.IPACL))

	service := service.NewService(s.db)
	s.handler = handler.NewHandler(s.router, s.db, service, enableAPIAuthentication)

	s.router.GET("/", s.ShowWebAdminHomePage)
	s.router.GET(shared.LivenessCheckPath, shared.LivenessProbe)
	s.router.GET(shared.ReadinessCheckPath, s.readiness.ReadinessProbe)
	s.router.GET(shared.MetricsPath, gin.WrapH(promhttp.Handler()))
	s.router.GET(shared.ConfigDumpPath, s.showConfiguration)
	s.router.GET(showHTTPForwardingPath, s.showHTTPForwarding)

	log.Info("Webadmin listening on ", s.config.WebAdmin.Listen)
	if s.config.WebAdmin.TLS.certFile != "" &&
		s.config.WebAdmin.TLS.keyFile != "" {

		log.Fatal(s.router.RunTLS(s.config.WebAdmin.Listen,
			s.config.WebAdmin.TLS.certFile, s.config.WebAdmin.TLS.keyFile))
	}
	log.Fatal(s.router.Run(s.config.WebAdmin.Listen))
}
