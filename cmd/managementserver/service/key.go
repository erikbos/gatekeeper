package service

import (
	"fmt"

	"github.com/dchest/uniuri"

	"github.com/erikbos/gatekeeper/cmd/managementserver/audit"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// KeyService is
type KeyService struct {
	db    *db.Database
	audit *audit.Audit
}

// NewKey returns a new key instance
func NewKey(database *db.Database, a *audit.Audit) *KeyService {

	return &KeyService{
		db:    database,
		audit: a,
	}
}

// Get returns details of an key
func (ks *KeyService) Get(organizationName, developerEmail, appName, key string) (*types.Key, types.Error) {

	return ks.db.Key.GetByKey(&organizationName, &key)
}

// GetByDeveloperAppID returns all keys of a developer app
func (ks *KeyService) GetByDeveloperAppID(organizationName, developerAppID string) (types.Keys, types.Error) {

	return ks.db.Key.GetByDeveloperAppID(organizationName, developerAppID)
}

// Create creates a key
func (ks *KeyService) Create(organizationName, developerEmail, developerAppName string, newKey types.Key, who audit.Requester) (*types.Key, types.Error) {

	developer, developerApp, err := ks.getDevApp(organizationName, developerEmail, developerAppName)
	if err != nil {
		return nil, err
	}
	if _, err := ks.db.Key.GetByKey(&organizationName, &newKey.ConsumerKey); err == nil {
		return nil, types.NewBadRequestError(
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

	if err := ks.db.Key.UpdateByKey(organizationName, &newKey); err != nil {
		return nil, err
	}
	env := &audit.Environment{
		Organization: organizationName,
		DeveloperID:  developer.DeveloperID,
		AppID:        developerApp.AppID,
	}
	ks.audit.Create(newKey, env, who)
	return &newKey, nil
}

// Update updates an existing key
func (ks *KeyService) Update(organizationName, developerEmail, developerAppName string,
	consumerKey string, updatedKey types.Key, who audit.Requester) (*types.Key, types.Error) {

	developer, developerApp, err := ks.getDevApp(organizationName, developerEmail, developerAppName)
	if err != nil {
		return nil, err
	}
	currentKey, err := ks.db.Key.GetByKey(&organizationName, &updatedKey.ConsumerKey)
	if err != nil {
		return nil, err
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

	if err := updatedKey.Validate(); err != nil {
		return nil, types.NewBadRequestError(err)
	}
	if err = ks.db.Key.UpdateByKey(organizationName, &updatedKey); err != nil {
		return nil, err
	}
	env := &audit.Environment{
		Organization: organizationName,
		DeveloperID:  developer.DeveloperID,
		AppID:        developerApp.AppID,
	}
	ks.audit.Update(currentKey, updatedKey, env, who)
	return &updatedKey, nil
}

// Delete deletes a key
func (ks *KeyService) Delete(organizationName, developerEmail, developerAppName,
	consumerKey string, who audit.Requester) (e types.Error) {

	developer, developerApp, err := ks.getDevApp(organizationName, developerEmail, developerAppName)
	if err != nil {
		return err
	}
	key, err := ks.db.Key.GetByKey(&organizationName, &consumerKey)
	if err != nil {
		return err
	}
	if err = ks.db.Key.DeleteByKey(organizationName, consumerKey); err != nil {
		return err
	}
	env := &audit.Environment{
		Organization: organizationName,
		DeveloperID:  developer.DeveloperID,
		AppID:        developerApp.AppID,
	}
	ks.audit.Delete(key, env, who)
	return nil
}

// getDevApp gets developer and application
func (ks *KeyService) getDevApp(organizationName, developerEmailaddress, appName string) (
	developer *types.Developer, application *types.DeveloperApp, err types.Error) {

	developer, err = ks.db.Developer.GetByEmail(organizationName, developerEmailaddress)
	if err != nil {
		return nil, nil, err
	}
	application, err = ks.db.DeveloperApp.GetByName(organizationName, developerEmailaddress, appName)
	if err != nil {
		return nil, nil, err
	}
	return
}

// generateConsumerKey returns a random string to be used as apikey (32 character base62)
func generateConsumerKey() string {

	return uniuri.NewLen(32)
}

// generateConsumerSecret returns a random string to be used as consumer key (16 character base62)
func generateConsumerSecret() string {

	return uniuri.New()
}
