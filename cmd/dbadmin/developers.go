package main

import (
	"fmt"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

// registerDeveloperRoutes registers all routes we handle
func (e *env) registerDeveloperRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers", e.GetAllDevelopers)
	r.POST("/v1/organizations/:organization/developers", types.AbortIfContentTypeNotJSON, e.PostCreateDeveloper)

	r.GET("/v1/organizations/:organization/developers/:developer", e.GetDeveloperByEmail)
	r.POST("/v1/organizations/:organization/developers/:developer", types.AbortIfContentTypeNotJSON, e.PostDeveloper)
	r.DELETE("/v1/organizations/:organization/developers/:developer", e.DeleteDeveloperByEmail)

	r.GET("/v1/organizations/:organization/developers/:developer/attributes", e.GetDeveloperAttributes)
	r.POST("/v1/organizations/:organization/developers/:developer/attributes", types.AbortIfContentTypeNotJSON, e.PostDeveloperAttributes)
	r.DELETE("/v1/organizations/:organization/developers/:developer/attributes", e.DeleteDeveloperAttributes)

	r.GET("/v1/organizations/:organization/developers/:developer/attributes/:attribute", e.GetDeveloperAttributeByName)
	r.POST("/v1/organizations/:organization/developers/:developer/attributes/:attribute", types.AbortIfContentTypeNotJSON, e.PostDeveloperAttributeByName)
	r.DELETE("/v1/organizations/:organization/developers/:developer/attributes/:attribute", e.DeleteDeveloperAttributeByName)
}

// GetAllDevelopers returns all developers in organization
// FIXME: add pagination support
func (e *env) GetAllDevelopers(c *gin.Context) {
	developers, err := e.db.GetDevelopersByOrganization(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"developers": developers})
}

// GetDeveloperByEmail returns full details of one developer
func (e *env) GetDeveloperByEmail(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, developer.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, developer)
}

// GetDeveloperAttributes returns attributes of a developer
func (e *env) GetDeveloperAttributes(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, developer.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developer.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of a developer
func (e *env) GetDeveloperAttributeByName(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(developer.Attributes); i++ {
		if developer.Attributes[i].Name == c.Param("attribute") {
			setLastModifiedHeader(c, developer.LastmodifiedAt)
			c.IndentedJSON(http.StatusOK, developer.Attributes[i])
			return
		}
	}
	returnJSONMessage(c, http.StatusNotFound,
		fmt.Errorf("Could not retrieve attribute '%s'", c.Param("attribute")))
}

// PostCreateDeveloper creates a new developer
func (e *env) PostCreateDeveloper(c *gin.Context) {
	var newDeveloper types.Developer
	if err := c.ShouldBindJSON(&newDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// we don't allow recreation of existing developer
	existingDeveloper, err := e.db.GetDeveloperByEmail(c.Param("organization"), newDeveloper.Email)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Developer '%s' already exists", existingDeveloper.Email))
		return
	}
	// Automatically assign new developer to organization
	newDeveloper.OrganizationName = c.Param("organization")
	// Generate primary key for new row
	newDeveloper.DeveloperID = generatePrimaryKeyOfDeveloper(newDeveloper.OrganizationName,
		newDeveloper.Email)
	// Developer starts without any apps
	newDeveloper.Apps = nil
	// New developers starts actived
	newDeveloper.Status = "active"
	newDeveloper.CreatedBy = e.whoAmI()
	newDeveloper.CreatedAt = types.GetCurrentTimeMilliseconds()
	newDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByName(&newDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newDeveloper)
}

// PostDeveloper updates an existing developer
func (e *env) PostDeveloper(c *gin.Context) {
	// Developer to update should exist
	currentDeveloper, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedDeveloper types.Developer
	if err := c.ShouldBindJSON(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update developer X while body says to update developer Y ;-)
	updatedDeveloper.Email = currentDeveloper.Email
	updatedDeveloper.DeveloperID = currentDeveloper.DeveloperID
	updatedDeveloper.OrganizationName = currentDeveloper.OrganizationName
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByName(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedDeveloper)
}

// PostDeveloperAttributes updates attributes of developer
func (e *env) PostDeveloperAttributes(c *gin.Context) {
	updatedDeveloper, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	var receivedAttributes struct {
		Attributes []types.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	updatedDeveloper.Attributes = receivedAttributes.Attributes
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByName(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedDeveloper.Attributes})
}

// DeleteDeveloperAttributes delete attributes of developer
func (e *env) DeleteDeveloperAttributes(c *gin.Context) {
	updatedDeveloper, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	deletedAttributes := updatedDeveloper.Attributes
	updatedDeveloper.Attributes = nil
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByName(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttributes)
}

// PostDeveloperAttributeByName update an attribute of developer
func (e *env) PostDeveloperAttributeByName(c *gin.Context) {
	updatedDeveloper, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
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
	attributeToUpdateIndex := types.FindIndexOfAttribute(
		updatedDeveloper.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find existing attribute, append new attribute
		updatedDeveloper.Attributes = append(updatedDeveloper.Attributes,
			types.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedDeveloper.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByName(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteDeveloperAttributeByName removes an attribute of developer
func (e *env) DeleteDeveloperAttributeByName(c *gin.Context) {
	updatedDeveloper, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	attributeToDelete := c.Param("attribute")
	updatedAttributes, index, oldValue := types.DeleteAttribute(updatedDeveloper.Attributes, attributeToDelete)
	if index == -1 {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedDeveloper.Attributes = updatedAttributes
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByName(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteDeveloperByEmail deletes of one developer
func (e *env) DeleteDeveloperByEmail(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerAppCount := e.db.GetDeveloperAppCountByDeveloperID(developer.DeveloperID)
	switch developerAppCount {
	case -1:
		returnJSONMessage(c, http.StatusInternalServerError,
			fmt.Errorf("Could not retrieve number of developerapps of developer (%s)",
				developer.Email))
	case 0:
		if err := e.db.DeleteDeveloperByEmail(developer.OrganizationName, developer.Email); err != nil {
			returnJSONMessage(c, http.StatusBadRequest, err)
			return
		}
		c.IndentedJSON(http.StatusOK, developer)
	default:
		returnJSONMessage(c, http.StatusForbidden,
			fmt.Errorf("Cannot delete developer '%s' with %d developer apps",
				developer.Email, developerAppCount))
	}
}

// GeneratePrimaryKeyOfDeveloper creates unique primary key for developer db row
func generatePrimaryKeyOfDeveloper(organization, developer string) string {
	return (fmt.Sprintf("%s@@@%s", organization, uniuri.New()))
}
