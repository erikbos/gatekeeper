package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	requestPerPathHits *prometheus.CounterVec
}

func New() *Metrics {

	return &Metrics{}
}

// registerMetrics registers our operational metrics
func (m *Metrics) RegisterWithPrometheus(metricNamespace string) {

	m.requestPerPathHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "api_requests_hits_total",
			Help:      "Number of hits per request path.",
		}, []string{"user", "method", "path", "status"})
	prometheus.MustRegister(m.requestPerPathHits)
}

func (m *Metrics) IncRequestPathHit(user, method, path, status string) {

	m.requestPerPathHits.WithLabelValues(user, method, path, status).Inc()
}
