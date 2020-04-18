package main

import (
	"fmt"
	"net/http"

	"github.com/erikbos/apiauth/pkg/shared"

	"github.com/gin-gonic/gin"
)

// registerAPIProductRoutes registers all routes we handle
func (e *env) registerAPIProductRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/apiproducts", e.GetAllAPIProducts)
	r.POST("/v1/organizations/:organization/apiproducts", shared.AbortIfContentTypeNotJSON, e.PostCreateAPIProduct)

	r.GET("/v1/organizations/:organization/apiproducts/:apiproduct", e.GetAPIProductByName)
	r.POST("/v1/organizations/:organization/apiproducts/:apiproduct", shared.AbortIfContentTypeNotJSON, e.PostAPIProduct)
	r.DELETE("/v1/organizations/:organization/apiproducts/:apiproduct", e.DeleteAPIProductByName)

	r.GET("/v1/organizations/:organization/apiproducts/:apiproduct/attributes", e.GetAPIProductAttributes)
	r.POST("/v1/organizations/:organization/apiproducts/:apiproduct/attributes", shared.AbortIfContentTypeNotJSON, e.PostAPIProductAttributes)
	r.DELETE("/v1/organizations/:organization/apiproducts/:apiproduct/attributes", e.DeleteAPIProductAttributes)

	r.GET("/v1/organizations/:organization/apiproducts/:apiproduct/attributes/:attribute", e.GetAPIProductAttributeByName)
	r.POST("/v1/organizations/:organization/apiproducts/:apiproduct/attributes/:attribute", shared.AbortIfContentTypeNotJSON, e.PostAPIProductAttributeByName)
	r.DELETE("/v1/organizations/:organization/apiproducts/:apiproduct/attributes/:attribute", e.DeleteAPIProductAttributeByName)
}

// GetAllAPIProducts returns all apiproduct names in an organization
func (e *env) GetAllAPIProducts(c *gin.Context) {
	apiproducts, err := e.db.GetAPIProductsByOrganization(c.Param("organization"))
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
func (e *env) GetAPIProductByName(c *gin.Context) {
	apiproduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, apiproduct.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, apiproduct)
}

// GetAPIProductAttributes returns attributes of a APIProduct
func (e *env) GetAPIProductAttributes(c *gin.Context) {
	apiproduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, apiproduct.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": apiproduct.Attributes})
}

// GetAPIProductAttributeByName returns one particular attribute of a APIProduct
func (e *env) GetAPIProductAttributeByName(c *gin.Context) {
	apiproduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(apiproduct.Attributes); i++ {
		if apiproduct.Attributes[i].Name == c.Param("attribute") {
			setLastModifiedHeader(c, apiproduct.LastmodifiedAt)
			c.IndentedJSON(http.StatusOK, apiproduct.Attributes[i])
			return
		}
	}
	returnJSONMessage(c, http.StatusNotFound,
		fmt.Errorf("Could not retrieve attribute '%s'", c.Param("attribute")))
}

// PostCreateAPIProduct creates a new APIProduct
func (e *env) PostCreateAPIProduct(c *gin.Context) {
	var newAPIProduct shared.APIProduct
	if err := c.ShouldBindJSON(&newAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// we don't allow recreation of existing APIProduct
	existingAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), newAPIProduct.Name)
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
	newAPIProduct.CreatedBy = e.whoAmI()
	newAPIProduct.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newAPIProduct.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateAPIProductByName(&newAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newAPIProduct)
}

// PostAPIProduct updates an existing APIProduct
func (e *env) PostAPIProduct(c *gin.Context) {
	// APIProduct to update should exist
	currentAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedAPIProduct shared.APIProduct
	if err := c.ShouldBindJSON(&updatedAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update APIProduct X while body says to update APIProduct Y ;-)
	updatedAPIProduct.Name = currentAPIProduct.Name
	updatedAPIProduct.Key = currentAPIProduct.Key
	updatedAPIProduct.OrganizationName = currentAPIProduct.OrganizationName
	updatedAPIProduct.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateAPIProductByName(&updatedAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedAPIProduct)
}

// PostAPIProductAttributes updates attributes of APIProduct
func (e *env) PostAPIProductAttributes(c *gin.Context) {
	updatedAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	var receivedAttributes struct {
		Attributes []shared.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	updatedAPIProduct.Attributes = receivedAttributes.Attributes
	updatedAPIProduct.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateAPIProductByName(&updatedAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedAPIProduct.Attributes})
}

// DeleteAPIProductAttributes delete attributes of APIProduct
func (e *env) DeleteAPIProductAttributes(c *gin.Context) {
	updatedAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	deletedAttributes := updatedAPIProduct.Attributes
	updatedAPIProduct.Attributes = nil
	updatedAPIProduct.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateAPIProductByName(&updatedAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttributes)
}

// PostAPIProductAttributeByName update an attribute of APIProduct
func (e *env) PostAPIProductAttributeByName(c *gin.Context) {
	updatedAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	attributeToUpdate := c.Param("attribute")
	var receivedValue struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// Find & update existing attribute in array
	attributeToUpdateIndex := shared.FindIndexOfAttribute(
		updatedAPIProduct.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find existing attribute, append new attribute
		updatedAPIProduct.Attributes = append(updatedAPIProduct.Attributes,
			shared.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedAPIProduct.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedAPIProduct.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateAPIProductByName(&updatedAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteAPIProductAttributeByName removes an attribute of APIProduct
func (e *env) DeleteAPIProductAttributeByName(c *gin.Context) {
	updatedAPIProduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
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
	updatedAPIProduct.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateAPIProductByName(&updatedAPIProduct); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteAPIProductByName deletes of one APIProduct
func (e *env) DeleteAPIProductByName(c *gin.Context) {
	apiproduct, err := e.db.GetAPIProductByName(c.Param("organization"), c.Param("apiproduct"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// FIX ME (we probably allow deletion only in case no dev app uses the product)
	// if err := e.db.DeleteAPIProductByName(apiproduct.OrganizationName, apiproduct.Name); err != nil {
	// 	returnJSONMessage(c, http.StatusBadRequest, err)
	// 	return
	// }
	c.IndentedJSON(http.StatusOK, apiproduct)
}

// GeneratePrimaryKeyOfAPIProduct creates unique primary key for apiproduct row
func generatePrimaryKeyOfAPIProduct(organization, name string) string {
	return (fmt.Sprintf("%s@@@%s", organization, name))
}
