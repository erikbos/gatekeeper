package service

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// DeveloperAppService is
type DeveloperAppService struct {
	db *db.Database
}

// NewDeveloperAppService returns a new developerApp instance
func NewDeveloperAppService(database *db.Database) *DeveloperAppService {

	return &DeveloperAppService{db: database}
}

// GetByOrganization returns all developerApp apps
func (das *DeveloperAppService) GetByOrganization(organizationName string) (developerApps types.DeveloperApps, err types.Error) {

	return das.db.DeveloperApp.GetByOrganization(organizationName)
}

// GetByName returns details of an developerApp
func (das *DeveloperAppService) GetByName(organizationName, developerAppName string) (developerApp *types.DeveloperApp, err types.Error) {

	return das.db.DeveloperApp.GetByName(organizationName, developerAppName)
}

// GetByID returns details of an developerApp
func (das *DeveloperAppService) GetByID(organizationName, developerAppName string) (developerApp *types.DeveloperApp, err types.Error) {

	return das.db.DeveloperApp.GetByID(organizationName, developerAppName)
}

// GetAttributes returns attributes of an developerApp
func (das *DeveloperAppService) GetAttributes(organizationName, developerAppName string) (attributes *types.Attributes, err types.Error) {

	developerApp, err := das.GetByName(organizationName, developerAppName)
	if err != nil {
		return nil, err
	}
	return &developerApp.Attributes, nil
}

// GetAttribute returns one particular attribute of an developerApp
func (das *DeveloperAppService) GetAttribute(organizationName, developerAppName, attributeName string) (value string, err types.Error) {

	developerApp, err := das.GetByName(organizationName, developerAppName)
	if err != nil {
		return "", err
	}
	return developerApp.Attributes.Get(attributeName)
}

// Create creates a new developerApp
func (das *DeveloperAppService) Create(organizationName, developerName string, newDeveloperApp types.DeveloperApp) (types.DeveloperApp, types.Error) {

	developer, err := das.db.Developer.GetByEmail(organizationName, developerName)
	if err != nil {
		return types.NullDeveloperApp, err
	}

	existingDeveloperApp, err := das.GetByName(organizationName, newDeveloperApp.Name)
	if err == nil {
		return types.NullDeveloperApp, types.NewBadRequestError(
			fmt.Errorf("DeveloperApp '%s' already exists", existingDeveloperApp.Name))
	}

	newDeveloperApp.AppID = generateAppID()
	newDeveloperApp.DeveloperID = developer.DeveloperID
	newDeveloperApp.OrganizationName = organizationName
	// New developer starts actived
	newDeveloperApp.Status = "active"
	newDeveloperApp.CreatedAt = shared.GetCurrentTimeMilliseconds()
	// newDeveloperApp.CreatedBy = h.GetSessionUser(c)
	newDeveloperApp.LastmodifiedAt = newDeveloperApp.CreatedAt
	// newDeveloperApp.LastmodifiedBy = h.GetSessionUser(c)

	// Automatically set default fields
	newDeveloperApp.CreatedAt = shared.GetCurrentTimeMilliseconds()

	err = das.updateDeveloperApp(&newDeveloperApp)
	return newDeveloperApp, err
}

// Update updates an existing developerApp
func (das *DeveloperAppService) Update(organizationName string, updatedDeveloperApp types.DeveloperApp) (types.DeveloperApp, types.Error) {

	developerAppToUpdate, err := das.db.DeveloperApp.GetByID(organizationName, updatedDeveloperApp.AppID)
	if err != nil {
		return types.NullDeveloperApp, types.NewItemNotFoundError(err)
	}

	// Copy over the fields we allow to be updated
	developerAppToUpdate.Name = updatedDeveloperApp.Name
	developerAppToUpdate.DisplayName = updatedDeveloperApp.DisplayName
	developerAppToUpdate.Attributes = updatedDeveloperApp.Attributes
	developerAppToUpdate.Status = updatedDeveloperApp.Status

	err = das.updateDeveloperApp(developerAppToUpdate)
	return *developerAppToUpdate, err
}

// UpdateAttributes updates attributes of an developerApp
func (das *DeveloperAppService) UpdateAttributes(organizationName string, developerAppName string, receivedAttributes types.Attributes) types.Error {

	updatedDeveloperApp, err := das.GetByName(organizationName, developerAppName)
	if err != nil {
		return err
	}
	updatedDeveloperApp.Attributes = receivedAttributes
	return das.updateDeveloperApp(updatedDeveloperApp)
}

// UpdateAttribute update an attribute of developerApp
func (das *DeveloperAppService) UpdateAttribute(organizationName string, developerAppName string,
	attributeValue types.Attribute) types.Error {

	updatedDeveloperApp, err := das.GetByName(organizationName, developerAppName)
	if err != nil {
		return err
	}
	updatedDeveloperApp.Attributes.Set(attributeValue)
	return das.updateDeveloperApp(updatedDeveloperApp)
}

// DeleteAttribute removes an attribute of an developerApp
func (das *DeveloperAppService) DeleteAttribute(organizationName, developerAppName, attributeToDelete string) (string, types.Error) {

	updatedDeveloperApp, err := das.GetByName(organizationName, developerAppName)
	if err != nil {
		return "", err
	}
	oldValue, err := updatedDeveloperApp.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}
	return oldValue, das.updateDeveloperApp(updatedDeveloperApp)
}

// updateDeveloperApp updates last-modified field(s) and updates developer app in database
func (das *DeveloperAppService) updateDeveloperApp(updatedDeveloperApp *types.DeveloperApp) types.Error {

	updatedDeveloperApp.Attributes.Tidy()
	updatedDeveloperApp.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	return das.db.DeveloperApp.Update(updatedDeveloperApp)
}

// Delete deletes an developerApp
func (das *DeveloperAppService) Delete(organizationName,
	developerID, developerAppName string) (deletedDeveloperApp types.DeveloperApp, e types.Error) {

	developer, err := das.db.Developer.GetByID(developerID)
	if err != nil {
		return types.NullDeveloperApp, err
	}

	developerApp, err := das.GetByName(organizationName, developerAppName)
	if err != nil {
		return types.NullDeveloperApp, err
	}
	developerAppKeys, err := das.db.Credential.GetByDeveloperAppID(developerApp.AppID)
	if err != nil {
		return types.NullDeveloperApp, err
	}

	developerAppKeyCount := len(developerAppKeys)
	if developerAppKeyCount > 0 {
		return types.NullDeveloperApp, types.NewBadRequestError(
			fmt.Errorf("Cannot delete developer app '%s' with %d apikeys",
				developerApp.Name, developerAppKeyCount))
	}

	err = das.db.DeveloperApp.DeleteByID(developerApp.OrganizationName, developerApp.AppID)
	if err != nil {
		return types.NullDeveloperApp, err
	}
	// Remove app from the apps field in developer entity as well
	for i := 0; i < len(developer.Apps); i++ {
		if developer.Apps[i] == developerAppName {
			developer.Apps = append(developer.Apps[:i], developer.Apps[i+1:]...)
			i-- // from the remove item index to start iterate next item
		}
	}
	// FIXME	developer.LastmodifiedBy = h.GetSessionUser(c)
	if err := das.db.Developer.Update(developer); err != nil {
		return types.NullDeveloperApp, err
	}
	return *developerApp, nil
}

// generateAppID creates unique primary key for developer app row
func generateAppID() string {
	return (uuid.New().String())
}
