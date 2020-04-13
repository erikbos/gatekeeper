package main

import (
	"fmt"
	"net/http"

	"github.com/erikbos/apiauth/pkg/shared"

	"github.com/gin-gonic/gin"
)

// registerOrganizationRoutes registers all routes we handle
func (e *env) registerOrganizationRoutes(r *gin.Engine) {
	r.GET("/v1/organizations", e.GetOrganizations)
	r.POST("/v1/organizations", shared.AbortIfContentTypeNotJSON, e.PostCreateOrganization)

	r.GET("/v1/organizations/:organization", e.GetOrganizationByName)
	r.POST("/v1/organizations/:organization", shared.AbortIfContentTypeNotJSON, e.PostOrganization)
	r.DELETE("/v1/organizations/:organization", e.DeleteOrganizationByName)

	r.GET("/v1/organizations/:organization/attributes", e.GetOrganizationAttributes)
	r.POST("/v1/organizations/:organization/attributes", shared.AbortIfContentTypeNotJSON, e.PostOrganizationAttributes)
	r.DELETE("/v1/organizations/:organization/attributes", e.DeleteOrganizationAttributes)

	r.GET("/v1/organizations/:organization/attributes/:attribute", e.GetOrganizationAttributeByName)
	r.POST("/v1/organizations/:organization/attributes/:attribute", shared.AbortIfContentTypeNotJSON, e.PostOrganizationAttributeByName)
	r.DELETE("/v1/organizations/:organization/attributes/:attribute", e.DeleteOrganizationAttributeByName)
}

// GetOrganizations returns all organizations
func (e *env) GetOrganizations(c *gin.Context) {
	organizations, err := e.db.GetOrganizations()
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"organizations": organizations})
}

// GetOrganizationByName returns details of an organization
func (e *env) GetOrganizationByName(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, organization.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, organization)
}

// GetOrganizationAttributes returns attributes of an organization
func (e *env) GetOrganizationAttributes(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, organization.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": organization.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of an organization
func (e *env) GetOrganizationAttributeByName(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(organization.Attributes); i++ {
		if organization.Attributes[i].Name == c.Param("attribute") {
			setLastModifiedHeader(c, organization.LastmodifiedAt)
			c.IndentedJSON(http.StatusOK, organization.Attributes[i])
			return
		}
	}
	returnJSONMessage(c, http.StatusNotFound,
		fmt.Errorf("Could not retrieve attribute '%s'", c.Param("attribute")))
}

// PostCreateOrganization creates an organization
func (e *env) PostCreateOrganization(c *gin.Context) {
	var newOrganization shared.Organization
	if err := c.ShouldBindJSON(&newOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	existingOrganization, err := e.db.GetOrganizationByName(newOrganization.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Organization '%s' already exists", existingOrganization.Name))
		return
	}
	// Automatically set default fields
	newOrganization.Key = newOrganization.Name
	newOrganization.CreatedBy = e.whoAmI()
	newOrganization.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(&newOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newOrganization)
}

// PostOrganization updates an existing organization
func (e *env) PostOrganization(c *gin.Context) {
	currentOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedOrganization shared.Organization
	if err := c.ShouldBindJSON(&updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update organization X while body says to update organization Y
	updatedOrganization.Name = currentOrganization.Name
	updatedOrganization.Key = currentOrganization.Key
	updatedOrganization.CreatedBy = currentOrganization.CreatedBy
	updatedOrganization.CreatedAt = currentOrganization.CreatedAt
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(&updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedOrganization)
}

// PostOrganizationAttributes updates attributes of an organization
func (e *env) PostOrganizationAttributes(c *gin.Context) {
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var receivedAttributes struct {
		Attributes []shared.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	updatedOrganization.Attributes = receivedAttributes.Attributes
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(&updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedOrganization.Attributes})
}

// PostOrganizationAttributeByName update an attribute of developer
func (e *env) PostOrganizationAttributeByName(c *gin.Context) {
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
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
		updatedOrganization.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find exist attribute, append new attribute
		updatedOrganization.Attributes = append(updatedOrganization.Attributes,
			shared.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedOrganization.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(&updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteOrganizationAttributeByName removes an attribute of an organization
func (e *env) DeleteOrganizationAttributeByName(c *gin.Context) {
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	attributeToDelete := c.Param("attribute")
	updatedAttributes, index, oldValue := shared.DeleteAttribute(updatedOrganization.Attributes, attributeToDelete)
	if index == -1 {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedOrganization.Attributes = updatedAttributes
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(&updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteOrganizationAttributes removes all attribute of an organization
func (e *env) DeleteOrganizationAttributes(c *gin.Context) {
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	DeleteDeveloperAttributes := updatedOrganization.Attributes
	updatedOrganization.Attributes = nil
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(&updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": DeleteDeveloperAttributes})
}

// DeleteOrganizationByName deletes an organization
func (e *env) DeleteOrganizationByName(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerCount := e.db.GetDeveloperCountByOrganization(organization.Name)
	switch developerCount {
	case -1:
		returnJSONMessage(c, http.StatusInternalServerError,
			fmt.Errorf("Could not retrieve number of developers in organization"))
	case 0:
		e.db.DeleteOrganizationByName(organization.Name)
		c.IndentedJSON(http.StatusOK, organization)
	default:
		returnJSONMessage(c, http.StatusForbidden,
			fmt.Errorf("Cannot delete organization '%s' with %d active developers",
				organization.Name, developerCount))
	}
}
