package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// registerListenerRoutes registers all routes we handle
func (h *Handler) registerListenerRoutes(r *gin.RouterGroup) {
	r.GET("/listeners", h.handler(h.getAllListeners))
	r.POST("/listeners", h.handler(h.createListener))

	r.GET("/listeners/:listener", h.handler(h.getListener))
	r.POST("/listeners/:listener", h.handler(h.updateListener))
	r.DELETE("/listeners/:listener", h.handler(h.deleteListener))

	r.GET("/listeners/:listener/attributes", h.handler(h.getListenerAttributes))
	r.POST("/listeners/:listener/attributes", h.handler(h.updateListenerAttributes))

	r.GET("/listeners/:listener/attributes/:attribute", h.handler(h.getListenerAttribute))
	r.POST("/listeners/:listener/attributes/:attribute", h.handler(h.updateListenerAttribute))
	r.DELETE("/listeners/:listener/attributes/:attribute", h.handler(h.deleteListenerAttribute))
}

const (
	// "attribute" = "attribute"

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
	value, err := listener.Attributes.Get(c.Param("attribute"))
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
	storedListener, err := h.service.Listener.Create(newListener, h.who(c))
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
	// listername in path must match listername in posted body
	if updatedListener.Name != c.Param(listenerParameter) {
		return handleNameMismatch()
	}
	storedListener, err := h.service.Listener.Update(updatedListener, h.who(c))
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
	if err := h.service.Listener.UpdateAttributes(c.Param(listenerParameter),
		receivedAttributes.Attributes, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttributes(receivedAttributes.Attributes)
}

// updateListenerAttribute update an attribute of developer
func (h *Handler) updateListenerAttribute(c *gin.Context) handlerResponse {

	var receivedValue AttributeValue
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		return handleBadRequest(err)
	}
	newAttribute := types.Attribute{
		Name:  c.Param("attribute"),
		Value: receivedValue.Value,
	}
	if err := h.service.Listener.UpdateAttribute(c.Param(listenerParameter),
		newAttribute, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttribute(newAttribute)
}

// deleteListenerAttribute removes an attribute of an listener
func (h *Handler) deleteListenerAttribute(c *gin.Context) handlerResponse {

	attributeToDelete := c.Param("attribute")
	oldValue, err := h.service.Listener.DeleteAttribute(c.Param(listenerParameter),
		attributeToDelete, h.who(c))
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

	deletedListener, err := h.service.Listener.Delete(c.Param(listenerParameter), h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(deletedListener)
}
