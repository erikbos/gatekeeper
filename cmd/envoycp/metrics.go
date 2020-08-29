package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metricsCollection struct {
	xdsDeployments *prometheus.CounterVec
	xdsMessages    *prometheus.CounterVec
}

// registerMetrics registers envoycp operational metrics
func (s *server) registerMetrics() {

	metricListenersCount := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "xds_listeners_total",
			Help:      "Total number of clusters.",
		}, func() float64 {
			return float64(s.dbentities.GetListenerCount())
		})

	metricRoutesCount := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "xds_routes_total",
			Help:      "Total number of routes.",
		}, func() float64 {
			return float64(s.dbentities.GetRouteCount())
		})

	metricClustersCount := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "xds_clusters_total",
			Help:      "Total number of clusters.",
		}, func() float64 {
			return float64(s.dbentities.GetClusterCount())
		})

	s.metrics.xdsDeployments = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "xds_deployments_total",
			Help:      "Total number of xds configuration deployments.",
		}, []string{"resource"})

	s.metrics.xdsMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "xds_resource_requests_total",
			Help:      "Total number of XDS messages.",
		}, []string{"messagetype"})

	prometheus.MustRegister(metricListenersCount)
	prometheus.MustRegister(metricRoutesCount)
	prometheus.MustRegister(metricClustersCount)
	prometheus.MustRegister(s.metrics.xdsDeployments)
	prometheus.MustRegister(s.metrics.xdsMessages)
}

// increaseCounterXDSMessage increases counter per XDS messageType
func (s *server) increaseCounterXDSMessage(messageType string) {

	s.metrics.xdsMessages.WithLabelValues(messageType).Inc()
}
