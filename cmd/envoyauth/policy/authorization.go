package policy

import (
	"errors"

	"github.com/bmatcuk/doublestar"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/envoyauth/request"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// CheckProductEntitlement loads developer, dev app, apiproduct details,
// as input request.apikey must be set
//
func (p *Policy) CheckProductEntitlement(request *request.State) error {

	if err := p.getAPIKeyDevDevAppDetails(request); err != nil {
		return err
	}
	if err := p.checkDevAndKeyValidity(request); err != nil {
		return err
	}
	var err error
	request.APIProduct, err = p.IsPathAllowed(request.URL.Path, request.Key)
	return err
}

// getAPIKeyDevDevAppDetails populates apikey, developer and developerapp details
func (p *Policy) getAPIKeyDevDevAppDetails(request *request.State) error {

	if request == nil {
		return errors.New("no request details available")
	}

	var err error
	request.Key, err = p.config.db.Key.GetByKey(request.ConsumerKey)
	if err != nil {
		p.config.metrics.IncUnknownAPIKey(request)
		return errors.New("cannot find apikey")
	}

	request.DeveloperApp, err = p.config.db.DeveloperApp.GetByID(request.Key.AppID)
	if err != nil {
		p.config.metrics.IncDatabaseFetchFailure(request)
		return errors.New("cannot find developer app of this apikey")
	}

	request.Developer, err = p.config.db.Developer.GetByID(request.DeveloperApp.DeveloperID)
	if err != nil {
		p.config.metrics.IncDatabaseFetchFailure(request)
		return errors.New("cannot find developer of developer app")
	}
	return nil
}

// checkDevAndKeyValidity checks devapp approval and expiry status
func (p *Policy) checkDevAndKeyValidity(request *request.State) error {

	if request == nil {
		return errors.New("no request details available")
	}

	if !request.Developer.IsActive() {
		return errors.New("developer not active")
	}

	if request.Developer.IsSuspended(request.Timestamp) {
		return errors.New("developer suspended")
	}

	if !request.Key.IsApproved() {
		// FIXME increase unapproved dev app counter (not an error state)
		return errors.New("unapproved apikey")
	}

	if request.Key.IsExpired(request.Timestamp) {
		// FIXME increase expired key counter (not an error state))
		return errors.New("expired apikey")
	}
	return nil
}

// IsPathAllowed checks whether paths is allowed by apikey,
// this means the apikey needs to contain a product that matchs the request path
func (p *Policy) IsPathAllowed(requestPath string,
	key *types.Key) (*types.APIProduct, error) {

	// Does this apikey have any products assigned?
	if len(key.APIProducts) == 0 {
		return nil, errors.New("no active products for apikey")
	}

	// Iterate over this key's apiproducts
	for _, apiproduct := range key.APIProducts {
		if apiproduct.IsApproved() {
			apiproductDetails, err := p.config.db.APIProduct.Get(apiproduct.Apiproduct)
			if err != nil {
				// apikey has product in it which we cannot find:
				// FIXME increase "unknown product in apikey" counter (not an error state)
			} else {
				// Iterate over all paths of apiproduct and try to match with path of request
				for _, productPath := range apiproductDetails.Paths {
					p.config.logger.Debug("IsRequestPathAllowed",
						zap.String("productpath", productPath),
						zap.String("requestpath", requestPath))

					if ok, _ := doublestar.Match(productPath, requestPath); ok {
						return apiproductDetails, nil
					}
				}
			}
		}
	}
	return nil, errors.New("not authorized for requested path")
}
