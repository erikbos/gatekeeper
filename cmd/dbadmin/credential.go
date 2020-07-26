package main

import (
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

func (s *server) registerCredentialRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys", s.GetDeveloperAppKeys)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/keys", shared.AbortIfContentTypeNotJSON, s.PostCreateDeveloperAppKey)

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", s.GetDeveloperAppKeyByKey)
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", shared.AbortIfContentTypeNotJSON, s.PostUpdateDeveloperAppKeyByKey)
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", s.DeleteDeveloperAppKeyByKey)
}

// GetDeveloperAppByKey returns all keys of one particular developer application
func (s *server) GetDeveloperAppKeys(c *gin.Context) {

	_, err := s.db.Developer.GetByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	developerApp, err := s.db.DeveloperApp.GetByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	AppCredentials, err := s.db.Credential.GetByDeveloperAppID(developerApp.AppID)
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"credentials": AppCredentials})
}

// GetDeveloperAppByKey returns one key of one particular developer application
func (s *server) GetDeveloperAppKeyByKey(c *gin.Context) {

	_, err := s.db.Developer.GetByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	_, err = s.db.DeveloperApp.GetByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	key := c.Param("key")
	organization := c.Param("organization")
	AppCredential, err := s.db.Credential.GetByKey(&organization, &key)
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	c.IndentedJSON(http.StatusOK, AppCredential)
}

// PostCreateDeveloperAppKey creates key for developerapp
func (s *server) PostCreateDeveloperAppKey(c *gin.Context) {

	developerApp, err := s.db.DeveloperApp.GetByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	newAppCredential := shared.DeveloperAppKey{
		ExpiresAt:        -1,
		IssuedAt:         shared.GetCurrentTimeMilliseconds(),
		AppID:            developerApp.AppID,
		OrganizationName: developerApp.OrganizationName,
		Status:           "approved",
	}

	var receivedCredential shared.DeveloperAppKey
	err = c.ShouldBindJSON(&receivedCredential)

	if err == nil && receivedCredential.ConsumerKey != "" {
		newAppCredential.ConsumerKey = receivedCredential.ConsumerKey
		newAppCredential.ConsumerSecret = receivedCredential.ConsumerSecret
	} else {
		newAppCredential.ConsumerKey = generateCredentialConsumerKey()
		newAppCredential.ConsumerSecret = generateCredentialConsumerSecret()
	}

	if err := s.db.Credential.UpdateByKey(&newAppCredential); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newAppCredential)
}

// PostUpdateDeveloperAppKeyByKey creates key for developerapp
func (s *server) PostUpdateDeveloperAppKeyByKey(c *gin.Context) {

	var receivedAppCredential shared.DeveloperAppKey
	if err := c.ShouldBindJSON(&receivedAppCredential); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	key := c.Param("key")
	organization := c.Param("organization")
	AppCredential, err := s.db.Credential.GetByKey(&organization, &key)
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	AppCredential.ConsumerSecret = receivedAppCredential.ConsumerSecret
	AppCredential.APIProducts = receivedAppCredential.APIProducts
	AppCredential.Attributes = receivedAppCredential.Attributes
	AppCredential.ExpiresAt = receivedAppCredential.ExpiresAt
	AppCredential.Status = receivedAppCredential.Status

	if err := s.db.Credential.UpdateByKey(AppCredential); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, AppCredential)
}

// DeleteDeveloperAppKeyByKey deletes apikey of developer app
func (s *server) DeleteDeveloperAppKeyByKey(c *gin.Context) {

	key := c.Param("key")
	organization := c.Param("organization")
	AppCredential, err := s.db.Credential.GetByKey(&organization, &key)
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	if err := s.db.Credential.DeleteByKey(c.Param("organization"), key); err != nil {
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
