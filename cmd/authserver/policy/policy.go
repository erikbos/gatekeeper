package policy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/authserver/request"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Policy holds input to be to evaluate one policy
type Policy struct {
	// Global state of our running application
	config *ChainConfig

	// Request information
	Request *request.Request

	// Current state of policy evaluation
	*ChainOutcome
}

// Response holds output of policy evaluation
type Response struct {
	// If true the request was Authenticated, subsequent policies should be evaluated
	Authenticated bool
	// If true the request should be Denied, no further policy evaluations required
	Denied bool
	// Statuscode to use when denying a request
	DeniedStatusCode int
	// Message to return when denying a request
	DeniedMessage string
	// Additional HTTP Headers to set when forwarding to upstream
	Headers map[string]string
	// Dynamic Metadata to set when forwarding to subsequent envoyproxy filter
	Metadata map[string]string
}

// Dynamic metadata keys which can be set by various policies
const (
	metadataAuthMethod            = "auth.method"
	metadataAuthMethodValueAPIKey = "apikey"
	metadataAuthMethodValueOAuth  = "oauth"
	metadataAuthAPIKey            = "auth.apikey"
	metadataAuthOAuthToken        = "auth.oauthtoken"
	metadataDeveloperEmail        = "developer.email"
	metadataDeveloperID           = "developer.id"
	metadataAppName               = "app.name"
	metadataAppID                 = "app.id"
	metadataAPIProductName        = "apiproduct.name"
	metadataGeoIPCountry          = "geoip.country"
	metadataGeoIPState            = "geoip.state"
)

// NewPolicy returns a new Policy instance
func NewPolicy(config *ChainConfig) *Policy {

	return &Policy{
		config: config,
	}
}

// Evaluate executes single policy statement
func (p *Policy) Evaluate(policy string, request *request.Request) *Response {

	switch policy {
	case "checkAPIKey":
		return p.checkAPIKey(request)
	case "checkOAuth2":
		return p.checkOAuth2(request)
	case "removeAPIKeyFromQP":
		return p.removeAPIKeyFromQP()
	case "lookupGeoIP":
		return p.lookupGeoIP(request)
	case "qps":
		return policyQPS1(request)
	case "sendAPIKey":
		return policySendAPIKey(request)
	case "sendDeveloperEmail":
		return policySendDeveloperEmail(request)
	case "sendDeveloperID":
		return policySendDeveloperID(request)
	case "sendDeveloperAppName":
		return policySendDeveloperAppName(request)
	case "sendDeveloperAppID":
		return policySendDeveloperAppID(request)
	case "checkIPAccessList":
		return policyCheckIPAccessList(request)
	case "checkReferer":
		return policycheckReferer(request)
	}
	return nil
}

// checkAPIKey tries to find key in querystring, loads dev app, dev details, and check whether path is allowed
func (p *Policy) checkAPIKey(request *request.Request) *Response {

	var err error
	request.ConsumerKey, err = getAPIkeyFromQueryString(request.QueryParameters)

	// In case we cannot find a query parameter we return immediately
	if err == nil && request.ConsumerKey == nil {
		return nil
	}
	// In case we cannot find a query parameter did not have a value we reject request
	if err != nil {
		return &Response{
			Denied:           true,
			DeniedStatusCode: http.StatusBadRequest,
			DeniedMessage:    fmt.Sprint(err),
		}
	}

	// In case we have an apikey we check whether product is allowed to be accessed
	err = p.CheckProductEntitlement(request)
	if err != nil {
		p.config.logger.Debug("CheckProductEntitlement() not allowed",
			zap.String("path", request.URL.Path), zap.String("reason", err.Error()))

		p.config.metrics.IncUnknownAPIKey(request)

		// apikey invalid or path not allowed
		return &Response{
			Denied:           true,
			DeniedStatusCode: http.StatusBadRequest,
			DeniedMessage:    fmt.Sprint(err),
		}
	}

	// apikey invalid or path not allowed
	return &Response{
		Authenticated: true,
		Metadata:      buildMetadata(request),
	}
}

// getAPIkeyFromQueryString extracts apikey from query parameters
func getAPIkeyFromQueryString(queryParameters url.Values) (*string, error) {

	// iterate over queryparameters be able to Find Them in alL CasEs
	for param, value := range queryParameters {

		// we allow both spellings
		param := strings.ToLower(param)
		if param == "apikey" || param == "key" {
			if len(value) == 1 {
				return &value[0], nil
			}
			return nil, errors.New("apikey parameter has no value")
		}
	}
	return nil, nil
}

// checkOAuth2 tries OAuth authentication, loads dev app, dev details, and check whether path is allowed
func (p *Policy) checkOAuth2(request *request.Request) *Response {

	authorizationHeader := request.HTTPRequest.Headers["authorization"]
	if authorizationHeader == "" {
		return nil
	}

	prefix := "Bearer "
	accessToken := ""
	if authorizationHeader != "" && strings.HasPrefix(authorizationHeader, prefix) {
		accessToken = authorizationHeader[len(prefix):]
	} else {
		// Cannot get bearer token from authorization header
		// Not a problem: apparently this request was not meant to be authenticated using OAuth
		return nil
	}

	// Load OAuth token details from data store
	tokenInfo, err := p.config.oauth.LoadAccessToken(context.Background(), accessToken)
	if err != nil {
		return &Response{
			Denied:           true,
			DeniedStatusCode: http.StatusInternalServerError,
			DeniedMessage:    fmt.Sprint(err),
		}
	}
	request.OauthToken = &accessToken

	// The temporary access token contains the apikey (Also Known As clientId)
	clientID := tokenInfo.GetClientID()
	request.ConsumerKey = &clientID

	err = p.CheckProductEntitlement(request)
	if err != nil {
		p.config.logger.Debug("CheckProductEntitlement() not allowed",
			zap.String("path", request.URL.Path), zap.String("reason", err.Error()))

		p.config.metrics.IncUnknownAPIKey(request)

		return &Response{
			Denied:           true,
			DeniedStatusCode: http.StatusForbidden,
			DeniedMessage:    fmt.Sprint(err),
		}
	}

	// Signal that we have authenticated this request
	return &Response{
		Authenticated: true,
		Metadata:      buildMetadata(request),
	}
}

// buildMetadata returns all authentication & apim metadata to be returned by authserver
func buildMetadata(request *request.Request) map[string]string {

	m := make(map[string]string, 10)

	if request.Developer != nil {
		if request.Developer.Email != "" {
			m[metadataDeveloperEmail] = request.Developer.Email
		}
		if request.Developer.DeveloperID != "" {
			m[metadataDeveloperID] = request.Developer.DeveloperID
		}
	}
	if request.DeveloperApp != nil {
		if request.DeveloperApp.AppID != "" {
			m[metadataAppID] = request.DeveloperApp.AppID
		}
		if request.DeveloperApp.Name != "" {
			m[metadataAppName] = request.DeveloperApp.Name
		}
	}
	if request.APIProduct != nil && request.APIProduct.Name != "" {
		m[metadataAPIProductName] = request.APIProduct.Name
	}
	if request.ConsumerKey != nil {
		m[metadataAuthMethod] = metadataAuthMethodValueAPIKey
		m[metadataAuthAPIKey] = *request.ConsumerKey
	}
	if request.OauthToken != nil {
		m[metadataAuthMethod] = metadataAuthMethodValueOAuth
		m[metadataAuthOAuthToken] = *request.OauthToken
	}

	return m
}

// removeAPIKeyFromQP sets path to requested path without any query parameters
func (p *Policy) removeAPIKeyFromQP() *Response {

	// We only update :path in case we know request was authenticated and hence goes upstream
	if p == nil || !p.ChainOutcome.Authenticated {
		return nil
	}

	p.Request.QueryParameters.Del("apikey")

	// We remove query parameters by having authserver overwrite the path
	return &Response{
		Headers: map[string]string{
			":path": p.Request.URL.Path + "?" + p.Request.QueryParameters.Encode(),
		},
	}
}

// lookupGeoIP lookup requestor's ip address in geoip database
func (p *Policy) lookupGeoIP(request *request.Request) *Response {

	if p.config == nil || p.config.geo == nil {
		return nil
	}

	country, state := p.config.geo.GetCountryAndState(request.IP)
	if country == "" {
		return nil
	}

	p.config.metrics.IncCountryHits(country)

	return &Response{
		Metadata: map[string]string{
			metadataGeoIPCountry: country,
			metadataGeoIPState:   state,
		},
	}
}

// policyQPS1 returns QPS quotakey to be used by Lyft ratelimiter
// QPS set as developer app attribute has priority over quota set as product attribute
//
func policyQPS1(request *request.Request) *Response {

	if request == nil || request.APIProduct == nil || request.DeveloperApp == nil {
		return nil
	}

	// quotaAttributeName := request.APIProduct.Name + "_quotaPerSecond"
	// quotaKey := *request.apikey + "_a_" + quotaAttributeName

	quotaAttributeName := request.APIProduct.Name + "_quotaPerSecond"
	value, err := request.DeveloperApp.Attributes.Get(quotaAttributeName)
	if err == nil && value != "" {
		return &Response{
			Metadata: map[string]string{
				"rl.requests_per_unit": value,
				"rl.unit":              "SECOND",
				"rl.descriptor":        "app",
				// "QPS-Quota-Key":        quotaKey,
			},
		}
	}
	value, err = request.APIProduct.Attributes.Get(quotaAttributeName)
	if err == nil && value != "" {
		return &Response{
			Metadata: map[string]string{
				"rl.requests_per_unit": value,
				"rl.unit":              "SECOND",
				"rl.descriptor":        "apiproduct",
				// "QPS-Quota-Key":        quotaKey,
			},
		}
	}
	// Nothing to add, no error
	return nil
}

// policySendAPIKey adds apikey as an upstream header
func policySendAPIKey(request *request.Request) *Response {

	if request != nil && request.ConsumerKey != nil {
		return &Response{
			Headers: map[string]string{
				"x-apikey": *request.ConsumerKey,
			},
		}
	}
	return nil
}

// policySendAPIKey adds developer's email address as an upstream header
func policySendDeveloperEmail(request *request.Request) *Response {

	if request != nil && request.Developer != nil {
		return &Response{
			Headers: map[string]string{
				"x-developer-email": request.Developer.Email,
			},
		}
	}
	return nil
}

// policySendAPIKey adds developerid as an upstream header
func policySendDeveloperID(request *request.Request) *Response {

	if request != nil && request.Developer != nil {
		return &Response{
			Headers: map[string]string{
				"x-developer-id": request.Developer.DeveloperID,
			},
		}
	}
	return nil
}

// policySendDeveloperAppName adds developer app name as an upstream header
func policySendDeveloperAppName(request *request.Request) *Response {

	if request != nil && request.DeveloperApp != nil {
		return &Response{
			Headers: map[string]string{
				"x-developer-app-name": request.DeveloperApp.Name,
			},
		}
	}
	return nil

}

// policySendDeveloperAppID adds developer app id as an upstream header
func policySendDeveloperAppID(request *request.Request) *Response {

	if request != nil && request.DeveloperApp != nil {
		return &Response{
			Headers: map[string]string{
				"x-developer-app-id": request.DeveloperApp.AppID,
			},
		}
	}
	return nil
}

// policyCheckIPAccessList checks requestor ip against IP ACL defined in developer app
func policyCheckIPAccessList(request *request.Request) *Response {

	ipAccessList, err := request.DeveloperApp.Attributes.Get("IPAccessList")
	if err == nil && ipAccessList != "" {
		if shared.CheckIPinAccessList(request.IP, ipAccessList) {
			// OK, we have a match
			return nil
		}
		return &Response{
			Denied:           true,
			DeniedStatusCode: http.StatusForbidden,
			DeniedMessage:    "Blocked by IP ACL",
		}
	}
	// No IPACL attribute or it's value was empty: we allow request
	return nil
}

// policycheckReferer checks request's Host header against host ACL defined in developer app
func policycheckReferer(request *request.Request) *Response {

	hostAccessList, err := request.DeveloperApp.Attributes.Get("Referer")
	if err == nil && hostAccessList != "" {
		if checkHostinAccessList(request.HTTPRequest.Headers[":authority"], hostAccessList) {
			return nil
		}
		return &Response{
			Denied:           true,
			DeniedStatusCode: http.StatusForbidden,
			DeniedMessage:    "Blocked by referer ACL",
		}
	}
	// No Host ACL attribute or it's value was empty: we allow request
	return nil
}

// checkHostinAccessList checks host string against a comma separated host regexp list
func checkHostinAccessList(hostName string, hostAccessList string) bool {

	if hostAccessList == "" {
		return false
	}
	for _, hostPattern := range strings.Split(hostAccessList, ",") {
		// Testing for matching error does not make sense: we cannot differentiate between bad regexp
		// or not-matching
		if ok, _ := doublestar.Match(hostPattern, hostName); ok {
			return true
		}
	}
	// hostname did not match any of the patterns in ACL, request rejected
	return false
}
