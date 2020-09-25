package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// registerListenerRoutes registers all routes we handle
func (h *Handler) registerListenerRoutes(r *gin.Engine) {
	r.GET("/v1/listeners", h.handler(h.getAllListeners))
	r.POST("/v1/listeners", h.handler(h.createListener))

	r.GET("/v1/listeners/:listener", h.handler(h.getListener))
	r.POST("/v1/listeners/:listener", h.handler(h.updateListener))
	r.DELETE("/v1/listeners/:listener", h.handler(h.deleteListener))

	r.GET("/v1/listeners/:listener/attributes", h.handler(h.getListenerAttributes))
	r.POST("/v1/listeners/:listener/attributes", h.handler(h.updateListenerAttributes))

	r.GET("/v1/listeners/:listener/attributes/:attribute", h.handler(h.getListenerAttribute))
	r.POST("/v1/listeners/:listener/attributes/:attribute", h.handler(h.updateListenerAttribute))
	r.DELETE("/v1/listeners/:listener/attributes/:attribute", h.handler(h.deleteListenerAttribute))
}

const (
	// Name of listener parameter in the route definition
	listenerParameter = "listener"
)

// getAllListeners returns all listeners
func (h *Handler) getAllListeners(c *gin.Context) handlerResponse {

	listeners, err := h.service.Listener.GetAll()
	if err != nil {
		return handleError(err)
	}
	return handleOK(StringMap{"listeners": listeners})
}

// getListener returns details of an listener
func (h *Handler) getListener(c *gin.Context) handlerResponse {

	listener, err := h.service.Listener.Get(c.Param(listenerParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(listener)
}

// getListenerAttributes returns attributes of an listener
func (h *Handler) getListenerAttributes(c *gin.Context) handlerResponse {

	listener, err := h.service.Listener.Get(c.Param(listenerParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOKAttributes(listener.Attributes)
}

// getListenerAttribute returns one particular attribute of an listener
func (h *Handler) getListenerAttribute(c *gin.Context) handlerResponse {

	listener, err := h.service.Listener.Get(c.Param(listenerParameter))
	if err != nil {
		return handleError(err)
	}
	value, err := listener.Attributes.Get(c.Param(attributeParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(value)
}

// createListener creates an listener
func (h *Handler) createListener(c *gin.Context) handlerResponse {

	var newListener types.Listener
	if err := c.ShouldBindJSON(&newListener); err != nil {
		return handleBadRequest(err)
	}

	// Automatically set default fields
	newListener.CreatedBy = h.GetSessionUser(c)
	newListener.LastmodifiedBy = h.GetSessionUser(c)

	storedListener, err := h.service.Listener.Create(newListener)
	if err != nil {
		return handleError(err)
	}
	return handleCreated(storedListener)
}

// updateListener updates an existing listener
func (h *Handler) updateListener(c *gin.Context) handlerResponse {

	var updatedListener types.Listener
	if err := c.ShouldBindJSON(&updatedListener); err != nil {
		return handleBadRequest(err)
	}

	// Automatically set default fields
	updatedListener.LastmodifiedBy = h.GetSessionUser(c)

	storedListener, err := h.service.Listener.Update(updatedListener)
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedListener)
}

// updateListenerAttributes updates attributes of an listener
func (h *Handler) updateListenerAttributes(c *gin.Context) handlerResponse {

	var receivedAttributes struct {
		Attributes types.Attributes `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		return handleBadRequest(err)
	}
	// FIXME we should set LastmodifiedBy
	// updatedListener.Attributes = receivedAttributes.Attributes
	// updatedListener.LastmodifiedBy = h.GetSessionUser(c)

	if err := h.service.Listener.UpdateAttributes(c.Param(listenerParameter), receivedAttributes.Attributes); err != nil {
		return handleError(err)
	}
	return handleOKAttributes(receivedAttributes.Attributes)
}

// updateListenerAttribute update an attribute of developer
func (h *Handler) updateListenerAttribute(c *gin.Context) handlerResponse {

	var receivedValue types.AttributeValue
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		return handleBadRequest(err)
	}

	newAttribute := types.Attribute{
		Name:  c.Param(attributeParameter),
		Value: receivedValue.Value,
	}

	// FIXME we should set LastmodifiedBy
	if err := h.service.Listener.UpdateAttribute(c.Param(listenerParameter), newAttribute); err != nil {
		return handleError(err)
	}
	return handleOKAttribute(newAttribute)
}

// deleteListenerAttribute removes an attribute of an listener
func (h *Handler) deleteListenerAttribute(c *gin.Context) handlerResponse {

	attributeToDelete := c.Param(attributeParameter)
	oldValue, err := h.service.Listener.DeleteAttribute(c.Param(listenerParameter), attributeToDelete)
	if err != nil {
		return handleBadRequest(err)
	}
	return handleOKAttribute(types.Attribute{
		Name:  attributeToDelete,
		Value: oldValue,
	})
}

// deleteListener deletes an listener
func (h *Handler) deleteListener(c *gin.Context) handlerResponse {

	deletedListener, err := h.service.Listener.Delete(c.Param(listenerParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(deletedListener)
}
