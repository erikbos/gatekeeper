package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/metrics"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

// registerMetricsRoute registers prometheus route
func registerMetricsRoute(g *gin.Engine, applicationName string) {

	m := metrics.New()
	m.RegisterWithPrometheus(applicationName)
	g.Use(metricsMiddleware(m))
}

// metricsMiddleware is gin middleware to maintain prometheus metrics on user, paths and status codes
func metricsMiddleware(m *metrics.Metrics) gin.HandlerFunc {

	return func(c *gin.Context) {

		c.Next()

		m.IncRequestPathHit(webadmin.GetUser(c),
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(c.Writer.Status()))
	}
}
