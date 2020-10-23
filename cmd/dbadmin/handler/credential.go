package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

func (h *Handler) registerCredentialRoutes(r *gin.RouterGroup) {
	r.GET("/organizations/:organization/developers/:developer/apps/:application/keys", h.handler(h.getDeveloperAppKeys))
	r.POST("/organizations/:organization/developers/:developer/apps/:application/keys", h.handler(h.createDeveloperAppKey))

	r.GET("/organizations/:organization/developers/:developer/apps/:application/keys/:key", h.handler(h.getDeveloperAppKeyByKey))
	r.POST("/organizations/:organization/developers/:developer/apps/:application/keys/:key", h.handler(h.updateDeveloperAppKeyByKey))
	r.DELETE("/organizations/:organization/developers/:developer/apps/:application/keys/:key", h.handler(h.deleteDeveloperAppKeyByKey))
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
	// We ignore error as it is not required to provided any data
	_ = c.ShouldBindJSON(&receivedCredential)

	developerApp, err := h.service.DeveloperApp.GetByName(c.Param(organizationParameter), c.Param("application"))
	if err != nil {
		return handleBadRequest(err)
	}
	storedAppCredential, err := h.service.Credential.Create(receivedCredential, developerApp, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedAppCredential)
}

// updateDeveloperAppKeyByKey updates existing key for developerapp
func (h *Handler) updateDeveloperAppKeyByKey(c *gin.Context) handlerResponse {

	var receivedAppCredential types.DeveloperAppKey
	if err := c.ShouldBindJSON(&receivedAppCredential); err != nil {
		return handleBadRequest(err)
	}
	// apikey in path must match consumer key in posted body
	if receivedAppCredential.ConsumerKey != c.Param(keyParameter) {
		return handleNameMismatch()
	}
	storedAppCredential, err := h.service.Credential.Update(receivedAppCredential, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedAppCredential)
}

// deleteDeveloperAppKeyByKey deletes apikey of developer app
func (h *Handler) deleteDeveloperAppKeyByKey(c *gin.Context) handlerResponse {

	deletedAppCredential, err := h.service.Credential.Delete(c.Param(organizationParameter),
		c.Param(keyParameter), h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(deletedAppCredential)
}
