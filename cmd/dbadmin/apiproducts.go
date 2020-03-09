package main

import (
	"fmt"
	"net/http"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

// registerAPIProductRoutes registers all routes we handle
func (e *env) registerAPIProductRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/apiproducts", e.GetAllAPIProducts)
	r.POST("/v1/organizations/:organization/apiproducts", e.CheckForJSONContentType, e.PostCreateAPIProduct)

	r.GET("/v1/organizations/:organization/apiproducts/:apiproduct", e.GetAPIProductByName)
	r.POST("/v1/organizations/:organization/apiproducts/:apiproduct", e.CheckForJSONContentType, e.PostAPIProduct)
	r.DELETE("/v1/organizations/:organization/apiproducts/:apiproduct", e.DeleteAPIProductByName)

	r.GET("/v1/organizations/:organization/apiproducts/:apiproduct/attributes", e.GetAPIProductAttributes)
	r.POST("/v1/organizations/:organization/apiproducts/:apiproduct/attributes", e.CheckForJSONContentType, e.PostAPIProductAttributes)
	r.DELETE("/v1/organizations/:organization/apiproducts/:apiproduct/attributes", e.DeleteAPIProductAttributes)

	r.GET("/v1/organizations/:organization/apiproducts/:apiproduct/attributes/:attribute", e.GetAPIProductAttributeByName)
	r.POST("/v1/organizations/:organization/apiproducts/:apiproduct/attributes/:attribute", e.CheckForJSONContentType, e.PostAPIProductAttributeByName)
	r.DELETE("/v1/organizations/:organization/apiproducts/:apiproduct/attributes/:attribute", e.DeleteAPIProductAttributeByName)
}

// GetAllAPIProducts returns all apiproduct names in an organization
func (e *env) GetAllAPIProducts(c *gin.Context) {
	apiproducts, err := e.db.GetAPIProductsByOrganization(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	var apiproductNames []string
	for _, product := range apiproducts {
		apiproductNames = append(apiproductNames, product.Name)
	}
	c.IndentedJSON(http.StatusOK, gin.H{"apiproducts": apiproductNames})
}

// GetAPIProductByName returns full details of one APIProduct
func (e *env) GetAPIProductByName(c *gin.Context) {
	apiproduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	e.SetLastModifiedHeader(c, apiproduct.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, apiproduct)
}

// GetAPIProductAttributes returns attributes of a APIProduct
func (e *env) GetAPIProductAttributes(c *gin.Context) {
	apiproduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	e.SetLastModifiedHeader(c, apiproduct.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": apiproduct.Attributes})
}

// GetAPIProductAttributeByName returns one particular attribute of a APIProduct
func (e *env) GetAPIProductAttributeByName(c *gin.Context) {
	apiproduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(apiproduct.Attributes); i++ {
		if apiproduct.Attributes[i].Name == c.Param("attribute") {
			e.SetLastModifiedHeader(c, apiproduct.LastmodifiedAt)
			c.IndentedJSON(http.StatusOK, apiproduct.Attributes[i])
			return
		}
	}
	e.returnJSONMessage(c, http.StatusNotFound,
		fmt.Errorf("Could not retrieve attribute '%s'", c.Param("attribute")))
}

// PostCreateAPIProduct creates a new APIProduct
func (e *env) PostCreateAPIProduct(c *gin.Context) {
	var newAPIProduct types.APIProduct
	if err := c.ShouldBindJSON(&newAPIProduct); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// we don't allow recreation of existing APIProduct
	existingAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), newAPIProduct.Name)
	if err == nil {
		e.returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("APIProduct '%s' already exists", existingAPIProduct.Name))
		return
	}
	// Automatically assign new APIProduct to organization
	newAPIProduct.OrganizationName = c.Param("organization")
	// Generate primary key for new row
	newAPIProduct.Key = e.GeneratePrimaryKeyOfAPIProduct(newAPIProduct.OrganizationName,
		newAPIProduct.Name)
	// Dedup provided attributes
	newAPIProduct.Attributes = e.removeDuplicateAttributes(newAPIProduct.Attributes)
	newAPIProduct.CreatedBy = e.whoAmI()
	newAPIProduct.CreatedAt = e.getCurrentTimeMilliseconds()
	newAPIProduct.LastmodifiedAt = newAPIProduct.CreatedAt
	newAPIProduct.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateAPIProductByName(newAPIProduct); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newAPIProduct)
}

// PostAPIProduct updates an existing APIProduct
func (e *env) PostAPIProduct(c *gin.Context) {
	// APIProduct to update should exist
	currentAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedAPIProduct types.APIProduct
	if err := c.ShouldBindJSON(&updatedAPIProduct); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update APIProduct X while body says to update APIProduct Y ;-)
	updatedAPIProduct.Name = currentAPIProduct.Name
	updatedAPIProduct.Key = currentAPIProduct.Key
	updatedAPIProduct.OrganizationName = currentAPIProduct.OrganizationName
	updatedAPIProduct.Attributes = e.removeDuplicateAttributes(updatedAPIProduct.Attributes)
	updatedAPIProduct.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedAPIProduct.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateAPIProductByName(updatedAPIProduct); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedAPIProduct)
}

// PostAPIProductAttributes updates attributes of APIProduct
func (e *env) PostAPIProductAttributes(c *gin.Context) {
	updatedAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	var receivedAttributes struct {
		Attributes []types.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	updatedAPIProduct.Attributes = e.removeDuplicateAttributes(receivedAttributes.Attributes)
	updatedAPIProduct.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedAPIProduct.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateAPIProductByName(updatedAPIProduct); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedAPIProduct.Attributes})
}

// DeleteAPIProductAttributes delete attributes of APIProduct
func (e *env) DeleteAPIProductAttributes(c *gin.Context) {
	updatedAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	deletedAttributes := updatedAPIProduct.Attributes
	updatedAPIProduct.Attributes = nil
	updatedAPIProduct.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedAPIProduct.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateAPIProductByName(updatedAPIProduct); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttributes)
}

// PostAPIProductAttributeByName update an attribute of APIProduct
func (e *env) PostAPIProductAttributeByName(c *gin.Context) {
	updatedAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	attributeToUpdate := c.Param("attribute")
	var receivedValue struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// Find & update existing attribute in array
	attributeToUpdateIndex := e.findAttributePositionInAttributeArray(
		updatedAPIProduct.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find existing attribute, append new attribute
		updatedAPIProduct.Attributes = append(updatedAPIProduct.Attributes,
			types.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedAPIProduct.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedAPIProduct.LastmodifiedBy = e.whoAmI()
	updatedAPIProduct.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	if err := e.db.UpdateAPIProductByName(updatedAPIProduct); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteAPIProductAttributeByName removes an attribute of APIProduct
func (e *env) DeleteAPIProductAttributeByName(c *gin.Context) {
	updatedAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Find attribute in array
	attributeToRemoveIndex := e.findAttributePositionInAttributeArray(
		updatedAPIProduct.Attributes, c.Param("attribute"))
	if attributeToRemoveIndex == -1 {
		e.returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", c.Param("attribute")))
		return
	}
	deletedAttribute := updatedAPIProduct.Attributes[attributeToRemoveIndex]
	// remove attribute
	updatedAPIProduct.Attributes =
		append(updatedAPIProduct.Attributes[:attributeToRemoveIndex],
			updatedAPIProduct.Attributes[attributeToRemoveIndex+1:]...)
	updatedAPIProduct.LastmodifiedBy = e.whoAmI()
	updatedAPIProduct.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	if err := e.db.UpdateAPIProductByName(updatedAPIProduct); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttribute)
}

// DeleteAPIProductByName deletes of one APIProduct
func (e *env) DeleteAPIProductByName(c *gin.Context) {
	apiproduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// if err := e.db.DeleteAPIProductByName(apiproduct.OrganizationName, apiproduct.Name); err != nil {
	// 	e.returnJSONMessage(c, http.StatusBadRequest, err)
	// 	return
	// }
	c.IndentedJSON(http.StatusOK, apiproduct)
}
