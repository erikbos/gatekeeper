package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// OrganizationService has the CRUD methods to manipulate the organization entity
type OrganizationService struct {
	db *db.Database
}

// NewOrganizationService returns a new organization instance
func NewOrganizationService(database *db.Database) *OrganizationService {

	return &OrganizationService{db: database}
}

// GetAll returns all organizations
func (os *OrganizationService) GetAll() (organizations types.Organizations, err types.Error) {

	return os.db.Organization.GetAll()
}

// Get returns details of an organization
func (os *OrganizationService) Get(name string) (organization *types.Organization, err types.Error) {

	return os.db.Organization.Get(name)
}

// GetAttributes returns attributes of an organization
func (os *OrganizationService) GetAttributes(OrganizatioName string) (attributes *types.Attributes, err types.Error) {

	organization, err := os.db.Organization.Get(OrganizatioName)
	if err != nil {
		return nil, err
	}
	return &organization.Attributes, nil
}

// GetAttribute returns one particular attribute of an organization
func (os *OrganizationService) GetAttribute(OrganizationName, attributeName string) (value string, err types.Error) {

	organization, err := os.db.Organization.Get(OrganizationName)
	if err != nil {
		return "", err
	}
	return organization.Attributes.Get(attributeName)
}

// Create creates an organization
func (os *OrganizationService) Create(newOrganization types.Organization) (types.Organization, types.Error) {

	existingOrganization, err := os.db.Organization.Get(newOrganization.Name)
	if err == nil {
		return types.NullOrganization, types.NewBadRequestError(
			fmt.Errorf("Organization '%s' already exists", existingOrganization.Name))
	}
	// Automatically set default fields
	newOrganization.CreatedAt = shared.GetCurrentTimeMilliseconds()

	err = os.updateOrganization(&newOrganization)
	return newOrganization, err
}

// Update updates an existing organization
func (os *OrganizationService) Update(updatedOrganization types.Organization) (types.Organization, types.Error) {

	organizationToUpdate, err := os.db.Organization.Get(updatedOrganization.Name)
	if err != nil {
		return types.NullOrganization, types.NewItemNotFoundError(err)
	}
	// Copy over fields we allow to be updated
	organizationToUpdate.DisplayName = updatedOrganization.DisplayName
	organizationToUpdate.Attributes = updatedOrganization.Attributes

	err = os.updateOrganization(organizationToUpdate)
	return *organizationToUpdate, err
}

// UpdateAttributes updates attributes of an organization
func (os *OrganizationService) UpdateAttributes(org string, receivedAttributes types.Attributes) types.Error {

	updatedOrganization, err := os.db.Organization.Get(org)
	if err != nil {
		return err
	}
	updatedOrganization.Attributes = receivedAttributes
	return os.updateOrganization(updatedOrganization)
}

// UpdateAttribute update an attribute of organization
func (os *OrganizationService) UpdateAttribute(org string, attributeValue types.Attribute) types.Error {

	updatedOrganization, err := os.db.Organization.Get(org)
	if err != nil {
		return err
	}
	updatedOrganization.Attributes.Set(attributeValue)
	return os.updateOrganization(updatedOrganization)
}

// DeleteAttribute removes an attribute of an organization
func (os *OrganizationService) DeleteAttribute(organizationName, attributeToDelete string) (string, types.Error) {

	updatedOrganization, err := os.db.Organization.Get(organizationName)
	if err != nil {
		return "", err
	}
	oldValue, err := updatedOrganization.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}
	return oldValue, os.updateOrganization(updatedOrganization)
}

// updateOrganization updates last-modified field(s) and updates organization in database
func (os *OrganizationService) updateOrganization(updatedOrganization *types.Organization) types.Error {

	updatedOrganization.Attributes.Tidy()
	updatedOrganization.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	return os.db.Organization.Update(updatedOrganization)
}

// Delete deletes an organization
func (os *OrganizationService) Delete(organizationName string) (deletedOrganization types.Organization, e types.Error) {

	organization, err := os.db.Organization.Get(organizationName)
	if err != nil {
		return types.NullOrganization, err
	}

	developerCount, err := os.db.Developer.GetCountByOrganization(organization.Name)
	if err != nil {
		return types.NullOrganization, err
	}
	switch developerCount {
	case -1:
		return types.NullOrganization, types.NewBadRequestError(
			fmt.Errorf("Could not retrieve number of developers in organization"))
	case 0:
		if err := os.db.Organization.Delete(organization.Name); err != nil {
			return types.NullOrganization, err
		}
		return *organization, nil
	default:
		return types.NullOrganization, types.NewPermissionDeniedError(
			fmt.Errorf("Cannot delete organization '%s' with %d active developers",
				organization.Name, developerCount))
	}
}
