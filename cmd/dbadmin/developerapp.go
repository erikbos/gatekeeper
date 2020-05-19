package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/erikbos/apiauth/pkg/shared"
)

func (s *server) registerDeveloperAppRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/apps", s.GetAllDevelopersApps)
	r.GET("/v1/organizations/:organization/apps/:application", s.GetDeveloperAppByName)

	r.GET("/v1/organizations/:organization/developers/:developer/apps", s.GetDeveloperAppsByDeveloperEmail)
	r.POST("/v1/organizations/:organization/developers/:developer/apps", shared.AbortIfContentTypeNotJSON, s.PostCreateDeveloperApp)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application", s.GetDeveloperAppByName)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application", shared.AbortIfContentTypeNotJSON, s.PostDeveloperApp)
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application", s.DeleteDeveloperAppByName)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes", s.GetDeveloperAppAttributes)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/attributes", shared.AbortIfContentTypeNotJSON, s.PostDeveloperAppAttributes)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", s.GetDeveloperAppAttributeByName)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostDeveloperAppAttributeByName)
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", s.DeleteDeveloperAppAttributeByName)
}

// GetAllDevelopersApps returns all developers in organization
// FIXME: add pagination support
func (s *server) GetAllDevelopersApps(c *gin.Context) {

	developerapps, err := s.db.GetDeveloperAppsByOrganization(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	// Should we return an array with names or full details?
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
func (s *server) GetDeveloperAppsByDeveloperEmail(c *gin.Context) {

	developer, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, developer.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, developer.Apps)
}

// GetDeveloperAppByName returns one named app of a developer
func (s *server) GetDeveloperAppByName(c *gin.Context) {

	developerApp, err := s.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	// // All apikeys belonging to this developer app
	// AppCredentials, err := s.db.GetAppCredentialByDeveloperAppID(developerApp.DeveloperAppID)
	// if err != nil {
	// 	returnJSONMessage(c, http.StatusNotFound, err)
	// 	return
	// }

	// developerApp.Credentials = AppCredentials

	setLastModifiedHeader(c, developerApp.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, developerApp)
}

// GetDeveloperAppAttributes returns attributes of a developer
func (s *server) GetDeveloperAppAttributes(c *gin.Context) {

	developerApp, err := s.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, developerApp.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developerApp.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of a developer
func (s *server) GetDeveloperAppAttributeByName(c *gin.Context) {

	developerApp, err := s.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	value, err := shared.GetAttribute(developerApp.Attributes, c.Param("attribute"))
	if err != nil {
		returnCanNotFindAttribute(c, c.Param("attribute"))
		return
	}

	setLastModifiedHeader(c, developerApp.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, value)
}

// PostCreateDeveloperApp creates a developer application
func (s *server) PostCreateDeveloperApp(c *gin.Context) {

	var newDeveloperApp shared.DeveloperApp
	if err := c.ShouldBindJSON(&newDeveloperApp); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	developer, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	existingDeveloperApp, err := s.db.GetDeveloperAppByName(c.Param("organization"), newDeveloperApp.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Developer app '%s' already exists", existingDeveloperApp.Name))
		return
	}

	newDeveloperApp.DeveloperAppID = generateDeveloperAppID()
	newDeveloperApp.DeveloperID = developer.DeveloperID
	newDeveloperApp.OrganizationName = c.Param("organization")

	// New developers starts actived
	newDeveloperApp.Status = "active"
	newDeveloperApp.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newDeveloperApp.CreatedBy = s.whoAmI()
	newDeveloperApp.LastmodifiedAt = newDeveloperApp.CreatedAt
	newDeveloperApp.LastmodifiedBy = s.whoAmI()

	if err := s.db.UpdateDeveloperAppByName(&newDeveloperApp); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	// Update apps field of developer to include name of this just created app
	developer.Apps = append(developer.Apps, newDeveloperApp.Name)
	developer.LastmodifiedBy = s.whoAmI()

	if err := s.db.UpdateDeveloperByName(&developer); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, newDeveloperApp)
}

// PostDeveloperApp updates an existing developer
func (s *server) PostDeveloperApp(c *gin.Context) {

	if _, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer")); err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	developerAppToUpdate, err := s.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	var updateRequest shared.DeveloperApp
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	// Copy over the fields we allow to be updated
	developerAppToUpdate.Name = updateRequest.Name
	developerAppToUpdate.DisplayName = updateRequest.DisplayName
	developerAppToUpdate.Attributes = updateRequest.Attributes
	developerAppToUpdate.Status = updateRequest.Status

	if err := s.db.UpdateDeveloperAppByName(&developerAppToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, developerAppToUpdate)
}

// PostDeveloperAppAttributes updates attribute of one particular app
func (s *server) PostDeveloperAppAttributes(c *gin.Context) {

	if _, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer")); err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerAppToUpdate, err := s.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
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

	developerAppToUpdate.Attributes = body.Attributes
	developerAppToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.UpdateDeveloperAppByName(&developerAppToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developerAppToUpdate.Attributes})
}

// PostDeveloperAppAttributeByName update an attribute of developer
func (s *server) PostDeveloperAppAttributeByName(c *gin.Context) {

	if _, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer")); err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	updatedDeveloperApp, err := s.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
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
	updatedDeveloperApp.Attributes = shared.UpdateAttribute(updatedDeveloperApp.Attributes,
		c.Param("attribute"), body.Value)

	updatedDeveloperApp.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateDeveloperAppByName(&updatedDeveloperApp); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": body.Value})
}

// DeleteDeveloperAppAttributeByName removes an attribute of developer
func (s *server) DeleteDeveloperAppAttributeByName(c *gin.Context) {

	if _, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer")); err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	updatedDeveloperApp, err := s.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	attributeToDelete := c.Param("attribute")
	updatedAttributes, index, oldValue := shared.DeleteAttribute(updatedDeveloperApp.Attributes, attributeToDelete)
	if index == -1 {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedDeveloperApp.Attributes = updatedAttributes
	updatedDeveloperApp.LastmodifiedBy = s.whoAmI()

	if err := s.db.UpdateDeveloperAppByName(&updatedDeveloperApp); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteDeveloperAppByName deletes a developer app
func (s *server) DeleteDeveloperAppByName(c *gin.Context) {

	developer, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerApp, err := s.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	AppCredentialCount := s.db.GetAppCredentialCountByDeveloperAppID(developerApp.DeveloperAppID)
	if AppCredentialCount == -1 {
		returnJSONMessage(c, http.StatusInternalServerError,
			fmt.Errorf("Could not retrieve number of api keys of developer app '%s'",
				developerApp.Name))
		return
	}
	if AppCredentialCount > 0 {
		returnJSONMessage(c, http.StatusForbidden,
			fmt.Errorf("Cannot delete developer app '%s' with %d apikeys",
				developerApp.Name, AppCredentialCount))
		return
	}
	err = s.db.DeleteDeveloperAppByID(developerApp.OrganizationName, developerApp.DeveloperAppID)
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}

	// Remove app from the apps field in developer entry as well
	for i := 0; i < len(developer.Apps); i++ {
		if developer.Apps[i] == c.Param("application") {
			developer.Apps = append(developer.Apps[:i], developer.Apps[i+1:]...)
			i-- // form the remove item index to start iterate next item
		}
	}

	developerApp.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateDeveloperByName(&developer); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, developerApp)
}

// generateDeveloperAppID creates unique primary key for developer app row
func generateDeveloperAppID() string {
	return (uuid.New().String())
}
