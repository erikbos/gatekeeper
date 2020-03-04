package main

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

func (e *env) registerDeveloperAppRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/apps", e.GetAllDevelopersApps)
	r.GET("/v1/organizations/:organization/apps/:application", e.GetDeveloperAppByName)
	// r.POST("/v1/organizations/:organization/apps/:application", e.CheckForJSONContentType, e.PostDeveloperApp)

	r.GET("/v1/organizations/:organization/developers/:developer/apps", e.GetDeveloperApps)
	// r.POST("/v1/organizations/:organization/developers/:developer/apps", e/CheckForJSONContentType, e.PostAllDeveloperApps)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application", e.GetDeveloperAppByName)
	// r.POST("/v1/organizations/:organization/developers/:developer/apps/:application", e.CheckForJSONContentType, e.PostDeveloperApp)
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application", e.DeleteDeveloperAppByName)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes", e.GetDeveloperAppAttributes)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/attributes", e.CheckForJSONContentType, e.PostDeveloperAppAttributes)
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application/attributes", e.DeleteDeveloperAppAttributes)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", e.GetDeveloperAppAttributeByName)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", e.CheckForJSONContentType, e.PostDeveloperAppAttributeByName)
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", e.DeleteDeveloperAppAttributeByName)
}

// GetAllDevelopersApps returns all developers in organization
// FIXME: add pagination support
func (e *env) GetAllDevelopersApps(c *gin.Context) {
	developerapps, err := e.db.GetDeveloperAppsByOrganization(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Should we return full details or just an array with names?
	if c.Query("expand") == "true" {
		c.IndentedJSON(http.StatusOK, gin.H{"apps": developerapps})
	} else {
		var developerAppNames []string
		for _, v := range developerapps {
			developerAppNames = append(developerAppNames, v.Name)
		}
		c.IndentedJSON(http.StatusOK, developerAppNames)
	}
}

// GetDeveloperApps returns apps of a developer
func (e *env) GetDeveloperApps(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, developer.Apps)
}

// GetDeveloperAppByName returns one named app of a developer
func (e *env) GetDeveloperAppByName(c *gin.Context) {
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developerApp.OrganizationName != c.Param("organization") {
		e.returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Organization (%s) does not match with developerapp's organization", c.Param("organization")))
		return
	}
	// All apikeys belonging to this developer app
	AppCredentials, err := e.db.GetAppCredentialByDeveloperAppID(developerApp.Key)
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerApp.Credentials = AppCredentials
	e.SetLastModified(c, developerApp.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, developerApp)
}

// GetDeveloperAppAttributes returns attributes of a developer
func (e *env) GetDeveloperAppAttributes(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Developer and application records do not have the same DevID"))
		return
	}
	e.SetLastModified(c, developerApp.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developerApp.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of a developer
func (e *env) GetDeveloperAppAttributeByName(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Developer and application records do not have the same DevID"))
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(developerApp.Attributes); i++ {
		if developerApp.Attributes[i].Name == c.Param("attribute") {
			e.SetLastModified(c, developerApp.LastmodifiedAt)
			c.IndentedJSON(http.StatusOK, developerApp.Attributes[i])
			return
		}
	}
	e.returnJSONMessage(c, http.StatusNotFound,
		fmt.Errorf("Could not retrieve attribute '%s'", c.Param("attribute")))
}

// PostDeveloperAppAttributes updates attribute of one particular app
func (e *env) PostDeveloperAppAttributes(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound, fmt.Errorf("Developer and application records do not have the same DevID"))
		return
	}
	var receivedAttributes struct {
		Attributes []types.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	log.Printf("QQQ")
	developerApp.Attributes = e.removeDuplicateAttributes(receivedAttributes.Attributes)
	developerApp.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	developerApp.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperAppByName(developerApp); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developerApp.Attributes})
}

// DeleteDeveloperAppAttributes delete attributes of developer app
func (e *env) DeleteDeveloperAppAttributes(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound, fmt.Errorf("Developer and application records do not have the same DevID"))
		return
	}
	deletedAttributes := developerApp.Attributes
	developerApp.Attributes = nil
	developerApp.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	developerApp.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperAppByName(developerApp); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttributes)
	// c.IndentedJSON(http.StatusOK, gin.H{"attribute": deletedAttributes})
}

// PostDeveloperAppAttributeByName update an attribute of developer
func (e *env) PostDeveloperAppAttributeByName(c *gin.Context) {
	updatedDeveloperApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if updatedDeveloperApp.OrganizationName != c.Param("organization") {
		e.returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Organization (%s) does not match with developerapp's organization", c.Param("organization")))
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
		updatedDeveloperApp.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find exist attribute, append new attribute
		updatedDeveloperApp.Attributes = append(updatedDeveloperApp.Attributes,
			types.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedDeveloperApp.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedDeveloperApp.LastmodifiedBy = e.whoAmI()
	updatedDeveloperApp.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	if err := e.db.UpdateDeveloperAppByName(updatedDeveloperApp); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteDeveloperAppAttributeByName removes an attribute of developer
func (e *env) DeleteDeveloperAppAttributeByName(c *gin.Context) {
	updatedDeveloperApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if updatedDeveloperApp.OrganizationName != c.Param("organization") {
		e.returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Organization (%s) does not match with developerapp's organization", c.Param("organization")))
		return
	}
	// Find attribute in array
	attributeToRemoveIndex := e.findAttributePositionInAttributeArray(
		updatedDeveloperApp.Attributes, c.Param("attribute"))
	if attributeToRemoveIndex == -1 {
		e.returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", c.Param("attribute")))
		return
	}
	deletedAttribute := updatedDeveloperApp.Attributes[attributeToRemoveIndex]
	// remove attribute
	updatedDeveloperApp.Attributes =
		append(updatedDeveloperApp.Attributes[:attributeToRemoveIndex],
			updatedDeveloperApp.Attributes[attributeToRemoveIndex+1:]...)
	updatedDeveloperApp.LastmodifiedBy = e.whoAmI()
	updatedDeveloperApp.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	if err := e.db.UpdateDeveloperAppByName(updatedDeveloperApp); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttribute)
}

// DeleteDeveloperAppByName deletes of one developer
func (e *env) DeleteDeveloperAppByName(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound, fmt.Errorf("Developer and application records do not have the same DevID"))
		return
	}
	if err := e.db.DeleteDeveloperAppByName(developerApp.OrganizationName, developerApp.Name); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.Status(http.StatusNoContent)
}
