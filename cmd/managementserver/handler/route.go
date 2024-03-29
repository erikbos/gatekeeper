package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// returns all routes
// (GET /v1/routes)
func (h *Handler) GetV1Routes(c *gin.Context) {

	routes, err := h.service.Route.GetAll()
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseRoutes(c, routes)
}

// creates a new route
// (POST /v1/routes)
func (h *Handler) PostV1Routes(c *gin.Context) {

	var receivedRoute Route
	if err := c.ShouldBindJSON(&receivedRoute); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newRoute := fromRoute(receivedRoute)
	createdDeveloper, err := h.service.Route.Create(newRoute, h.who(c))
	if err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	h.responseRouteCreated(c, createdDeveloper)
}

// deletes an route
// (DELETE /v1/routes/{route_name})
func (h *Handler) DeleteV1RoutesRouteName(c *gin.Context, routeName RouteName) {

	route, err := h.service.Route.Get(string(routeName))
	if err != nil {
		responseError(c, err)
		return
	}
	if err := h.service.Route.Delete(string(routeName), h.who(c)); err != nil {
		responseError(c, err)
		return
	}
	h.responseRoute(c, route)
}

// returns full details of one route
// (GET /v1/routes/{route_name})
func (h *Handler) GetV1RoutesRouteName(c *gin.Context, routeName RouteName) {

	route, err := h.service.Route.Get(string(routeName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseRoute(c, route)
}

// (POST /v1/routes/{route_name})
func (h *Handler) PostV1RoutesRouteName(c *gin.Context, routeName RouteName) {

	var receivedRoute Route
	if err := c.ShouldBindJSON(&receivedRoute); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	if receivedRoute.Name != "" && receivedRoute.Name != string(routeName) {
		responseErrorNameValueMisMatch(c)
		return
	}
	updatedRoute := fromRoute(receivedRoute)
	storedRoute, err := h.service.Route.Update(updatedRoute, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseRoutesUpdated(c, storedRoute)
}

// returns attributes of a route
// (GET /v1/routes/{route_name}/attributes)
func (h *Handler) GetV1RoutesRouteNameAttributes(c *gin.Context, routeName RouteName) {

	route, err := h.service.Route.Get(string(routeName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, route.Attributes)
}

// replaces attributes of an route
// (POST /v1/routes/{route_name}/attributes)
func (h *Handler) PostV1RoutesRouteNameAttributes(c *gin.Context, routeName RouteName) {

	var receivedAttributes Attributes
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	route, err := h.service.Route.Get(string(routeName))
	if err != nil {
		responseError(c, err)
		return
	}
	route.Attributes = fromAttributesRequest(receivedAttributes.Attribute)
	storedRoute, err := h.service.Route.Update(*route, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, storedRoute.Attributes)
}

// deletes one attribute of an route
// (DELETE /v1/routes/{route_name}/attributes/{attribute_name})
func (h *Handler) DeleteV1RoutesRouteNameAttributesAttributeName(
	c *gin.Context, routeName RouteName, attributeName AttributeName) {

	route, err := h.service.Route.Get(string(routeName))
	if err != nil {
		responseError(c, err)
		return
	}
	oldValue, err := route.Attributes.Delete(string(attributeName))
	if err != nil {
		responseError(c, err)
	}
	_, err = h.service.Route.Update(*route, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeDeleted(c, types.NewAttribute(string(attributeName), oldValue))
}

// returns one attribute of an route
// (GET /v1/routes/{route_name}/attributes/{attribute_name})
func (h *Handler) GetV1RoutesRouteNameAttributesAttributeName(
	c *gin.Context, routeName RouteName, attributeName AttributeName) {

	route, err := h.service.Route.Get(string(routeName))
	if err != nil {
		responseError(c, err)
		return
	}
	attributeValue, err := route.Attributes.Get(string(attributeName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeRetrieved(c, types.NewAttribute(string(attributeName), attributeValue))
}

// updates an attribute of an route
// (POST /v1/routes/{route_name}/attributes/{attribute_name})
func (h *Handler) PostV1RoutesRouteNameAttributesAttributeName(
	c *gin.Context, routeName RouteName, attributeName AttributeName) {

	var receivedValue Attribute
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	route, err := h.service.Route.Get(string(routeName))
	if err != nil {
		responseError(c, err)
		return
	}
	newAttribute := types.NewAttribute(string(attributeName), *receivedValue.Value)
	if err := route.Attributes.Set(newAttribute); err != nil {
		responseError(c, err)
	}
	_, err = h.service.Route.Update(*route, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeUpdated(c, newAttribute)
}

// API responses

func (h *Handler) responseRoutes(c *gin.Context, routes types.Routes) {

	allRoutes := make([]Route, len(routes))
	for i := range routes {
		allRoutes[i] = h.ToRouteResponse(&routes[i])
	}
	c.IndentedJSON(http.StatusOK, Routes{
		Route: &allRoutes,
	})
}

func (h *Handler) responseRoute(c *gin.Context, route *types.Route) {

	c.IndentedJSON(http.StatusOK, h.ToRouteResponse(route))
}

func (h *Handler) responseRouteCreated(c *gin.Context, route *types.Route) {

	c.IndentedJSON(http.StatusCreated, h.ToRouteResponse(route))
}

func (h *Handler) responseRoutesUpdated(c *gin.Context, route *types.Route) {

	c.IndentedJSON(http.StatusOK, h.ToRouteResponse(route))
}

// type conversion

func (h *Handler) ToRouteResponse(r *types.Route) Route {

	route := Route{
		Attributes:     toAttributesResponse(r.Attributes),
		CreatedAt:      &r.CreatedAt,
		CreatedBy:      &r.CreatedBy,
		DisplayName:    &r.DisplayName,
		LastModifiedBy: &r.LastModifiedBy,
		LastModifiedAt: &r.LastModifiedAt,
		Name:           r.Name,
		PathType:       &r.PathType,
		Path:           &r.Path,
		RouteGroup:     &r.RouteGroup,
	}
	return route
}

func fromRoute(r Route) types.Route {

	route := types.Route{}
	if r.Attributes != nil {
		route.Attributes = fromAttributesRequest(r.Attributes)
	}
	if r.CreatedAt != nil {
		route.CreatedAt = *r.CreatedAt
	}
	if r.CreatedBy != nil {
		route.CreatedBy = *r.CreatedBy
	}
	if r.DisplayName != nil {
		route.DisplayName = *r.DisplayName
	}
	if r.LastModifiedBy != nil {
		route.LastModifiedBy = *r.LastModifiedBy
	}
	if r.LastModifiedAt != nil {
		route.LastModifiedAt = *r.LastModifiedAt
	}
	if r.Name != "" {
		route.Name = r.Name
	}
	if r.PathType != nil {
		route.PathType = *r.PathType
	}
	if r.Path != nil {
		route.Path = *r.Path
	}
	if r.RouteGroup != nil {
		route.RouteGroup = *r.RouteGroup
	}
	return route
}
