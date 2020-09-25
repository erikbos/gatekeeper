package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// APIProductStore the apiproduct information storage interface
	APIProductStore interface {
		// GetAll retrieves all api products
		GetAll() (types.APIProducts, types.Error)

		// GetByOrganization retrieves all api products belonging to an organization
		GetByOrganization(organizationName string) (types.APIProducts, types.Error)

		// Get returns an apiproduct
		Get(organizationName, apiproductName string) (*types.APIProduct, types.Error)

		// Update UPSERTs an apiproduct in database
		Update(p *types.APIProduct) types.Error

		// Delete deletes an apiproduct
		Delete(organizationName, apiProduct string) types.Error
	}
)
