package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/apiauth/pkg/shared"
)

// registerVirtualHostRoutes registers all virtualhosts we handle
func (s *server) registerVirtualHostRoutes(r *gin.Engine) {
	r.GET("/v1/virtualhosts", s.GetVirtualHosts)
	r.POST("/v1/virtualhosts", shared.AbortIfContentTypeNotJSON, s.PostCreateVirtualHost)

	r.GET("/v1/virtualhosts/:virtualhost", s.GetVirtualHostByName)
	r.POST("/v1/virtualhosts/:virtualhost", shared.AbortIfContentTypeNotJSON, s.PostVirtualHost)
	r.DELETE("/v1/virtualhosts/:virtualhost", s.DeleteVirtualHostByName)

	r.GET("/v1/virtualhosts/:virtualhost/attributes", s.GetVirtualHostAttributes)
	r.POST("/v1/virtualhosts/:virtualhost/attributes", shared.AbortIfContentTypeNotJSON, s.PostVirtualHostAttributes)
	r.DELETE("/v1/virtualhosts/:virtualhost/attributes", s.DeleteVirtualHostAttributes)

	r.GET("/v1/virtualhosts/:virtualhost/attributes/:attribute", s.GetVirtualHostAttributeByName)
	r.POST("/v1/virtualhosts/:virtualhost/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostVirtualHostAttributeByName)
	r.DELETE("/v1/virtualhosts/:virtualhost/attributes/:attribute", s.DeleteVirtualHostAttributeByName)
}

// GetVirtualHosts returns all virtualhosts
func (s *server) GetVirtualHosts(c *gin.Context) {
	virtualhosts, err := s.db.GetVirtualHosts()
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"virtualhosts": virtualhosts})
}

// GetVirtualHostByName returns details of an route
func (s *server) GetVirtualHostByName(c *gin.Context) {
	route, err := s.db.GetVirtualHostByName(c.Param("virtualhost"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, route)
}

// GetVirtualHostAttributes returns attributes of a route
func (s *server) GetVirtualHostAttributes(c *gin.Context) {
	route, err := s.db.GetVirtualHostByName(c.Param("virtualhost"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, route.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": route.Attributes})
}

// GetVirtualHostAttributeByName returns one particular attribute of a route
func (s *server) GetVirtualHostAttributeByName(c *gin.Context) {
	route, err := s.db.GetVirtualHostByName(c.Param("virtualhost"))
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

// PostCreateVirtualHost creates a route
func (s *server) PostCreateVirtualHost(c *gin.Context) {
	var newVirtualHost shared.VirtualHost
	if err := c.ShouldBindJSON(&newVirtualHost); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	existingVirtualHost, err := s.db.GetVirtualHostByName(newVirtualHost.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("VirtualHost '%s' already exists", existingVirtualHost.Name))
		return
	}
	// Automatically set default fields
	newVirtualHost.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newVirtualHost.CreatedBy = s.whoAmI()
	newVirtualHost.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateVirtualHostByName(&newVirtualHost); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newVirtualHost)
}

// PostVirtualHost updates an existing route
func (s *server) PostVirtualHost(c *gin.Context) {
	currentVirtualHost, err := s.db.GetVirtualHostByName(c.Param("virtualhost"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedVirtualHost shared.VirtualHost
	if err := c.ShouldBindJSON(&updatedVirtualHost); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update route X while body says to update route Y
	updatedVirtualHost.Name = currentVirtualHost.Name
	updatedVirtualHost.CreatedBy = currentVirtualHost.CreatedBy
	updatedVirtualHost.CreatedAt = currentVirtualHost.CreatedAt
	updatedVirtualHost.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateVirtualHostByName(&updatedVirtualHost); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedVirtualHost)
}

// PostVirtualHostAttributes updates attributes of a route
func (s *server) PostVirtualHostAttributes(c *gin.Context) {
	updatedVirtualHost, err := s.db.GetVirtualHostByName(c.Param("virtualhost"))
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
	updatedVirtualHost.Attributes = receivedAttributes.Attributes
	updatedVirtualHost.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateVirtualHostByName(&updatedVirtualHost); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedVirtualHost.Attributes})
}

// DeleteVirtualHostAttributes delete attributes of APIProduct
func (s *server) DeleteVirtualHostAttributes(c *gin.Context) {
	updatedVirtualHost, err := s.db.GetVirtualHostByName(c.Param("virtualhost"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	deletedAttributes := updatedVirtualHost.Attributes
	updatedVirtualHost.Attributes = nil
	updatedVirtualHost.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateVirtualHostByName(&updatedVirtualHost); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttributes)
}

// PostVirtualHostAttributeByName update an attribute of APIProduct
func (s *server) PostVirtualHostAttributeByName(c *gin.Context) {
	updatedVirtualHost, err := s.db.GetVirtualHostByName(c.Param("virtualhost"))
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
		updatedVirtualHost.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find existing attribute, append new attribute
		updatedVirtualHost.Attributes = append(updatedVirtualHost.Attributes,
			shared.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedVirtualHost.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedVirtualHost.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateVirtualHostByName(&updatedVirtualHost); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteVirtualHostAttributeByName removes an attribute of route
func (s *server) DeleteVirtualHostAttributeByName(c *gin.Context) {
	updatedVirtualHost, err := s.db.GetVirtualHostByName(c.Param("virtualhost"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	attributeToDelete := c.Param("attribute")
	updatedAttributes, index, oldValue := shared.DeleteAttribute(updatedVirtualHost.Attributes, attributeToDelete)
	if index == -1 {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedVirtualHost.Attributes = updatedAttributes
	updatedVirtualHost.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateVirtualHostByName(&updatedVirtualHost); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteVirtualHostByName deletes a route
func (s *server) DeleteVirtualHostByName(c *gin.Context) {
	route, err := s.db.GetVirtualHostByName(c.Param("virtualhost"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	if err := s.db.DeleteVirtualHostByName(route.Name); err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
	}
	c.IndentedJSON(http.StatusOK, route)
}
