package main

import (
	// "fmt"
	"fmt"
	"net/http"

	// "github.com/erikbos/apiauth/pkg/types"
	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

func (e *env) registerDeveloperAppRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers/:developer/apps", e.GetDeveloperApps)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application", e.GetDeveloperAppByName)
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application", e.DeleteDeveloperAppByName)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes", e.GetDeveloperAppAttributes)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", e.GetDeveloperAppAttributeByName)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", e.CheckForJSONContentType, e.PostDeveloperAppAttributeByName)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", e.GetDeveloperAppByKey)

	// r.POST("/v1/organizations/:organization/developers/:developer/apps", e.PostAllDeveloperApps)
	// r.POST("/v1/organizations/:organization/developers/:developer/apps/:application", e.PostOneDeveloperApp)
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

// GetDeveloperAppByKey returns keys of one particular developer application
func (e *env) GetDeveloperAppByKey(c *gin.Context) {
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
	AppCredential, err := e.db.GetAppCredentialByKey(c.Param("key"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, AppCredential)
}

// PostDeveloperAppAttributeByName updates attribute of one particular app
func (e *env) PostDeveloperAppAttributeByName(c *gin.Context) {
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
	developerApp.Attributes = e.removeDuplicateAttributes(receivedAttributes.Attributes)
	developerApp.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	developerApp.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperAppByName(developerApp); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developerApp.Attributes})
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
