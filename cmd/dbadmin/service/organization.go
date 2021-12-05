package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/audit"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// OrganizationService is
type OrganizationService struct {
	db    *db.Database
	audit *audit.Audit
}

// NewOrganization returns a new organization instance
func NewOrganization(database *db.Database, a *audit.Audit) *OrganizationService {

	return &OrganizationService{
		db:    database,
		audit: a,
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
func (os *OrganizationService) Create(newOrganization types.Organization, who audit.Requester) (*types.Organization, types.Error) {

	if _, err := os.db.Organization.Get(newOrganization.Name); err == nil {
		return nil, types.NewBadRequestError(
			fmt.Errorf("organization '%s' already exists", newOrganization.Name))
	}
	// Automatically set default fields
	newOrganization.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newOrganization.CreatedBy = who.User

	if err := os.updateOrganization(&newOrganization, who); err != nil {
		return nil, err
	}
	os.audit.Create(newOrganization, who)
	return &newOrganization, nil
}

// Update updates an existing organization
func (os *OrganizationService) Update(updatedOrganization types.Organization, who audit.Requester) (*types.Organization, types.Error) {

	currentOrganization, err := os.db.Organization.Get(updatedOrganization.Name)
	if err != nil {
		return nil, err
	}
	// Copy over fields we do not allow to be updated
	updatedOrganization.Name = currentOrganization.Name
	updatedOrganization.CreatedAt = currentOrganization.CreatedAt
	updatedOrganization.CreatedBy = currentOrganization.CreatedBy

	if err = os.updateOrganization(&updatedOrganization, who); err != nil {
		return nil, err
	}
	os.audit.Update(currentOrganization, updatedOrganization, who)
	return &updatedOrganization, nil
}

// updateOrganization updates last-modified field(s) and updates cluster in database
func (os *OrganizationService) updateOrganization(updatedOrganization *types.Organization, who audit.Requester) types.Error {

	updatedOrganization.Attributes.Tidy()
	updatedOrganization.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedOrganization.LastModifiedBy = who.User

	if err := updatedOrganization.Validate(); err != nil {
		return types.NewBadRequestError(err)
	}
	return os.db.Organization.Update(updatedOrganization)
}

// Delete deletes an organization
func (os *OrganizationService) Delete(organizationName string, who audit.Requester) (e types.Error) {

	organization, err := os.db.Organization.Get(organizationName)
	if err != nil {
		return err
	}
	if err = os.db.Organization.Delete(organizationName); err != nil {
		return err
	}
	os.audit.Delete(organization, who)
	return nil
}
