package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/apiauth/pkg/shared"
)

// registerAPIProductRoutes registers all routes we handle
func (s *server) registerAPIProductRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/apiproducts", s.GetAllAPIProducts)
	r.POST("/v1/organizations/:organization/apiproducts", shared.AbortIfContentTypeNotJSON, s.PostCreateAPIProduct)

	r.GET("/v1/organizations/:organization/apiproducts/:apiproduct", s.GetAPIProductByName)
	r.POST("/v1/organizations/:organization/apiproducts/:apiproduct", shared.AbortIfContentTypeNotJSON, s.PostAPIProduct)
	r.DELETE("/v1/organizations/:organization/apiproducts/:apiproduct", s.DeleteAPIProductByName)

	r.GET("/v1/organizations/:organization/apiproducts/:apiproduct/attributes", s.GetAPIProductAttributes)
	r.POST("/v1/organizations/:organization/apiproducts/:apiproduct/attributes", shared.AbortIfContentTypeNotJSON, s.PostAPIProductAttributes)

	r.GET("/v1/organizations/:organization/apiproducts/:apiproduct/attributes/:attribute", s.GetAPIProductAttributeByName)
	r.POST("/v1/organizations/:organization/apiproducts/:apiproduct/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostAPIProductAttributeByName)
	r.DELETE("/v1/organizations/:organization/apiproducts/:apiproduct/attributes/:attribute", s.DeleteAPIProductAttributeByName)
}

// GetAllAPIProducts returns all apiproduct names in an organization
func (s *server) GetAllAPIProducts(c *gin.Context) {

	apiproducts, err := s.db.GetAPIProductsByOrganization(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	var apiproductNames []string
	for _, product := range apiproducts {
		apiproductNames = append(apiproductNames, product.Name)
	}

	c.IndentedJSON(http.StatusOK, gin.H{"apiproducts": apiproductNames})
}

// GetAPIProductByName returns full details of one APIProduct
func (s *server) GetAPIProductByName(c *gin.Context) {

	apiproduct, err := s.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, apiproduct.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, apiproduct)
}

// GetAPIProductAttributes returns attributes of a APIProduct
func (s *server) GetAPIProductAttributes(c *gin.Context) {

	apiproduct, err := s.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, apiproduct.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": apiproduct.Attributes})
}

// GetAPIProductAttributeByName returns one particular attribute of a APIProduct
func (s *server) GetAPIProductAttributeByName(c *gin.Context) {

	apiproduct, err := s.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	value, err := shared.GetAttribute(apiproduct.Attributes, c.Param("attribute"))
	if err != nil {
		returnCanNotFindAttribute(c, c.Param("attribute"))
		return
	}

	setLastModifiedHeader(c, apiproduct.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, value)
}

// PostCreateAPIProduct creates a new APIProduct
func (s *server) PostCreateAPIProduct(c *gin.Context) {

	var newAPIProduct shared.APIProduct
	if err := c.ShouldBindJSON(&newAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	// we don't allow recreation of existing APIProduct
	existingAPIProduct, err := s.db.GetAPIProductByName(c.Param("organization"), newAPIProduct.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("APIProduct '%s' already exists", existingAPIProduct.Name))
		return
	}

	// Automatically assign new APIProduct to organization
	newAPIProduct.OrganizationName = c.Param("organization")
	// Generate primary key for new row
	newAPIProduct.Key = generatePrimaryKeyOfAPIProduct(newAPIProduct.OrganizationName,
		newAPIProduct.Name)
	newAPIProduct.CreatedBy = s.whoAmI()
	newAPIProduct.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newAPIProduct.LastmodifiedBy = s.whoAmI()

	if err := s.db.UpdateAPIProductByName(&newAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, newAPIProduct)
}

// PostAPIProduct updates an existing APIProduct
func (s *server) PostAPIProduct(c *gin.Context) {
	// APIProduct to update should exist
	apiproductToUpdate, err := s.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	var updateRequest shared.APIProduct
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	// Copy over the fields we allow to be updated
	apiproductToUpdate.DisplayName = updateRequest.DisplayName
	apiproductToUpdate.Description = updateRequest.Description
	apiproductToUpdate.RouteSet = updateRequest.RouteSet
	apiproductToUpdate.APIResources = updateRequest.APIResources
	apiproductToUpdate.Attributes = updateRequest.Attributes

	apiproductToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.UpdateAPIProductByName(&apiproductToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, apiproductToUpdate)
}

// PostAPIProductAttributes updates attributes of APIProduct
func (s *server) PostAPIProductAttributes(c *gin.Context) {

	apiproductToUpdate, err := s.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	var body struct {
		Attributes []shared.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	if len(body.Attributes) == 0 {
		returnJSONMessage(c, http.StatusBadRequest, errors.New("No attributes posted"))
		return
	}

	apiproductToUpdate.Attributes = body.Attributes

	apiproductToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.UpdateAPIProductByName(&apiproductToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{
			"attribute": apiproductToUpdate.Attributes,
		})
}

// PostAPIProductAttributeByName update an attribute of APIProduct
func (s *server) PostAPIProductAttributeByName(c *gin.Context) {

	apiproductToUpdate, err := s.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	var body struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	attributeToUpdate := c.Param("attribute")
	apiproductToUpdate.Attributes = shared.UpdateAttribute(apiproductToUpdate.Attributes,
		attributeToUpdate, body.Value)

	apiproductToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.UpdateAPIProductByName(&apiproductToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{
			"name":  attributeToUpdate,
			"value": body.Value,
		})
}

// DeleteAPIProductAttributeByName removes an attribute of APIProduct
func (s *server) DeleteAPIProductAttributeByName(c *gin.Context) {

	updatedAPIProduct, err := s.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	attributeToDelete := c.Param("attribute")
	updatedAttributes, index, oldValue :=
		shared.DeleteAttribute(updatedAPIProduct.Attributes, attributeToDelete)
	if index == -1 {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedAPIProduct.Attributes = updatedAttributes

	updatedAPIProduct.LastmodifiedBy = s.whoAmI()

	if err := s.db.UpdateAPIProductByName(&updatedAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{
			"name":  attributeToDelete,
			"value": oldValue,
		})
}

// DeleteAPIProductByName deletes of one APIProduct
func (s *server) DeleteAPIProductByName(c *gin.Context) {

	apiproduct, err := s.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	// FIX ME (we probably allow deletion only in case no dev app uses the product)
	if err := s.db.DeleteAPIProductByName(apiproduct.OrganizationName, apiproduct.Name); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, apiproduct)
}

// GeneratePrimaryKeyOfAPIProduct creates unique primary key for apiproduct row
func generatePrimaryKeyOfAPIProduct(organization, name string) string {
	return (fmt.Sprintf("%s@@@%s", organization, name))
}
