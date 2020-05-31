package main

import (
	"errors"
	"net/url"
	"strings"

	"github.com/bmatcuk/doublestar"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

type function func(policy string, request *requestInfo) (map[string]string, error)

// handleVhostPolicy executes a single policy to optionally add upstream headers
func (a *authorizationServer) handleVhostPolicy(policy string, request *requestInfo) (map[string]string, error) {

	a.metrics.virtualHostPolicy.WithLabelValues(request.httpRequest.Host, policy).Inc()

	switch policy {
	case "checkapikey":
		return a.checkAPIKey(request)
	case "geoiplookup":
		return a.geoIPLookup(request)
	}

	a.metrics.virtualHostPolicyUnknown.WithLabelValues(request.httpRequest.Host, policy).Inc()
	return nil, nil
}

// checkAPIKey tries to find key in querystring, loads dev app, dev details, and check whether path is allowed
func (a *authorizationServer) checkAPIKey(request *requestInfo) (map[string]string, error) {

	var err error
	request.apikey, err = getAPIkeyFromQueryString(request.queryParameters)
	if err != nil || request.apikey == "" {
		return nil, err
	}

	err = a.CheckProductEntitlement(request.vhost.OrganizationName, request)
	if err != nil {
		log.Debugf("CheckProductEntitlement() not allowed '%s' (%s)", request.URL.Path, err.Error())
		a.increaseCounterApikeyNotfound(request)
		return nil, err
	}
	return nil, nil
}

// getAPIkeyFromQueryString extracts apikey from query parameters
func getAPIkeyFromQueryString(queryParameters url.Values) (string, error) {

	// iterate over queryparameters be able to Find Them in alL CasEs
	for param, value := range queryParameters {
		param := strings.ToLower(param)

		// we allow both spellings
		if param == "apikey" || param == "key" {
			if len(value) == 1 {
				return value[0], nil
			}
			return "", errors.New("apikey has no value")
		}
	}
	return "", errors.New("querystring does not contain apikey")
}

// geoIPLookup lookup requestor's ip address in geoip database
func (a *authorizationServer) geoIPLookup(request *requestInfo) (map[string]string, error) {

	if a.g == nil {
		return nil, nil
	}
	country, state := a.g.GetCountryAndState(request.IP)
	if country == "" {
		return nil, nil
	}

	a.metrics.requestsPerCountry.WithLabelValues(country).Inc()

	return map[string]string{
			"geoip-country": country,
			"geoip-state":   state,
		},
		nil
}

func (a *authorizationServer) CheckProductEntitlement(organization string, request *requestInfo) error {

	err := a.getEntitlementDetails(request)
	if err != nil {
		return err
	}

	if err = checkDevAndKeyValidity(request); err != nil {
		return err
	}

	request.APIProduct, err = a.IsRequestPathAllowed(organization, request.URL.Path, request.appCredential)
	if err != nil {
		return err
	}

	return nil
}

// getEntitlementDetails populates apikey, developer and developerapp details
func (a *authorizationServer) getEntitlementDetails(request *requestInfo) error {
	var err error

	request.appCredential, err = a.db.GetAppCredentialByKey(request.vhost.OrganizationName, request.apikey)
	if err != nil {
		// FIX ME increase unknown apikey counter (not an error state)
		return errors.New("Could not find apikey")
	}

	request.developerApp, err = a.db.GetDeveloperAppByID(request.vhost.OrganizationName, request.appCredential.AppID)
	if err != nil {
		// FIX ME increase counter as every apikey should link to dev app (error state)
		return errors.New("Could not find developer app of this apikey")
	}

	request.developer, err = a.db.GetDeveloperByID(request.developerApp.DeveloperID)
	if err != nil {
		// FIX ME increase counter as every devapp should link to developer (error state)
		return errors.New("Could not find developer of this apikey")
	}

	return nil
}

// checkDevAndKeyValidity checks devapp approval and expiry status
func checkDevAndKeyValidity(request *requestInfo) error {

	now := shared.GetCurrentTimeMilliseconds()

	if request.developer.SuspendedTill != -1 &&
		now < request.developer.SuspendedTill {

		return errors.New("Developer suspended")
	}

	if request.appCredential.Status != "approved" {
		// FIXME increase unapproved dev app counter (not an error state)
		return errors.New("Unapproved apikey")
	}

	if request.appCredential.ExpiresAt != -1 {
		if now > request.appCredential.ExpiresAt {
			// FIXME increase expired dev app credentials counter (not an error state))
			return errors.New("Expired apikey")
		}
	}
	return nil
}

// IsRequestPathAllowed
// - iterate over products in apikey
// - 	iterate over path(s) of each product:
// - 		if requestor path matches paths(s)
// -			- return 200
// - if not 403

func (a *authorizationServer) IsRequestPathAllowed(organization, requestPath string,
	appcredential shared.DeveloperAppKey) (shared.APIProduct, error) {

	// Does this apikey have any products assigned?
	if len(appcredential.APIProducts) == 0 {
		return shared.APIProduct{}, errors.New("No active products")
	}

	// Iterate over this key's apiproducts
	for _, apiproduct := range appcredential.APIProducts {
		if apiproduct.Status == "approved" {
			apiproductDetails, err := a.db.GetAPIProductByName(organization, apiproduct.Apiproduct)
			if err != nil {
				// apikey has non existing product in it:
				// FIXME increase "unknown product in apikey" counter (not an error state)
			} else {
				// Iterate over all paths of apiproduct and try to match with path of request
				for _, productPath := range apiproductDetails.Paths {
					// log.Debugf("IsRequestPathAllowed() Matching path %s in %s", requestPath, productPath)
					if ok, _ := doublestar.Match(productPath, requestPath); ok {
						// log.Debugf("IsRequestPathAllowed: match!")
						return apiproductDetails, nil
					}
				}
			}
		}
	}
	return shared.APIProduct{}, errors.New("No product active for requested path")
}
