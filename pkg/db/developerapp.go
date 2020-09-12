package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// DeveloperAppStore the developer app information storage interface
	DeveloperAppStore interface {
		// GetByOrganization retrieves all developer apps belonging to an organization
		GetByOrganization(organizationName string) (types.DeveloperApps, error)

		// GetByName returns a developer app
		GetByName(organization, developerAppName string) (*types.DeveloperApp, error)

		// GetByID returns a developer app
		GetByID(organization, developerAppID string) (*types.DeveloperApp, error)

		// GetCountByDeveloperID retrieves number of apps belonging to a developer
		GetCountByDeveloperID(developerID string) int

		// UpdateByName UPSERTs a developer app
		UpdateByName(app *types.DeveloperApp) error

		// DeleteByID deletes a developer app
		DeleteByID(organizationName, developerAppID string) error
	}
)
