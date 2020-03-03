package main

import (
	"fmt"
	"net/http"

	// "github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

func (e *env) registerCredentialRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys", e.GetDeveloperAppKeys)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", e.GetDeveloperAppKeyByKey)
}

// GetDeveloperAppByKey returns keys of one particular developer application
func (e *env) GetDeveloperAppKeys(c *gin.Context) {
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
	AppCredentials, err := e.db.GetAppCredentialByDeveloperAppID(developerApp.Key)
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"credentials": AppCredentials})
}

// GetDeveloperAppByKey returns keys of one particular developer application
func (e *env) GetDeveloperAppKeyByKey(c *gin.Context) {
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
