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

	return ds.db.Developer.GetAll(organizationName)
}

// Get returns details of an developer, in case name contains a @ assumption is developerId was provided
func (ds *DeveloperService) Get(organizationName, developerName string) (developer *types.Developer, err types.Error) {

	if strings.Contains(developerName, "@") {
		return ds.db.Developer.GetByEmail(organizationName, developerName)
	}
	return ds.db.Developer.GetByID(organizationName, developerName)
}

// Create creates a new developer
func (ds *DeveloperService) Create(organizationName string, newDeveloper types.Developer,
	who Requester) (*types.Developer, types.Error) {

	if _, err := ds.Get(organizationName, newDeveloper.Email); err == nil {
		return nil, types.NewBadRequestError(
			fmt.Errorf("developer '%s' already exists", newDeveloper.Email))
	}
	// Automatically set default fields
	newDeveloper.DeveloperID = generateDeveloperID(newDeveloper.Email)
	newDeveloper.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newDeveloper.CreatedBy = who.User
	newDeveloper.Activate()

	if err := ds.updateDeveloper(organizationName, &newDeveloper, who); err != nil {
		return nil, err
	}
	ds.changelog.Create(newDeveloper, who)
	return &newDeveloper, nil
}

// Update updates an existing developer
func (ds *DeveloperService) Update(organizationName, developerEmail string, updatedDeveloper types.Developer, who Requester) (
	*types.Developer, types.Error) {

	currentDeveloper, err := ds.db.Developer.GetByEmail(organizationName, developerEmail)
	if err != nil {
		return nil, err
	}

	// Copy over fields we do not allow to be updated
	updatedDeveloper.Apps = currentDeveloper.Apps
	updatedDeveloper.DeveloperID = currentDeveloper.DeveloperID
	updatedDeveloper.CreatedAt = currentDeveloper.CreatedAt
	updatedDeveloper.CreatedBy = currentDeveloper.CreatedBy

	if err = ds.updateDeveloper(organizationName, &updatedDeveloper, who); err != nil {
		return nil, err
	}
	ds.changelog.Update(currentDeveloper, updatedDeveloper, who)
	return &updatedDeveloper, nil
}

// updateDeveloper updates last-modified field(s) and updates developer in database
func (ds *DeveloperService) updateDeveloper(organizationName string,
	updatedDeveloper *types.Developer, who Requester) types.Error {

	updatedDeveloper.Attributes.Tidy()
	updatedDeveloper.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedDeveloper.LastModifiedBy = who.User

	if err := updatedDeveloper.Validate(); err != nil {
		return types.NewBadRequestError(err)
	}
	return ds.db.Developer.Update(organizationName, updatedDeveloper)
}

// Delete deletes an developer
func (ds *DeveloperService) Delete(organizationName, developerName string, who Requester) (e types.Error) {

	developer, err := ds.Get(organizationName, developerName)
	if err != nil {
		return err
	}
	appCountOfDeveloper, err := ds.db.DeveloperApp.GetCountByDeveloperID(organizationName, developer.DeveloperID)
	if err != nil {
		return err
	}
	if appCountOfDeveloper > 0 {
		return types.NewBadRequestError(
			fmt.Errorf("cannot delete developer '%s' with %d active applications",
				developer.Email, appCountOfDeveloper))
	}
	if err := ds.db.Developer.DeleteByID(organizationName, developer.DeveloperID); err != nil {
		return err
	}
	ds.changelog.Delete(developer, who)
	return nil
}

// generateDeveloperID generates a DeveloperID
func generateDeveloperID(developer string) string {
	return uniuri.New()
}
