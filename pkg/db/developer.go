package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// DeveloperStore the developer information storage interface
	DeveloperStore interface {
		// GetByOrganization retrieves all developer belonging to an organization
		GetByOrganization(organizationName string) ([]shared.Developer, error)

		// GetCountByOrganization retrieves number of developer belonging to an organization
		GetCountByOrganization(organizationName string) int

		// GetByEmail retrieves a developer
		GetByEmail(developerOrganization, developerEmail string) (*shared.Developer, error)

		// GetByID retrieves a developer
		GetByID(developerID string) (*shared.Developer, error)

		// UpdateByName UPSERTs a developer
		UpdateByName(dev *shared.Developer) error

		// DeleteByEmail deletes a developer
		DeleteByEmail(organizationName, developerEmail string) error
	}
)
