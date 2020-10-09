package service

import (
	"fmt"

	"github.com/dchest/uniuri"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// CredentialService is
type CredentialService struct {
	db        *db.Database
	changelog *Changelog
}

// NewCredential returns a new credential instance
func NewCredential(database *db.Database, c *Changelog) *CredentialService {

	return &CredentialService{db: database, changelog: c}
}

// Get returns details of an credential
func (cs *CredentialService) Get(organizationName, key string) (credential *types.DeveloperAppKey, err types.Error) {

	return cs.db.Credential.GetByKey(&organizationName, &key)
}

// GetByDeveloperAppID returns all credentials of a developer app
func (cs *CredentialService) GetByDeveloperAppID(developerAppID string) (clusters types.DeveloperAppKeys, err types.Error) {

	return cs.db.Credential.GetByDeveloperAppID(developerAppID)
}

// Create creates a credential
func (cs *CredentialService) Create(newCredential types.DeveloperAppKey,
	who Requester) (types.DeveloperAppKey, types.Error) {

	credentialToUpdate, err := cs.db.Credential.GetByKey(&newCredential.OrganizationName, &newCredential.ConsumerKey)
	if err != nil {
		return types.NullDeveloperAppKey, types.NewBadRequestError(
			fmt.Errorf("consumerKey '%s' already exists", newCredential.ConsumerKey))
	}

	// Generate adopt user supplied consumerkey if present
	if newCredential.ConsumerKey == "" {
		newCredential.ConsumerKey = generateCredentialConsumerKey()
	}

	// adopt user supplied consumersecret if present
	if newCredential.ConsumerSecret == "" {
		newCredential.ConsumerSecret = generateCredentialConsumerSecret()
	}

	err = cs.db.Credential.UpdateByKey(&newCredential)
	cs.changelog.Create(newCredential, who)
	return *credentialToUpdate, err
}

// Update updates an existing credential
func (cs *CredentialService) Update(updatedCredential types.DeveloperAppKey,
	who Requester) (types.DeveloperAppKey, types.Error) {

	currentCredential, err := cs.db.Credential.GetByKey(&updatedCredential.OrganizationName, &updatedCredential.ConsumerKey)
	if err != nil {
		return types.NullDeveloperAppKey, types.NewItemNotFoundError(err)
	}
	// Copy over fields we do not allow to be updated
	updatedCredential.IssuedAt = currentCredential.IssuedAt
	updatedCredential.ConsumerKey = currentCredential.ConsumerKey
	updatedCredential.ConsumerSecret = currentCredential.ConsumerSecret
	updatedCredential.AppID = currentCredential.AppID
	updatedCredential.OrganizationName = currentCredential.OrganizationName

	err = cs.db.Credential.UpdateByKey(&updatedCredential)
	cs.changelog.Update(currentCredential, updatedCredential, who)
	return updatedCredential, err
}

// Delete deletes an credential
func (cs *CredentialService) Delete(organizationName, consumerKey string,
	who Requester) (deletedCredential types.DeveloperAppKey, e types.Error) {

	credential, err := cs.db.Credential.GetByKey(&organizationName, &consumerKey)
	if err != nil {
		return types.NullDeveloperAppKey, err
	}
	err = cs.db.Credential.DeleteByKey(organizationName, consumerKey)
	if err != nil {
		return types.NullDeveloperAppKey, err
	}
	cs.changelog.Delete(credential, who)
	return *credential, nil

}

// generateCredentialConsumerKey returns a random string to be used as apikey (32 character base62)
func generateCredentialConsumerKey() string {

	return uniuri.NewLen(32)
}

// generateCredentialConsumerSecret returns a random string to be used as consumer key (16 character base62)
func generateCredentialConsumerSecret() string {

	return uniuri.New()
}
