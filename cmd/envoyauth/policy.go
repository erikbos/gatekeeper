package main

import (
	"errors"
	"net"
	"net/http"
	"strings"

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
		// case "checkHostHeader":
		// 	return policyCheckIPAccessList(request)
		// case "sendbasicauth":
		// 	return policyCheckIPAccessList(request)
	}
	// FIXME insert counter for unknown policy name in an active product
	// label: product, policyname
	return nil, nil
}

// policyQPS1 returns QPS quotakey to be used by Lyft ratelimiter
// QPS set as developer app attribute has priority over quota set as product attribute
//
func policyQPS1(request *requestInfo) (map[string]string, error) {
	quotaAttributeName := request.APIProduct.Name + "_quotaPerSecond"

	returnValue := make(map[string]string)

	value, err := shared.GetAttribute(request.developerApp.Attributes, quotaAttributeName)
	if err == nil && value != "" {
		returnValue[request.apikey+"_a_"+quotaAttributeName] = value
	}
	value, err = shared.GetAttribute(request.APIProduct.Attributes, quotaAttributeName)
	if err == nil && value != "" {
		returnValue[request.apikey+"_p_"+quotaAttributeName] = value
	}
	return returnValue, nil
}

//
func policySendAPIKey(request *requestInfo) (map[string]string, error) {
	returnValue := make(map[string]string, 1)

	returnValue["x-apikey"] = request.apikey

	return returnValue, nil
}

//
func policySendDeveloperEmail(request *requestInfo) (map[string]string, error) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-email"] = request.developer.Email

	return returnValue, nil
}

//
func policySendDeveloperID(request *requestInfo) (map[string]string, error) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-id"] = request.developer.DeveloperID

	return returnValue, nil
}

//
func policySendDeveloperAppName(request *requestInfo) (map[string]string, error) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-app-name"] = request.developerApp.Name

	return returnValue, nil
}

//
func policySendDeveloperAppID(request *requestInfo) (map[string]string, error) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-app-id"] = request.developerApp.AppID

	return returnValue, nil
}

func policyCheckIPAccessList(request *requestInfo) (map[string]string, error) {
	ipAccessList, err := shared.GetAttribute(request.developerApp.Attributes, "IPAccessList")

	if err == nil && ipAccessList != "" {
		if checkIPinAccessList(request.IP, ipAccessList) {
			return nil, nil
		}
		return nil, errors.New("Blocked by IP ACL")
	}
	// No IPACL attribute or it's empty: we allow request
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
		}
		// else FIXME increase unparsable ACL counter
	}
	// ip did not match any of the subnets found, request rejected
	return false
}
