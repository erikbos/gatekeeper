package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	applicationName string
	xdsEntities     *prometheus.GaugeVec
	xdsSnapshots    *prometheus.CounterVec
	xdsMessages     *prometheus.CounterVec
}

func New(applicationName string) *Metrics {

	return &Metrics{
		applicationName: applicationName,
	}
}

// GinHandler returns a Gin handler for Prometheus metrics endpoint
func (m *Metrics) GinHandler() gin.HandlerFunc {

	return gin.WrapH(promhttp.Handler())
}

// RegisterWithPrometheus registers envoycp operational metrics
func (m *Metrics) RegisterWithPrometheus() {

	m.xdsEntities = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.applicationName,
			Name:      "xds_entities_total",
			Help:      "Total number of xds entities.",
		}, []string{"messagetype"})
	prometheus.MustRegister(m.xdsEntities)

	m.xdsSnapshots = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.applicationName,
			Name:      "xds_snapshots_total",
			Help:      "Total number of xds snapshots created.",
		}, []string{"resource"})
	prometheus.MustRegister(m.xdsSnapshots)

	m.xdsMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.applicationName,
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
