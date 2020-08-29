package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// registerListenerRoutes registers all listeners we handle
func (s *server) registerListenerRoutes(r *gin.Engine) {
	r.GET("/v1/listeners", s.GetAllListeners)
	r.POST("/v1/listeners", shared.AbortIfContentTypeNotJSON, s.PostCreateListener)

	r.GET("/v1/listeners/:listener", s.GetListenerByName)
	r.POST("/v1/listeners/:listener", shared.AbortIfContentTypeNotJSON, s.PostListener)
	r.DELETE("/v1/listeners/:listener", s.DeleteListenerByName)

	r.GET("/v1/listeners/:listener/attributes", s.GetListenerAttributes)
	r.POST("/v1/listeners/:listener/attributes", shared.AbortIfContentTypeNotJSON, s.PostListenerAttributes)

	r.GET("/v1/listeners/:listener/attributes/:attribute", s.GetListenerAttributeByName)
	r.POST("/v1/listeners/:listener/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostListenerAttributeByName)
	r.DELETE("/v1/listeners/:listener/attributes/:attribute", s.DeleteListenerAttributeByName)
}

// GetAllListeners returns all listeners
func (s *server) GetAllListeners(c *gin.Context) {

	listeners, err := s.db.Listener.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"listeners": listeners})
}

// GetListenerByName returns details of an route
func (s *server) GetListenerByName(c *gin.Context) {

	route, err := s.db.Listener.GetByName(c.Param("listener"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, route)
}

// GetListenerAttributes returns attributes of a virtual host
func (s *server) GetListenerAttributes(c *gin.Context) {

	route, err := s.db.Listener.GetByName(c.Param("listener"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": route.Attributes})
}

// GetListenerAttributeByName returns one particular attribute of a virtual host
func (s *server) GetListenerAttributeByName(c *gin.Context) {

	listener, err := s.db.Listener.GetByName(c.Param("listener"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	value, err := listener.Attributes.Get(c.Param("attribute"))
	if err != nil {
		returnCanNotFindAttribute(c, c.Param("attribute"))
		return
	}

	setLastModifiedHeader(c, listener.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, value)
}

// PostCreateListener creates a virtual host
func (s *server) PostCreateListener(c *gin.Context) {

	var newListener shared.Listener
	if err := c.ShouldBindJSON(&newListener); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	existingListener, err := s.db.Listener.GetByName(newListener.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Listener '%s' already exists", existingListener.Name))
		return
	}

	// Automatically set default fields
	newListener.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newListener.CreatedBy = s.whoAmI()
	newListener.LastmodifiedBy = s.whoAmI()

	if err := s.db.Listener.UpdateByName(&newListener); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, newListener)
}

// PostListener updates an existing virtual host
func (s *server) PostListener(c *gin.Context) {

	listenerToUpdate, err := s.db.Listener.GetByName(c.Param("listener"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	var updateRequest shared.Listener
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	// Copy over the fields we allow to be updated
	listenerToUpdate.VirtualHosts = updateRequest.VirtualHosts
	listenerToUpdate.Port = updateRequest.Port
	listenerToUpdate.DisplayName = updateRequest.DisplayName
	listenerToUpdate.Attributes = updateRequest.Attributes
	listenerToUpdate.RouteGroup = updateRequest.RouteGroup
	listenerToUpdate.Policies = updateRequest.Policies
	listenerToUpdate.OrganizationName = updateRequest.OrganizationName

	if err := s.db.Listener.UpdateByName(listenerToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, listenerToUpdate)
}

// PostListenerAttributes updates attributes of a virtual host
func (s *server) PostListenerAttributes(c *gin.Context) {

	listenerToUpdate, err := s.db.Listener.GetByName(c.Param("listener"))
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
	listenerToUpdate.Attributes = body.Attributes
	listenerToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Listener.UpdateByName(listenerToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"attribute": listenerToUpdate.Attributes})
}

// PostListenerAttributeByName update an attribute of virtual host
func (s *server) PostListenerAttributeByName(c *gin.Context) {

	listenerToUpdate, err := s.db.Listener.GetByName(c.Param("listener"))
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
	listenerToUpdate.Attributes.Set(attributeToUpdate, body.Value)
	listenerToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Listener.UpdateByName(listenerToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": body.Value})
}

// DeleteListenerAttributeByName removes an attribute of virtual host
func (s *server) DeleteListenerAttributeByName(c *gin.Context) {

	updatedListener, err := s.db.Listener.GetByName(c.Param("listener"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	attributeToDelete := c.Param("attribute")
	deleted, oldValue := updatedListener.Attributes.Delete(attributeToDelete)
	if !deleted {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedListener.LastmodifiedBy = s.whoAmI()

	if err := s.db.Listener.UpdateByName(updatedListener); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteListenerByName deletes a virtual host
func (s *server) DeleteListenerByName(c *gin.Context) {

	listener, err := s.db.Listener.GetByName(c.Param("listener"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	if err := s.db.Listener.DeleteByName(listener.Name); err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
	}
	c.IndentedJSON(http.StatusOK, listener)
}
