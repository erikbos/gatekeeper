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
	db    *db.Database
	audit *Auditlog
}

// NewDeveloperApp returns a new developerApp instance
func NewDeveloperApp(database *db.Database, a *Auditlog) *DeveloperAppService {

	return &DeveloperAppService{
		db:    database,
		audit: a,
	}
}

// GetAll returns all developerApp apps
func (das *DeveloperAppService) GetAll(organizationName string) (developerApps types.DeveloperApps, err types.Error) {

	return das.db.DeveloperApp.GetAll(organizationName)
}

// GetAll returns all developer apps of one develoepr
func (das *DeveloperAppService) GetAllByEmail(organizationName, developerEmail string) (developerApps types.DeveloperApps, err types.Error) {

	developer, err := das.db.Developer.GetByEmail(organizationName, developerEmail)
	if err != nil {
		return types.NullDeveloperApps, err
	}

	return das.db.DeveloperApp.GetAllByDeveloperID(organizationName, developer.DeveloperID)
}

// GetByName returns details of an developerApp
func (das *DeveloperAppService) GetByName(organizationName, developerEmail, developerAppName string) (developerApp *types.DeveloperApp, err types.Error) {

	if _, err := das.db.Developer.GetByEmail(organizationName, developerEmail); err != nil {
		return nil, err
	}
	return das.db.DeveloperApp.GetByName(organizationName, developerEmail, developerAppName)
}

// GetByID returns details of an developerApp
func (das *DeveloperAppService) GetByID(organizationName, developerAppID string) (developerApp *types.DeveloperApp, err types.Error) {

	return das.db.DeveloperApp.GetByID(organizationName, developerAppID)
}

// Create creates a new developerApp
func (das *DeveloperAppService) Create(organizationName, developerEmail string,
	newDeveloperApp types.DeveloperApp, who Requester) (*types.DeveloperApp, types.Error) {

	developer, err := das.db.Developer.GetByEmail(organizationName, developerEmail)
	if err != nil {
		return nil, err
	}

	existingDeveloperApp, err := das.GetByName(organizationName, developerEmail, newDeveloperApp.Name)
	if err == nil {
		return nil, types.NewBadRequestError(
			fmt.Errorf("developerApp '%s' already exists", existingDeveloperApp.Name))
	}

	// Automatically set default fields
	newDeveloperApp.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newDeveloperApp.CreatedBy = who.User

	newDeveloperApp.AppID = generateAppID()
	newDeveloperApp.DeveloperID = developer.DeveloperID
	// New developer starts approved
	newDeveloperApp.Approve()

	if err = das.updateDeveloperApp(organizationName, &newDeveloperApp, who); err != nil {
		return nil, err
	}
	das.audit.Create(newDeveloperApp, who)

	// Add app to the apps field in developer entity
	developer.Apps = append(developer.Apps, newDeveloperApp.Name)
	if err := das.db.Developer.Update(organizationName, developer); err != nil {
		return &newDeveloperApp, err
	}

	return &newDeveloperApp, nil
}

// Update updates an existing developerApp
func (das *DeveloperAppService) Update(organizationName, developerEmail string, updatedDeveloperApp types.DeveloperApp,
	who Requester) (*types.DeveloperApp, types.Error) {

	currentDeveloperApp, err := das.db.DeveloperApp.GetByName(organizationName, developerEmail, updatedDeveloperApp.Name)
	if err != nil {
		return nil, err
	}

	// Copy over the fields we do not allow to be updated
	updatedDeveloperApp.Name = currentDeveloperApp.Name
	updatedDeveloperApp.DeveloperID = currentDeveloperApp.DeveloperID
	updatedDeveloperApp.AppID = currentDeveloperApp.AppID
	updatedDeveloperApp.CreatedAt = currentDeveloperApp.CreatedAt
	updatedDeveloperApp.CreatedBy = currentDeveloperApp.CreatedBy

	if err = das.updateDeveloperApp(organizationName, &updatedDeveloperApp, who); err != nil {
		return nil, err
	}
	das.audit.Update(currentDeveloperApp, updatedDeveloperApp, who)
	return &updatedDeveloperApp, nil
}

// updateDeveloperApp updates last-modified field(s) and updates developer app in database
func (das *DeveloperAppService) updateDeveloperApp(organizationName string,
	updatedDeveloperApp *types.DeveloperApp, who Requester) types.Error {

	updatedDeveloperApp.Attributes.Tidy()
	updatedDeveloperApp.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedDeveloperApp.LastModifiedBy = who.User

	if err := updatedDeveloperApp.Validate(); err != nil {
		return types.NewBadRequestError(err)
	}
	return das.db.DeveloperApp.Update(organizationName, updatedDeveloperApp)
}

// Delete deletes an developerApp
func (das *DeveloperAppService) Delete(organizationName, developerEmail, developerAppName string,
	who Requester) (e types.Error) {

	developer, err := das.db.Developer.GetByEmail(organizationName, developerEmail)
	if err != nil {
		return err
	}
	developerApp, err := das.GetByName(organizationName, developerEmail, developerAppName)
	if err != nil {
		return err
	}
	developerAppKeys, _ := das.db.Key.GetByDeveloperAppID(organizationName, developerApp.AppID)
	if len(developerAppKeys) != 0 {
		for _, k := range developerAppKeys {
			if err := das.db.Key.DeleteByKey(organizationName, k.ConsumerKey); err != nil {
				return err
			}
		}
	}
	// TODO
	// FIXME, this needs to move to db layer
	err = das.db.DeveloperApp.DeleteByID(organizationName, developerApp.AppID)
	if err != nil {
		return err
	}
	// Remove app from the apps field in developer entity as well
	for i := 0; i < len(developer.Apps); i++ {
		if developer.Apps[i] == developerAppName {
			developer.Apps = append(developer.Apps[:i], developer.Apps[i+1:]...)
			i-- // from the remove item index to start iterate next item
		}
	}
	// FIXME	developer.LastmodifiedBy = h.GetSessionUser(c)
	if err := das.db.Developer.Update(organizationName, developer); err != nil {
		return err
	}
	das.audit.Delete(developerApp, who)
	return nil
}

// generateAppID creates unique primary key for developer app row
func generateAppID() string {
	return (uuid.New().String())
}
