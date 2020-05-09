package main

import (
	"net/http"

	"github.com/erikbos/apiauth/pkg/shared"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
)

func (e *env) registerCredentialRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys", e.GetDeveloperAppKeys)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/keys", shared.AbortIfContentTypeNotJSON, e.PostCreateDeveloperAppKey)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", e.GetDeveloperAppKeyByKey)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", shared.AbortIfContentTypeNotJSON, e.PostUpdateDeveloperAppKeyByKey)
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", e.DeleteDeveloperAppKeyByKey)
}

// GetDeveloperAppByKey returns all keys of one particular developer application
func (e *env) GetDeveloperAppKeys(c *gin.Context) {
	_, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	AppCredentials, err := e.db.GetAppCredentialByDeveloperAppID(developerApp.DeveloperAppID)
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"credentials": AppCredentials})
}

// GetDeveloperAppByKey returns one key of one particular developer application
func (e *env) GetDeveloperAppKeyByKey(c *gin.Context) {
	_, err := e.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	_, err = e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	AppCredential, err := e.db.GetAppCredentialByKey(c.Param("organization"), c.Param("key"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
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
	// 	returnJSONMessage(c, http.StatusBadRequest, err)
	// 	return
	// }
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	var newAppCredential shared.AppCredential
	newAppCredential.ConsumerKey = generateCredentialConsumerKey()
	//	newAppCredential.APIProducts = []shared.APIProductStatus{"status": "approved", "apiproduct", "teleporter2020"}
	newAppCredential.ConsumerSecret = generateCredentialConsumerSecret()
	newAppCredential.ExpiresAt = -1
	newAppCredential.IssuedAt = shared.GetCurrentTimeMilliseconds()
	newAppCredential.OrganizationAppID = developerApp.DeveloperAppID
	newAppCredential.OrganizationName = developerApp.OrganizationName
	newAppCredential.Status = "approved"

	if err := e.db.UpdateAppCredentialByKey(&newAppCredential); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newAppCredential)
}

// PostUpdateDeveloperAppKeyByKey creates key for developerapp
func (e *env) PostUpdateDeveloperAppKeyByKey(c *gin.Context) {
	var receivedAppCredential shared.AppCredential
	if err := c.ShouldBindJSON(&receivedAppCredential); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	AppCredential, err := e.db.GetAppCredentialByKey(c.Param("organization"), c.Param("key"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	AppCredential.ConsumerSecret = receivedAppCredential.ConsumerSecret
	AppCredential.APIProducts = receivedAppCredential.APIProducts
	AppCredential.Attributes = receivedAppCredential.Attributes
	AppCredential.ExpiresAt = receivedAppCredential.ExpiresAt
	AppCredential.Status = receivedAppCredential.Status
	if err := e.db.UpdateAppCredentialByKey(&AppCredential); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, AppCredential)
}

// DeleteDeveloperAppKeyByKey deletes apikey of developer app
func (e *env) DeleteDeveloperAppKeyByKey(c *gin.Context) {
	AppCredential, err := e.db.GetAppCredentialByKey(c.Param("organization"), c.Param("key"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	if err := e.db.DeleteAppCredentialByKey(c.Param("organization"), c.Param("key")); err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	c.IndentedJSON(http.StatusOK, AppCredential)
}

// GenerateCredentialConsumerKey returns a random string to be used as apikey (32 character base62)
func generateCredentialConsumerKey() string {
	return uniuri.NewLen(32)
}

// GenerateCredentialConsumerSecret returns a random string to be used as consumer key (16 character base62)
func generateCredentialConsumerSecret() string {
	return uniuri.New()
}
