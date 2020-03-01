package main

import (
	"fmt"
	"net/http"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

// registerOrganizationRoutes registers all routes we handle
func (e *env) registerOrganizationRoutes(r *gin.Engine) {
	r.GET("/v1/organizations", e.EnforeJSONContentType, e.GetOrganizations)
	r.POST("/v1/organizations", e.PostCreateOrganization)
	r.GET("/v1/organizations/:organization", e.GetOrganizationByName)
	r.POST("/v1/organizations/:organization", e.PostUpdateOrganization)
	r.DELETE("/v1/organizations/:organization", e.DeleteOrganizationByName)
	r.GET("/v1/organizations/:organization/attributes", e.GetOrganizationAttributes)
	r.POST("/v1/organizations/:organization/attributes", e.PostUpdateOrganizationAttributes)
	r.DELETE("/v1/organizations/:organization/attributes", e.DeleteOrganizationAttributes)
	r.GET("/v1/organizations/:organization/attributes/:attribute", e.GetOrganizationAttributeByName)
	r.POST("/v1/organizations/:organization/attributes/:attribute", e.PostUpdateOrganizationAttributeByName)
	r.DELETE("/v1/organizations/:organization/attributes/:attribute", e.DeleteOrganizationAttributeByName)
}

// GetOrganizations returns all organizations
func (e *env) GetOrganizations(c *gin.Context) {
	organizations, err := e.db.GetOrganizations()
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"organizations": organizations})
}

// GetOrganizationByName returns details of an organization
func (e *env) GetOrganizationByName(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, organization)
}

// GetOrganizationAttributes returns attributes of an organization
func (e *env) GetOrganizationAttributes(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": organization.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of an organization
func (e *env) GetOrganizationAttributeByName(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(organization.Attributes); i++ {
		if organization.Attributes[i].Name == c.Param("attribute") {
			c.IndentedJSON(http.StatusOK, gin.H{"value": organization.Attributes[i].Value})
			return
		}
	}
	e.returnJSONMessage(c, http.StatusNotFound, "Could not retrieve attribute")
}

// PostCreateOrganization creates an organization
func (e *env) PostCreateOrganization(c *gin.Context) {
	var newOrganization types.Organization
	if err := c.ShouldBindJSON(&newOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	// we don't allow creation of an existing organization
	existingOrganization, err := e.db.GetOrganizationByName(newOrganization.Name)
	if err == nil {
		e.returnJSONMessage(c, http.StatusUnauthorized,
			fmt.Sprintf("Organization %s already exists", existingOrganization.Name))
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
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusCreated, newOrganization)
}

// PostUpdateOrganization updates an existing organization
func (e *env) PostUpdateOrganization(c *gin.Context) {
	var updatedOrganization types.Organization
	if err := c.ShouldBindJSON(&updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	// Organization to update should exist
	currentOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
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
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, updatedOrganization)
}

// PostUpdateOrganizationAttributes updates attributes of an organization
func (e *env) PostUpdateOrganizationAttributes(c *gin.Context) {
	var receivedAttributes struct {
		Attributes []types.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	updatedOrganization.Attributes = e.removeDuplicateAttributes(receivedAttributes.Attributes)
	updatedOrganization.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedOrganization.Attributes})
}

// PostUpdateOrganizationAttributeByName update an attribute of developer
func (e *env) PostUpdateOrganizationAttributeByName(c *gin.Context) {
	var receivedValue struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	// Find & update existing attribute in array
	attributeToUpdateIndex := e.findAttributePositionInAttributeArray(
		updatedOrganization.Attributes, c.Param("attribute"))
	if attributeToUpdateIndex == -1 {
		// We did not find exist attribute, so we cannot update its value
		updatedOrganization.Attributes = append(updatedOrganization.Attributes,
			types.AttributeKeyValues{Name: c.Param("attribute"), Value: receivedValue.Value})
	} else {
		updatedOrganization.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedOrganization.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, receivedValue)
}

// DeleteOrganizationAttributeByName removes an attribute of an organization
func (e *env) DeleteOrganizationAttributeByName(c *gin.Context) {
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	// Find attribute in array
	attributeToRemoveIndex := e.findAttributePositionInAttributeArray(
		updatedOrganization.Attributes, c.Param("attribute"))
	if attributeToRemoveIndex == -1 {
		e.returnJSONMessage(c, http.StatusNotFound, "could not find attribute")
		return
	}
	// remove attribute
	updatedOrganization.Attributes =
		append(updatedOrganization.Attributes[:attributeToRemoveIndex],
			updatedOrganization.Attributes[attributeToRemoveIndex+1:]...)
	updatedOrganization.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteOrganizationAttributes removes all attribute of an organization
func (e *env) DeleteOrganizationAttributes(c *gin.Context) {
	updatedOrganization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	updatedOrganization.Attributes = nil
	updatedOrganization.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedOrganization.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateOrganizationByName(updatedOrganization); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteOrganizationByName deletes an organization
func (e *env) DeleteOrganizationByName(c *gin.Context) {
	organization, err := e.db.GetOrganizationByName(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	developerCount := e.db.GetDeveloperCountByOrganization(organization.Name)
	switch developerCount {
	case -1:
		e.returnJSONMessage(c, http.StatusInternalServerError,
			"could not retrieve number of developers in organization")
	case 0:
		e.db.DeleteOrganizationByName(organization.Name)
		c.Status(http.StatusNoContent)
	default:
		e.returnJSONMessage(c, http.StatusForbidden,
			fmt.Sprintf("Cannot delete organization %s with %d developers",
				organization.Name, developerCount))
	}
}
