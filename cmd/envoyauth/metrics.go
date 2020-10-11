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

// registerMetrics registers our operational metrics
func (a *authorizationServer) registerMetrics() {
	a.metrics.configLoads = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "config_table_loads_total",
			Help:      "Total sum of listener/route/cluster table loads.",
		}, []string{"resource"})
	prometheus.MustRegister(a.metrics.configLoads)

	a.metrics.connectInfoFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "connection_info_failures_total",
			Help:      "Total number of connection info failures.",
		})
	prometheus.MustRegister(a.metrics.connectInfoFailures)

	a.metrics.requestsPerCountry = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "requests_percountry_total",
			Help:      "Total number of requests per country.",
		}, []string{"country"})
	prometheus.MustRegister(a.metrics.requestsPerCountry)

	a.metrics.requestsApikeyNotFound = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "requests_apikey_notfound_total",
			Help:      "Total number of requests with an unknown apikey.",
		}, []string{"hostname", "protocol", "method"})
	prometheus.MustRegister(a.metrics.requestsApikeyNotFound)

	a.metrics.requestsAccepted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "requests_accepted_total",
			Help:      "Total number of requests accepted.",
		}, []string{"hostname", "protocol", "method", "apiproduct"})
	prometheus.MustRegister(a.metrics.requestsAccepted)

	a.metrics.requestsRejected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "requests_rejected_total",
			Help:      "Total number of requests rejected.",
		}, []string{"hostname", "protocol", "method", "apiproduct"})
	prometheus.MustRegister(a.metrics.requestsRejected)

	a.metrics.authLatencyHistogram = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: applicationName,
			Name:      "request_latency",
			Help:      "Authentication latency in seconds.",
			Objectives: map[float64]float64{
				0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001,
			},
		})
	prometheus.MustRegister(a.metrics.authLatencyHistogram)

	a.metrics.Policy = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "policy_hits_total",
			Help:      "Total number of policy hits.",
		}, []string{"scope", "policy"})
	prometheus.MustRegister(a.metrics.Policy)

	a.metrics.PolicyUnknown = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "policy_unknown_total",
			Help:      "Total number of unknown policy hits.",
		}, []string{"scope", "policy"})
	prometheus.MustRegister(a.metrics.PolicyUnknown)
}

// increaseCounterApikeyNotfound requests with unknown apikey
func (a *authorizationServer) increaseCounterApikeyNotfound(r *requestInfo) {

	a.metrics.requestsApikeyNotFound.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method).Inc()
}

// increaseCounterRequestRejected counts requests that are rejected
func (a *authorizationServer) increaseCounterRequestRejected(r *requestInfo) {

	var product string

	if r.APIProduct != nil {
		product = r.APIProduct.Name
	}

	a.metrics.requestsRejected.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method,
		product).Inc()
}

// IncreaseCounterRequestAccept counts requests that are accepted
func (a *authorizationServer) IncreaseCounterRequestAccept(r *requestInfo) {

	a.metrics.requestsAccepted.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method,
		r.APIProduct.Name).Inc()
}

// IncreaseCounterRequestAccept counts requests that are accepted
func (a *authorizationServer) IncreaseMetricConfigLoad(what string) {

	a.metrics.configLoads.WithLabelValues(what).Inc()
}

func (a *authorizationServer) IncreaseMetricPolicy(scope, name string) {

	a.metrics.Policy.WithLabelValues(scope, name).Inc()
}

func (a *authorizationServer) IncreaseMetricPolicyUnknown(scope, name string) {

	a.metrics.PolicyUnknown.WithLabelValues(scope, name).Inc()
}
