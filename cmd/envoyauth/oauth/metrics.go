package oauth

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	clientStoreHits          prometheus.Counter
	clientStoreMisses        prometheus.Counter
	tokenStoreIssueSuccesses prometheus.Counter
	tokenStoreIssueFailures  prometheus.Counter
	tokenStoreLookupHits     *prometheus.CounterVec
	tokenStoreLookupMisses   *prometheus.CounterVec
}

func newMetrics() *metrics {

	return &metrics{}
}

// registerMetrics registers our operational metrics
func (m *metrics) RegisterWithPrometheus(applicationName string) {
	m.clientStoreHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "oauth_clientstore_hits_total",
			Help:      "Number of OAuth client store hits.",
		})
	prometheus.MustRegister(m.clientStoreHits)

	m.clientStoreMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "oauth_clientstore_misses_total",
			Help:      "Number of OAuth client store misses.",
		})
	prometheus.MustRegister(m.clientStoreMisses)

	m.tokenStoreIssueSuccesses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "oauth_tokenstore_issue_successes_total",
			Help:      "Number of OAuth succesful token store issue requests.",
		})
	prometheus.MustRegister(m.tokenStoreIssueSuccesses)

	m.tokenStoreIssueFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "oauth_tokenstore_issue_failures_total",
			Help:      "Number of OAuth token store issue failures.",
		})
	prometheus.MustRegister(m.tokenStoreIssueFailures)

	m.tokenStoreLookupHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "oauth_tokenstore_lookup_hits_total",
			Help:      "Number of OAuth token store lookup hits.",
		}, []string{"method"})
	prometheus.MustRegister(m.tokenStoreLookupHits)

	m.tokenStoreLookupMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "oauth_tokenstore_lookup_misses_total",
			Help:      "Number of OAuth token store lookup misses.",
		}, []string{"method"})
	prometheus.MustRegister(m.tokenStoreLookupMisses)
}

func (m *metrics) IncClientStoreHits() {
	m.clientStoreHits.Inc()
}

func (m *metrics) IncClientStoreMisses() {
	m.clientStoreMisses.Inc()
}

func (m *metrics) IncTokenStoreIssueSuccesses() {
	m.tokenStoreIssueSuccesses.Inc()
}

func (m *metrics) IncTokenStoreIssueFailures() {
	m.tokenStoreIssueFailures.Inc()
}

func (m *metrics) IncTokenStoreLookupHits(method string) {
	m.tokenStoreLookupHits.WithLabelValues(method).Inc()
}

func (m *metrics) IncTokenStoreLookupMisses(method string) {
	m.tokenStoreLookupMisses.WithLabelValues(method).Inc()
}
