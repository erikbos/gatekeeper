package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/apiauth/pkg/shared"
)

// registerClusterRoutes registers all routes we handle
func (s *server) registerClusterRoutes(r *gin.Engine) {
	r.GET("/v1/clusters", s.GetClusters)
	r.POST("/v1/clusters", shared.AbortIfContentTypeNotJSON, s.PostCreateCluster)

	r.GET("/v1/clusters/:cluster", s.GetClusterByName)
	r.POST("/v1/clusters/:cluster", shared.AbortIfContentTypeNotJSON, s.PostCluster)
	r.DELETE("/v1/clusters/:cluster", s.DeleteClusterByName)

	r.GET("/v1/clusters/:cluster/attributes", s.GetClusterAttributes)
	r.POST("/v1/clusters/:cluster/attributes", shared.AbortIfContentTypeNotJSON, s.PostClusterAttributes)
	r.DELETE("/v1/clusters/:cluster/attributes", s.DeleteClusterAttributes)

	r.GET("/v1/clusters/:cluster/attributes/:attribute", s.GetClusterAttributeByName)
	r.POST("/v1/clusters/:cluster/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostClusterAttributeByName)
	r.DELETE("/v1/clusters/:cluster/attributes/:attribute", s.DeleteClusterAttributeByName)
}

// GetClusters returns all clusters
func (s *server) GetClusters(c *gin.Context) {
	clusters, err := s.db.GetClusters()
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"clusters": clusters})
}

// GetClusterByName returns details of an cluster
func (s *server) GetClusterByName(c *gin.Context) {
	cluster, err := s.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, cluster.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, cluster)
}

// GetClusterAttributes returns attributes of a cluster
func (s *server) GetClusterAttributes(c *gin.Context) {
	cluster, err := s.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, cluster.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": cluster.Attributes})
}

// GetClusterAttributeByName returns one particular attribute of a cluster
func (s *server) GetClusterAttributeByName(c *gin.Context) {
	cluster, err := s.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(cluster.Attributes); i++ {
		if cluster.Attributes[i].Name == c.Param("attribute") {
			setLastModifiedHeader(c, cluster.LastmodifiedAt)
			c.IndentedJSON(http.StatusOK, cluster.Attributes[i])
			return
		}
	}
	returnJSONMessage(c, http.StatusNotFound,
		fmt.Errorf("Could not retrieve attribute '%s'", c.Param("attribute")))
}

// PostCreateCluster creates a cluster
func (s *server) PostCreateCluster(c *gin.Context) {
	var newCluster shared.Cluster
	if err := c.ShouldBindJSON(&newCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	existingCluster, err := s.db.GetClusterByName(newCluster.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Cluster '%s' already exists", existingCluster.Name))
		return
	}
	// Automatically set default fields
	newCluster.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newCluster.CreatedBy = s.whoAmI()
	newCluster.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateClusterByName(&newCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newCluster)
}

// PostCluster updates an existing cluster
func (s *server) PostCluster(c *gin.Context) {
	currentCluster, err := s.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedCluster shared.Cluster
	if err := c.ShouldBindJSON(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update cluster X while body says to update cluster Y
	updatedCluster.Name = currentCluster.Name
	updatedCluster.CreatedBy = currentCluster.CreatedBy
	updatedCluster.CreatedAt = currentCluster.CreatedAt
	updatedCluster.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateClusterByName(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedCluster)
}

// PostClusterAttributes updates attributes of a cluster
func (s *server) PostClusterAttributes(c *gin.Context) {
	updatedCluster, err := s.db.GetClusterByName(c.Param("cluster"))
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
	updatedCluster.Attributes = receivedAttributes.Attributes
	updatedCluster.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateClusterByName(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedCluster.Attributes})
}

// DeleteClusterAttributes delete attributes of APIProduct
func (s *server) DeleteClusterAttributes(c *gin.Context) {
	updatedCluster, err := s.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	deletedAttributes := updatedCluster.Attributes
	updatedCluster.Attributes = nil
	updatedCluster.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateClusterByName(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttributes)
}

// PostClusterAttributeByName update an attribute of APIProduct
func (s *server) PostClusterAttributeByName(c *gin.Context) {
	updatedCluster, err := s.db.GetClusterByName(c.Param("cluster"))
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
		updatedCluster.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find existing attribute, append new attribute
		updatedCluster.Attributes = append(updatedCluster.Attributes,
			shared.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedCluster.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedCluster.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateClusterByName(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteClusterAttributeByName removes an attribute of cluster
func (s *server) DeleteClusterAttributeByName(c *gin.Context) {
	updatedCluster, err := s.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	attributeToDelete := c.Param("attribute")
	updatedAttributes, index, oldValue := shared.DeleteAttribute(updatedCluster.Attributes, attributeToDelete)
	if index == -1 {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedCluster.Attributes = updatedAttributes
	updatedCluster.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateClusterByName(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteClusterByName deletes a cluster
func (s *server) DeleteClusterByName(c *gin.Context) {
	cluster, err := s.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	if err := s.db.DeleteClusterByName(cluster.Name); err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
	}
	c.IndentedJSON(http.StatusOK, cluster)
}
