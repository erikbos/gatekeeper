package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// registerRouteRoutes registers all routes we handle
func (h *Handler) registerRouteRoutes(r *gin.RouterGroup) {
	r.GET("/routes", h.handler(h.getAllRoutes))
	r.POST("/routes", h.handler(h.createRoute))

	r.GET("/routes/:route", h.handler(h.getRoute))
	r.POST("/routes/:route", h.handler(h.updateRoute))
	r.DELETE("/routes/:route", h.handler(h.deleteRoute))

	r.GET("/routes/:route/attributes", h.handler(h.getRouteAttributes))
	r.POST("/routes/:route/attributes", h.handler(h.updateRouteAttributes))

	r.GET("/routes/:route/attributes/:attribute", h.handler(h.getRouteAttribute))
	r.POST("/routes/:route/attributes/:attribute", h.handler(h.updateRouteAttribute))
	r.DELETE("/routes/:route/attributes/:attribute", h.handler(h.deleteRouteAttribute))
}

const (
	// Name of route parameter in the route definition
	routeParameter = "route"
)

// getAllRoutes returns all routes
func (h *Handler) getAllRoutes(c *gin.Context) handlerResponse {

	routes, err := h.service.Route.GetAll()
	if err != nil {
		return handleError(err)
	}
	return handleOK(StringMap{"routes": routes})
}

// getRoute returns details of an route
func (h *Handler) getRoute(c *gin.Context) handlerResponse {

	route, err := h.service.Route.Get(c.Param(routeParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(route)
}

// getRouteAttributes returns attributes of an route
func (h *Handler) getRouteAttributes(c *gin.Context) handlerResponse {

	route, err := h.service.Route.Get(c.Param(routeParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOKAttributes(route.Attributes)
}

// getRouteAttribute returns one particular attribute of an route
func (h *Handler) getRouteAttribute(c *gin.Context) handlerResponse {

	route, err := h.service.Route.Get(c.Param(routeParameter))
	if err != nil {
		return handleError(err)
	}
	value, err := route.Attributes.Get(c.Param("attribute"))
	if err != nil {
		return handleError(err)
	}
	return handleOK(value)
}

// createRoute creates an route
func (h *Handler) createRoute(c *gin.Context) handlerResponse {

	var newRoute types.Route
	if err := c.ShouldBindJSON(&newRoute); err != nil {
		return handleBadRequest(err)
	}
	storedRoute, err := h.service.Route.Create(newRoute, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleCreated(storedRoute)
}

// updateRoute updates an existing route
func (h *Handler) updateRoute(c *gin.Context) handlerResponse {

	var updatedRoute types.Route
	if err := c.ShouldBindJSON(&updatedRoute); err != nil {
		return handleBadRequest(err)
	}
	// routename in path must match routename in posted body
	if updatedRoute.Name != c.Param(routeParameter) {
		return handleNameMismatch()
	}
	storedRoute, err := h.service.Route.Update(updatedRoute, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedRoute)
}

// updateRouteAttributes updates attributes of an route
func (h *Handler) updateRouteAttributes(c *gin.Context) handlerResponse {

	var receivedAttributes struct {
		Attributes types.Attributes `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		return handleBadRequest(err)
	}
	if err := h.service.Route.UpdateAttributes(c.Param(routeParameter), receivedAttributes.Attributes, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttributes(receivedAttributes.Attributes)
}

// updateRouteAttribute update an attribute of developer
func (h *Handler) updateRouteAttribute(c *gin.Context) handlerResponse {

	var receivedValue AttributeValue
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		return handleBadRequest(err)
	}
	newAttribute := types.Attribute{
		Name:  c.Param("attribute"),
		Value: receivedValue.Value,
	}
	if err := h.service.Route.UpdateAttribute(c.Param(routeParameter), newAttribute, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttribute(newAttribute)
}

// deleteRouteAttribute removes an attribute of an route
func (h *Handler) deleteRouteAttribute(c *gin.Context) handlerResponse {

	attributeToDelete := c.Param("attribute")
	oldValue, err := h.service.Route.DeleteAttribute(c.Param(routeParameter), attributeToDelete, h.who(c))
	if err != nil {
		return handleBadRequest(err)
	}
	return handleOKAttribute(types.Attribute{
		Name:  attributeToDelete,
		Value: oldValue,
	})
}

// deleteRoute deletes an route
func (h *Handler) deleteRoute(c *gin.Context) handlerResponse {

	deletedRoute, err := h.service.Route.Delete(c.Param(routeParameter), h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(deletedRoute)
}
