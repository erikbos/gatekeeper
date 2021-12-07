package cache

import (
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// KeyCache holds our database config
type KeyCache struct {
	key   db.Key
	cache *Cache
}

// NewKeyCache creates key instance
func NewKeyCache(cache *Cache, key db.Key) *KeyCache {
	return &KeyCache{
		key:   key,
		cache: cache,
	}
}

// GetByKey returns details of a single apikey
func (s *KeyCache) GetByKey(organizationName, key *string) (*types.Key, types.Error) {

	getByKey := func() (interface{}, types.Error) {
		return s.key.GetByKey(organizationName, key)
	}
	var retrievedKey types.Key
	if err := s.cache.fetchEntity(types.TypeKeyName, *key, &retrievedKey, getByKey); err != nil {
		return nil, err
	}
	return &retrievedKey, nil
}

// GetByDeveloperAppID returns an array with apikey details of a developer app
func (s *KeyCache) GetByDeveloperAppID(organizationName, developerAppID string) (types.Keys, types.Error) {

	getByAppID := func() (interface{}, types.Error) {
		return s.key.GetByDeveloperAppID(organizationName, developerAppID)
	}
	var retrievedKeys types.Keys
	if err := s.cache.fetchEntity(types.TypeKeyName, developerAppID, &retrievedKeys, getByAppID); err != nil {
		return nil, err
	}
	return retrievedKeys, nil
}

// GetCountByAPIProductName counts the number of times an apiproduct has been assigned to keys
func (s *KeyCache) GetCountByAPIProductName(organizationName, apiProductName string) (int, types.Error) {

	getCountOfAPIProduct := func() (interface{}, types.Error) {
		return s.key.GetCountByAPIProductName(organizationName, apiProductName)
	}
	var apiProductCount int
	if err := s.cache.fetchEntity(types.TypeKeyName, apiProductName, &apiProductCount, getCountOfAPIProduct); err != nil {
		return 0, err
	}
	return apiProductCount, nil
}

// UpdateByKey UPSERTs keys in database
func (s *KeyCache) UpdateByKey(organizationName string, c *types.Key) types.Error {

	s.cache.deleteEntry(types.TypeKeyName, c.ConsumerKey)
	return s.key.UpdateByKey(organizationName, c)
}

// DeleteByKey deletes keys
func (s *KeyCache) DeleteByKey(organizationName, consumerKey string) types.Error {

	s.cache.deleteEntry(types.TypeKeyName, consumerKey)
	return s.key.DeleteByKey(organizationName, consumerKey)
}
