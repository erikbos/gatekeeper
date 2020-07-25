package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// APIProductStore the apiproduct information storage interface
	APIProductStore interface {
		// GetByOrganization retrieves all api products belonging to an organization
		GetByOrganization(organizationName string) ([]shared.APIProduct, error)

		// GetByName returns an apiproduct
		GetByName(organizationName, apiproductName string) (*shared.APIProduct, error)

		// UpdateByName UPSERTs an apiproduct in database
		UpdateByName(p *shared.APIProduct) error

		// DeleteByName deletes an apiproduct
		DeleteByName(organizationName, apiProduct string) error
	}
)
