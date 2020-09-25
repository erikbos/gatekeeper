package cassandra

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metricsCollection struct {
	hostname            string
	LookupHitsCounter   *prometheus.CounterVec
	LookupMissesCounter *prometheus.CounterVec
	LookupFailedCounter *prometheus.CounterVec
	LookupHistogram     prometheus.Summary
}

func (m *metricsCollection) register(serviceName, hostName string) {

	m.hostname = hostName

	m.LookupHitsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "database_lookup_hits_total",
			Help:      "Number of successful database lookups.",
		}, []string{"hostname", "table"})

	m.LookupMissesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "database_lookup_misses_total",
			Help:      "Number of unsuccessful database lookups.",
		}, []string{"hostname", "table"})

	m.LookupFailedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "database_lookup_failed_total",
			Help:      "Number of failed database lookups.",
		}, []string{"hostname", "table"})

	m.LookupHistogram = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: serviceName + "_database_lookup_latency",
			Help: "Database lookup latency in seconds.",
			Objectives: map[float64]float64{
				0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001,
			},
		})

	prometheus.MustRegister(m.LookupHitsCounter)
	prometheus.MustRegister(m.LookupMissesCounter)
	prometheus.MustRegister(m.LookupFailedCounter)
	prometheus.MustRegister(m.LookupHistogram)
}

// QueryHit increase positive metric counter
func (m *metricsCollection) QueryHit(tableName string) {
	m.LookupHitsCounter.WithLabelValues(m.hostname, tableName).Inc()
}

// QueryMiss increases negative metric counter
func (m *metricsCollection) QueryMiss(tableName string) {
	m.LookupMissesCounter.WithLabelValues(m.hostname, tableName).Inc()
}

// QueryFailed increases failed metric counter
func (m *metricsCollection) QueryFailed(tableName string) {
	m.LookupFailedCounter.WithLabelValues(m.hostname, tableName).Inc()
}
