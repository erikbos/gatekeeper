package main

import (
	"fmt"
	"net/http"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

// registerDeveloperRoutes registers all routes we handle
func (e *env) registerDeveloperRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers", e.GetAllDevelopers)
	r.POST("/v1/organizations/:organization/developers", e.PostCreateDeveloper)

	r.GET("/v1/organizations/:organization/developers/:developer", e.GetDeveloperByEmail)
	r.POST("/v1/organizations/:organization/developers/:developer", e.PostUpdateDeveloper)
	r.DELETE("/v1/organizations/:organization/developers/:developer", e.DeleteDeveloperByEmail)

	r.GET("/v1/organizations/:organization/developers/:developer/attributes", e.GetDeveloperAttributes)
	r.POST("/v1/organizations/:organization/developers/:developer/attributes", e.PostUpdateDeveloperAttributes)
	r.DELETE("/v1/organizations/:organization/developers/:developer/attributes", e.DeleteDeveloperAttributes)

	r.GET("/v1/organizations/:organization/developers/:developer/attributes/:attribute", e.GetDeveloperAttributeByName)
	r.POST("/v1/organizations/:organization/developers/:developer/attributes/:attribute", e.PostUpdateDeveloperAttributeByName)
	r.DELETE("/v1/organizations/:organization/developers/:developer/attributes/:attribute", e.DeleteDeveloperAttributeByName)
}

// GetAllDevelopers returns all developers in organization
// FIXME: add pagination support
func (e *env) GetAllDevelopers(c *gin.Context) {
	developers, err := e.db.GetDevelopersByOrganization(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"developers": developers})
}

// GetDeveloperByEmail returns full details of one developer
func (e *env) GetDeveloperByEmail(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, developer)
}

// GetDeveloperAttributes returns attributes of a developer
func (e *env) GetDeveloperAttributes(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developer.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of a developer
func (e *env) GetDeveloperAttributeByName(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(developer.Attributes); i++ {
		if developer.Attributes[i].Name == c.Param("attribute") {
			c.IndentedJSON(http.StatusOK, gin.H{"value": developer.Attributes[i].Value})
			return
		}
	}
	e.returnJSONMessage(c, http.StatusNotFound, "Could not retrieve attribute")
}

// GetDeveloperAppAttributes returns attributes of a developer
func (e *env) GetDeveloperAppAttributes(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developerApp.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of a developer
func (e *env) GetDeveloperAppAttributeByName(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(developerApp.Attributes); i++ {
		if developerApp.Attributes[i].Name == c.Param("attribute") {
			c.IndentedJSON(http.StatusOK, developerApp.Attributes[i])
			return
		}
	}
	e.returnJSONMessage(c, http.StatusNotFound, "Could not retrieve attributes")
}

// PostCreateDeveloper creates a new developer
func (e *env) PostCreateDeveloper(c *gin.Context) {
	var newDeveloper types.Developer
	if err := c.ShouldBindJSON(&newDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	// we don't allow recreation of existing developer
	existingDeveloper, err := e.db.GetDeveloperByEmail(newDeveloper.Email)
	if err == nil {
		e.returnJSONMessage(c, http.StatusUnauthorized,
			fmt.Sprintf("Developer %s already exists", existingDeveloper.Email))
		return
	}
	// Automatically assign new developer to organization
	newDeveloper.OrganizationName = c.Param("organization")
	// Dedup provided attributes
	newDeveloper.Attributes = e.removeDuplicateAttributes(newDeveloper.Attributes)
	// New developers starts actived
	newDeveloper.Status = "active"
	newDeveloper.CreatedBy = e.whoAmI()
	newDeveloper.CreatedAt = e.getCurrentTimeMilliseconds()
	newDeveloper.LastmodifiedAt = newDeveloper.CreatedAt
	newDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.CreateDeveloper(newDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, newDeveloper)
}

// PostUpdateDeveloper updates an existing developer
func (e *env) PostUpdateDeveloper(c *gin.Context) {
	var updatedDeveloper types.Developer
	if err := c.ShouldBindJSON(&updatedDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	// Developer to update should exist
	currentDeveloper, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	// We don't allow POSTing to update developer X while body says to update developer Y
	updatedDeveloper.Email = currentDeveloper.Email
	updatedDeveloper.DeveloperID = currentDeveloper.DeveloperID
	updatedDeveloper.OrganizationName = currentDeveloper.OrganizationName
	updatedDeveloper.Attributes = e.removeDuplicateAttributes(updatedDeveloper.Attributes)
	updatedDeveloper.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByEmail(updatedDeveloper.Email, updatedDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, updatedDeveloper)
}

// PostUpdateDeveloperAttributes updates attributes of developer
func (e *env) PostUpdateDeveloperAttributes(c *gin.Context) {
	var receivedAttributes struct {
		Attributes []types.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	updatedDeveloper, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// We don't allow updating developer whom is member of a different organization
	if c.Param("organization") != updatedDeveloper.OrganizationName {
		e.returnJSONMessage(c, http.StatusBadRequest, "Organization mismatch")
		return
	}
	updatedDeveloper.Attributes = e.removeDuplicateAttributes(receivedAttributes.Attributes)
	updatedDeveloper.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByEmail(updatedDeveloper.Email, updatedDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedDeveloper.Attributes})
}

// DeleteDeveloperAttributes updates attributes of developer
func (e *env) DeleteDeveloperAttributes(c *gin.Context) {
	updatedDeveloper, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// We don't allow updating developer whom is member of a different organization
	if c.Param("organization") != updatedDeveloper.OrganizationName {
		e.returnJSONMessage(c, http.StatusBadRequest, "Organization mismatch")
		return
	}
	updatedDeveloper.Attributes = nil
	updatedDeveloper.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByEmail(updatedDeveloper.Email, updatedDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedDeveloper.Attributes})
}

// PostUpdateDeveloperAttributeByName update an attribute of developer
func (e *env) PostUpdateDeveloperAttributeByName(c *gin.Context) {
	var receivedValue struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	updatedDeveloper, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// We don't allow updating developer who is member of different organization
	if c.Param("organization") != updatedDeveloper.OrganizationName {
		e.returnJSONMessage(c, http.StatusBadRequest, "Organization mismatch")
		return
	}
	// Find & update existing attribute in array
	attributeToUpdateIndex := e.findAttributePositionInAttributeArray(
		updatedDeveloper.Attributes, c.Param("attribute"))
	if attributeToUpdateIndex == -1 {
		// We did not find exist attribute, so we cannot update its value
		updatedDeveloper.Attributes = append(updatedDeveloper.Attributes,
			types.AttributeKeyValues{Name: c.Param("attribute"), Value: receivedValue.Value})
	} else {
		updatedDeveloper.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	updatedDeveloper.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	if err := e.db.UpdateDeveloperByEmail(updatedDeveloper.Email, updatedDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, receivedValue)
}

// DeleteDeveloperAttributeByName removes an attribute of developer
func (e *env) DeleteDeveloperAttributeByName(c *gin.Context) {
	updatedDeveloper, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// We don't allow updating developer who is member of different organization
	if c.Param("organization") != updatedDeveloper.OrganizationName {
		e.returnJSONMessage(c, http.StatusBadRequest, "Organization mismatch")
		return
	}
	// Find attribute in array
	attributeToRemoveIndex := e.findAttributePositionInAttributeArray(
		updatedDeveloper.Attributes, c.Param("attribute"))
	if attributeToRemoveIndex == -1 {
		e.returnJSONMessage(c, http.StatusNotFound, "could not find attribute")
	}
	// remove attribute
	updatedDeveloper.Attributes =
		append(updatedDeveloper.Attributes[:attributeToRemoveIndex],
			updatedDeveloper.Attributes[attributeToRemoveIndex+1:]...)
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	updatedDeveloper.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	if err := e.db.UpdateDeveloperByEmail(updatedDeveloper.Email, updatedDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteDeveloperByEmail deletes of one developer
func (e *env) DeleteDeveloperByEmail(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// We don't allow updating developer who is member of different organization
	if c.Param("organization") != developer.OrganizationName {
		e.returnJSONMessage(c, http.StatusBadRequest, "Organization mismatch")
		return
	}
	if err := e.db.DeleteDeveloperByEmail(developer.Email); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
