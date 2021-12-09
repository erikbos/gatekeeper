package metrics

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

// Prometheus metrics endpoint
const MetricsPath = "/metrics"

type Metrics struct {
	applicationName    string
	requestPerPathHits *prometheus.CounterVec
}

func New(applicationName string) *Metrics {

	m := &Metrics{
		applicationName: applicationName,
	}
	m.RegisterWithPrometheus()
	return m
}

// Middleware is gin middleware to maintain prometheus metrics on user, paths and status codes
func (m *Metrics) Middleware() gin.HandlerFunc {

	return (func(c *gin.Context) {

		c.Next()

		m.IncRequestPathHit(webadmin.GetUser(c),
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(c.Writer.Status()))
	})
}

// GinHandler returns a Gin handler for Prometheus metrics endpoint
func (m *Metrics) GinHandler() gin.HandlerFunc {

	return gin.WrapH(promhttp.Handler())
}

// RegisterMetrics registers our operational metrics
func (m *Metrics) RegisterWithPrometheus() {

	m.requestPerPathHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.applicationName,
			Name:      "api_requests_hits_total",
			Help:      "Number of hits per request path.",
		}, []string{"user", "method", "path", "status"})
	prometheus.MustRegister(m.requestPerPathHits)
}

func (m *Metrics) IncRequestPathHit(user, method, path, status string) {

	m.requestPerPathHits.WithLabelValues(user, method, path, status).Inc()
}
