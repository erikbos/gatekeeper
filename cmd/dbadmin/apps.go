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

	r.GET("/v1/organizations/:organization/developers/:developer/apps", e.GetDeveloperAppsByDeveloperEmail)
	r.POST("/v1/organizations/:organization/developers/:developer/apps", e.CheckForJSONContentType, e.PostCreateDeveloperApp)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application", e.GetDeveloperAppByName)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application", e.CheckForJSONContentType, e.PostDeveloperApp)
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

// GetDeveloperAppsByDeveloperEmail returns apps of a developer
func (e *env) GetDeveloperAppsByDeveloperEmail(c *gin.Context) {
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

// PostCreateDeveloperApp updates an existing developer
func (e *env) PostCreateDeveloperApp(c *gin.Context) {
	var newDeveloperApp types.DeveloperApp
	if err := c.ShouldBindJSON(&newDeveloperApp); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	existingDeveloperApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), newDeveloperApp.Name)
	if err == nil {
		e.returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Developer app '%s' already exists", existingDeveloperApp.Name))
		return
	}
	newDeveloperApp.AppID = e.GenerateDeveloperAppPrimaryKey()
	newDeveloperApp.DeveloperAppID = e.GenerateDeveloperAppID(c.Param("organization"), newDeveloperApp.AppID)
	// Automatically assign new developer to organization
	newDeveloperApp.OrganizationName = c.Param("organization")
	// Dedup provided attributes
	newDeveloperApp.Attributes = e.removeDuplicateAttributes(newDeveloperApp.Attributes)
	// New developers starts actived
	newDeveloperApp.Status = "active"
	newDeveloperApp.CreatedBy = e.whoAmI()
	newDeveloperApp.CreatedAt = e.getCurrentTimeMilliseconds()
	newDeveloperApp.LastmodifiedAt = newDeveloperApp.CreatedAt
	newDeveloperApp.LastmodifiedBy = e.whoAmI()
	newDeveloperApp.ParentID = developer.DeveloperID
	newDeveloperApp.ParentStatus = developer.Status
	if err := e.db.UpdateDeveloperAppByName(newDeveloperApp); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, newDeveloperApp)
}

// PostDeveloperApp updates an existing developer
func (e *env) PostDeveloperApp(c *gin.Context) {
	currentDeveloperApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if c.Param("developer") != currentDeveloperApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound, fmt.Errorf("Developer and application records do not have the same DevID"))
		return
	}
	var updatedDeveloperApp types.DeveloperApp
	if err := c.ShouldBindJSON(&updatedDeveloperApp); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update developer X while body says to update developer Y
	updatedDeveloperApp.Key = currentDeveloperApp.Key
	updatedDeveloperApp.Name = currentDeveloperApp.Name
	updatedDeveloperApp.AppID = currentDeveloperApp.AppID
	updatedDeveloperApp.DeveloperAppID = currentDeveloperApp.DeveloperAppID
	updatedDeveloperApp.ParentID = currentDeveloperApp.ParentID
	updatedDeveloperApp.ParentStatus = currentDeveloperApp.ParentStatus

	updatedDeveloperApp.Attributes = e.removeDuplicateAttributes(updatedDeveloperApp.Attributes)
	updatedDeveloperApp.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedDeveloperApp.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperAppByName(updatedDeveloperApp); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedDeveloperApp)
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
	AppCredentialCount := e.db.GetAppCredentialCountByDeveloperAppID(developerApp.DeveloperAppID)
	log.Printf("apikey count: %d", AppCredentialCount)
	switch AppCredentialCount {
	case -1:
		e.returnJSONMessage(c, http.StatusInternalServerError,
			fmt.Errorf("Could not retrieve number of developerapps of developer (%s)",
				developer.Email))
	case 0:
		// Not yet
		// e.db.DeleteDeveloperAppByName(developerApp.OrganizationName, developerApp.Name)
		c.IndentedJSON(http.StatusOK, developerApp)
	default:
		e.returnJSONMessage(c, http.StatusForbidden,
			fmt.Errorf("Cannot delete developer app '%s' with %d apikeys",
				developerApp.Name, AppCredentialCount))
	}
}
