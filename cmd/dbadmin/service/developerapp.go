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
	db        *db.Database
	changelog *Changelog
}

// NewDeveloperApp returns a new developerApp instance
func NewDeveloperApp(database *db.Database, c *Changelog) *DeveloperAppService {

	return &DeveloperAppService{
		db:        database,
		changelog: c,
	}
}

// GetAll returns all developerApp apps
func (das *DeveloperAppService) GetAll() (developerApps types.DeveloperApps, err types.Error) {

	return das.db.DeveloperApp.GetAll()
}

// GetAll returns all developer apps of one develoepr
func (das *DeveloperAppService) GetAllByEmail(developerEmail string) (developerApps types.DeveloperApps, err types.Error) {

	developer, err := das.db.Developer.GetByEmail(developerEmail)
	if err != nil {
		return types.NullDeveloperApps, err
	}

	return das.db.DeveloperApp.GetAllByDeveloperID(developer.DeveloperID)
}

// GetByName returns details of an developerApp
func (das *DeveloperAppService) GetByName(developerAppName string) (developerApp *types.DeveloperApp, err types.Error) {

	return das.db.DeveloperApp.GetByName(developerAppName)
}

// GetByID returns details of an developerApp
func (das *DeveloperAppService) GetByID(developerAppID string) (developerApp *types.DeveloperApp, err types.Error) {

	return das.db.DeveloperApp.GetByID(developerAppID)
}

// GetAttributes returns attributes of an developerApp
func (das *DeveloperAppService) GetAttributes(developerAppName string) (attributes *types.Attributes, err types.Error) {

	developerApp, err := das.GetByName(developerAppName)
	if err != nil {
		return nil, err
	}
	return &developerApp.Attributes, nil
}

// GetAttribute returns one particular attribute of an developerApp
func (das *DeveloperAppService) GetAttribute(developerAppName, attributeName string) (value string, err types.Error) {

	developerApp, err := das.GetByName(developerAppName)
	if err != nil {
		return "", err
	}
	return developerApp.Attributes.Get(attributeName)
}

// Create creates a new developerApp
func (das *DeveloperAppService) Create(developerEmail string,
	newDeveloperApp types.DeveloperApp, who Requester) (types.DeveloperApp, types.Error) {

	developer, err := das.db.Developer.GetByEmail(developerEmail)
	if err != nil {
		return types.NullDeveloperApp, err
	}

	existingDeveloperApp, err := das.GetByName(newDeveloperApp.Name)
	if err == nil {
		return types.NullDeveloperApp, types.NewBadRequestError(
			fmt.Errorf("developerApp '%s' already exists", existingDeveloperApp.Name))
	}

	// Automatically set default fields
	newDeveloperApp.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newDeveloperApp.CreatedBy = who.User

	newDeveloperApp.AppID = generateAppID()
	newDeveloperApp.DeveloperID = developer.DeveloperID
	// New developer starts approved
	newDeveloperApp.Approve()

	if err = das.updateDeveloperApp(&newDeveloperApp, who); err != nil {
		return types.NullDeveloperApp, err
	}
	das.changelog.Create(newDeveloperApp, who)

	// Add app to the apps field in developer entity
	developer.Apps = append(developer.Apps, newDeveloperApp.Name)
	if err := das.db.Developer.Update(developer); err != nil {
		return newDeveloperApp, err
	}

	return newDeveloperApp, nil
}

// Update updates an existing developerApp
func (das *DeveloperAppService) Update(updatedDeveloperApp types.DeveloperApp,
	who Requester) (types.DeveloperApp, types.Error) {

	currentDeveloperApp, err := das.db.DeveloperApp.GetByName(updatedDeveloperApp.Name)
	if err != nil {
		return types.NullDeveloperApp, err
	}

	// Copy over the fields we do not allow to be updated
	updatedDeveloperApp.Name = currentDeveloperApp.Name
	updatedDeveloperApp.DeveloperID = currentDeveloperApp.DeveloperID
	updatedDeveloperApp.AppID = currentDeveloperApp.AppID
	updatedDeveloperApp.CreatedAt = currentDeveloperApp.CreatedAt
	updatedDeveloperApp.CreatedBy = currentDeveloperApp.CreatedBy

	if err = das.updateDeveloperApp(&updatedDeveloperApp, who); err != nil {
		return types.NullDeveloperApp, err
	}
	das.changelog.Update(currentDeveloperApp, updatedDeveloperApp, who)
	return updatedDeveloperApp, nil
}

// UpdateAttributes updates attributes of an developerApp
func (das *DeveloperAppService) UpdateAttributes(developerAppName string,
	receivedAttributes types.Attributes, who Requester) types.Error {

	currentDeveloperApp, err := das.GetByName(developerAppName)
	if err != nil {
		return err
	}
	updatedDeveloperApp := copyDeveloperApp(*currentDeveloperApp)
	updatedDeveloperApp.Attributes = receivedAttributes

	if err = das.updateDeveloperApp(updatedDeveloperApp, who); err != nil {
		return err
	}
	das.changelog.Update(currentDeveloperApp, updatedDeveloperApp, who)
	return err
}

// UpdateAttribute update an attribute of developerApp
func (das *DeveloperAppService) UpdateAttribute(developerAppName string,
	attributeValue types.Attribute, who Requester) types.Error {

	currentDeveloperApp, err := das.GetByName(developerAppName)
	if err != nil {
		return err
	}
	updatedDeveloperApp := copyDeveloperApp(*currentDeveloperApp)
	if err := updatedDeveloperApp.Attributes.Set(attributeValue); err != nil {
		return err
	}

	if err = das.updateDeveloperApp(updatedDeveloperApp, who); err != nil {
		return err
	}
	das.changelog.Update(currentDeveloperApp, updatedDeveloperApp, who)
	return err
}

// DeleteAttribute removes an attribute of an developerApp
func (das *DeveloperAppService) DeleteAttribute(developerAppName,
	attributeToDelete string, who Requester) (string, types.Error) {

	currentDeveloperApp, err := das.GetByName(developerAppName)
	if err != nil {
		return "", err
	}
	updatedDeveloperApp := currentDeveloperApp
	oldValue, err := updatedDeveloperApp.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}
	if err = das.updateDeveloperApp(updatedDeveloperApp, who); err != nil {
		return "", err
	}
	das.changelog.Update(currentDeveloperApp, updatedDeveloperApp, who)
	return oldValue, err
}

// updateDeveloperApp updates last-modified field(s) and updates developer app in database
func (das *DeveloperAppService) updateDeveloperApp(updatedDeveloperApp *types.DeveloperApp, who Requester) types.Error {

	updatedDeveloperApp.Attributes.Tidy()
	updatedDeveloperApp.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedDeveloperApp.LastModifiedBy = who.User
	return das.db.DeveloperApp.Update(updatedDeveloperApp)
}

// Delete deletes an developerApp
func (das *DeveloperAppService) Delete(developerID, developerAppName string,
	who Requester) (deletedDeveloperApp types.DeveloperApp, e types.Error) {

	developer, err := das.db.Developer.GetByID(developerID)
	if err != nil {
		return types.NullDeveloperApp, err
	}

	developerApp, err := das.GetByName(developerAppName)
	if err != nil {
		return types.NullDeveloperApp, err
	}
	developerAppKeys, _ := das.db.Key.GetByDeveloperAppID(developerApp.AppID)
	if len(developerAppKeys) != 0 {
		for _, k := range developerAppKeys {
			das.db.Key.DeleteByKey(k.ConsumerKey)
		}
	}

	// developerAppKeyCount := len(developerAppKeys)
	// for
	// if developerAppKeyCount > 0 {
	// 	return types.NullDeveloperApp, types.NewBadRequestError(
	// 		fmt.Errorf("cannot delete developer app '%s' with %d apikeys",
	// 			developerApp.Name, developerAppKeyCount))
	// }

	err = das.db.DeveloperApp.DeleteByID(developerApp.AppID)
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
	das.changelog.Delete(developerApp, who)
	return *developerApp, nil
}

// generateAppID creates unique primary key for developer app row
func generateAppID() string {
	return (uuid.New().String())
}

func copyDeveloperApp(d types.DeveloperApp) *types.DeveloperApp {

	return &types.DeveloperApp{
		AppID:          d.AppID,
		Attributes:     d.Attributes,
		CreatedAt:      d.CreatedAt,
		CreatedBy:      d.CreatedBy,
		DeveloperID:    d.DeveloperID,
		DisplayName:    d.DisplayName,
		LastModifiedBy: d.LastModifiedBy,
		LastModifiedAt: d.LastModifiedAt,
		Name:           d.Name,
		Status:         d.Status,
	}
}
