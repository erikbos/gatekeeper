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
func (ds *APIProductService) GetAll(organizationName string) (apiproducts types.APIProducts, err types.Error) {

	return ds.db.APIProduct.GetAll(organizationName)
}

// Get returns details of an apiproduct
func (ds *APIProductService) Get(organizationName, apiproductName string) (apiproduct *types.APIProduct, err types.Error) {

	return ds.db.APIProduct.Get(organizationName, apiproductName)
}

// Create creates a new apiproduct
func (ds *APIProductService) Create(organizationName string, newAPIProduct types.APIProduct,
	who Requester) (types.APIProduct, types.Error) {

	if _, err := ds.Get(organizationName, newAPIProduct.Name); err == nil {
		return types.NullAPIProduct, types.NewBadRequestError(
			fmt.Errorf("apiproduct '%s' already exists", newAPIProduct.Name))
	}
	// Automatically set default fields
	newAPIProduct.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newAPIProduct.CreatedBy = who.User

	if newAPIProduct.ApprovalType == "" {
		newAPIProduct.ApprovalType = "auto"
	}

	if err := ds.updateAPIProduct(organizationName, &newAPIProduct, who); err != nil {
		return types.NullAPIProduct, err
	}
	ds.changelog.Create(newAPIProduct, who)
	return newAPIProduct, nil
}

// Update updates an existing apiproduct
func (ds *APIProductService) Update(organizationName string, updatedAPIProduct types.APIProduct,
	who Requester) (types.APIProduct, types.Error) {

	currentAPIProduct, err := ds.db.APIProduct.Get(organizationName, updatedAPIProduct.Name)
	if err != nil {
		return types.NullAPIProduct, err
	}

	// Copy over fields we do not allow to be updated
	updatedAPIProduct.Name = currentAPIProduct.Name
	updatedAPIProduct.CreatedAt = currentAPIProduct.CreatedAt
	updatedAPIProduct.CreatedBy = currentAPIProduct.CreatedBy

	if err = ds.updateAPIProduct(organizationName, &updatedAPIProduct, who); err != nil {
		return types.NullAPIProduct, err
	}
	ds.changelog.Update(currentAPIProduct, updatedAPIProduct, who)
	return updatedAPIProduct, nil
}

// updateAPIProduct updates last-modified field(s) and updates apiproduct in database
func (ds *APIProductService) updateAPIProduct(organizationName string,
	updatedAPIProduct *types.APIProduct, who Requester) types.Error {

	updatedAPIProduct.Attributes.Tidy()
	updatedAPIProduct.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedAPIProduct.LastModifiedBy = who.User
	return ds.db.APIProduct.Update(organizationName, updatedAPIProduct)
}

// Delete deletes an apiproduct
func (ds *APIProductService) Delete(organizationName, apiproductName string,
	who Requester) (e types.Error) {

	apiproduct, err := ds.Get(organizationName, apiproductName)
	if err != nil {
		return err
	}
	keyWithAPIProduct, err := ds.db.Key.GetCountByAPIProductName(organizationName, apiproductName)
	if err != nil {
		return err
	}
	if keyWithAPIProduct > 0 {
		return types.NewBadRequestError(
			fmt.Errorf("cannot delete api product '%s' assigned to %d keys",
				apiproductName, keyWithAPIProduct))
	}
	if err := ds.db.APIProduct.Delete(organizationName, apiproductName); err != nil {
		return err
	}
	ds.changelog.Delete(apiproduct, who)
	return nil
}
