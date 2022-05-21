package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// returns all apiproducts
// (GET /v1/organizations/{organization_name}/apiproducts)
func (h *Handler) GetV1OrganizationsOrganizationNameApiproducts(c *gin.Context,
	organizationName OrganizationName, params GetV1OrganizationsOrganizationNameApiproductsParams) {

	apiproducts, err := h.service.APIProduct.GetAll(string(organizationName))
	if err != nil {
		responseError(c, err)
		return
	}
	// Do we have to return full developer details?
	if params.Expand != nil && *params.Expand {
		h.responseAPIproducts(c, apiproducts)
		return
	}
	h.responseAPIproductNames(c, apiproducts)
}

// creates a new apiproduct
// (POST /v1/organizations/{organization_name}/apiproducts)
func (h *Handler) PostV1OrganizationsOrganizationNameApiproducts(c *gin.Context,
	organizationName OrganizationName) {

	var receivedAPIProduct APIProduct
	if err := c.ShouldBindJSON(&receivedAPIProduct); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newAPIProduct := fromAPIproduct(receivedAPIProduct)
	createdDeveloper, err := h.service.APIProduct.Create(string(organizationName), newAPIProduct, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAPIProductCreated(c, createdDeveloper)
}

// deletes an apiproduct
// (DELETE /v1/organizations/{organization_name}/apiproducts/{apiproduct_name})
func (h *Handler) DeleteV1OrganizationsOrganizationNameApiproductsApiproductName(
	c *gin.Context, organizationName OrganizationName, apiproductName ApiproductName) {

	deletedApiproduct, err := h.service.APIProduct.Get(string(organizationName), string(apiproductName))
	if err != nil {
		responseError(c, err)
		return
	}
	if err := h.service.APIProduct.Delete(
		string(organizationName), string(apiproductName), h.who(c)); err != nil {
		responseError(c, err)
		return
	}
	h.responseAPIproduct(c, deletedApiproduct)
}

// returns full details of one apiproduct
// (GET /v1/organizations/{organization_name}/apiproducts/{apiproduct_name})
func (h *Handler) GetV1OrganizationsOrganizationNameApiproductsApiproductName(
	c *gin.Context, organizationName OrganizationName, apiproductName ApiproductName) {

	apiproduct, err := h.service.APIProduct.Get(string(organizationName), string(apiproductName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAPIproduct(c, apiproduct)
}

// updates existing apiproduct
// (POST /v1/organizations/{organization_name}/apiproducts/{apiproduct_name})
func (h *Handler) PostV1OrganizationsOrganizationNameApiproductsApiproductName(
	c *gin.Context, organizationName OrganizationName, apiproductName ApiproductName,
	params PostV1OrganizationsOrganizationNameApiproductsApiproductNameParams) {

	var receivedAPIProduct APIProduct
	if err := c.ShouldBindJSON(&receivedAPIProduct); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	updatedAPIProduct := fromAPIproduct(receivedAPIProduct)
	storedAPIProduct, err := h.service.APIProduct.Update(string(organizationName), updatedAPIProduct, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAPIproductUpdated(c, storedAPIProduct)
}

// returns attributes of a apiproduct
// (GET /v1/organizations/{organization_name}/apiproducts/{apiproduct_name}/attributes)
func (h *Handler) GetV1OrganizationsOrganizationNameApiproductsApiproductNameAttributes(
	c *gin.Context, organizationName OrganizationName, apiproductName ApiproductName) {

	apiproduct, err := h.service.APIProduct.Get(string(organizationName), string(apiproductName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, apiproduct.Attributes)
}

// replaces attributes of an apiproduct
// (POST /v1/organizations/{organization_name}/apiproducts/{apiproduct_name}/attributes)
func (h *Handler) PostV1OrganizationsOrganizationNameApiproductsApiproductNameAttributes(
	c *gin.Context, organizationName OrganizationName, apiproductName ApiproductName) {

	var receivedAttributes Attributes
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	apiproduct, err := h.service.APIProduct.Get(string(organizationName), string(apiproductName))
	if err != nil {
		responseError(c, err)
		return
	}
	apiproduct.Attributes = fromAttributesRequest(receivedAttributes.Attribute)
	storedAPIProduct, err := h.service.APIProduct.Update(string(organizationName), *apiproduct, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, storedAPIProduct.Attributes)
}

// deletes one attribute of an apiproduct
// (DELETE /v1/organizations/{organization_name}/apiproducts/{apiproduct_name}/attributes/{attribute_name})
func (h *Handler) DeleteV1OrganizationsOrganizationNameApiproductsApiproductNameAttributesAttributeName(
	c *gin.Context, organizationName OrganizationName, apiproductName ApiproductName, attributeName AttributeName) {

	apiproduct, err := h.service.APIProduct.Get(string(organizationName), string(apiproductName))
	if err != nil {
		responseError(c, err)
		return
	}
	oldValue, err := apiproduct.Attributes.Delete(string(attributeName))
	if err != nil {
		responseError(c, err)
		return
	}
	_, err = h.service.APIProduct.Update(string(organizationName), *apiproduct, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeDeleted(c, types.NewAttribute(string(attributeName), oldValue))
}

// returns one attribute of an apiproduct
// (GET /v1/organizations/{organization_name}/apiproducts/{apiproduct_name}/attributes/{attribute_name})
func (h *Handler) GetV1OrganizationsOrganizationNameApiproductsApiproductNameAttributesAttributeName(
	c *gin.Context, organizationName OrganizationName, apiproductName ApiproductName, attributeName AttributeName) {

	apiproduct, err := h.service.APIProduct.Get(string(organizationName), string(apiproductName))
	if err != nil {
		responseError(c, err)
		return
	}
	attributeValue, err := apiproduct.Attributes.Get(string(attributeName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeRetrieved(c, types.NewAttribute(string(attributeName), attributeValue))
}

// updates an attribute of an apiproduct
// (POST /v1/organizations/{organization_name}/apiproducts/{apiproduct_name}/attributes/{attribute_name})
func (h *Handler) PostV1OrganizationsOrganizationNameApiproductsApiproductNameAttributesAttributeName(
	c *gin.Context, organizationName OrganizationName, apiproductName ApiproductName, attributeName AttributeName) {

	var receivedValue Attribute
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	apiproduct, err := h.service.APIProduct.Get(string(organizationName), string(apiproductName))
	if err != nil {
		responseError(c, err)
		return
	}
	newAttribute := types.NewAttribute(string(attributeName), *receivedValue.Value)
	if err := apiproduct.Attributes.Set(newAttribute); err != nil {
		responseError(c, err)
		return
	}
	_, err = h.service.APIProduct.Update(string(organizationName), *apiproduct, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeUpdated(c, newAttribute)
}

// API responses

func (h *Handler) responseAPIproductNames(c *gin.Context, apiproducts types.APIProducts) {

	APIproductNames := make([]string, len(apiproducts))
	for i := range apiproducts {
		APIproductNames[i] = apiproducts[i].Name
	}
	c.IndentedJSON(http.StatusOK, APIproductNames)
}

func (h *Handler) responseAPIproducts(c *gin.Context, apiproducts types.APIProducts) {

	allApiproducts := make([]APIProduct, len(apiproducts))
	for i := range apiproducts {
		allApiproducts[i] = h.ToAPIproductResponse(&apiproducts[i])
	}
	c.IndentedJSON(http.StatusOK, APIProducts{
		ApiProduct: &allApiproducts,
	})
}

func (h *Handler) responseAPIproduct(c *gin.Context, apiproduct *types.APIProduct) {

	c.IndentedJSON(http.StatusOK, h.ToAPIproductResponse(apiproduct))
}

func (h *Handler) responseAPIProductCreated(c *gin.Context, apiproduct *types.APIProduct) {

	c.IndentedJSON(http.StatusCreated, h.ToAPIproductResponse(apiproduct))
}

func (h *Handler) responseAPIproductUpdated(c *gin.Context, apiproduct *types.APIProduct) {

	c.IndentedJSON(http.StatusOK, h.ToAPIproductResponse(apiproduct))
}

// type conversion

func (h *Handler) ToAPIproductResponse(a *types.APIProduct) APIProduct {

	p := APIProduct{
		ApprovalType:   &a.ApprovalType,
		Attributes:     toAttributesResponse(a.Attributes),
		CreatedAt:      &a.CreatedAt,
		CreatedBy:      &a.CreatedBy,
		Description:    &a.Description,
		DisplayName:    &a.DisplayName,
		LastModifiedBy: &a.LastModifiedBy,
		LastModifiedAt: &a.LastModifiedAt,
		Name:           &a.Name,
	}
	if a.APIResources != nil {
		p.ApiResources = &a.APIResources
	} else {
		p.ApiResources = &[]string{}
	}
	if a.Scopes != nil {
		p.Scopes = &a.Scopes
	} else {
		p.Scopes = &[]string{}
	}
	return p
}

func fromAPIproduct(p APIProduct) types.APIProduct {

	product := types.APIProduct{}
	if p.ApprovalType != nil {
		product.ApprovalType = *p.ApprovalType
	}
	if p.Attributes != nil {
		product.Attributes = fromAttributesRequest(p.Attributes)
	}
	if p.CreatedAt != nil {
		product.CreatedAt = *p.CreatedAt
	}
	if p.CreatedBy != nil {
		product.CreatedBy = *p.CreatedBy
	}
	if p.Description != nil {
		product.Description = *p.Description
	}
	if p.DisplayName != nil {
		product.DisplayName = *p.DisplayName
	}
	if p.Name != nil {
		product.Name = *p.Name
	}
	if p.ApiResources != nil {
		product.APIResources = *p.ApiResources
	}
	if p.LastModifiedBy != nil {
		product.LastModifiedBy = *p.LastModifiedBy
	}
	if p.LastModifiedAt != nil {
		product.LastModifiedAt = *p.LastModifiedAt
	}
	if p.Scopes != nil {
		product.Scopes = *p.Scopes
	}
	return product
}
