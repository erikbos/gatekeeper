package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// DeveloperStore the developer information storage interface
	DeveloperStore interface {
		// GetByOrganization retrieves all developer belonging to an organization
		GetByOrganization(organizationName string) (types.Developers, types.Error)

		// GetCountByOrganization retrieves number of developer belonging to an organization
		GetCountByOrganization(organizationName string) (int, types.Error)

		// GetByEmail retrieves a developer
		GetByEmail(developerOrganization, developerEmail string) (*types.Developer, types.Error)

		// GetByID retrieves a developer
		GetByID(developerID string) (*types.Developer, types.Error)

		// Update UPSERTs a developer
		Update(dev *types.Developer) types.Error

		// DeleteByID deletes a developer
		DeleteByID(organizationName, developerID string) types.Error
	}
)
