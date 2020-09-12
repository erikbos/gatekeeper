package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// APIProductStore the apiproduct information storage interface
	APIProductStore interface {
		// GetAll retrieves all api products
		GetAll() (types.APIProducts, error)

		// GetByOrganization retrieves all api products belonging to an organization
		GetByOrganization(organizationName string) (types.APIProducts, error)

		// GetByName returns an apiproduct
		GetByName(organizationName, apiproductName string) (*types.APIProduct, error)

		// UpdateByName UPSERTs an apiproduct in database
		UpdateByName(p *types.APIProduct) error

		// DeleteByName deletes an apiproduct
		DeleteByName(organizationName, apiProduct string) error
	}
)
