package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all our metrics
type Metrics struct {
	accessLogNodeHits  *prometheus.CounterVec
	accessLogVHostHits *prometheus.CounterVec
	accessLogLatency   prometheus.Summary
}

// New returns a new Metrics instance
func New() *Metrics {

	return &Metrics{}
}

// Handler returns HTTP handler function that exposes metrics
func Handler() http.Handler {
	return promhttp.Handler()
}

// RegisterWithPrometheus registers our operational metrics
func (m *Metrics) RegisterWithPrometheus(metricNamespace string) {

	m.accessLogNodeHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "accesslog_received_node_total",
			Help:      "Total number of access log entries per node received.",
		}, []string{"id", "cluster"})
	prometheus.MustRegister(m.accessLogNodeHits)

	m.accessLogVHostHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "accesslog_received_vhost_total",
			Help:      "Total number of access log entries per vhost received.",
		}, []string{"hostname"})
	prometheus.MustRegister(m.accessLogVHostHits)

	m.accessLogLatency = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: metricNamespace,
			Name:      "accesslog_latency",
			Help:      "Access logging latency in seconds.",
			Objectives: map[float64]float64{
				0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001,
			},
		})
	prometheus.MustRegister(m.accessLogLatency)
}

// IncAccessLogNodeHits
func (m *Metrics) IncAccessLogNodeHits(nodeID, nodeCluster string) {

	m.accessLogNodeHits.WithLabelValues(nodeID, nodeCluster).Inc()
}

// IncAccessLogVHostHits counts access log entries that are received
func (m *Metrics) IncAccessLogVHostHits(hostname string) {

	m.accessLogVHostHits.WithLabelValues(hostname).Inc()
}

// ObserveAccesLogLatency add observation to access log latency histogram
func (m *Metrics) ObserveAccesLogLatency(d time.Duration) {

	m.accessLogLatency.Observe(float64(d))
}
