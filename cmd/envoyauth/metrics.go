package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	configLoads            *prometheus.CounterVec
	authLatencyHistogram   prometheus.Summary
	connectInfoFailures    prometheus.Counter
	requestsPerCountry     *prometheus.CounterVec
	requestsApikeyNotFound *prometheus.CounterVec
	requestsAccepted       *prometheus.CounterVec
	requestsRejected       *prometheus.CounterVec
	Policy                 *prometheus.CounterVec
	PolicyUnknown          *prometheus.CounterVec
}

func newMetrics() *metrics {

	return &metrics{}
}

// registerMetrics registers our operational metrics
func (m *metrics) RegisterWithPrometheus() {
	m.configLoads = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "config_table_loads_total",
			Help:      "Total sum of listener/route/cluster table loads.",
		}, []string{"resource"})
	prometheus.MustRegister(m.configLoads)

	m.connectInfoFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "connection_info_failures_total",
			Help:      "Total number of connection info failures.",
		})
	prometheus.MustRegister(m.connectInfoFailures)

	m.requestsPerCountry = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "requests_percountry_total",
			Help:      "Total number of requests per country.",
		}, []string{"country"})
	prometheus.MustRegister(m.requestsPerCountry)

	m.requestsApikeyNotFound = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "requests_apikey_notfound_total",
			Help:      "Total number of requests with an unknown apikey.",
		}, []string{"hostname", "protocol", "method"})
	prometheus.MustRegister(m.requestsApikeyNotFound)

	m.requestsAccepted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "requests_accepted_total",
			Help:      "Total number of requests accepted.",
		}, []string{"hostname", "protocol", "method", "apiproduct"})
	prometheus.MustRegister(m.requestsAccepted)

	m.requestsRejected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "requests_rejected_total",
			Help:      "Total number of requests rejected.",
		}, []string{"hostname", "protocol", "method", "apiproduct"})
	prometheus.MustRegister(m.requestsRejected)

	m.authLatencyHistogram = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: applicationName,
			Name:      "request_latency",
			Help:      "Authentication latency in seconds.",
			Objectives: map[float64]float64{
				0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001,
			},
		})
	prometheus.MustRegister(m.authLatencyHistogram)

	m.Policy = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "policy_hits_total",
			Help:      "Total number of policy hits.",
		}, []string{"scope", "policy"})
	prometheus.MustRegister(m.Policy)

	m.PolicyUnknown = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "policy_unknown_total",
			Help:      "Total number of unknown policy hits.",
		}, []string{"scope", "policy"})
	prometheus.MustRegister(m.PolicyUnknown)
}

// increaseCounterApikeyNotfound requests with unknown apikey
func (m *metrics) increaseCounterApikeyNotfound(r *requestInfo) {

	m.requestsApikeyNotFound.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method).Inc()
}

// increaseCounterRequestRejected counts requests that are rejected
func (m *metrics) increaseCounterRequestRejected(r *requestInfo) {

	var product string

	if r.APIProduct != nil {
		product = r.APIProduct.Name
	}

	m.requestsRejected.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method,
		product).Inc()
}

// IncreaseCounterRequestAccept counts requests that are accepted
func (m *metrics) IncreaseCounterRequestAccept(r *requestInfo) {

	m.requestsAccepted.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method,
		r.APIProduct.Name).Inc()
}

// IncreaseCounterRequestAccept counts requests that are accepted
func (m *metrics) IncreaseMetricConfigLoad(what string) {

	m.configLoads.WithLabelValues(what).Inc()
}

func (m *metrics) IncreaseMetricPolicy(scope, name string) {

	m.Policy.WithLabelValues(scope, name).Inc()
}

func (m *metrics) IncreaseMetricPolicyUnknown(scope, name string) {

	m.PolicyUnknown.WithLabelValues(scope, name).Inc()
}
