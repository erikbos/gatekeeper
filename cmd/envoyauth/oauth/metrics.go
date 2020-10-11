package oauth

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	tokenIssueRequests *prometheus.CounterVec
	tokenInfoRequests  *prometheus.CounterVec
}

func newMetrics() *metrics {

	return &metrics{}
}

// registerMetrics registers our operational metrics
func (m *metrics) RegisterWithPrometheus(applicationName string) {
	m.tokenIssueRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "token_issue_total",
			Help:      "Number of OAuth token issue requests.",
		}, []string{"status"})
	prometheus.MustRegister(m.tokenIssueRequests)

	m.tokenInfoRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "token_info_total",
			Help:      "Number of OAuth token info requests.",
		}, []string{"status"})
	prometheus.MustRegister(m.tokenInfoRequests)
}

func (m *metrics) IncTokenIssueRequests(statuscode int) {

	m.tokenIssueRequests.WithLabelValues(fmt.Sprintf("%d", statuscode)).Inc()
}

func (m *metrics) IncTokenInfoRequests(statuscode int) {

	m.tokenInfoRequests.WithLabelValues(fmt.Sprintf("%d", statuscode)).Inc()
}
