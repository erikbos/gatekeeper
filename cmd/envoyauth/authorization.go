package main

import (
	"errors"

	"github.com/bmatcuk/doublestar"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// CheckProductEntitlement loads developer, dev app, apiproduct details,
// as input request.apikey must be set
//
func (s *server) CheckProductEntitlement(request *requestDetails) error {

	if err := s.getAPIKeyDevDevAppDetails(request); err != nil {
		return err
	}
	if err := s.checkDevAndKeyValidity(request); err != nil {
		return err
	}
	var err error
	request.APIProduct, err = s.IsPathAllowed(request.URL.Path, request.developerAppKey)
	return err
}

// getAPIKeyDevDevAppDetails populates apikey, developer and developerapp details
func (s *server) getAPIKeyDevDevAppDetails(request *requestDetails) error {

	if request == nil {
		return errors.New("No request details available")
	}

	var err error
	request.developerAppKey, err = s.db.Credential.GetByKey(request.consumerKey)
	if err != nil {
		s.metrics.IncUnknownAPIKey(request)
		return errors.New("Cannot find apikey")
	}

	request.developerApp, err = s.db.DeveloperApp.GetByID(request.developerAppKey.AppID)
	if err != nil {
		s.metrics.IncDatabaseFetchFailure(request)
		return errors.New("Cannot find developer app of this apikey")
	}

	request.developer, err = s.db.Developer.GetByID(request.developerApp.DeveloperID)
	if err != nil {
		s.metrics.IncDatabaseFetchFailure(request)
		return errors.New("Cannot find developer of developer app")
	}
	return nil
}

// checkDevAndKeyValidity checks devapp approval and expiry status
func (s *server) checkDevAndKeyValidity(request *requestDetails) error {

	if request == nil {
		return errors.New("No request details available")
	}

	if !request.developer.IsActive() {
		return errors.New("Developer not active")
	}

	if request.developer.IsSuspended(request.timestamp) {
		return errors.New("Developer suspended")
	}

	if !request.developerAppKey.IsApproved() {
		// FIXME increase unapproved dev app counter (not an error state)
		return errors.New("Unapproved apikey")
	}

	if request.developerAppKey.IsExpired(request.timestamp) {
		// FIXME increase expired dev app credentials counter (not an error state))
		return errors.New("Expired apikey")
	}
	return nil
}

// IsPathAllowed checks whether paths is allowed by apikey,
// this means the apikey needs to contain a product that matchs the request path
func (s *server) IsPathAllowed(requestPath string,
	credential *types.DeveloperAppKey) (*types.APIProduct, error) {

	// Does this apikey have any products assigned?
	if len(credential.APIProducts) == 0 {
		return nil, errors.New("No active products for apikey")
	}

	// Iterate over this key's apiproducts
	for _, apiproduct := range credential.APIProducts {
		if apiproduct.IsApproved() {
			apiproductDetails, err := s.db.APIProduct.Get(apiproduct.Apiproduct)
			if err != nil {
				// apikey has product in it which we cannot find:
				// FIXME increase "unknown product in apikey" counter (not an error state)
			} else {
				// Iterate over all paths of apiproduct and try to match with path of request
				for _, productPath := range apiproductDetails.Paths {
					s.logger.Debug("IsRequestPathAllowed",
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
