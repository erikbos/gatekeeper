package handler

import (
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

func (h *Handler) registerCredentialRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys", h.handler(h.getDeveloperAppKeys))
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/keys", h.handler(h.createDeveloperAppKey))

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", h.handler(h.getDeveloperAppKeyByKey))
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", h.handler(h.updateDeveloperAppKeyByKey))
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", h.handler(h.deleteDeveloperAppKeyByKey))
}

const (
	// Name of organization parameter in the route definition
	keyParameter = "key"
)

// getDeveloperAppKeys returns all keys of one particular developer application
func (h *Handler) getDeveloperAppKeys(c *gin.Context) handlerResponse {

	_, err := h.service.Developer.Get(c.Param(organizationParameter), c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	developerApp, err := h.service.DeveloperApp.GetByName(c.Param(organizationParameter), c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}
	AppCredentials, err := h.service.Credential.GetByDeveloperAppID(developerApp.AppID)
	if err != nil {
		return handleError(err)
	}
	return handleOK(StringMap{"credentials": AppCredentials})
}

// getDeveloperAppKeyByKey returns one key of one particular developer application
func (h *Handler) getDeveloperAppKeyByKey(c *gin.Context) handlerResponse {

	_, err := h.service.Developer.Get(c.Param(organizationParameter), c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	_, err = h.service.DeveloperApp.GetByName(c.Param(organizationParameter), c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}
	AppCredentials, err := h.service.Credential.Get(c.Param(organizationParameter), c.Param(keyParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(AppCredentials)
}

// createDeveloperAppKey creates key for developerapp
func (h *Handler) createDeveloperAppKey(c *gin.Context) handlerResponse {

	var receivedCredential types.DeveloperAppKey
	errJSON := c.ShouldBindJSON(&receivedCredential)

	developerApp, err := h.service.DeveloperApp.GetByName(c.Param(organizationParameter), c.Param("application"))
	if err != nil {
		return handleBadRequest(err)
	}

	newAppCredential := types.DeveloperAppKey{
		ExpiresAt:        -1,
		IssuedAt:         shared.GetCurrentTimeMilliseconds(),
		AppID:            developerApp.AppID,
		OrganizationName: developerApp.OrganizationName,
		Status:           "approved",
	}
	if errJSON == nil && receivedCredential.ConsumerKey != "" {
		newAppCredential.ConsumerKey = receivedCredential.ConsumerKey
		newAppCredential.ConsumerSecret = receivedCredential.ConsumerSecret
	} else {
		newAppCredential.ConsumerKey = generateConsumerKey()
		newAppCredential.ConsumerSecret = generateConsumerSecret()
	}

	storedAppCredential, err := h.service.Credential.Update(newAppCredential)
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedAppCredential)
}

// updateDeveloperAppKeyByKey creates key for developerapp
func (h *Handler) updateDeveloperAppKeyByKey(c *gin.Context) handlerResponse {

	var receivedAppCredential types.DeveloperAppKey
	if err := c.ShouldBindJSON(&receivedAppCredential); err != nil {
		return handleBadRequest(err)
	}

	key := c.Param(keyParameter)
	organization := c.Param(organizationParameter)
	AppCredential, err := h.service.Credential.Get(organization, key)
	if err != nil {
		return handleError(err)
	}

	AppCredential.ConsumerSecret = receivedAppCredential.ConsumerSecret
	AppCredential.APIProducts = receivedAppCredential.APIProducts
	AppCredential.Attributes = receivedAppCredential.Attributes
	AppCredential.ExpiresAt = receivedAppCredential.ExpiresAt
	AppCredential.Status = receivedAppCredential.Status

	storedAppCredential, err := h.service.Credential.Update(*AppCredential)
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedAppCredential)
}

// deleteDeveloperAppKeyByKey deletes apikey of developer app
func (h *Handler) deleteDeveloperAppKeyByKey(c *gin.Context) handlerResponse {

	deletedAppCredential, err := h.service.Credential.Delete(c.Param(organizationParameter), c.Param(keyParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(deletedAppCredential)
}

// generateConsumerKey returns a random string to be used as apikey (32 character base62)
func generateConsumerKey() string {

	return uniuri.NewLen(32)
}

// generateConsumerSecret returns a random string to be used as consumer key (16 character base62)
func generateConsumerSecret() string {

	return uniuri.New()
}
