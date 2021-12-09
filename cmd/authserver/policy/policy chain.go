package policy

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/authserver/metrics"
	"github.com/erikbos/gatekeeper/cmd/authserver/oauth"
	"github.com/erikbos/gatekeeper/cmd/authserver/request"
	"github.com/erikbos/gatekeeper/pkg/db"
)

// ChainConfig hold chain policy configuration
type ChainConfig struct {
	db      *db.Database
	oauth   *oauth.Server
	geo     *Geoip
	metrics *metrics.Metrics
	logger  *zap.Logger
}

// Chain holds the input to evaluating a series of policies
type Chain struct {
	config *ChainConfig

	// Request information
	Request *request.State

	// "vhost" or "apiproduct"
	scope string
}

// NewChainConfig returns a ChainConfig object holding policy configuration
func NewChainConfig(db *db.Database, oauth *oauth.Server, geo *Geoip,
	metrics *metrics.Metrics, logger *zap.Logger) *ChainConfig {

	return &ChainConfig{
		db:      db,
		oauth:   oauth,
		geo:     geo,
		metrics: metrics,
		logger:  logger,
	}
}

// NewChain returns a new Chain object
func NewChain(r *request.State, scope string, config *ChainConfig) *Chain {

	return &Chain{
		Request: r,
		scope:   scope,
		config:  config,
	}
}

//
const (
	PolicyScopeVhost      = "listener"
	PolicyScopeAPIProduct = "apiproduct"
)

// ChainOutcome holds the output of a policy chain evaluation
type ChainOutcome struct {
	// If true the request was Authenticated, subsequent policies should be evaluated
	Authenticated bool
	// If true the request should be Denied, no further policy evaluations required
	Denied bool
	// Statuscode to use when denying a request
	DeniedStatusCode int
	// Message to return when denying a request
	DeniedMessage string
	// Additional HTTP headers to set when forwarding to upstream
	UpstreamHeaders map[string]string
	// Dynamic metadata to set when forwarding to subsequent envoyproxy filter
	UpstreamDynamicMetadata map[string]string
}

// Evaluate invokes all policy functions one by one, to:
// - check whether call should be allowed or reject
// - set HTTP response payload message
// - set additional upstream headers
func (p Chain) Evaluate() *ChainOutcome {

	// Take policies from vhost configuration
	policies := p.Request.Listener.Policies
	// Or apiproduct policies in case requested
	if p.scope == PolicyScopeAPIProduct {
		policies = p.Request.APIProduct.Policies
	}

	policyChainResult := ChainOutcome{
		// By default we intend to reject request (unauthenticated)
		// This should be overwritten by one of authentication policies to:
		// 1) allow the request
		// 2) reject, with specific deny message
		Authenticated:           false,
		Denied:                  true,
		DeniedStatusCode:        http.StatusForbidden,
		DeniedMessage:           "No credentials provided",
		UpstreamHeaders:         make(map[string]string, 5),
		UpstreamDynamicMetadata: make(map[string]string, 15),
	}

	p.config.logger.Debug("Evaluating policy chain",
		zap.String("scope", p.scope),
		zap.String("policies", policies))

	for _, policyName := range strings.Split(policies, ",") {

		trimmedPolicyName := strings.TrimSpace(policyName)

		policy := NewPolicy(p.config)

		policy.ChainOutcome = &policyChainResult
		policy.Request = p.Request
		policyResult := policy.Evaluate(trimmedPolicyName, p.Request)

		p.config.logger.Debug("Evaluating policy",
			zap.String("scope", p.scope),
			zap.String("policy", trimmedPolicyName),
			zap.Reflect("result", policyResult))

		if policyResult != nil {
			// Register this policy evaluation successed
			p.config.metrics.IncPolicyHits(p.scope, trimmedPolicyName)

			// Add policy generated headers to upstream
			for key, value := range policyResult.Headers {
				policyChainResult.UpstreamHeaders[key] = value
			}
			// Add policy generated metadata
			for key, value := range policyResult.Metadata {
				policyChainResult.UpstreamDynamicMetadata[key] = value
			}
			if policyResult.Authenticated {
				policyChainResult.Authenticated = true
			}

			// In case policy wants to deny request we do so with provided status code
			if policyResult.Denied {
				policyChainResult.Denied = policyResult.Denied
				policyChainResult.DeniedStatusCode = policyResult.DeniedStatusCode
				policyChainResult.DeniedMessage = policyResult.DeniedMessage

				return &policyChainResult
			}
		} else {
			// Register this policy evaluation failed
			p.config.metrics.IncPolicyMisses(p.scope, trimmedPolicyName)
		}
	}
	return &policyChainResult
}
