package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/shared"
)

var (
	version   string
	buildTime string
)

const (
	myName = "envoyauth"
)

type authorizationServer struct {
	config       *APIAuthConfig
	ginEngine    *gin.Engine
	readiness    shared.Readiness
	virtualhosts []shared.VirtualHost
	routes       []shared.Route
	db           *db.Database
	c            *db.Cache
	g            *shared.Geoip
	metrics      struct {
		xdsDeployments       *prometheus.CounterVec
		authLatencyHistogram prometheus.Summary
	}
}

func main() {
	shared.StartLogging(myName, version, buildTime)

	a := authorizationServer{}
	a.config = loadConfiguration()
	// FIXME we should check if we have all required parameters (use viper package?)

	shared.SetLoggingConfiguration(a.config.LogLevel)
	a.readiness.RegisterMetrics(myName)

	var err error
	a.db, err = db.Connect(a.config.Database, &a.readiness, myName)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	a.c = db.CacheInit(myName, a.config.Cache.Size, a.config.Cache.TTL, a.config.Cache.NegativeTTL)

	a.g, err = shared.OpenGeoipDatabase(a.config.Geoip.Filename)
	if err != nil {
		log.Fatalf("Geoip db load failed: %v", err)
	}

	a.registerMetrics()
	go StartWebAdminServer(&a)
	go a.GetVirtualHostConfigFromDatabase()
	go a.GetRouteConfigFromDatabase()

	startGRPCAuthorizationServer(a)
}

// registerMetrics registers our operational metrics
func (a *authorizationServer) registerMetrics() {
	// metricVirtualHostsCount := prometheus.NewGaugeFunc(
	// 	prometheus.GaugeOpts{
	// 		Name: myName + "_xds_virtualhosts_total",
	// 		Help: "Total number of clusters.",
	// 	}, s.GetVirtualHostCount)

	// metricRoutesCount := prometheus.NewGaugeFunc(
	// 	prometheus.GaugeOpts{
	// 		Name: myName + "_xds_routes_total",
	// 		Help: "Total number of routes.",
	// 	}, s.GetRouteCount)

	// metricClustersCount := prometheus.NewGaugeFunc(
	// 	prometheus.GaugeOpts{
	// 		Name: myName + "_xds_clusters_total",
	// 		Help: "Total number of clusters.",
	// 	}, s.GetClusterCount)
	a.metrics.authLatencyHistogram = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: myName + "_request_latency",
			Help: "Authentication latency in seconds.",
			Objectives: map[float64]float64{
				0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001,
			},
		})

	a.metrics.xdsDeployments = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: myName + "_config_table_loads_total",
			Help: "Total number of vhost/route table loads.",
		}, []string{"resource"})

	// prometheus.MustRegister(metricVirtualHostsCount)
	// prometheus.MustRegister(metricRoutesCount)
	// prometheus.MustRegister(metricClustersCount)
	prometheus.MustRegister(a.metrics.authLatencyHistogram)
	prometheus.MustRegister(a.metrics.xdsDeployments)
}
