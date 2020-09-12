package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// DeveloperStore the developer information storage interface
	DeveloperStore interface {
		// GetByOrganization retrieves all developer belonging to an organization
		GetByOrganization(organizationName string) (types.Developers, error)

		// GetCountByOrganization retrieves number of developer belonging to an organization
		GetCountByOrganization(organizationName string) int

		// GetByEmail retrieves a developer
		GetByEmail(developerOrganization, developerEmail string) (*types.Developer, error)

		// GetByID retrieves a developer
		GetByID(developerID string) (*types.Developer, error)

		// UpdateByName UPSERTs a developer
		UpdateByName(dev *types.Developer) error

		// DeleteByEmail deletes a developer
		DeleteByEmail(organizationName, developerEmail string) error
	}
)
