package cache

import (
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// CredentialCache holds our database config
type CredentialCache struct {
	credential db.Credential
	cache      *Cache
}

// NewCredentialCache creates credential instance
func NewCredentialCache(cache *Cache, credential db.Credential) *CredentialCache {
	return &CredentialCache{
		credential: credential,
		cache:      cache,
	}
}

// GetByKey returns details of a single apikey
func (s *CredentialCache) GetByKey(key *string) (*types.DeveloperAppKey, types.Error) {

	getByKey := func() (interface{}, types.Error) {
		return s.credential.GetByKey(key)
	}
	var credential types.DeveloperAppKey
	if err := s.cache.fetchEntity(types.TypeCredentialName, *key, &credential, getByKey); err != nil {
		return nil, err
	}
	return &credential, nil
}

// GetByDeveloperAppID returns an array with apikey details of a developer app
func (s *CredentialCache) GetByDeveloperAppID(developerAppID string) (types.DeveloperAppKeys, types.Error) {

	getByAppID := func() (interface{}, types.Error) {
		return s.credential.GetByDeveloperAppID(developerAppID)
	}
	var credentials types.DeveloperAppKeys
	if err := s.cache.fetchEntity(types.TypeCredentialName, developerAppID, &credentials, getByAppID); err != nil {
		return nil, err
	}
	return credentials, nil
}

// UpdateByKey UPSERTs credentials in database
func (s *CredentialCache) UpdateByKey(c *types.DeveloperAppKey) types.Error {

	s.cache.deleteEntry(types.TypeCredentialName, c.ConsumerKey)
	return s.credential.UpdateByKey(c)
}

// DeleteByKey deletes credentials
func (s *CredentialCache) DeleteByKey(consumerKey string) types.Error {

	s.cache.deleteEntry(types.TypeCredentialName, consumerKey)
	return s.credential.DeleteByKey(consumerKey)
}
