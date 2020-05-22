package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metricsCollection struct {
	configLoads              *prometheus.CounterVec
	authLatencyHistogram     prometheus.Summary
	connectInfoFailures      prometheus.Counter
	requestsPerCountry       *prometheus.CounterVec
	requestsApikeyNotFound   *prometheus.CounterVec
	requestsAccepted         *prometheus.CounterVec
	requestsRejected         *prometheus.CounterVec
	devApp                   *prometheus.CounterVec
	apiProductPolicy         *prometheus.CounterVec
	apiProductPolicyUnknown  *prometheus.CounterVec
	virtualHostPolicy        *prometheus.CounterVec
	virtualHostPolicyUnknown *prometheus.CounterVec
}

// registerMetrics registers our operational metrics
func (a *authorizationServer) registerMetrics() {
	a.metrics.configLoads = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "config_table_loads_total",
			Help:      "Total number of vhost/route table loads.",
		}, []string{"resource"})

	a.metrics.connectInfoFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "connection_info_failures_total",
			Help:      "Total number of connection info failures.",
		})

	a.metrics.requestsPerCountry = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "requests_percountry_total",
			Help:      "Total number of requests per country.",
		}, []string{"country"})

	a.metrics.requestsApikeyNotFound = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "requests_apikey_notfound_total",
			Help:      "Total number of requests with an unknown apikey.",
		}, []string{"hostname", "protocol", "method"})

	a.metrics.requestsAccepted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "requests_accepted_total",
			Help:      "Total number of requests accepted.",
		}, []string{"hostname", "protocol", "method", "apiproduct"})

	a.metrics.requestsRejected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "requests_rejected_total",
			Help:      "Total number of requests rejected.",
		}, []string{"hostname", "protocol", "method", "apiproduct"})

	a.metrics.authLatencyHistogram = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: myName,
			Name:      "request_latency",
			Help:      "Authentication latency in seconds.",
			Objectives: map[float64]float64{
				0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001,
			},
		})

	a.metrics.apiProductPolicy = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "apiproduct_policy_hits_total",
			Help:      "Total number of product policy hits.",
		}, []string{"apiproduct", "policy"})

	a.metrics.apiProductPolicyUnknown = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "apiproduct_policy_unknown_total",
			Help:      "Total number of unknown product policy hits.",
		}, []string{"apiproduct", "policy"})

	a.metrics.virtualHostPolicy = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "virtualhost_policy_hits_total",
			Help:      "Total number of virtualhost policy hits.",
		}, []string{"virtualhost", "policy"})

	a.metrics.virtualHostPolicyUnknown = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "virtualhost_policy_unknown_total",
			Help:      "Total number of unknown virtualhost policy hits.",
		}, []string{"virtualhost", "policy"})

	prometheus.MustRegister(a.metrics.configLoads)
	prometheus.MustRegister(a.metrics.connectInfoFailures)
	prometheus.MustRegister(a.metrics.requestsPerCountry)
	prometheus.MustRegister(a.metrics.requestsApikeyNotFound)
	prometheus.MustRegister(a.metrics.requestsAccepted)
	prometheus.MustRegister(a.metrics.requestsRejected)
	prometheus.MustRegister(a.metrics.authLatencyHistogram)
	prometheus.MustRegister(a.metrics.apiProductPolicy)
	prometheus.MustRegister(a.metrics.apiProductPolicyUnknown)
	prometheus.MustRegister(a.metrics.virtualHostPolicy)
	prometheus.MustRegister(a.metrics.virtualHostPolicyUnknown)
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

	a.metrics.requestsRejected.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method,
		r.APIProduct.Name).Inc()
}

// IncreaseCounterRequestAccept counts requests that are accpeted
func (a *authorizationServer) IncreaseCounterRequestAccept(r *requestInfo) {

	a.metrics.requestsAccepted.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method,
		r.APIProduct.Name).Inc()
}

// IncreaseCounterRequestAccept counts requests that are accpeted
func (a *authorizationServer) IncreaseCounterConfigLoad(what string) {

	a.metrics.configLoads.WithLabelValues(what).Inc()
}