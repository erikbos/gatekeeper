package service

import (
	"fmt"

	"github.com/dchest/uniuri"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// KeyService is
type KeyService struct {
	db        *db.Database
	changelog *Changelog
}

// NewKey returns a new key instance
func NewKey(database *db.Database, c *Changelog) *KeyService {

	return &KeyService{
		db:        database,
		changelog: c,
	}
}

// GetAll returns all keys
func (ks *KeyService) GetAll() (keys types.Keys, err types.Error) {

	return ks.db.Key.GetAll()
}

// Get returns details of an key
func (ks *KeyService) Get(key string) (*types.Key, types.Error) {

	return ks.db.Key.GetByKey(&key)
}

// GetByDeveloperAppID returns all keys of a developer app
func (ks *KeyService) GetByDeveloperAppID(developerAppID string) (types.Keys, types.Error) {

	return ks.db.Key.GetByDeveloperAppID(developerAppID)
}

// Create creates a key
func (ks *KeyService) Create(newKey types.Key, developerApp *types.DeveloperApp,
	who Requester) (types.Key, types.Error) {

	if _, err := ks.db.Key.GetByKey(&newKey.ConsumerKey); err == nil {
		return types.NullDeveloperAppKey, types.NewBadRequestError(
			fmt.Errorf("consumerKey '%s' already exists", newKey.ConsumerKey))
	}

	// Generate consumerkey if not provided
	if newKey.ConsumerKey == "" {
		newKey.ConsumerKey = generateConsumerKey()
	}
	// Generate consumersecret if not provided
	if newKey.ConsumerSecret == "" {
		newKey.ConsumerSecret = generateConsumerSecret()
	}
	// Generate issuedate if not provided
	if newKey.IssuedAt == 0 {
		newKey.IssuedAt = shared.GetCurrentTimeMilliseconds()
	}
	// Set expiry if not provided
	if newKey.ExpiresAt == 0 {
		newKey.ExpiresAt = -1
	}
	newKey.Approved()

	// Populate fields we do not allow to be updated
	newKey.AppID = developerApp.AppID

	if err := ks.db.Key.UpdateByKey(&newKey); err != nil {
		return types.NullDeveloperAppKey, err
	}
	ks.changelog.Create(newKey, who)
	return newKey, nil
}

// Update updates an existing key
func (ks *KeyService) Update(consumerKey string, updatedKey *types.Key,
	who Requester) (types.Key, types.Error) {

	currentKey, err := ks.db.Key.GetByKey(&updatedKey.ConsumerKey)
	if err != nil {
		return types.NullDeveloperAppKey, err
	}
	// Copy over fields we do not allow to be updated
	updatedKey.IssuedAt = currentKey.IssuedAt
	updatedKey.ConsumerKey = currentKey.ConsumerKey
	updatedKey.ConsumerSecret = currentKey.ConsumerSecret
	updatedKey.AppID = currentKey.AppID
	// If no status provided we will use existing status
	if updatedKey.Status == "" {
		updatedKey.Status = currentKey.Status
	}

	if err = ks.db.Key.UpdateByKey(updatedKey); err != nil {
		return types.NullDeveloperAppKey, err
	}
	ks.changelog.Update(currentKey, updatedKey, who)
	return *updatedKey, nil
}

// Delete deletes a key
func (ks *KeyService) Delete(consumerKey string,
	who Requester) (deletedKey types.Key, e types.Error) {

	key, err := ks.db.Key.GetByKey(&consumerKey)
	if err != nil {
		return types.NullDeveloperAppKey, err
	}
	if err = ks.db.Key.DeleteByKey(consumerKey); err != nil {
		return types.NullDeveloperAppKey, err
	}
	ks.changelog.Delete(key, who)
	return *key, nil

}

// generateConsumerKey returns a random string to be used as apikey (32 character base62)
func generateConsumerKey() string {

	return uniuri.NewLen(32)
}

// generateConsumerSecret returns a random string to be used as consumer key (16 character base62)
func generateConsumerSecret() string {

	return uniuri.New()
}
