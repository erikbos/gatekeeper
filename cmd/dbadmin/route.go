package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/apiauth/pkg/shared"
)

// registerRouteRoutes registers all routes we handle
func (e *env) registerRouteRoutes(r *gin.Engine) {
	r.GET("/v1/routes", e.GetRoutes)
	r.POST("/v1/routes", shared.AbortIfContentTypeNotJSON, e.PostCreateRoute)

	r.GET("/v1/routes/:route", e.GetRouteByName)
	r.POST("/v1/routes/:route", shared.AbortIfContentTypeNotJSON, e.PostRoute)
	r.DELETE("/v1/routes/:route", e.DeleteRouteByName)

	r.GET("/v1/routes/:route/attributes", e.GetRouteAttributes)
	r.POST("/v1/routes/:route/attributes", shared.AbortIfContentTypeNotJSON, e.PostRouteAttributes)
	r.DELETE("/v1/routes/:route/attributes", e.DeleteRouteAttributes)

	r.GET("/v1/routes/:route/attributes/:attribute", e.GetRouteAttributeByName)
	r.POST("/v1/routes/:route/attributes/:attribute", shared.AbortIfContentTypeNotJSON, e.PostRouteAttributeByName)
	r.DELETE("/v1/routes/:route/attributes/:attribute", e.DeleteRouteAttributeByName)
}

// GetRoutes returns all routes
func (e *env) GetRoutes(c *gin.Context) {
	routes, err := e.db.GetRoutes()
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"routes": routes})
}

// GetRouteByName returns details of an route
func (e *env) GetRouteByName(c *gin.Context) {
	route, err := e.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, route)
}

// GetRouteAttributes returns attributes of a route
func (e *env) GetRouteAttributes(c *gin.Context) {
	route, err := e.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": route.Attributes})
}

// GetRouteAttributeByName returns one particular attribute of a route
func (e *env) GetRouteAttributeByName(c *gin.Context) {
	route, err := e.db.GetRouteByName(c.Param("route"))
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
func (e *env) PostCreateRoute(c *gin.Context) {
	var newRoute shared.Route
	if err := c.ShouldBindJSON(&newRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	existingRoute, err := e.db.GetRouteByName(newRoute.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Route '%s' already exists", existingRoute.Name))
		return
	}
	// Automatically set default fields
	newRoute.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newRoute.CreatedBy = e.whoAmI()
	newRoute.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateRouteByName(&newRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newRoute)
}

// PostRoute updates an existing route
func (e *env) PostRoute(c *gin.Context) {
	currentRoute, err := e.db.GetRouteByName(c.Param("route"))
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
	updatedRoute.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateRouteByName(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedRoute)
}

// PostRouteAttributes updates attributes of a route
func (e *env) PostRouteAttributes(c *gin.Context) {
	updatedRoute, err := e.db.GetRouteByName(c.Param("route"))
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
	updatedRoute.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateRouteByName(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedRoute.Attributes})
}

// DeleteRouteAttributes delete attributes of APIProduct
func (e *env) DeleteRouteAttributes(c *gin.Context) {
	updatedRoute, err := e.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	deletedAttributes := updatedRoute.Attributes
	updatedRoute.Attributes = nil
	updatedRoute.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateRouteByName(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttributes)
}

// PostRouteAttributeByName update an attribute of APIProduct
func (e *env) PostRouteAttributeByName(c *gin.Context) {
	updatedRoute, err := e.db.GetRouteByName(c.Param("route"))
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
	updatedRoute.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateRouteByName(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteRouteAttributeByName removes an attribute of route
func (e *env) DeleteRouteAttributeByName(c *gin.Context) {
	updatedRoute, err := e.db.GetRouteByName(c.Param("route"))
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
	updatedRoute.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateRouteByName(&updatedRoute); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteRouteByName deletes a route
func (e *env) DeleteRouteByName(c *gin.Context) {
	route, err := e.db.GetRouteByName(c.Param("route"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	if err := e.db.DeleteRouteByName(route.Name); err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
	}
	c.IndentedJSON(http.StatusOK, route)
}
