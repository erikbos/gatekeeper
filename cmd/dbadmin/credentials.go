package main

import (
	"fmt"
	"net/http"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

func (e *env) registerCredentialRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys", e.GetDeveloperAppKeys)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/keys", e.CheckForJSONContentType, e.PostCreateDeveloperAppKey)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", e.GetDeveloperAppKeyByKey)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", e.CheckForJSONContentType, e.PostUpdateDeveloperAppKeyByKey)
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", e.DeleteDeveloperAppKeyByKey)
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

// PostCreateDeveloperAppKey creates key for developerapp
func (e *env) PostCreateDeveloperAppKey(c *gin.Context) {
	// var receivedKeypair struct {
	// 	ConsumerKey    string `json:"consumerKey"`
	// 	ConsumerSecret string `json:"ConsumerSecret"`
	// }
	// if err := c.ShouldBindJSON(&receivedKeypair); err != nil {
	// 	e.returnJSONMessage(c, http.StatusBadRequest, err)
	// 	return
	// }
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
	var newAppCredential types.AppCredential
	newAppCredential.ConsumerKey = e.GenerateCredentialConsumerKey()
	//	newAppCredential.APIProducts = []types.APIProductStatus{"status": "approved", "apiproduct", "teleporter2020"}
	newAppCredential.ConsumerSecret = e.GenerateCredentialConsumerSecret()
	newAppCredential.ExpiresAt = -1
	newAppCredential.IssuedAt = e.getCurrentTimeMilliseconds()
	newAppCredential.OrganizationAppID = developerApp.DeveloperAppID
	newAppCredential.OrganizationName = developerApp.OrganizationName
	newAppCredential.Status = "approved"

	if err := e.db.UpdateAppCredentialByKey(newAppCredential); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newAppCredential)
}

// PostUpdateDeveloperAppKeyByKey creates key for developerapp
func (e *env) PostUpdateDeveloperAppKeyByKey(c *gin.Context) {
	var receivedAppCredential types.AppCredential
	if err := c.ShouldBindJSON(&receivedAppCredential); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
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
	AppCredential.ConsumerSecret = receivedAppCredential.ConsumerSecret
	AppCredential.APIProducts = receivedAppCredential.APIProducts
	AppCredential.Attributes = receivedAppCredential.Attributes
	AppCredential.ExpiresAt = receivedAppCredential.ExpiresAt
	AppCredential.Status = receivedAppCredential.Status
	if err := e.db.UpdateAppCredentialByKey(AppCredential); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, AppCredential)
}

// DeleteDeveloperAppKeyByKey deletes apikey of developer app
func (e *env) DeleteDeveloperAppKeyByKey(c *gin.Context) {
	AppCredential, err := e.db.GetAppCredentialByKey(c.Param("key"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	if err := e.db.DeleteAppCredentialByKey(c.Param("key")); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, AppCredential)
}
