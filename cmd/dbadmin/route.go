package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/apiauth/pkg/shared"
)

// registerRouteRoutes registers all routes we handle
func (s *server) registerRouteRoutes(r *gin.Engine) {
	r.GET("/v1/routes", s.GetRoutes)
	r.POST("/v1/routes", shared.AbortIfContentTypeNotJSON, s.PostCreateRoute)

	r.GET("/v1/routes/:route", s.GetRouteByName)
	r.POST("/v1/routes/:route", shared.AbortIfContentTypeNotJSON, s.PostRoute)
	r.DELETE("/v1/routes/:route", s.DeleteRouteByName)

	r.GET("/v1/routes/:route/attributes", s.GetRouteAttributes)
	r.POST("/v1/routes/:route/attributes", shared.AbortIfContentTypeNotJSON, s.PostRouteAttributes)
	r.DELETE("/v1/routes/:route/attributes", s.DeleteRouteAttributes)

	r.GET("/v1/routes/:route/attributes/:attribute", s.GetRouteAttributeByName)
	r.POST("/v1/routes/:route/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostRouteAttributeByName)
	r.DELETE("/v1/routes/:route/attributes/:attribute", s.DeleteRouteAttributeByName)
}

// GetRoutes returns all routes
func (s *server) GetRoutes(c *gin.Context) {
	routes, err := s.db.GetRoutes()
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"routes": routes})
}

// GetRouteByName returns details of an route
func (s *server) GetRouteByName(c *gin.Context) {
	route, err := s.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, route)
}

// GetRouteAttributes returns attributes of a route
func (s *server) GetRouteAttributes(c *gin.Context) {
	route, err := s.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": route.Attributes})
}

// GetRouteAttributeByName returns one particular attribute of a route
func (s *server) GetRouteAttributeByName(c *gin.Context) {
	route, err := s.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(route.Attributes); i++ {
		if route.Attributes[i].Name == c.Param("attribute") {
			setLastModifiedHeader(c, route.LastmodifiedAt)
			c.IndentedJSON(http.StatusOK, route.Attributes[i])
			return
		}
	}
	returnJSONMessage(c, http.StatusNotFound,
		fmt.Errorf("Could not retrieve attribute '%s'", c.Param("attribute")))
}

// PostCreateRoute creates a route
func (s *server) PostCreateRoute(c *gin.Context) {
	var newRoute shared.Route
	if err := c.ShouldBindJSON(&newRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	existingRoute, err := s.db.GetRouteByName(newRoute.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Route '%s' already exists", existingRoute.Name))
		return
	}
	// Automatically set default fields
	newRoute.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newRoute.CreatedBy = s.whoAmI()
	newRoute.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateRouteByName(&newRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newRoute)
}

// PostRoute updates an existing route
func (s *server) PostRoute(c *gin.Context) {
	currentRoute, err := s.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedRoute shared.Route
	if err := c.ShouldBindJSON(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update route X while body says to update route Y
	updatedRoute.Name = currentRoute.Name
	updatedRoute.CreatedBy = currentRoute.CreatedBy
	updatedRoute.CreatedAt = currentRoute.CreatedAt
	updatedRoute.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateRouteByName(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedRoute)
}

// PostRouteAttributes updates attributes of a route
func (s *server) PostRouteAttributes(c *gin.Context) {
	updatedRoute, err := s.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	var receivedAttributes struct {
		Attributes []shared.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	updatedRoute.Attributes = receivedAttributes.Attributes
	updatedRoute.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateRouteByName(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedRoute.Attributes})
}

// DeleteRouteAttributes delete attributes of APIProduct
func (s *server) DeleteRouteAttributes(c *gin.Context) {
	updatedRoute, err := s.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	deletedAttributes := updatedRoute.Attributes
	updatedRoute.Attributes = nil
	updatedRoute.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateRouteByName(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttributes)
}

// PostRouteAttributeByName update an attribute of APIProduct
func (s *server) PostRouteAttributeByName(c *gin.Context) {
	updatedRoute, err := s.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	attributeToUpdate := c.Param("attribute")
	var receivedValue struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// Find & update existing attribute in array
	attributeToUpdateIndex := shared.FindIndexOfAttribute(
		updatedRoute.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find existing attribute, append new attribute
		updatedRoute.Attributes = append(updatedRoute.Attributes,
			shared.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedRoute.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedRoute.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateRouteByName(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteRouteAttributeByName removes an attribute of route
func (s *server) DeleteRouteAttributeByName(c *gin.Context) {
	updatedRoute, err := s.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	attributeToDelete := c.Param("attribute")
	updatedAttributes, index, oldValue := shared.DeleteAttribute(updatedRoute.Attributes, attributeToDelete)
	if index == -1 {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedRoute.Attributes = updatedAttributes
	updatedRoute.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateRouteByName(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteRouteByName deletes a route
func (s *server) DeleteRouteByName(c *gin.Context) {
	route, err := s.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	if err := s.db.DeleteRouteByName(route.Name); err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
	}
	c.IndentedJSON(http.StatusOK, route)
}
