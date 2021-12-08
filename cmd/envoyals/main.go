package main

import (
	"flag"
	"log"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/envoyals/metrics"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

var (
	version   string // Git version of build, set by Makefile
	buildTime string // Build time, set by Makefile
)

type server struct {
	config   *EnvoyALSConfig
	webadmin *webadmin.Webadmin
	metrics  *metrics.Metrics
	logger   *zap.Logger
}

func main() {
	const applicationName = "envoyals"

	filename := flag.String("config", "envoyals-config.yaml", "Configuration filename")
	flag.Parse()

	var s server
	var err error
	if s.config, err = loadConfiguration(filename); err != nil {
		log.Fatalf("Cannot parse configuration file: (%s)", err)
	}

	logConfig := &shared.Logger{
		Level:    s.config.Logger.Level,
		Filename: s.config.Logger.Filename,
	}
	s.logger = shared.NewLogger(logConfig)
	s.logger.Info("Starting",
		zap.String("application", applicationName),
		zap.String("version", version),
		zap.String("buildtime", buildTime))

	s.metrics = metrics.New(applicationName)
	s.metrics.RegisterWithPrometheus()

	go startWebAdmin(&s, applicationName)

	accessLogLogger := shared.NewLogger(&s.config.AccessLogger.Logger)

	accessLogServer := NewAccessLogServer(s.config.AccessLogger.MaxStreamDuration,
		s.metrics, accessLogLogger)
	accessLogServer.Start(s.config.AccessLogger.Listen)
}

// startWebAdmin starts the admin web UI
func startWebAdmin(s *server, applicationName string) {

	webAdminLogger := shared.NewLogger(&s.config.WebAdmin.Logger)
	s.webadmin = webadmin.New(s.config.WebAdmin, applicationName, webAdminLogger)

	// Enable showing indexpage on / that shows all possible routes
	s.webadmin.Router.GET("/", webadmin.ShowAllRoutes(s.webadmin.Router, applicationName))
	s.webadmin.Router.GET(webadmin.LivenessCheckPath, webadmin.LivenessProbe)
	// s.webadmin.Router.GET(webadmin.ReadinessCheckPath, s.readiness.ReadinessProbe)
	s.webadmin.Router.GET(webadmin.MetricsPath, s.metrics.GinHandler())
	s.webadmin.Router.GET(webadmin.ConfigDumpPath, webadmin.ShowStartupConfiguration(s.config))

	s.webadmin.Start()
}
