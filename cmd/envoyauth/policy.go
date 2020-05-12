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
func handlePolicies(request *requestInfo, headersToReturn map[string]string) (int, error) {

	for _, policy := range strings.Split(request.APIProduct.Scopes, ",") {
		headersToAdd, err := handlePolicy(policy, request)

		// Stop and return error in case policy indicates we should stop
		if err != nil {
			return http.StatusForbidden, err
		}

		// Add policy generated headers for upstream
		for key, value := range headersToAdd {
			// log.Infof("q2 %s", key)
			headersToReturn[key] = value
		}
	}
	return http.StatusOK, nil
}

func handlePolicy(policy string, request *requestInfo) (map[string]string, error) {

	// FIXME insert policy counter
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
	// FIXME insert counter for unknown policy name in an apiproduct
	// label: product, policyname
	return nil, nil
}

// policyQPS1 returns QPS quotakey to be used by Lyft ratelimiter
// QPS set as developer app attribute has priority over quota set as product attribute
//
func policyQPS1(request *requestInfo) (map[string]string, error) {
	headerToAdd := make(map[string]string)

	quotaAttributeName := request.APIProduct.Name + "_quotaPerSecond"
	quotaKeyHeader := request.apikey + "_a_" + quotaAttributeName

	value, err := shared.GetAttribute(request.developerApp.Attributes, quotaAttributeName)
	if err == nil && value != "" {
		headerToAdd[quotaKeyHeader] = value
		headerToAdd["quotaSource"] = "app"

		return headerToAdd, nil
	}
	value, err = shared.GetAttribute(request.APIProduct.Attributes, quotaAttributeName)
	if err == nil && value != "" {
		headerToAdd[quotaKeyHeader] = value
		headerToAdd["quotaSource"] = "apiproduct"
	}
	return headerToAdd, nil
}

//
func policySendAPIKey(request *requestInfo) (map[string]string, error) {
	headerToAdd := make(map[string]string, 1)

	headerToAdd["x-apikey"] = request.apikey

	return headerToAdd, nil
}

//
func policySendDeveloperEmail(request *requestInfo) (map[string]string, error) {
	headerToAdd := make(map[string]string, 1)

	headerToAdd["x-developer-email"] = request.developer.Email

	return headerToAdd, nil
}

//
func policySendDeveloperID(request *requestInfo) (map[string]string, error) {
	headerToAdd := make(map[string]string, 1)

	headerToAdd["x-developer-id"] = request.developer.DeveloperID

	return headerToAdd, nil
}

//
func policySendDeveloperAppName(request *requestInfo) (map[string]string, error) {
	headerToAdd := make(map[string]string, 1)

	headerToAdd["x-developer-app-name"] = request.developerApp.Name

	return headerToAdd, nil
}

//
func policySendDeveloperAppID(request *requestInfo) (map[string]string, error) {
	headerToAdd := make(map[string]string, 1)

	headerToAdd["x-developer-app-id"] = request.developerApp.AppID

	return headerToAdd, nil
}

func policyCheckIPAccessList(request *requestInfo) (map[string]string, error) {
	ipAccessList, err := shared.GetAttribute(request.developerApp.Attributes, "IPAccessList")

	if err == nil && ipAccessList != "" {
		if checkIPinAccessList(request.IP, ipAccessList) {
			return nil, nil
		}
		return nil, errors.New("Blocked by IP ACL")
	}
	// No IPACL attribute or it's value was empty: we allow request
	return nil, nil
}

func checkIPinAccessList(ip net.IP, ipAccessList string) bool {
	if ipAccessList == "" {
		return false
	}
	for _, subnet := range strings.Split(ipAccessList, ",") {
		if _, network, err := net.ParseCIDR(subnet); err == nil {
			if network.Contains(ip) {
				return true
			}
		} else {
			log.Debugf("FIXME increase unparsable ip ACL counter")
		}
	}
	// source ip did not match any of the ACL subnets, request rejected
	return false
}

func policyCheckHostHeader(request *requestInfo) (map[string]string, error) {
	hostAccessList, err := shared.GetAttribute(request.developerApp.Attributes, "HostWhiteList")

	log.Infof("Host header: %s", request.httpRequest.Headers[":authority"])
	if err == nil && hostAccessList != "" {
		if checkHostinAccessList(request.httpRequest.Headers[":authority"], hostAccessList) {
			return nil, nil
		}
		return nil, errors.New("Blocked by host ACL")
	}
	// No Host ACL attribute or it's value was empty: we allow request
	return nil, nil
}

func checkHostinAccessList(hostHeader string, hostAccessList string) bool {
	if hostAccessList == "" {
		return false
	}
	for _, hostPattern := range strings.Split(hostAccessList, ",") {
		// Testing for this error value does not make sense: we cannot see difference betwen bad regexp
		// or not-matching
		if ok, _ := doublestar.Match(hostPattern, hostHeader); ok {
			return true
		}
	}
	// ip did not match any of the subnets found, request rejected
	return false
}
