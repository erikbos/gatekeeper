package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// OrganizationStore the organization information storage interface
	OrganizationStore interface {
		// GetAll retrieves all organizations
		GetAll() (types.Organizations, error)

		// GetByName retrieves an organization
		GetByName(organizationName string) (*types.Organization, error)

		// UpdateByName UPSERTs an organization
		UpdateByName(o *types.Organization) error

		// DeleteByName deletes an organization
		DeleteByName(organizationToDelete string) error
	}
)
