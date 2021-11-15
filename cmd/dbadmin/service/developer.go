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
func (ds *DeveloperService) GetAll(organizationName string) (developers types.Developers, err types.Error) {

	return ds.db.Developer.GetAll()
}

// Get returns details of an developer, in case name contains a @ assumption is developerId was provided
func (ds *DeveloperService) Get(organizationName, developerName string) (developer *types.Developer, err types.Error) {

	if strings.Contains(developerName, "@") {
		return ds.db.Developer.GetByEmail(developerName)
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
func (ds *DeveloperService) Create(organizationName string, newDeveloper types.Developer,
	who Requester) (types.Developer, types.Error) {

	if _, err := ds.Get(organizationName, newDeveloper.Email); err == nil {
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
func (ds *DeveloperService) Update(organizationName, developerEmail string, updatedDeveloper types.Developer, who Requester) (
	types.Developer, types.Error) {

	currentDeveloper, err := ds.db.Developer.GetByEmail(developerEmail)
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
func (ds *DeveloperService) UpdateAttributes(organizationName, developerName string,
	receivedAttributes types.Attributes, who Requester) types.Error {

	currentDeveloper, err := ds.Get(organizationName, developerName)
	if err != nil {
		return err
	}
	updatedDeveloper := copyDeveloper(*currentDeveloper)
	updatedDeveloper.Attributes = receivedAttributes
	if err = ds.updateDeveloper(updatedDeveloper, who); err != nil {
		return err
	}
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return nil
}

// UpdateAttribute update an attribute of developer
func (ds *DeveloperService) UpdateAttribute(organizationName, developerName string,
	attributeValue types.Attribute, who Requester) types.Error {

	currentDeveloper, err := ds.Get(organizationName, developerName)
	if err != nil {
		return err
	}
	updatedDeveloper := copyDeveloper(*currentDeveloper)
	if err := updatedDeveloper.Attributes.Set(attributeValue); err != nil {
		return err
	}

	if err = ds.updateDeveloper(updatedDeveloper, who); err != nil {
		return err
	}
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return err
}

// DeleteAttribute removes an attribute of an developer
func (ds *DeveloperService) DeleteAttribute(
	organizationName, developerName, attributeToDelete string, who Requester) (string, types.Error) {

	currentDeveloper, err := ds.Get(organizationName, developerName)
	if err != nil {
		return "", err
	}
	updatedDeveloper := copyDeveloper(*currentDeveloper)
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
	updatedDeveloper.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedDeveloper.LastModifiedBy = who.User
	return ds.db.Developer.Update(updatedDeveloper)
}

// Delete deletes an developer
func (ds *DeveloperService) Delete(organizationName, developerName string,
	who Requester) (deletedDeveloper *types.Developer, e types.Error) {

	developer, err := ds.Get(organizationName, developerName)
	if err != nil {
		return nil, err
	}
	appCountOfDeveloper, err := ds.db.DeveloperApp.GetCountByDeveloperID(developer.DeveloperID)
	if err != nil {
		return nil, err
	}
	if appCountOfDeveloper > 0 {
		return &types.NullDeveloper, types.NewBadRequestError(
			fmt.Errorf("cannot delete developer '%s' with %d active applications",
				developer.Email, appCountOfDeveloper))
	}
	if err := ds.db.Developer.DeleteByID(developer.DeveloperID); err != nil {
		return nil, err
	}
	ds.changelog.Delete(developer, who)
	return developer, nil
}

// generateDeveloperID generates a DeveloperID
func generateDeveloperID(developer string) string {
	return uniuri.New()
}

func copyDeveloper(d types.Developer) *types.Developer {

	return &types.Developer{
		Apps:           d.Apps,
		Attributes:     d.Attributes,
		CreatedAt:      d.CreatedAt,
		CreatedBy:      d.CreatedBy,
		DeveloperID:    d.DeveloperID,
		Email:          d.Email,
		FirstName:      d.FirstName,
		LastModifiedBy: d.LastModifiedBy,
		LastModifiedAt: d.LastModifiedAt,
		LastName:       d.LastName,
		Status:         d.Status,
		UserName:       d.UserName,
	}
}
