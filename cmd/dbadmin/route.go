package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// registerRouteRoutes registers all routes we handle
func (s *server) registerRouteRoutes(r *gin.Engine) {
	r.GET("/v1/routes", s.GetAllRoutes)
	r.POST("/v1/routes", shared.AbortIfContentTypeNotJSON, s.PostCreateRoute)

	r.GET("/v1/routes/:route", s.GetRouteByName)
	r.POST("/v1/routes/:route", shared.AbortIfContentTypeNotJSON, s.PostRoute)
	r.DELETE("/v1/routes/:route", s.DeleteRouteByName)

	r.GET("/v1/routes/:route/attributes", s.GetRouteAttributes)
	r.POST("/v1/routes/:route/attributes", shared.AbortIfContentTypeNotJSON, s.PostRouteAttributes)

	r.GET("/v1/routes/:route/attributes/:attribute", s.GetRouteAttributeByName)
	r.POST("/v1/routes/:route/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostRouteAttributeByName)
	r.DELETE("/v1/routes/:route/attributes/:attribute", s.DeleteRouteAttributeByName)
}

// GetAllRoutes returns all routes
func (s *server) GetAllRoutes(c *gin.Context) {

	routes, err := s.db.Route.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"routes": routes})
}

// GetRouteByName returns details of an route
func (s *server) GetRouteByName(c *gin.Context) {

	route, err := s.db.Route.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, route)
}

// GetRouteAttributes returns attributes of a route
func (s *server) GetRouteAttributes(c *gin.Context) {

	route, err := s.db.Route.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": route.Attributes})
}

// GetRouteAttributeByName returns one particular attribute of a route
func (s *server) GetRouteAttributeByName(c *gin.Context) {

	route, err := s.db.Route.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	value, err := route.Attributes.Get(c.Param("attribute"))
	if err != nil {
		returnCanNotFindAttribute(c, c.Param("attribute"))
		return
	}

	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, value)
}

// PostCreateRoute creates a route
func (s *server) PostCreateRoute(c *gin.Context) {

	var newRoute shared.Route
	if err := c.ShouldBindJSON(&newRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	existingRoute, err := s.db.Route.GetRouteByName(newRoute.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Route '%s' already exists", existingRoute.Name))
		return
	}

	// Automatically set default fields
	newRoute.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newRoute.CreatedBy = s.whoAmI()
	newRoute.LastmodifiedBy = s.whoAmI()

	if err := s.db.Route.UpdateRouteByName(&newRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, newRoute)
}

// PostRoute updates an existing route
func (s *server) PostRoute(c *gin.Context) {

	routeToUpdate, err := s.db.Route.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	var updateRequest shared.Route
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	// Copy over the fields we allow to be updated
	routeToUpdate.DisplayName = updateRequest.DisplayName
	routeToUpdate.RouteGroup = updateRequest.RouteGroup
	routeToUpdate.Path = updateRequest.Path
	routeToUpdate.PathType = updateRequest.PathType
	routeToUpdate.Cluster = updateRequest.Cluster
	routeToUpdate.Attributes = updateRequest.Attributes

	if err := s.db.Route.UpdateRouteByName(routeToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, routeToUpdate)
}

// PostRouteAttributes updates attributes of a route
func (s *server) PostRouteAttributes(c *gin.Context) {

	routeToUpdate, err := s.db.Route.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	var body struct {
		Attributes shared.Attributes `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	if len(body.Attributes) == 0 {
		returnJSONMessage(c, http.StatusBadRequest, errors.New("No attributes posted"))
		return
	}

	routeToUpdate.Attributes = body.Attributes
	routeToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Route.UpdateRouteByName(routeToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"attribute": routeToUpdate.Attributes})
}

// PostRouteAttributeByName update an attribute of route
func (s *server) PostRouteAttributeByName(c *gin.Context) {

	routeToUpdate, err := s.db.Route.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	var body struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	attributeToUpdate := c.Param("attribute")
	routeToUpdate.Attributes.Set(attributeToUpdate, body.Value)
	routeToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Route.UpdateRouteByName(routeToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": body.Value})
}

// DeleteRouteAttributeByName removes an attribute of route
func (s *server) DeleteRouteAttributeByName(c *gin.Context) {

	routeToUpdate, err := s.db.Route.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	attributeToDelete := c.Param("attribute")
	deleted, oldValue := routeToUpdate.Attributes.Delete(attributeToDelete)
	if !deleted {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	routeToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Route.UpdateRouteByName(routeToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteRouteByName deletes a route
func (s *server) DeleteRouteByName(c *gin.Context) {

	route, err := s.db.Route.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	if err := s.db.Route.DeleteRouteByName(route.Name); err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
	}

	c.IndentedJSON(http.StatusOK, route)
}
