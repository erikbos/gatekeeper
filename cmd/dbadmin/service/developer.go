package service

import (
	"fmt"
	"strings"

	"github.com/dchest/uniuri"

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

	return &DeveloperService{
		db:        database,
		changelog: c,
	}
}

// GetAll returns all developers
func (ds *DeveloperService) GetAll() (developers types.Developers, err types.Error) {

	return ds.db.Developer.GetAll()
}

// Get returns details of an developer
func (ds *DeveloperService) Get(developerName string) (developer *types.Developer, err types.Error) {

	if strings.Contains(developerName, "@") {
		return ds.db.Developer.GetByEmail(developerName)
	}
	return ds.db.Developer.GetByID(developerName)
}

// GetAttributes returns attributes of an developer
func (ds *DeveloperService) GetAttributes(developerName string) (attributes *types.Attributes, err types.Error) {

	developer, err := ds.Get(developerName)
	if err != nil {
		return nil, err
	}
	return &developer.Attributes, nil
}

// GetAttribute returns one particular attribute of an developer
func (ds *DeveloperService) GetAttribute(developerName, attributeName string) (value string, err types.Error) {

	developer, err := ds.Get(developerName)
	if err != nil {
		return "", err
	}
	return developer.Attributes.Get(attributeName)
}

// Create creates a new developer
func (ds *DeveloperService) Create(newDeveloper types.Developer,
	who Requester) (types.Developer, types.Error) {

	if _, err := ds.Get(newDeveloper.Email); err == nil {
		return types.NullDeveloper, types.NewBadRequestError(
			fmt.Errorf("developer '%s' already exists", newDeveloper.Email))
	}
	// Automatically set default fields
	newDeveloper.DeveloperID = generateDeveloperID(newDeveloper.Email)
	newDeveloper.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newDeveloper.CreatedBy = who.User
	newDeveloper.Activate()

	if err := ds.updateDeveloper(&newDeveloper, who); err != nil {
		return types.NullDeveloper, err
	}
	ds.changelog.Create(newDeveloper, who)
	return newDeveloper, nil
}

// Update updates an existing developer
func (ds *DeveloperService) Update(updatedDeveloper types.Developer, who Requester) (
	types.Developer, types.Error) {

	currentDeveloper, err := ds.db.Developer.GetByEmail(updatedDeveloper.Email)
	if err != nil {
		return types.NullDeveloper, err
	}

	// Copy over fields we do not allow to be updated
	updatedDeveloper.Apps = currentDeveloper.Apps
	updatedDeveloper.DeveloperID = currentDeveloper.DeveloperID
	updatedDeveloper.CreatedAt = currentDeveloper.CreatedAt
	updatedDeveloper.CreatedBy = currentDeveloper.CreatedBy

	if err = ds.updateDeveloper(&updatedDeveloper, who); err != nil {
		return types.NullDeveloper, err
	}
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return updatedDeveloper, nil
}

// UpdateAttributes updates attributes of an developer
func (ds *DeveloperService) UpdateAttributes(developerName string,
	receivedAttributes types.Attributes, who Requester) types.Error {

	currentDeveloper, err := ds.Get(developerName)
	if err != nil {
		return err
	}
	updatedDeveloper := currentDeveloper
	if err = updatedDeveloper.Attributes.SetMultiple(receivedAttributes); err != nil {
		return err
	}

	if err = ds.updateDeveloper(updatedDeveloper, who); err != nil {
		return err
	}
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return nil
}

// UpdateAttribute update an attribute of developer
func (ds *DeveloperService) UpdateAttribute(developerName string,
	attributeValue types.Attribute, who Requester) types.Error {

	currentDeveloper, err := ds.Get(developerName)
	if err != nil {
		return err
	}
	updatedDeveloper := currentDeveloper
	updatedDeveloper.Attributes.Set(attributeValue)

	if err = ds.updateDeveloper(updatedDeveloper, who); err != nil {
		return err
	}
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return err
}

// DeleteAttribute removes an attribute of an developer
func (ds *DeveloperService) DeleteAttribute(developerName,
	attributeToDelete string, who Requester) (string, types.Error) {

	currentDeveloper, err := ds.Get(developerName)
	if err != nil {
		return "", err
	}
	updatedDeveloper := currentDeveloper
	oldValue, err := updatedDeveloper.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}

	if err = ds.updateDeveloper(updatedDeveloper, who); err != nil {
		return "", err
	}
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return oldValue, nil
}

// updateDeveloper updates last-modified field(s) and updates developer in database
func (ds *DeveloperService) updateDeveloper(updatedDeveloper *types.Developer, who Requester) types.Error {

	updatedDeveloper.Attributes.Tidy()
	updatedDeveloper.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedDeveloper.LastmodifiedBy = who.User
	return ds.db.Developer.Update(updatedDeveloper)
}

// Delete deletes an developer
func (ds *DeveloperService) Delete(developerName string,
	who Requester) (deletedDeveloper types.Developer, e types.Error) {

	developer, err := ds.Get(developerName)
	if err != nil {
		return types.NullDeveloper, err
	}
	appCountOfDeveloper, err := ds.db.DeveloperApp.GetCountByDeveloperID(developer.DeveloperID)
	if err != nil {
		return types.NullDeveloper, err
	}
	if appCountOfDeveloper > 0 {
		return types.NullDeveloper, types.NewBadRequestError(
			fmt.Errorf("cannot delete developer '%s' with %d active developer apps",
				developer.Email, appCountOfDeveloper))
	}
	if err := ds.db.Developer.DeleteByID(developer.DeveloperID); err != nil {
		return types.NullDeveloper, err
	}
	ds.changelog.Delete(developer, who)
	return *developer, nil
}

// generateDeveloperID generates a DeveloperID
func generateDeveloperID(developer string) string {
	return uniuri.New()
}
