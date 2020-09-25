package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// APIProductService is
type APIProductService struct {
	db *db.Database
}

// NewAPIProductService returns a new apiproduct instance
func NewAPIProductService(database *db.Database) *APIProductService {

	return &APIProductService{db: database}
}

// GetByOrganization returns all apiproducts
func (ds *APIProductService) GetByOrganization(organizationName string) (apiproducts types.APIProducts, err types.Error) {

	return ds.db.APIProduct.GetByOrganization(organizationName)
}

// Get returns details of an apiproduct
func (ds *APIProductService) Get(organizationName, apiproductName string) (apiproduct *types.APIProduct, err types.Error) {

	return ds.db.APIProduct.Get(organizationName, apiproductName)
}

// GetAttributes returns attributes of an apiproduct
func (ds *APIProductService) GetAttributes(organizationName, apiproductName string) (attributes *types.Attributes, err types.Error) {

	apiproduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return nil, err
	}
	return &apiproduct.Attributes, nil
}

// GetAttribute returns one particular attribute of an apiproduct
func (ds *APIProductService) GetAttribute(organizationName, apiproductName, attributeName string) (value string, err types.Error) {

	apiproduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return "", err
	}
	return apiproduct.Attributes.Get(attributeName)
}

// Create creates a new apiproduct
func (ds *APIProductService) Create(organizationName string, newAPIProduct types.APIProduct) (types.APIProduct, types.Error) {

	existingAPIProduct, err := ds.Get(organizationName, newAPIProduct.Name)
	if err == nil {
		return types.NullAPIProduct, types.NewBadRequestError(
			fmt.Errorf("APIProduct '%s' already exists", existingAPIProduct.Name))
	}
	// Automatically set default fields
	// Automatically assign new APIProduct to organization
	newAPIProduct.OrganizationName = organizationName
	newAPIProduct.CreatedAt = shared.GetCurrentTimeMilliseconds()

	err = ds.updateAPIProduct(&newAPIProduct)
	return newAPIProduct, err
}

// Update updates an existing apiproduct
func (ds *APIProductService) Update(organizationName string, updatedAPIProduct types.APIProduct) (types.APIProduct, types.Error) {

	apiproductToUpdate, err := ds.db.APIProduct.Get(organizationName, updatedAPIProduct.Name)
	if err != nil {
		return types.NullAPIProduct, types.NewItemNotFoundError(err)
	}
	// Copy over the fields we allow to be updated
	apiproductToUpdate.DisplayName = updatedAPIProduct.DisplayName
	apiproductToUpdate.Description = updatedAPIProduct.Description
	apiproductToUpdate.RouteGroup = updatedAPIProduct.RouteGroup
	apiproductToUpdate.Paths = updatedAPIProduct.Paths
	apiproductToUpdate.Policies = updatedAPIProduct.Policies
	apiproductToUpdate.Attributes = updatedAPIProduct.Attributes

	err = ds.updateAPIProduct(apiproductToUpdate)
	return *apiproductToUpdate, err
}

// UpdateAttributes updates attributes of an apiproduct
func (ds *APIProductService) UpdateAttributes(organizationName string, apiproductName string, receivedAttributes types.Attributes) types.Error {

	updatedAPIProduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return err
	}
	updatedAPIProduct.Attributes = receivedAttributes
	return ds.updateAPIProduct(updatedAPIProduct)
}

// UpdateAttribute update an attribute of apiproduct
func (ds *APIProductService) UpdateAttribute(organizationName string,
	apiproductName string, attributeValue types.Attribute) types.Error {

	updatedAPIProduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return err
	}
	updatedAPIProduct.Attributes.Set(attributeValue)
	return ds.updateAPIProduct(updatedAPIProduct)
}

// DeleteAttribute removes an attribute of an apiproduct
func (ds *APIProductService) DeleteAttribute(organizationName, apiproductName, attributeToDelete string) (string, types.Error) {

	updatedAPIProduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return "", err
	}
	oldValue, err := updatedAPIProduct.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}
	return oldValue, ds.updateAPIProduct(updatedAPIProduct)
}

// updateAPIProduct updates last-modified field(s) and updates apiproduct in database
func (ds *APIProductService) updateAPIProduct(updatedAPIProduct *types.APIProduct) types.Error {

	updatedAPIProduct.Attributes.Tidy()
	updatedAPIProduct.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	return ds.db.APIProduct.Update(updatedAPIProduct)
}

// Delete deletes an apiproduct
func (ds *APIProductService) Delete(organizationName, apiproductName string) (deletedAPIProduct types.APIProduct, e types.Error) {

	apiproduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return types.NullAPIProduct, err
	}

	// FIX ME (we probably allow deletion only in case no dev app uses the product)

	return *apiproduct, nil
}
