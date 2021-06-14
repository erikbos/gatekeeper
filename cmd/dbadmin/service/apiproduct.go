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

	return &APIProductService{
		db:        database,
		changelog: c,
	}
}

// GetAll returns all apiproducts
func (ds *APIProductService) GetAll() (apiproducts types.APIProducts, err types.Error) {

	return ds.db.APIProduct.GetAll()
}

// Get returns details of an apiproduct
func (ds *APIProductService) Get(apiproductName string) (apiproduct *types.APIProduct, err types.Error) {

	return ds.db.APIProduct.Get(apiproductName)
}

// GetAttributes returns attributes of an apiproduct
func (ds *APIProductService) GetAttributes(apiproductName string) (attributes *types.Attributes, err types.Error) {

	apiproduct, err := ds.Get(apiproductName)
	if err != nil {
		return nil, err
	}
	return &apiproduct.Attributes, nil
}

// GetAttribute returns one particular attribute of an apiproduct
func (ds *APIProductService) GetAttribute(apiproductName, attributeName string) (value string, err types.Error) {

	apiproduct, err := ds.Get(apiproductName)
	if err != nil {
		return "", err
	}
	return apiproduct.Attributes.Get(attributeName)
}

// Create creates a new apiproduct
func (ds *APIProductService) Create(newAPIProduct types.APIProduct,
	who Requester) (types.APIProduct, types.Error) {

	if _, err := ds.Get(newAPIProduct.Name); err == nil {
		return types.NullAPIProduct, types.NewBadRequestError(
			fmt.Errorf("apiproduct '%s' already exists", newAPIProduct.Name))
	}
	// Automatically set default fields
	newAPIProduct.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newAPIProduct.CreatedBy = who.User

	if err := ds.updateAPIProduct(&newAPIProduct, who); err != nil {
		return types.NullAPIProduct, err
	}
	ds.changelog.Create(newAPIProduct, who)
	return newAPIProduct, nil
}

// Update updates an existing apiproduct
func (ds *APIProductService) Update(updatedAPIProduct types.APIProduct,
	who Requester) (types.APIProduct, types.Error) {

	currentAPIProduct, err := ds.db.APIProduct.Get(updatedAPIProduct.Name)
	if err != nil {
		return types.NullAPIProduct, err
	}

	// Copy over fields we do not allow to be updated
	updatedAPIProduct.Name = currentAPIProduct.Name
	updatedAPIProduct.CreatedAt = currentAPIProduct.CreatedAt
	updatedAPIProduct.CreatedBy = currentAPIProduct.CreatedBy

	if err = ds.updateAPIProduct(&updatedAPIProduct, who); err != nil {
		return types.NullAPIProduct, err
	}
	ds.changelog.Update(currentAPIProduct, updatedAPIProduct, who)
	return updatedAPIProduct, nil
}

// UpdateAttributes updates attributes of an apiproduct
func (ds *APIProductService) UpdateAttributes(apiproductName string,
	receivedAttributes types.Attributes, who Requester) types.Error {

	currentAPIProduct, err := ds.Get(apiproductName)
	if err != nil {
		return err
	}
	updatedAPIProduct := currentAPIProduct
	if err = updatedAPIProduct.Attributes.SetMultiple(receivedAttributes); err != nil {
		return err
	}

	if err = ds.updateAPIProduct(updatedAPIProduct, who); err != nil {
		return err
	}
	ds.changelog.Update(currentAPIProduct, updatedAPIProduct, who)
	return nil
}

// UpdateAttribute update an attribute of apiproduct
func (ds *APIProductService) UpdateAttribute(apiproductName string,
	attributeValue types.Attribute, who Requester) types.Error {

	currentAPIProduct, err := ds.Get(apiproductName)
	if err != nil {
		return err
	}
	updatedAPIProduct := currentAPIProduct
	if err := updatedAPIProduct.Attributes.Set(attributeValue); err != nil {
		return err
	}

	if err := ds.updateAPIProduct(updatedAPIProduct, who); err != nil {
		return err
	}
	ds.changelog.Update(currentAPIProduct, updatedAPIProduct, who)
	return nil
}

// DeleteAttribute removes an attribute of an apiproduct
func (ds *APIProductService) DeleteAttribute(apiproductName,
	attributeToDelete string, who Requester) (string, types.Error) {

	currentAPIProduct, err := ds.Get(apiproductName)
	if err != nil {
		return "", err
	}
	updatedAPIProduct := currentAPIProduct
	oldValue, err := updatedAPIProduct.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}

	if err = ds.updateAPIProduct(updatedAPIProduct, who); err != nil {
		return "", err
	}
	ds.changelog.Update(currentAPIProduct, updatedAPIProduct, who)
	return oldValue, nil
}

// updateAPIProduct updates last-modified field(s) and updates apiproduct in database
func (ds *APIProductService) updateAPIProduct(updatedAPIProduct *types.APIProduct, who Requester) types.Error {

	updatedAPIProduct.Attributes.Tidy()
	updatedAPIProduct.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedAPIProduct.LastmodifiedBy = who.User
	return ds.db.APIProduct.Update(updatedAPIProduct)
}

// Delete deletes an apiproduct
func (ds *APIProductService) Delete(apiproductName string,
	who Requester) (deletedAPIProduct types.APIProduct, e types.Error) {

	apiproduct, err := ds.Get(apiproductName)
	if err != nil {
		return types.NullAPIProduct, err
	}

	// FIX ME (we probably allow deletion only in case no dev app uses the product)
	ds.changelog.Delete(apiproduct, who)
	return *apiproduct, nil
}
