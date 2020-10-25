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
func (a *server) CheckProductEntitlement(request *requestDetails) error {

	if err := a.getAPIKeyDevDevAppDetails(request); err != nil {
		return err
	}
	if err := a.checkDevAndKeyValidity(request); err != nil {
		return err
	}
	var err error
	request.APIProduct, err = a.IsRequestPathAllowed(request.URL.Path, request.appCredential)
	return err
}

// getAPIKeyDevDevAppDetails populates apikey, developer and developerapp details
func (a *server) getAPIKeyDevDevAppDetails(r *requestDetails) error {
	var err error

	r.appCredential, err = a.db.Credential.GetByKey(r.consumerKey)
	if err != nil {
		// FIX ME increase unknown apikey counter (not an error state)
		return errors.New("Cannot find apikey")
	}

	r.developerApp, err = a.db.DeveloperApp.GetByID(r.appCredential.AppID)
	if err != nil {
		// FIX ME increase counter as every apikey should link to dev app (error state)
		return errors.New("Cannot find developer app of this apikey")
	}

	r.developer, err = a.db.Developer.GetByID(r.developerApp.DeveloperID)
	if err != nil {
		// FIX ME increase counter as every devapp should link to developer (error state)
		return errors.New("Cannot find developer of developer app")
	}

	return nil
}

// checkDevAndKeyValidity checks devapp approval and expiry status
func (a *server) checkDevAndKeyValidity(request *requestDetails) error {

	now := shared.GetCurrentTimeMilliseconds()

	if request.developer.SuspendedTill != -1 &&
		now < request.developer.SuspendedTill {

		return errors.New("Developer suspended")
	}

	if !request.appCredential.IsApproved() {
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

// IsRequestPathAllowed checks whether paths is allowed by apikey
// this means the apikey needs to contain a product that matchs the path
func (a *server) IsRequestPathAllowed(requestPath string,
	credential *types.DeveloperAppKey) (*types.APIProduct, error) {

	// Does this apikey have any products assigned?
	if len(credential.APIProducts) == 0 {
		return nil, errors.New("No active products")
	}

	// Iterate over this key's apiproducts
	for _, apiproduct := range credential.APIProducts {
		if apiproduct.IsApproved() {
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
