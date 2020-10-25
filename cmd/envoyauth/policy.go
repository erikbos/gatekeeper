package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bmatcuk/doublestar"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Policy holds input to be to evaluate one policy
type Policy struct {

	// Global state of our running application
	authServer *server

	// Request information
	request *requestDetails

	// Current state of policy evaluation
	*PolicyChainResponse
}

// PolicyResponse holds output of policy evaluation
type PolicyResponse struct {
	// If true the request was authenticated, subsequent policies should be evaluated
	authenticated bool
	// If true the request should be denied, no further policy evaluations required
	denied bool
	// Statuscode to use when denying a request
	deniedStatusCode int
	// Message to return when denying a request
	deniedMessage string
	// Additional HTTP headers to set when forwarding to upstream
	headers map[string]string
	// Dynamic metadata to set when forwarding to subsequent envoyproxy filter
	metadata map[string]string
}

// These are dynamic metadata keys set by various policies
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

// Evaluate executes single policy statement
func (p *Policy) Evaluate(policy string, request *requestDetails) *PolicyResponse {

	switch policy {
	case "checkAPIKey":
		return checkAPIKey(request, p.authServer)
	case "checkOAuth2":
		return checkOAuth2(request, p.authServer)
	case "removeAPIKeyFromQP":
		return p.removeAPIKeyFromQP()
	case "lookupGeoIP":
		return lookupGeoIP(request, p.authServer)
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
func checkAPIKey(request *requestDetails, authServer *server) *PolicyResponse {

	var err error
	request.consumerKey, err = getAPIkeyFromQueryString(request.queryParameters)

	// In case we cannot find a query parameter we return immediately
	if err == nil && request.consumerKey == nil {
		return nil
	}
	// In case we cannot find a query parameter did not have a value we reject request
	if err != nil {
		return &PolicyResponse{
			denied:           true,
			deniedStatusCode: http.StatusBadRequest,
			deniedMessage:    fmt.Sprint(err),
		}
	}

	// In case we have an apikey we check whether product is allowed to be accessed
	err = authServer.CheckProductEntitlement(request)
	if err != nil {
		authServer.logger.Debug("CheckProductEntitlement() not allowed",
			zap.String("path", request.URL.Path), zap.String("reason", err.Error()))

		authServer.metrics.increaseCounterApikeyNotfound(request)

		// apikey invalid or path not allowed
		return &PolicyResponse{
			denied:           true,
			deniedStatusCode: http.StatusBadRequest,
			deniedMessage:    fmt.Sprint(err),
		}
	}

	// apikey invalid or path not allowed
	return &PolicyResponse{
		authenticated: true,
		metadata:      buildMetadata(request),
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
func checkOAuth2(request *requestDetails, authServer *server) *PolicyResponse {

	authorizationHeader := request.httpRequest.Headers["authorization"]
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
	tokenInfo, err := authServer.oauth.LoadAccessToken(accessToken)
	if err != nil {
		return &PolicyResponse{
			denied:           true,
			deniedStatusCode: http.StatusInternalServerError,
			deniedMessage:    fmt.Sprint(err),
		}
	}
	request.oauthToken = &accessToken

	// The temporary access token contains the apikey (Also Known As clientId)
	clientID := tokenInfo.GetClientID()
	request.consumerKey = &clientID

	err = authServer.CheckProductEntitlement(request)
	if err != nil {
		authServer.logger.Debug("CheckProductEntitlement() not allowed",
			zap.String("path", request.URL.Path), zap.String("reason", err.Error()))

		authServer.metrics.increaseCounterApikeyNotfound(request)

		return &PolicyResponse{
			denied:           true,
			deniedStatusCode: http.StatusForbidden,
			deniedMessage:    fmt.Sprint(err),
		}
	}

	// Signal that we have authenticated this request
	return &PolicyResponse{
		authenticated: true,
		metadata:      buildMetadata(request),
	}
}

// buildMetadata returns all authentication & apim metadata to be returned by envoyauth
func buildMetadata(request *requestDetails) map[string]string {

	m := make(map[string]string, 10)

	if request.developer != nil {
		if request.developer.Email != "" {
			m[metadataDeveloperEmail] = request.developer.Email
		}
		if request.developer.DeveloperID != "" {
			m[metadataDeveloperID] = request.developer.DeveloperID
		}
	}
	if request.developerApp != nil {
		if request.developerApp.AppID != "" {
			m[metadataAppID] = request.developerApp.AppID
		}
		if request.developerApp.Name != "" {
			m[metadataAppName] = request.developerApp.Name
		}
	}
	if request.APIProduct != nil && request.APIProduct.Name != "" {
		m[metadataAPIProductName] = request.APIProduct.Name
	}
	if request.consumerKey != nil {
		m[metadataAuthMethod] = metadataAuthMethodValueAPIKey
		m[metadataAuthAPIKey] = *request.consumerKey
	}
	if request.oauthToken != nil {
		m[metadataAuthMethod] = metadataAuthMethodValueOAuth
		m[metadataAuthOAuthToken] = *request.oauthToken
	}

	return m
}

// removeAPIKeyFromQP sets path to requested path without any query parameters
func (p *Policy) removeAPIKeyFromQP() *PolicyResponse {

	// We only update :path in case we know request goes upstream
	if !p.PolicyChainResponse.authenticated {
		return nil
	}

	p.request.queryParameters.Del("apikey")

	// We remove query parameters by having envoyauth overwrite the path
	return &PolicyResponse{
		headers: map[string]string{
			":path": p.request.URL.Path + "?" + p.request.queryParameters.Encode(),
		},
	}
}

// lookupGeoIP lookup requestor's ip address in geoip database
func lookupGeoIP(request *requestDetails, authServer *server) *PolicyResponse {

	if authServer.geoip == nil {
		return nil
	}

	country, state := authServer.geoip.GetCountryAndState(request.IP)
	if country == "" {
		return nil
	}

	authServer.metrics.requestsPerCountry.WithLabelValues(country).Inc()

	return &PolicyResponse{
		metadata: map[string]string{
			metadataGeoIPCountry: country,
			metadataGeoIPState:   state,
		},
	}
}

// policyQPS1 returns QPS quotakey to be used by Lyft ratelimiter
// QPS set as developer app attribute has priority over quota set as product attribute
//
func policyQPS1(request *requestDetails) *PolicyResponse {

	if request == nil || request.APIProduct == nil || request.developerApp == nil {
		return nil
	}

	// quotaAttributeName := request.APIProduct.Name + "_quotaPerSecond"
	// quotaKey := *request.apikey + "_a_" + quotaAttributeName

	quotaAttributeName := request.APIProduct.Name + "_quotaPerSecond"
	value, err := request.developerApp.Attributes.Get(quotaAttributeName)
	if err == nil && value != "" {
		return &PolicyResponse{
			metadata: map[string]string{
				"rl.requests_per_unit": value,
				"rl.unit":              "SECOND",
				"rl.descriptor":        "app",
				// "QPS-Quota-Key":        quotaKey,
			},
		}
	}
	value, err = request.APIProduct.Attributes.Get(quotaAttributeName)
	if err == nil && value != "" {
		return &PolicyResponse{
			metadata: map[string]string{
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
func policySendAPIKey(request *requestDetails) *PolicyResponse {

	if request != nil && request.consumerKey != nil {
		return &PolicyResponse{
			headers: map[string]string{
				"x-apikey": *request.consumerKey,
			},
		}
	}
	return nil
}

// policySendAPIKey adds developer's email address as an upstream header
func policySendDeveloperEmail(request *requestDetails) *PolicyResponse {

	if request != nil && request.developer != nil {
		return &PolicyResponse{
			headers: map[string]string{
				"x-developer-email": request.developer.Email,
			},
		}
	}
	return nil
}

// policySendAPIKey adds developerid as an upstream header
func policySendDeveloperID(request *requestDetails) *PolicyResponse {

	if request != nil && request.developer != nil {
		return &PolicyResponse{
			headers: map[string]string{
				"x-developer-id": request.developer.DeveloperID,
			},
		}
	}
	return nil
}

// policySendDeveloperAppName adds developer app name as an upstream header
func policySendDeveloperAppName(request *requestDetails) *PolicyResponse {

	if request != nil && request.developerApp != nil {
		return &PolicyResponse{
			headers: map[string]string{
				"x-developer-app-name": request.developerApp.Name,
			},
		}
	}
	return nil

}

// policySendDeveloperAppID adds developer app id as an upstream header
func policySendDeveloperAppID(request *requestDetails) *PolicyResponse {

	if request != nil && request.developerApp != nil {
		return &PolicyResponse{
			headers: map[string]string{
				"x-developer-app-id": request.developerApp.AppID,
			},
		}
	}
	return nil
}

// policyCheckIPAccessList checks requestor ip against IP ACL defined in developer app
func policyCheckIPAccessList(request *requestDetails) *PolicyResponse {

	ipAccessList, err := request.developerApp.Attributes.Get("IPAccessList")
	if err == nil && ipAccessList != "" {
		if shared.CheckIPinAccessList(request.IP, ipAccessList) {
			// OK, we have a match
			return nil
		}
		return &PolicyResponse{
			denied:           true,
			deniedStatusCode: http.StatusForbidden,
			deniedMessage:    "Blocked by IP ACL",
		}
	}
	// No IPACL attribute or it's value was empty: we allow request
	return nil
}

// policycheckReferer checks request's Host header against host ACL defined in developer app
func policycheckReferer(request *requestDetails) *PolicyResponse {

	hostAccessList, err := request.developerApp.Attributes.Get("Referer")
	if err == nil && hostAccessList != "" {
		if checkHostinAccessList(request.httpRequest.Headers[":authority"], hostAccessList) {
			return nil
		}
		return &PolicyResponse{
			denied:           true,
			deniedStatusCode: http.StatusForbidden,
			deniedMessage:    "Blocked by referer ACL",
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
