package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// CredentialStore the cluster information storage interface
	CredentialStore interface {
		// GetByKey returns details of a single apikey
		GetByKey(organizationName, key *string) (*shared.DeveloperAppKey, error)

		// GetByDeveloperAppID returns an array with apikey details of a developer app
		GetByDeveloperAppID(developerAppID string) ([]shared.DeveloperAppKey, error)

		// GetCountByDeveloperAppID retrieves number of keys beloning to developer app
		GetCountByDeveloperAppID(developerAppID string) int

		// UpdateByKey UPSERTs credentials
		UpdateByKey(c *shared.DeveloperAppKey) error

		// DeleteByKey deletes credentials
		DeleteByKey(organizationName, consumerKey string) error
	}
)
