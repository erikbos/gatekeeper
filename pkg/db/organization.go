package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// OrganizationStore the organization information storage interface
	OrganizationStore interface {
		// GetAll retrieves all organizations
		GetAll() (types.Organizations, types.Error)

		// Get retrieves an organization
		Get(organizationName string) (*types.Organization, types.Error)

		// Update UPSERTs an organization
		Update(o *types.Organization) types.Error

		// Delete deletes an organization
		Delete(organizationToDelete string) types.Error
	}
)
