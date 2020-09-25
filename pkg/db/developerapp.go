package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// DeveloperAppStore the developer app information storage interface
	DeveloperAppStore interface {
		// GetByOrganization retrieves all developer apps belonging to an organization
		GetByOrganization(organizationName string) (types.DeveloperApps, types.Error)

		// GetByName returns a developer app
		GetByName(organization, developerAppName string) (*types.DeveloperApp, types.Error)

		// GetByID returns a developer app
		GetByID(organization, developerAppID string) (*types.DeveloperApp, types.Error)

		// GetCountByDeveloperID retrieves number of apps belonging to a developer
		GetCountByDeveloperID(developerID string) (int, types.Error)

		// UpdateByName UPSERTs a developer app
		Update(app *types.DeveloperApp) types.Error

		// DeleteByID deletes a developer app
		DeleteByID(organizationName, developerAppID string) types.Error
	}
)
