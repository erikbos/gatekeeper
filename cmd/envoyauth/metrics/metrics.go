package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/cmd/envoyauth/request"
)

// Metrics holds all our metrics
type Metrics struct {
	authAccepted                  *prometheus.CounterVec
	authRejected                  *prometheus.CounterVec
	authLatency                   prometheus.Summary
	configLoads                   *prometheus.CounterVec
	connectInfoFailures           prometheus.Counter
	UnknownAPIkey                 *prometheus.CounterVec
	PolicyUsage                   *prometheus.CounterVec
	PolicyUnknown                 *prometheus.CounterVec
	CountryHits                   *prometheus.CounterVec
	OAuthClientStoreHits          prometheus.Counter
	OAuthClientStoreMisses        prometheus.Counter
	OAuthTokenStoreIssueSuccesses prometheus.Counter
	OAuthTokenStoreIssueFailures  prometheus.Counter
	OAuthTokenStoreLookupHits     *prometheus.CounterVec
	OAuthTokenStoreLookupMisses   *prometheus.CounterVec
}

// NewMetrics returns a new Metrics instance
func NewMetrics() *Metrics {

	return &Metrics{}
}

// RegisterWithPrometheus registers our operational metrics
func (m *Metrics) RegisterWithPrometheus(metricNamespace string) {

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

	m.OAuthClientStoreHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "oauth_clientstore_hits_total",
			Help:      "Number of OAuth client store hits.",
		})
	prometheus.MustRegister(m.OAuthClientStoreHits)

	m.OAuthClientStoreMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "oauth_clientstore_misses_total",
			Help:      "Number of OAuth client store misses.",
		})
	prometheus.MustRegister(m.OAuthClientStoreMisses)

	m.OAuthTokenStoreIssueSuccesses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "oauth_tokenstore_issue_successes_total",
			Help:      "Number of OAuth succesful token store issue requests.",
		})
	prometheus.MustRegister(m.OAuthTokenStoreIssueSuccesses)

	m.OAuthTokenStoreIssueFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "oauth_tokenstore_issue_failures_total",
			Help:      "Number of OAuth token store issue failures.",
		})
	prometheus.MustRegister(m.OAuthTokenStoreIssueFailures)

	m.OAuthTokenStoreLookupHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "oauth_tokenstore_lookup_hits_total",
			Help:      "Number of OAuth token store lookup hits.",
		}, []string{"method"})
	prometheus.MustRegister(m.OAuthTokenStoreLookupHits)

	m.OAuthTokenStoreLookupMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "oauth_tokenstore_lookup_misses_total",
			Help:      "Number of OAuth token store lookup misses.",
		}, []string{"method"})
	prometheus.MustRegister(m.OAuthTokenStoreLookupMisses)
}

// IncAuthenticationAccepted counts requests that are accepted
func (m *Metrics) IncAuthenticationAccepted(r *request.State) {

	m.authAccepted.WithLabelValues(
		r.HTTPRequest.Host,
		r.HTTPRequest.Protocol,
		r.HTTPRequest.Method,
		r.APIProduct.Name).Inc()
}

// IncAuthenticationRejected counts requests that are rejected
func (m *Metrics) IncAuthenticationRejected(r *request.State) {

	var product string

	if r.APIProduct != nil {
		product = r.APIProduct.Name
	}

	m.authRejected.WithLabelValues(
		r.HTTPRequest.Host,
		r.HTTPRequest.Protocol,
		r.HTTPRequest.Method,
		product).Inc()
}

// IncConnectionInfoFailure increases connection info failures metric
func (m *Metrics) IncConnectionInfoFailure() {

	m.connectInfoFailures.Inc()
}

// IncreaseMetricConfigLoad increases configuration loads metric
func (m *Metrics) IncreaseMetricConfigLoad(what string) {

	m.configLoads.WithLabelValues(what).Inc()
}

// NewTimerAuthLatency returns timer to record latency of authentication
func (m *Metrics) NewTimerAuthLatency() *prometheus.Timer {

	return prometheus.NewTimer(m.authLatency)
}

// IncUnknownAPIKey increases unknown apikey metric
func (m *Metrics) IncUnknownAPIKey(r *request.State) {

	m.UnknownAPIkey.WithLabelValues(
		r.HTTPRequest.Host,
		r.HTTPRequest.Protocol,
		r.HTTPRequest.Method).Inc()
}

// IncDatabaseFetchFailure increases database retrieval failure metric
func (m *Metrics) IncDatabaseFetchFailure(r *request.State) {

	m.UnknownAPIkey.WithLabelValues(
		r.HTTPRequest.Host,
		r.HTTPRequest.Protocol,
		r.HTTPRequest.Method).Inc()
}

func (m *Metrics) IncPolicyUsage(scope, name string) {

	m.PolicyUsage.WithLabelValues(scope, name).Inc()
}

func (m *Metrics) IncPolicyUnknown(scope, name string) {

	m.PolicyUnknown.WithLabelValues(scope, name).Inc()
}

func (m *Metrics) IncCountryHits(country string) {

	m.CountryHits.WithLabelValues(country).Inc()
}

func (m *Metrics) IncOAuthClientStoreHits() {

	m.OAuthClientStoreHits.Inc()
}

func (m *Metrics) IncOAuthClientStoreMisses() {

	m.OAuthClientStoreMisses.Inc()
}

func (m *Metrics) IncOAuthTokenStoreIssueSuccesses() {

	m.OAuthTokenStoreIssueSuccesses.Inc()
}

func (m *Metrics) IncOAuthTokenStoreIssueFailures() {

	m.OAuthTokenStoreIssueFailures.Inc()
}

func (m *Metrics) IncOAuthTokenStoreLookupHits(method string) {

	m.OAuthTokenStoreLookupHits.WithLabelValues(method).Inc()
}

func (m *Metrics) IncOAuthTokenStoreLookupMisses(method string) {

	m.OAuthTokenStoreLookupMisses.WithLabelValues(method).Inc()
}
