package main

import (
	"fmt"
	"net/http"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

// registerOrganizationRoutes registers all routes we handle
func (e *env) registerOrganizationRoutes(r *gin.Engine) {
	r.GET("/v1/organizations", e.GetOrganizations)
	r.POST("/v1/organizations", e.CheckForJSONContentType, e.PostCreateOrganization)

	r.GET("/v1/organizations/:organization", e.GetOrganizationByName)
	r.POST("/v1/organizations/:organization", e.CheckForJSONContentType, e.PostOrganization)
	r.DELETE("/v1/organizations/:organization", e.DeleteOrganizationByName)

	r.GET("/v1/organizations/:organization/attributes", e.GetOrganizationAttributes)
	r.POST("/v1/organizations/:organization/attributes", e.CheckForJSONContentType, e.PostOrganizationAttributes)
	r.DELETE("/v1/organizations/:organization/attributes", e.DeleteOrganizationAttributes)

	r.GET("/v1/organizations/:organization/attributes/:attribute", e.GetOrganizationAttributeByName)
	r.POST("/v1/organizations/:organization/attributes/:attribute", e.CheckForJSONContentType, e.PostOrganizationAttributeByName)
	r.DELETE("/v1/organizations/:organization/attributes/:attribute", e.DeleteOrganizationAttributeByName)
}

// GetOrganizations returns all organizations
func (e *env) GetOrganizations(c *gin.Context) {
	organizations, err := e.db.GetOrganizations()
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"organizations": organizations})
}

// GetOrganizationByName returns details of an organization
func (e *env) GetOrganizationByName(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	e.SetLastModifiedHeader(c, organization.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, organization)
}

// GetOrganizationAttributes returns attributes of an organization
func (e *env) GetOrganizationAttributes(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	e.SetLastModifiedHeader(c, organization.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": organization.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of an organization
func (e *env) GetOrganizationAttributeByName(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(organization.Attributes); i++ {
		if organization.Attributes[i].Name == c.Param("attribute") {
			e.SetLastModifiedHeader(c, organization.LastmodifiedAt)
			c.IndentedJSON(http.StatusOK, organization.Attributes[i])
			return
		}
	}
	e.returnJSONMessage(c, http.StatusNotFound,
		fmt.Errorf("Could not retrieve attribute '%s'", c.Param("attribute")))
}

// PostCreateOrganization creates an organization
func (e *env) PostCreateOrganization(c *gin.Context) {
	var newOrganization types.Organization
	if err := c.ShouldBindJSON(&newOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	existingOrganization, err := e.db.GetOrganizationByName(newOrganization.Name)
	if err == nil {
		e.returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Organization '%s' already exists", existingOrganization.Name))
		return
	}
	// Automatically set default fields
	newOrganization.Key = newOrganization.Name
	newOrganization.Attributes = e.removeDuplicateAttributes(newOrganization.Attributes)
	newOrganization.CreatedBy = e.whoAmI()
	newOrganization.CreatedAt = e.getCurrentTimeMilliseconds()
	newOrganization.LastmodifiedAt = newOrganization.CreatedAt
	newOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(newOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newOrganization)
}

// PostOrganization updates an existing organization
func (e *env) PostOrganization(c *gin.Context) {
	currentOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedOrganization types.Organization
	if err := c.ShouldBindJSON(&updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update organization X while body says to update organization Y
	updatedOrganization.Name = currentOrganization.Name
	updatedOrganization.Key = currentOrganization.Key
	updatedOrganization.CreatedBy = currentOrganization.CreatedBy
	updatedOrganization.CreatedAt = currentOrganization.CreatedAt
	updatedOrganization.Attributes = e.removeDuplicateAttributes(updatedOrganization.Attributes)
	updatedOrganization.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedOrganization)
}

// PostOrganizationAttributes updates attributes of an organization
func (e *env) PostOrganizationAttributes(c *gin.Context) {
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var receivedAttributes struct {
		Attributes []types.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	updatedOrganization.Attributes = e.removeDuplicateAttributes(receivedAttributes.Attributes)
	updatedOrganization.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedOrganization.Attributes})
}

// PostOrganizationAttributeByName update an attribute of developer
func (e *env) PostOrganizationAttributeByName(c *gin.Context) {
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
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
		updatedOrganization.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find exist attribute, append new attribute
		updatedOrganization.Attributes = append(updatedOrganization.Attributes,
			types.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedOrganization.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedOrganization.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteOrganizationAttributeByName removes an attribute of an organization
func (e *env) DeleteOrganizationAttributeByName(c *gin.Context) {
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// Find attribute in array
	attributeToRemoveIndex := e.findAttributePositionInAttributeArray(
		updatedOrganization.Attributes, c.Param("attribute"))
	if attributeToRemoveIndex == -1 {
		e.returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", c.Param("attribute")))
		return
	}
	deletedAttribute := updatedOrganization.Attributes[attributeToRemoveIndex]
	// remove attribute
	updatedOrganization.Attributes =
		append(updatedOrganization.Attributes[:attributeToRemoveIndex],
			updatedOrganization.Attributes[attributeToRemoveIndex+1:]...)
	updatedOrganization.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttribute)
}

// DeleteOrganizationAttributes removes all attribute of an organization
func (e *env) DeleteOrganizationAttributes(c *gin.Context) {
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	DeleteDeveloperAttributes := updatedOrganization.Attributes
	updatedOrganization.Attributes = nil
	updatedOrganization.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": DeleteDeveloperAttributes})
}

// DeleteOrganizationByName deletes an organization
func (e *env) DeleteOrganizationByName(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerCount := e.db.GetDeveloperCountByOrganization(organization.Name)
	switch developerCount {
	case -1:
		e.returnJSONMessage(c, http.StatusInternalServerError,
			fmt.Errorf("Could not retrieve number of developers in organization"))
	case 0:
		e.db.DeleteOrganizationByName(organization.Name)
		c.IndentedJSON(http.StatusOK, organization)
	default:
		e.returnJSONMessage(c, http.StatusForbidden,
			fmt.Errorf("Cannot delete organization '%s' with %d active developers",
				organization.Name, developerCount))
	}
}
