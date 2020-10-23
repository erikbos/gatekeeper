package handler

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	requestPerPathHits *prometheus.CounterVec
}

func newMetrics() *metrics {

	return &metrics{}
}

// registerMetrics registers our operational metrics
func (m *metrics) RegisterWithPrometheus(applicationName string) {
	m.requestPerPathHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "api_requests_hits_total",
			Help:      "Number of hits per request path.",
		}, []string{"user", "method", "path", "status"})
	prometheus.MustRegister(m.requestPerPathHits)
}

func (m *metrics) IncRequestPathHit(user, method, path, status string) {
	m.requestPerPathHits.WithLabelValues(user, method, path, status).Inc()
}
