package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	xdsEntities  *prometheus.GaugeVec
	xdsSnapshots *prometheus.CounterVec
	xdsMessages  *prometheus.CounterVec
}

func newMetrics(s *server) *metrics {

	return &metrics{}
}

// Start registers envoycp operational metrics
func (m *metrics) Start() {

	m.xdsEntities = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "xds_entities_total",
			Help:      "Total number of entities.",
		}, []string{"messagetype"})

	m.xdsSnapshots = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "xds_snapshots_total",
			Help:      "Total number of xds snapshots created.",
		}, []string{"resource"})

	m.xdsMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "xds_resource_requests_total",
			Help:      "Total number of XDS messages.",
		}, []string{"messagetype"})

	prometheus.MustRegister(m.xdsEntities)
	prometheus.MustRegister(m.xdsSnapshots)
	prometheus.MustRegister(m.xdsMessages)
}

// SetEntityCount sets number of listeners we know
func (m *metrics) SetEntityCount(label string, count int) {

	m.xdsEntities.WithLabelValues(label).Set(float64(count))
}

// IncXDSSnapshotCreateCount increases number of snapshots taken
func (m *metrics) IncXDSSnapshotCreateCount(messageType string) {

	m.xdsSnapshots.WithLabelValues(messageType).Inc()
}

// IncXDSMessageCount increases counter per XDS messageType
func (m *metrics) IncXDSMessageCount(messageType string) {

	m.xdsMessages.WithLabelValues(messageType).Inc()
}
