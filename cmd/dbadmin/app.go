package main

import (
	// "fmt"
	"net/http"

	// "github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

func (e *env) registerDeveloperAppRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers/:developer/apps", e.GetDeveloperApps)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application", e.GetDeveloperAppByName)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes", e.GetDeveloperAppAttributes)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", e.GetDeveloperAppAttributeByName)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", e.GetDeveloperAppByKey)

	// r.POST("/v1/organizations/:organization/developers/:developer/apps", e.PostAllDeveloperApps)
	// r.POST("/v1/organizations/:organization/developers/:developer/apps/:application", e.PostOneDeveloperApp)
}

// GetDeveloperApps returns apps of a developer
func (e *env) GetDeveloperApps(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, developer.Apps)
}

// GetDeveloperAppByName returns one named app of a developer
func (e *env) GetDeveloperAppByName(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
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
		e.returnJSONMessage(c, http.StatusNotFound, "Developer and application records do not have the same DevID!")
		return
	}
	// All apikeys belonging to this developer app
	AppCredentials, err := e.db.GetAppCredentialByDeveloperAppID(developerApp.Key)
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	developerApp.Credentials = AppCredentials
	c.IndentedJSON(http.StatusOK, developerApp)
}

// GetDeveloperAppByKey returns keys of one particular developer application
func (e *env) GetDeveloperAppByKey(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
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
		e.returnJSONMessage(c, http.StatusNotFound, "Developer and application records do not have the same DevID!")
		return
	}
	AppCredential, err := e.db.GetAppCredentialByKey(c.Param("key"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, AppCredential)
}
