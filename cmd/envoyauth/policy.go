package main

import (
	"net/http"
	"strings"

	"github.com/erikbos/apiauth/pkg/shared"
)

// handlePolicies invokes all policy functions to retrieve additional upstream headers
func handlePolicies(entitlement sessionState, headersToReturn map[string]string) {

	for _, policy := range strings.Split(entitlement.APIProduct.Scopes, ",") {
		// log.Infof("q %s", policy)
		headersToAdd, _ := handlePolicy(policy, entitlement)

		for key, value := range headersToAdd {
			// log.Infof("q2 %s", key)
			headersToReturn[key] = value
		}
	}
}

func handlePolicy(policy string, entitlement sessionState) (map[string]string, int) {
	switch policy {
	case "qps":
		return policyQPS1(entitlement)
	case "sendDeveloperEmail":
		return policySendDeveloperEmail(entitlement)
	case "sendDeveloperID":
		return policySendDeveloperID(entitlement)
	case "sendDeveloperAppName":
		return policySendDeveloperAppName(entitlement)
	case "sendDeveloperAppID":
		return policySendDeveloperAppID(entitlement)
	}
	// FIXME insert counter for unknown policy name
	return nil, http.StatusOK
}

//
func policySendDeveloperEmail(entitlement sessionState) (map[string]string, int) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-email"] = entitlement.developer.Email

	return returnValue, http.StatusOK
}

//
func policySendDeveloperID(entitlement sessionState) (map[string]string, int) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-id"] = entitlement.developer.DeveloperID

	return returnValue, http.StatusOK
}

//
func policySendDeveloperAppName(entitlement sessionState) (map[string]string, int) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-app-name"] = entitlement.developerApp.Name

	return returnValue, http.StatusOK
}

//
func policySendDeveloperAppID(entitlement sessionState) (map[string]string, int) {
	returnValue := make(map[string]string, 1)

	returnValue["x-developer-app-id"] = entitlement.developerApp.AppID

	return returnValue, http.StatusOK
}

// policyQPS1 returns QPS quotakey to be used by Lyft ratelimiter
// QPS set as developer app attribute has priority over quota set as product attribute
//
func policyQPS1(entitlement sessionState) (map[string]string, int) {
	quotaAttributeName := entitlement.APIProduct.Name + "_quotaPerSecond"

	returnValue := make(map[string]string)

	value, err := shared.GetAttribute(entitlement.developerApp.Attributes, quotaAttributeName)
	if err == nil && value != "" {
		returnValue[entitlement.apikey+"_a_"+quotaAttributeName] = value
	}
	value, err = shared.GetAttribute(entitlement.APIProduct.Attributes, quotaAttributeName)
	if err == nil && value != "" {
		returnValue[entitlement.apikey+"_p_"+quotaAttributeName] = value
	}
	return returnValue, http.StatusOK
}
