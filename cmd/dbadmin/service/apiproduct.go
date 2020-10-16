package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// APIProductService is
type APIProductService struct {
	db        *db.Database
	changelog *Changelog
}

// NewAPIProduct returns a new apiproduct instance
func NewAPIProduct(database *db.Database, c *Changelog) *APIProductService {

	return &APIProductService{db: database, changelog: c}
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
func (ds *APIProductService) Create(organizationName string, newAPIProduct types.APIProduct,
	who Requester) (types.APIProduct, types.Error) {

	existingAPIProduct, err := ds.Get(organizationName, newAPIProduct.Name)
	if err == nil {
		return types.NullAPIProduct, types.NewBadRequestError(
			fmt.Errorf("APIProduct '%s' already exists", existingAPIProduct.Name))
	}
	// Automatically set default fields
	newAPIProduct.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newAPIProduct.CreatedBy = who.User

	// Automatically assign new APIProduct to organization
	newAPIProduct.OrganizationName = organizationName

	err = ds.updateAPIProduct(&newAPIProduct, who)
	ds.changelog.Create(newAPIProduct, who)
	return newAPIProduct, err
}

// Update updates an existing apiproduct
func (ds *APIProductService) Update(organizationName string, updatedAPIProduct types.APIProduct,
	who Requester) (types.APIProduct, types.Error) {

	currentAPIProduct, err := ds.db.APIProduct.Get(organizationName, updatedAPIProduct.Name)
	if err != nil {
		return types.NullAPIProduct, types.NewItemNotFoundError(err)
	}

	// Copy over fields we do not allow to be updated
	updatedAPIProduct.Name = currentAPIProduct.Name
	updatedAPIProduct.CreatedAt = currentAPIProduct.CreatedAt
	updatedAPIProduct.CreatedBy = currentAPIProduct.CreatedBy

	err = ds.updateAPIProduct(&updatedAPIProduct, who)
	ds.changelog.Update(currentAPIProduct, updatedAPIProduct, who)
	return updatedAPIProduct, err
}

// UpdateAttributes updates attributes of an apiproduct
func (ds *APIProductService) UpdateAttributes(organizationName string, apiproductName string,
	receivedAttributes types.Attributes, who Requester) types.Error {

	currentAPIProduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return err
	}
	updatedAPIProduct := currentAPIProduct
	if err = updatedAPIProduct.Attributes.SetMultiple(receivedAttributes); err != nil {
		return err
	}

	err = ds.updateAPIProduct(updatedAPIProduct, who)
	ds.changelog.Update(currentAPIProduct, updatedAPIProduct, who)
	return err
}

// UpdateAttribute update an attribute of apiproduct
func (ds *APIProductService) UpdateAttribute(organizationName string,
	apiproductName string, attributeValue types.Attribute, who Requester) types.Error {

	currentAPIProduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return err
	}
	updatedAPIProduct := currentAPIProduct
	updatedAPIProduct.Attributes.Set(attributeValue)

	err = ds.updateAPIProduct(updatedAPIProduct, who)
	ds.changelog.Update(currentAPIProduct, updatedAPIProduct, who)
	return err
}

// DeleteAttribute removes an attribute of an apiproduct
func (ds *APIProductService) DeleteAttribute(organizationName, apiproductName,
	attributeToDelete string, who Requester) (string, types.Error) {

	currentAPIProduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return "", err
	}
	updatedAPIProduct := currentAPIProduct
	oldValue, err := updatedAPIProduct.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}

	err = ds.updateAPIProduct(updatedAPIProduct, who)
	ds.changelog.Update(currentAPIProduct, updatedAPIProduct, who)
	return oldValue, err
}

// updateAPIProduct updates last-modified field(s) and updates apiproduct in database
func (ds *APIProductService) updateAPIProduct(updatedAPIProduct *types.APIProduct, who Requester) types.Error {

	updatedAPIProduct.Attributes.Tidy()
	updatedAPIProduct.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedAPIProduct.LastmodifiedBy = who.User
	return ds.db.APIProduct.Update(updatedAPIProduct)
}

// Delete deletes an apiproduct
func (ds *APIProductService) Delete(organizationName, apiproductName string,
	who Requester) (deletedAPIProduct types.APIProduct, e types.Error) {

	apiproduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return types.NullAPIProduct, err
	}

	// FIX ME (we probably allow deletion only in case no dev app uses the product)
	ds.changelog.Delete(apiproduct, who)
	return *apiproduct, nil
}
