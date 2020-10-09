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
	db        *db.Database
	changelog *Changelog
}

// NewDeveloper returns a new developer instance
func NewDeveloper(database *db.Database, c *Changelog) *DeveloperService {

	return &DeveloperService{db: database, changelog: c}
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
func (ds *DeveloperService) Create(organizationName string,
	newDeveloper types.Developer, who Requester) (types.Developer, types.Error) {

	existingDeveloper, err := ds.Get(organizationName, newDeveloper.Email)
	if err == nil {
		return types.NullDeveloper, types.NewBadRequestError(
			fmt.Errorf("Developer '%s' already exists", existingDeveloper.Email))
	}
	// Automatically set default fields
	newDeveloper.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newDeveloper.CreatedBy = who.User

	err = ds.updateDeveloper(&newDeveloper, who)
	ds.changelog.Create(newDeveloper, who)
	return newDeveloper, err
}

// Update updates an existing developer
func (ds *DeveloperService) Update(organizationName string,
	updatedDeveloper types.Developer, who Requester) (
	types.Developer, types.Error) {

	currentDeveloper, err := ds.db.Developer.GetByID(updatedDeveloper.DeveloperID)
	if err != nil {
		return types.NullDeveloper, types.NewItemNotFoundError(err)
	}
	// Copy over fields we do not allow to be updated
	updatedDeveloper.DeveloperID = currentDeveloper.DeveloperID
	updatedDeveloper.CreatedAt = currentDeveloper.CreatedAt
	updatedDeveloper.CreatedBy = currentDeveloper.CreatedBy

	err = ds.updateDeveloper(&updatedDeveloper, who)
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return updatedDeveloper, err
}

// UpdateAttributes updates attributes of an developer
func (ds *DeveloperService) UpdateAttributes(organizationName string, developerName string,
	receivedAttributes types.Attributes, who Requester) types.Error {

	currentDeveloper, err := ds.Get(organizationName, developerName)
	if err != nil {
		return err
	}
	updatedDeveloper := currentDeveloper
	updatedDeveloper.Attributes.SetMultiple(receivedAttributes)

	err = ds.updateDeveloper(updatedDeveloper, who)
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return err
}

// UpdateAttribute update an attribute of developer
func (ds *DeveloperService) UpdateAttribute(organizationName string,
	developerName string, attributeValue types.Attribute, who Requester) types.Error {

	currentDeveloper, err := ds.Get(organizationName, developerName)
	if err != nil {
		return err
	}
	updatedDeveloper := currentDeveloper
	updatedDeveloper.Attributes.Set(attributeValue)

	err = ds.updateDeveloper(updatedDeveloper, who)
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return err
}

// DeleteAttribute removes an attribute of an developer
func (ds *DeveloperService) DeleteAttribute(organizationName, developerName,
	attributeToDelete string, who Requester) (string, types.Error) {

	currentDeveloper, err := ds.Get(organizationName, developerName)
	if err != nil {
		return "", err
	}
	updatedDeveloper := currentDeveloper
	oldValue, err := updatedDeveloper.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}

	err = ds.updateDeveloper(updatedDeveloper, who)
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return oldValue, err
}

// updateDeveloper updates last-modified field(s) and updates developer in database
func (ds *DeveloperService) updateDeveloper(updatedDeveloper *types.Developer, who Requester) types.Error {

	updatedDeveloper.Attributes.Tidy()
	updatedDeveloper.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedDeveloper.LastmodifiedBy = who.User
	return ds.db.Developer.Update(updatedDeveloper)
}

// Delete deletes an developer
func (ds *DeveloperService) Delete(organizationName, developerName string,
	who Requester) (deletedDeveloper types.Developer, e types.Error) {

	developer, err := ds.Get(organizationName, developerName)
	if err != nil {
		return types.NullDeveloper, err
	}
	appCountOfDeveloper, err := ds.db.DeveloperApp.GetCountByDeveloperID(developer.DeveloperID)
	if err != nil {
		return types.NullDeveloper, err
	}
	if appCountOfDeveloper > 0 {
		return types.NullDeveloper, types.NewBadRequestError(
			fmt.Errorf("Cannot delete developer '%s' with %d active developer apps",
				developer.Email, appCountOfDeveloper))
	}
	if err := ds.db.Developer.DeleteByID(developer.OrganizationName, developer.DeveloperID); err != nil {
		return types.NullDeveloper, err
	}
	ds.changelog.Delete(developer, who)
	return *developer, nil
}
