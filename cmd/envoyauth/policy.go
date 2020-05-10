package main

import (
	"net/http"
	"strings"

	"github.com/erikbos/apiauth/pkg/shared"
)

// handlePolicies invokes all policy functions to retrieve additional upstream headers
func handlePolicies(request *requestInfo, headersToReturn map[string]string) {

	for _, policy := range strings.Split(request.APIProduct.Scopes, ",") {
		// log.Infof("q %s", policy)
		headersToAdd, _ := handlePolicy(policy, request)

		for key, value := range headersToAdd {
			// log.Infof("q2 %s", key)
			headersToReturn[key] = value
		}
	}
}

func handlePolicy(policy string, request *requestInfo) (map[string]string, int) {
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
	}
	// FIXME insert counter for unknown policy name in an active product
	// label: product, policyname
	return nil, http.StatusOK
}

//
func policySendAPIKey(request *requestInfo) (map[string]string, int) {
	returnValue := make(map[string]string, 1)

	returnValue["x-apikey"] = request.apikey

	return returnValue, http.StatusOK
}

//
func policySendDeveloperEmail(request *requestInfo) (map[string]string, int) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-email"] = request.developer.Email

	return returnValue, http.StatusOK
}

//
func policySendDeveloperID(request *requestInfo) (map[string]string, int) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-id"] = request.developer.DeveloperID

	return returnValue, http.StatusOK
}

//
func policySendDeveloperAppName(request *requestInfo) (map[string]string, int) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-app-name"] = request.developerApp.Name

	return returnValue, http.StatusOK
}

//
func policySendDeveloperAppID(request *requestInfo) (map[string]string, int) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-app-id"] = request.developerApp.AppID

	return returnValue, http.StatusOK
}

// policyQPS1 returns QPS quotakey to be used by Lyft ratelimiter
// QPS set as developer app attribute has priority over quota set as product attribute
//
func policyQPS1(request *requestInfo) (map[string]string, int) {
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
	return returnValue, http.StatusOK
}
