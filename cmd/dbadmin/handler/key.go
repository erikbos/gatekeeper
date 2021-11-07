package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// func (h *Handler) registerKeyRoutes(r *gin.RouterGroup) {
// 	r.GET("/developers/:developer/apps/:application/keys", h.handler(h.getDeveloperAppKeys))
// 	r.POST("/developers/:developer/apps/:application/keys", h.handler(h.createDeveloperAppKey))

// 	r.GET("/developers/:developer/apps/:application/keys/:key", h.handler(h.getDeveloperAppKeyByKey))
// 	r.POST("/developers/:developer/apps/:application/keys/:key", h.handler(h.updateDeveloperAppKeyByKey))
// 	r.DELETE("/developers/:developer/apps/:application/keys/:key", h.handler(h.deleteDeveloperAppKeyByKey))
// }

// getDeveloperAppKeys returns all keys of one particular developer application
func (h *Handler) getDeveloperAppKeys(c *gin.Context) handlerResponse {

	_, err := h.service.Developer.Get(c.Param("developer"))
	if err != nil {
		return handleError(err)
	}
	developerApp, err := h.service.DeveloperApp.GetByName(c.Param("application"))
	if err != nil {
		return handleError(err)
	}
	keys, err := h.service.Key.GetByDeveloperAppID(developerApp.AppID)
	if err != nil {
		return handleError(err)
	}
	return handleOK(StringMap{"keys": keys})
}

// getDeveloperAppKeyByKey returns one key of one particular developer application
func (h *Handler) getDeveloperAppKeyByKey(c *gin.Context) handlerResponse {

	_, err := h.service.Developer.Get(c.Param("developer"))
	if err != nil {
		return handleError(err)
	}
	_, err = h.service.DeveloperApp.GetByName(c.Param("application"))
	if err != nil {
		return handleError(err)
	}
	key, err := h.service.Key.Get(c.Param("key"))
	if err != nil {
		return handleError(err)
	}
	return handleOK(key)
}

// createDeveloperAppKey creates key for developerapp
func (h *Handler) createDeveloperAppKey(c *gin.Context) handlerResponse {

	var receivedKey types.Key
	// We ignore error as it is not required to provided any data
	_ = c.ShouldBindJSON(&receivedKey)

	developerApp, err := h.service.DeveloperApp.GetByName(c.Param("application"))
	if err != nil {
		return handleBadRequest(err)
	}
	storedKey, err := h.service.Key.Create(receivedKey, developerApp, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedKey)
}

// updateDeveloperAppKeyByKey updates existing key for developerapp
func (h *Handler) updateDeveloperAppKeyByKey(c *gin.Context) handlerResponse {

	var receivedKey types.Key
	if err := c.ShouldBindJSON(&receivedKey); err != nil {
		return handleBadRequest(err)
	}
	// apikey in path must match consumer key in posted body
	if receivedKey.ConsumerKey != c.Param("key") {
		return handleNameMismatch()
	}
	storedKey, err := h.service.Key.Update(receivedKey, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedKey)
}

// deleteDeveloperAppKeyByKey deletes apikey of developer app
func (h *Handler) deleteDeveloperAppKeyByKey(c *gin.Context) handlerResponse {

	deletedKey, err := h.service.Key.Delete(c.Param("key"), h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(deletedKey)
}
