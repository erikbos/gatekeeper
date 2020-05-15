package main

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/bmatcuk/doublestar"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/apiauth/pkg/shared"
)

// handlePolicies invokes all policy functions to set additional upstream headers
func (a *authorizationServer) handlePolicies(request *requestInfo, newUpstreamHeaders map[string]string) (int, error) {

	for _, policy := range strings.Split(request.APIProduct.Scopes, ",") {
		headersToAdd, err := a.handlePolicy(policy, request)

		// Stop and return error in case policy indicates we should stop
		if err != nil {
			return http.StatusForbidden, err
		}

		// Add policy generated headers for upstream
		for key, value := range headersToAdd {
			// log.Infof("q2 %s", key)
			newUpstreamHeaders[key] = value
		}
	}
	return http.StatusOK, nil
}

// handlePolicy executes a single policy to optionally add upstream headers
func (a *authorizationServer) handlePolicy(policy string, request *requestInfo) (map[string]string, error) {

	a.metrics.apiProductPolicy.WithLabelValues(request.APIProduct.Name, policy).Inc()

	switch policy {
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
	case "checkHostHeader":
		return policyCheckHostHeader(request)
	}

	a.metrics.apiProductPolicyUnknown.WithLabelValues(request.APIProduct.Name, policy).Inc()
	return nil, nil
}

// getCountryAndStateOfRequestorIP lookup requestor's ip address in geoip database
func (a *authorizationServer) getCountryAndStateOfRequestorIP(
	request *requestInfo, newUpstreamHeaders map[string]string) {

	country, state := a.g.GetCountryAndState(request.IP)
	newUpstreamHeaders["geoip-country"] = country
	newUpstreamHeaders["geoip-state"] = state

	// log.Debugf("Check() rx ip country: %s, state: %s", country, state)

	a.metrics.requestsPerCountry.WithLabelValues(country).Inc()
}

// policyQPS1 returns QPS quotakey to be used by Lyft ratelimiter
// QPS set as developer app attribute has priority over quota set as product attribute
//
func policyQPS1(request *requestInfo) (map[string]string, error) {

	quotaAttributeName := request.APIProduct.Name + "_quotaPerSecond"
	quotaKey := request.apikey + "_a_" + quotaAttributeName

	value, err := shared.GetAttribute(request.developerApp.Attributes, quotaAttributeName)
	if err == nil && value != "" {
		return map[string]string{
				"QPS-Quota-Key": quotaKey,
				"QPS-Rate":      value,
				"QPS-Source":    "app",
			},
			nil
	}
	value, err = shared.GetAttribute(request.APIProduct.Attributes, quotaAttributeName)
	if err == nil && value != "" {
		return map[string]string{
				"QPS-Quota-Key": quotaKey,
				"QPS-Rate":      value,
				"QPS-Source":    "apiproduct",
			},
			nil
	}
	// Nothing to add, no error
	return nil, nil
}

// policySendAPIKey adds apikey as an upstream header
func policySendAPIKey(request *requestInfo) (map[string]string, error) {

	return map[string]string{
			"x-apikey": request.apikey,
		},
		nil
}

// policySendAPIKey adds developer's email address as an upstream header
func policySendDeveloperEmail(request *requestInfo) (map[string]string, error) {

	return map[string]string{
			"x-developer-email": request.developer.Email,
		},
		nil
}

// policySendAPIKey adds developerid as an upstream header
func policySendDeveloperID(request *requestInfo) (map[string]string, error) {

	return map[string]string{
			"x-developer-id": request.developer.DeveloperID,
		},
		nil
}

// policySendDeveloperAppName adds developer app name as an upstream header
func policySendDeveloperAppName(request *requestInfo) (map[string]string, error) {

	return map[string]string{
			"x-developer-app-name": request.developerApp.Name,
		},
		nil
}

// policySendDeveloperAppID adds developer app id as an upstream header
func policySendDeveloperAppID(request *requestInfo) (map[string]string, error) {

	return map[string]string{
			"x-developer-app-id": request.developerApp.AppID,
		},
		nil
}

// policyCheckIPAccessList checks requestor ip against IP ACL defined in developer app
func policyCheckIPAccessList(request *requestInfo) (map[string]string, error) {

	ipAccessList, err := shared.GetAttribute(request.developerApp.Attributes, "IPAccessList")
	if err == nil && ipAccessList != "" {
		if checkIPinAccessList(request.IP, ipAccessList) {
			// OK, we have a match
			return nil, nil
		}
		return nil, errors.New("Blocked by IP ACL")
	}
	// No IPACL attribute or it's value was empty: we allow request
	return nil, nil
}

// checkIPinAccessList checks ip against all subnets in IP ACL
func checkIPinAccessList(ip net.IP, ipAccessList string) bool {

	if ipAccessList == "" {
		return false
	}
	for _, subnet := range strings.Split(ipAccessList, ",") {
		if _, network, err := net.ParseCIDR(subnet); err == nil {
			if network.Contains(ip) {
				// OK, we have a match
				return true
			}
		} else {
			log.Debugf("FIXME increase unparsable ip ACL counter")
		}
	}
	// source ip did not match any of the ACL subnets, request rejected
	return false
}

// policyCheckHostHeader checks request's Host header against host ACL defined in developer app
func policyCheckHostHeader(request *requestInfo) (map[string]string, error) {

	hostAccessList, err := shared.GetAttribute(request.developerApp.Attributes, "HostWhiteList")
	if err == nil && hostAccessList != "" {
		if checkHostinAccessList(request.httpRequest.Headers[":authority"], hostAccessList) {
			return nil, nil
		}
		return nil, errors.New("Blocked by host ACL")
	}
	// No Host ACL attribute or it's value was empty: we allow request
	return nil, nil
}

// policyCheckHostHeader checks host string against a comma separated host regexp list
func checkHostinAccessList(hostHeader string, hostAccessList string) bool {

	if hostAccessList == "" {
		return false
	}
	for _, hostPattern := range strings.Split(hostAccessList, ",") {
		// Testing for matching error  does not make sense: we cannot differentiate betwen bad regexp
		// or not-matching
		if ok, _ := doublestar.Match(hostPattern, hostHeader); ok {
			return true
		}
	}
	// ip did not match any of the subnets found, request rejected
	return false
}
