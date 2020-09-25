package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// CredentialStore the cluster information storage interface
	CredentialStore interface {
		// GetByKey returns details of a single apikey
		GetByKey(organizationName, key *string) (*types.DeveloperAppKey, types.Error)

		// GetByDeveloperAppID returns an array with apikey details of a developer app
		GetByDeveloperAppID(developerAppID string) (types.DeveloperAppKeys, types.Error)

		// UpdateByKey UPSERTs credentials
		UpdateByKey(c *types.DeveloperAppKey) types.Error

		// DeleteByKey deletes credentials
		DeleteByKey(organizationName, consumerKey string) types.Error
	}
)
