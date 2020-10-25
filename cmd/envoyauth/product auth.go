package main

import (
	"errors"

	"github.com/bmatcuk/doublestar"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// CheckProductEntitlement loads developer, dev app, apiproduct details,
// as input request.apikey must be set
//
func (a *authorizationServer) CheckProductEntitlement(request *requestInfo) error {

	if err := a.getAPIKeyDevDevAppDetails(request); err != nil {
		return err
	}
	if err := checkDevAndKeyValidity(request); err != nil {
		return err
	}
	var err error
	request.APIProduct, err = a.IsRequestPathAllowed(request.URL.Path, request.appCredential)
	return err
}

// getAPIKeyDevDevAppDetails populates apikey, developer and developerapp details
func (a *authorizationServer) getAPIKeyDevDevAppDetails(request *requestInfo) error {
	var err error

	request.appCredential, err = a.db.Credential.GetByKey(request.apikey)
	if err != nil {
		// FIX ME increase unknown apikey counter (not an error state)
		return errors.New("Cannot find apikey")
	}

	request.developerApp, err = a.db.DeveloperApp.GetByID(request.appCredential.AppID)
	if err != nil {
		// FIX ME increase counter as every apikey should link to dev app (error state)
		return errors.New("Cannot find developer app of this apikey")
	}

	request.developer, err = a.db.Developer.GetByID(request.developerApp.DeveloperID)
	if err != nil {
		// FIX ME increase counter as every devapp should link to developer (error state)
		return errors.New("Cannot find developer of developer app")
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

func (a *authorizationServer) IsRequestPathAllowed(requestPath string,
	credential *types.DeveloperAppKey) (*types.APIProduct, error) {

	// Does this apikey have any products assigned?
	if len(credential.APIProducts) == 0 {
		return nil, errors.New("No active products")
	}

	// Iterate over this key's apiproducts
	for _, apiproduct := range credential.APIProducts {
		if apiproduct.Status == "approved" {

			apiproductDetails, err := a.db.APIProduct.Get(apiproduct.Apiproduct)
			if err != nil {
				// apikey has product in it which we cannot find:
				// FIXME increase "unknown product in apikey" counter (not an error state)
			} else {
				// Iterate over all paths of apiproduct and try to match with path of request
				for _, productPath := range apiproductDetails.Paths {
					a.logger.Debug("IsRequestPathAllowed",
						zap.String("productpath", productPath),
						zap.String("requestpath", requestPath))

					if ok, _ := doublestar.Match(productPath, requestPath); ok {
						return apiproductDetails, nil
					}
				}
			}
		}
	}
	return nil, errors.New("Not authorized for requested path")
}
