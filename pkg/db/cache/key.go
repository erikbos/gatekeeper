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

// GetAll retrieves all keys
func (s *KeyCache) GetAll() (types.Keys, types.Error) {

	getKeys := func() (interface{}, types.Error) {
		return s.key.GetAll()
	}
	var keys types.Keys
	// TODO/FIXME
	if err := s.cache.fetchEntity(types.TypeKeyName, "--all-keys--", &keys, getKeys); err != nil {
		return nil, err
	}
	return keys, nil
}

// GetByKey returns details of a single apikey
func (s *KeyCache) GetByKey(key *string) (*types.Key, types.Error) {

	getByKey := func() (interface{}, types.Error) {
		return s.key.GetByKey(key)
	}
	var retrievedKey types.Key
	if err := s.cache.fetchEntity(types.TypeKeyName, *key, &retrievedKey, getByKey); err != nil {
		return nil, err
	}
	return &retrievedKey, nil
}

// GetByDeveloperAppID returns an array with apikey details of a developer app
func (s *KeyCache) GetByDeveloperAppID(developerAppID string) (types.Keys, types.Error) {

	getByAppID := func() (interface{}, types.Error) {
		return s.key.GetByDeveloperAppID(developerAppID)
	}
	var retrievedKeys types.Keys
	if err := s.cache.fetchEntity(types.TypeKeyName, developerAppID, &retrievedKeys, getByAppID); err != nil {
		return nil, err
	}
	return retrievedKeys, nil
}

// GetCountByAPIProductName counts the number of times an apiproduct has been assigned to keys
func (s *KeyCache) GetCountByAPIProductName(apiProductName string) (int, types.Error) {

	getCountOfApiProduct := func() (interface{}, types.Error) {
		return s.key.GetCountByAPIProductName(apiProductName)
	}
	var apiProductCount int
	if err := s.cache.fetchEntity(types.TypeKeyName, apiProductName, &apiProductCount, getCountOfApiProduct); err != nil {
		return 0, err
	}
	return apiProductCount, nil
}

// UpdateByKey UPSERTs keys in database
func (s *KeyCache) UpdateByKey(c *types.Key) types.Error {

	s.cache.deleteEntry(types.TypeKeyName, c.ConsumerKey)
	return s.key.UpdateByKey(c)
}

// DeleteByKey deletes keys
func (s *KeyCache) DeleteByKey(consumerKey string) types.Error {

	s.cache.deleteEntry(types.TypeKeyName, consumerKey)
	return s.key.DeleteByKey(consumerKey)
}
