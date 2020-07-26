package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// OrganizationStore the organization information storage interface
	OrganizationStore interface {
		// GetAll retrieves all organizations
		GetAll() ([]shared.Organization, error)

		// GetByName retrieves an organization
		GetByName(organizationName string) (*shared.Organization, error)

		// UpdateByName UPSERTs an organization
		UpdateByName(o *shared.Organization) error

		// DeleteByName deletes an organization
		DeleteByName(organizationToDelete string) error
	}
)
