package cassandra

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metricsCollection struct {
	hostname       string
	querySuccesful *prometheus.CounterVec
	queryNotFound  *prometheus.CounterVec
	queryFailed    *prometheus.CounterVec
	queryHistogram prometheus.Summary
}

func (m *metricsCollection) register(serviceName, hostName string) {

	m.hostname = hostName

	m.querySuccesful = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "database_lookup_hits_total",
			Help:      "Number of successful database lookups.",
		}, []string{"hostname", "table"})

	m.queryNotFound = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "database_lookup_misses_total",
			Help:      "Number of unsuccessful database lookups.",
		}, []string{"hostname", "table"})

	m.queryFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "database_lookup_failed_total",
			Help:      "Number of failed database lookups.",
		}, []string{"hostname", "table"})

	m.queryHistogram = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: serviceName + "_database_lookup_latency",
			Help: "Database lookup latency in seconds.",
			Objectives: map[float64]float64{
				0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001,
			},
		})

	prometheus.MustRegister(m.querySuccesful)
	prometheus.MustRegister(m.queryNotFound)
	prometheus.MustRegister(m.queryFailed)
	prometheus.MustRegister(m.queryHistogram)
}

// QuerySuccessful increase sucessful query counter
func (m *metricsCollection) QuerySuccessful(tableName string) {
	m.querySuccesful.WithLabelValues(m.hostname, tableName).Inc()
}

// QueryNotFound increases failed query counter
func (m *metricsCollection) QueryNotFound(tableName string) {
	m.queryNotFound.WithLabelValues(m.hostname, tableName).Inc()
}

// QueryFailed increases failed metric counter
func (m *metricsCollection) QueryFailed(tableName string) {
	m.queryFailed.WithLabelValues(m.hostname, tableName).Inc()
}
