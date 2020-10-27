package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	authAccepted        *prometheus.CounterVec
	authRejected        *prometheus.CounterVec
	authLatency         prometheus.Summary
	configLoads         *prometheus.CounterVec
	connectInfoFailures prometheus.Counter
	UnknownAPIkey       *prometheus.CounterVec
	PolicyUsage         *prometheus.CounterVec
	PolicyUnknown       *prometheus.CounterVec
	CountryHits         *prometheus.CounterVec
}

func newMetrics() *metrics {

	return &metrics{}
}

// registerMetrics registers our operational metrics
func (m *metrics) RegisterWithPrometheus(metricNamespace string) {

	m.authAccepted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "requests_accepted_total",
			Help:      "Total number of authentication requests accepted.",
		}, []string{"hostname", "protocol", "method", "apiproduct"})
	prometheus.MustRegister(m.authAccepted)

	m.authRejected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "requests_rejected_total",
			Help:      "Total number of authentication requests rejected.",
		}, []string{"hostname", "protocol", "method", "apiproduct"})
	prometheus.MustRegister(m.authRejected)

	m.authLatency = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: metricNamespace,
			Name:      "request_latency",
			Help:      "Authentication latency in seconds.",
			Objectives: map[float64]float64{
				0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001,
			},
		})
	prometheus.MustRegister(m.authLatency)

	m.configLoads = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "config_table_loads_total",
			Help:      "Total sum of listener/route/cluster table loads.",
		}, []string{"resource"})
	prometheus.MustRegister(m.configLoads)

	m.connectInfoFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "connection_info_failures_total",
			Help:      "Total number of connection info failures.",
		})
	prometheus.MustRegister(m.connectInfoFailures)

	m.UnknownAPIkey = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "requests_unknown_apikey_total",
			Help:      "Total number of requests with an unknown apikey.",
		}, []string{"hostname", "protocol", "method"})
	prometheus.MustRegister(m.UnknownAPIkey)

	m.PolicyUsage = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "policy_hits_total",
			Help:      "Total number of policy hits.",
		}, []string{"scope", "policy"})
	prometheus.MustRegister(m.PolicyUsage)

	m.PolicyUnknown = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "policy_unknown_total",
			Help:      "Total number of unknown policy hits.",
		}, []string{"scope", "policy"})
	prometheus.MustRegister(m.PolicyUnknown)

	m.CountryHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "requests_per_country_total",
			Help:      "Total number of requests per country.",
		}, []string{"country"})
	prometheus.MustRegister(m.CountryHits)
}

// IncAuthenticationAccepted counts requests that are accepted
func (m *metrics) IncAuthenticationAccepted(r *requestDetails) {

	m.authAccepted.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method,
		r.APIProduct.Name).Inc()
}

// IncAuthenticationRejected counts requests that are rejected
func (m *metrics) IncAuthenticationRejected(r *requestDetails) {

	var product string

	if r.APIProduct != nil {
		product = r.APIProduct.Name
	}

	m.authRejected.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method,
		product).Inc()
}

// IncConnectionInfoFailure increases connection info failures metric
func (m *metrics) IncConnectionInfoFailure() {

	m.connectInfoFailures.Inc()
}

// IncUnknownAPIKey increases unknown apikey metric
func (m *metrics) IncUnknownAPIKey(r *requestDetails) {

	m.UnknownAPIkey.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method).Inc()
}

// IncDatabaseFetchFailure increases database retrieval failure metric
func (m *metrics) IncDatabaseFetchFailure(r *requestDetails) {

	m.UnknownAPIkey.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method).Inc()
}

// IncreaseCounterRequestAccept counts requests that are accepted
func (m *metrics) IncreaseMetricConfigLoad(what string) {

	m.configLoads.WithLabelValues(what).Inc()
}

func (m *metrics) IncPolicyUsage(scope, name string) {

	m.PolicyUsage.WithLabelValues(scope, name).Inc()
}

func (m *metrics) IncPolicyUnknown(scope, name string) {

	m.PolicyUnknown.WithLabelValues(scope, name).Inc()
}

func (m *metrics) IncCountryHits(country string) {

	m.CountryHits.WithLabelValues(country).Inc()
}
