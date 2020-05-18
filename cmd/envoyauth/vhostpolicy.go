package main

import (
	"errors"

	"github.com/bmatcuk/doublestar"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/apiauth/pkg/shared"
)

// handleVhostPolicy executes a single policy to optionally add upstream headers
func (a *authorizationServer) handleVhostPolicy(policy string, request *requestInfo) (map[string]string, error) {

	a.metrics.virtualHostPolicy.WithLabelValues(request.httpRequest.Host, policy).Inc()

	switch policy {
	case "checkapi":
		return a.checkAPIKey(request)
	case "geoiplookup":
		return a.geoIPLookup(request)
	}

	a.metrics.virtualHostPolicyUnknown.WithLabelValues(request.httpRequest.Host, policy).Inc()
	return nil, nil
}

func (a *authorizationServer) checkAPIKey(request *requestInfo) (map[string]string, error) {

	request.apikey = request.queryParameters["apikey"][0]

	err := a.CheckProductEntitlement("petstore", request)
	if err != nil {
		log.Debugf("CheckProductEntitlement() not allowed '%s' (%s)", request.URL.Path, err.Error())
		a.increaseCounterApikeyNotfound(request)
		return nil, err
	}
	return nil, nil
}

// geoIPLookup lookup requestor's ip address in geoip database
func (a *authorizationServer) geoIPLookup(request *requestInfo) (map[string]string, error) {

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

	err := a.getEntitlementDetails(organization, request)
	if err != nil {
		return err
	}
	if err = checkAppCredentialValidity(request.appCredential); err != nil {
		return err
	}
	request.APIProduct, err = a.IsRequestPathAllowed(organization, request.URL.Path, request.appCredential)
	if err != nil {
		return errors.New("No product match")
	}
	return nil
}

// getEntitlementDetails populates apikey, developer and developerapp details
func (a *authorizationServer) getEntitlementDetails(organization string, request *requestInfo) error {
	var err error

	request.appCredential, err = a.db.GetAppCredentialByKey(organization, request.apikey)
	if err != nil {
		// FIX ME increase unknown apikey counter (not an error state)
		return errors.New("Could not find apikey")
	}

	request.developerApp, err = a.db.GetDeveloperAppByID(organization, request.appCredential.OrganizationAppID)
	if err != nil {
		// FIX ME increase counter as every apikey should link to dev app (error state)
		return errors.New("Could not find developer app of this apikey")
	}

	request.developer, err = a.db.GetDeveloperByID(request.developerApp.ParentID)
	if err != nil {
		// FIX ME increase counter as every devapp should link to developer (error state)
		return errors.New("Could not find developer of this apikey")
	}

	return nil
}

// checkAppCredentialValidity checks devapp approval and expiry status
func checkAppCredentialValidity(appcredential shared.AppCredential) error {
	if appcredential.Status != "approved" {
		// FIXME increase unapproved dev app counter (not an error state)
		return errors.New("Unapproved apikey")
	}
	if appcredential.ExpiresAt != -1 {
		if shared.GetCurrentTimeMilliseconds() > appcredential.ExpiresAt {
			// FIXME increase expired dev app credentials counter (not an error state))
			return errors.New("Expired apikey")
		}
	}
	return nil
}

// IsRequestPathAllowed
// - iterate over products in apikey
// - 	iterate over resource path(s) of each product:
// - 		if requestor path matches apiresource_path(s)
// -			- return 200
// - if not 403

func (a *authorizationServer) IsRequestPathAllowed(organization, requestPath string,
	appcredential shared.AppCredential) (shared.APIProduct, error) {

	// Iterate over this key's apiproducts
	for _, apiproduct := range appcredential.APIProducts {
		if apiproduct.Status == "approved" {
			// Retrieve details of each api product embedded in key
			apiproduct, err := a.db.GetAPIProductByName(organization, apiproduct.Apiproduct)
			if err != nil {
				// FIXME increase unknown product in apikey counter (not an error state)
			} else {
				// Iterate over apiresource(paths) of apiproduct
				for _, productPath := range apiproduct.APIResources {
					// log.Debugf("IsRequestPathAllowed() Matching path %s in %s", requestPath, productPath)
					if ok, _ := doublestar.Match(productPath, requestPath); ok {
						// log.Debugf("IsRequestPathAllowed: match!")
						return apiproduct, nil
					}
				}
			}
		}
	}
	return shared.APIProduct{}, errors.New("No access")
}
