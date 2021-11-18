package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// OrganizationService is
type OrganizationService struct {
	db        *db.Database
	changelog *Changelog
}

// NewOrganization returns a new organization instance
func NewOrganization(database *db.Database, c *Changelog) *OrganizationService {

	return &OrganizationService{
		db:        database,
		changelog: c,
	}
}

// GetAll returns all organizations
func (os *OrganizationService) GetAll() (organizations types.Organizations, err types.Error) {

	return os.db.Organization.GetAll()
}

// Get returns details of an organization
func (os *OrganizationService) Get(organizationName string) (organization *types.Organization, err types.Error) {

	return os.db.Organization.Get(organizationName)
}

// Create creates an organization
func (os *OrganizationService) Create(newOrganization types.Organization, who Requester) (types.Organization, types.Error) {

	if _, err := os.db.Organization.Get(newOrganization.Name); err == nil {
		return types.NullOrganization, types.NewBadRequestError(
			fmt.Errorf("organization '%s' already exists", newOrganization.Name))
	}
	// Automatically set default fields
	newOrganization.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newOrganization.CreatedBy = who.User

	if err := os.updateOrganization(&newOrganization, who); err != nil {
		return types.NullOrganization, err
	}
	os.changelog.Create(newOrganization, who)
	return newOrganization, nil
}

// Update updates an existing organization
func (os *OrganizationService) Update(updatedOrganization types.Organization, who Requester) (types.Organization, types.Error) {

	currentOrganization, err := os.db.Organization.Get(updatedOrganization.Name)
	if err != nil {
		return types.NullOrganization, err
	}
	// Copy over fields we do not allow to be updated
	updatedOrganization.Name = currentOrganization.Name
	updatedOrganization.CreatedAt = currentOrganization.CreatedAt
	updatedOrganization.CreatedBy = currentOrganization.CreatedBy

	if err = os.updateOrganization(&updatedOrganization, who); err != nil {
		return types.NullOrganization, err
	}
	os.changelog.Update(currentOrganization, updatedOrganization, who)
	return updatedOrganization, nil
}

// updateOrganization updates last-modified field(s) and updates cluster in database
func (os *OrganizationService) updateOrganization(updatedOrganization *types.Organization, who Requester) types.Error {

	updatedOrganization.Attributes.Tidy()
	updatedOrganization.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedOrganization.LastModifiedBy = who.User

	if err := updatedOrganization.Validate(); err != nil {
		return types.NewBadRequestError(err)
	}
	return os.db.Organization.Update(updatedOrganization)
}

// Delete deletes an organization
func (os *OrganizationService) Delete(organizationName string, who Requester) (e types.Error) {

	organization, err := os.db.Organization.Get(organizationName)
	if err != nil {
		return err
	}
	if err = os.db.Organization.Delete(organizationName); err != nil {
		return err
	}
	os.changelog.Delete(organization, who)
	return nil
}
