package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	xdsEntities  *prometheus.GaugeVec
	xdsSnapshots *prometheus.CounterVec
	xdsMessages  *prometheus.CounterVec
}

func New() *Metrics {

	return &Metrics{}
}

// RegisterWithPrometheus registers envoycp operational metrics
func (m *Metrics) RegisterWithPrometheus(metricNamespace string) {

	m.xdsEntities = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "xds_entities_total",
			Help:      "Total number of xds entities.",
		}, []string{"messagetype"})
	prometheus.MustRegister(m.xdsEntities)

	m.xdsSnapshots = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "xds_snapshots_total",
			Help:      "Total number of xds snapshots created.",
		}, []string{"resource"})
	prometheus.MustRegister(m.xdsSnapshots)

	m.xdsMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "xds_resource_requests_total",
			Help:      "Total number of xds messages.",
		}, []string{"messagetype"})
	prometheus.MustRegister(m.xdsMessages)
}

// SetEntityCount sets number of listeners we know
func (m *Metrics) SetEntityCount(label string, count int) {

	m.xdsEntities.WithLabelValues(label).Set(float64(count))
}

// IncXDSSnapshotCreateCount increases number of snapshots taken
func (m *Metrics) IncXDSSnapshotCreateCount(messageType string) {

	m.xdsSnapshots.WithLabelValues(messageType).Inc()
}

// IncXDSMessageCount increases counter per XDS messageType
func (m *Metrics) IncXDSMessageCount(messageType string) {

	m.xdsMessages.WithLabelValues(messageType).Inc()
}
