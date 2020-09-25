package service

import (
	"fmt"
	"strings"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// DeveloperService is
type DeveloperService struct {
	db *db.Database
}

// NewDeveloperService returns a new developer instance
func NewDeveloperService(database *db.Database) *DeveloperService {

	return &DeveloperService{db: database}
}

// GetByOrganization returns all developers
func (ds *DeveloperService) GetByOrganization(organizationName string) (developers types.Developers, err types.Error) {

	return ds.db.Developer.GetByOrganization(organizationName)
}

// Get returns details of an developer
func (ds *DeveloperService) Get(organizationName, developerName string) (developer *types.Developer, err types.Error) {

	if strings.Contains(developerName, "@") {
		return ds.db.Developer.GetByEmail(organizationName, developerName)
	}
	return ds.db.Developer.GetByID(developerName)
}

// GetAttributes returns attributes of an developer
func (ds *DeveloperService) GetAttributes(organizationName, developerName string) (attributes *types.Attributes, err types.Error) {

	developer, err := ds.Get(organizationName, developerName)
	if err != nil {
		return nil, err
	}
	return &developer.Attributes, nil
}

// GetAttribute returns one particular attribute of an developer
func (ds *DeveloperService) GetAttribute(organizationName, developerName, attributeName string) (value string, err types.Error) {

	developer, err := ds.Get(organizationName, developerName)
	if err != nil {
		return "", err
	}
	return developer.Attributes.Get(attributeName)
}

// Create creates a new developer
func (ds *DeveloperService) Create(organizationName string, newDeveloper types.Developer) (types.Developer, types.Error) {

	existingDeveloper, err := ds.Get(organizationName, newDeveloper.Email)
	if err == nil {
		return types.NullDeveloper, types.NewBadRequestError(
			fmt.Errorf("Developer '%s' already exists", existingDeveloper.Email))
	}

	// Automatically set default fields
	newDeveloper.CreatedAt = shared.GetCurrentTimeMilliseconds()
	err = ds.updateDeveloper(&newDeveloper)
	return newDeveloper, err
}

// Update updates an existing developer
func (ds *DeveloperService) Update(organizationName string, updatedDeveloper types.Developer) (types.Developer, types.Error) {

	developerToUpdate, err := ds.db.Developer.GetByID(updatedDeveloper.DeveloperID)
	if err != nil {
		return types.NullDeveloper, types.NewItemNotFoundError(err)
	}

	// Copy over the fields we allow to be updated
	developerToUpdate.Email = updatedDeveloper.Email
	developerToUpdate.FirstName = updatedDeveloper.FirstName
	developerToUpdate.LastName = updatedDeveloper.LastName
	developerToUpdate.Attributes = updatedDeveloper.Attributes
	developerToUpdate.Apps = updatedDeveloper.Apps
	developerToUpdate.Status = updatedDeveloper.Status
	developerToUpdate.SuspendedTill = updatedDeveloper.SuspendedTill

	err = ds.updateDeveloper(developerToUpdate)
	return *developerToUpdate, err
}

// UpdateAttributes updates attributes of an developer
func (ds *DeveloperService) UpdateAttributes(organizationName string, developerName string, receivedAttributes types.Attributes) types.Error {

	updatedDeveloper, err := ds.Get(organizationName, developerName)
	if err != nil {
		return err
	}
	updatedDeveloper.Attributes = receivedAttributes
	return ds.updateDeveloper(updatedDeveloper)
}

// UpdateAttribute update an attribute of developer
func (ds *DeveloperService) UpdateAttribute(organizationName string,
	developerName string, attributeValue types.Attribute) types.Error {

	updatedDeveloper, err := ds.Get(organizationName, developerName)
	if err != nil {
		return err
	}
	updatedDeveloper.Attributes.Set(attributeValue)
	return ds.updateDeveloper(updatedDeveloper)
}

// DeleteAttribute removes an attribute of an developer
func (ds *DeveloperService) DeleteAttribute(organizationName, developerName,
	attributeToDelete string) (string, types.Error) {

	updatedDeveloper, err := ds.Get(organizationName, developerName)
	if err != nil {
		return "", err
	}
	oldValue, err := updatedDeveloper.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}
	return oldValue, ds.updateDeveloper(updatedDeveloper)
}

// updateDeveloper updates last-modified field(s) and updates developer in database
func (ds *DeveloperService) updateDeveloper(updatedDeveloper *types.Developer) types.Error {

	updatedDeveloper.Attributes.Tidy()
	updatedDeveloper.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	return ds.db.Developer.Update(updatedDeveloper)
}

// Delete deletes an developer
func (ds *DeveloperService) Delete(organizationName, developerName string) (deletedDeveloper types.Developer, e types.Error) {

	developer, err := ds.Get(organizationName, developerName)
	if err != nil {
		return types.NullDeveloper, err
	}
	developerAppCount, err := ds.db.DeveloperApp.GetCountByDeveloperID(developer.DeveloperID)
	if err != nil {
		return types.NullDeveloper, err
	}
	if developerAppCount > 0 {
		return types.NullDeveloper, types.NewBadRequestError(
			fmt.Errorf("Cannot delete developer '%s' with %d active developer apps",
				developer.Email, developerAppCount))
	}
	if err := ds.db.Developer.DeleteByID(developer.OrganizationName, developer.DeveloperID); err != nil {
		return types.NullDeveloper, err
	}
	return *developer, nil
}
