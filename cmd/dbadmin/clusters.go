package main

import (
	"fmt"
	"net/http"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
)

// registerClusterRoutes registers all routes we handle
func (e *env) registerClusterRoutes(r *gin.Engine) {
	r.GET("/v1/clusters", e.GetClusters)
	r.POST("/v1/clusters", types.AbortIfContentTypeNotJSON, e.PostCreateCluster)

	r.GET("/v1/clusters/:cluster", e.GetClusterByName)
	r.POST("/v1/clusters/:cluster", types.AbortIfContentTypeNotJSON, e.PostCluster)
	r.DELETE("/v1/clusters/:cluster", e.DeleteClusterByName)

	r.GET("/v1/clusters/:cluster/attributes", e.GetClusterAttributes)
	r.POST("/v1/clusters/:cluster/attributes", types.AbortIfContentTypeNotJSON, e.PostClusterAttributes)
	r.DELETE("/v1/clusters/:cluster/attributes", e.DeleteClusterAttributes)

	r.GET("/v1/clusters/:cluster/attributes/:attribute", e.GetClusterAttributeByName)
	r.POST("/v1/clusters/:cluster/attributes/:attribute", types.AbortIfContentTypeNotJSON, e.PostClusterAttributeByName)
	r.DELETE("/v1/clusters/:cluster/attributes/:attribute", e.DeleteClusterAttributeByName)
}

// GetClusters returns all clusters
func (e *env) GetClusters(c *gin.Context) {
	clusters, err := e.db.GetClusters()
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"clusters": clusters})
}

// GetClusterByName returns details of an cluster
func (e *env) GetClusterByName(c *gin.Context) {
	cluster, err := e.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, cluster.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, cluster)
}

// GetClusterAttributes returns attributes of a cluster
func (e *env) GetClusterAttributes(c *gin.Context) {
	cluster, err := e.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, cluster.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": cluster.Attributes})
}

// GetClusterAttributeByName returns one particular attribute of a cluster
func (e *env) GetClusterAttributeByName(c *gin.Context) {
	cluster, err := e.db.GetClusterByName(c.Param("cluster"))
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
func (e *env) PostCreateCluster(c *gin.Context) {
	var newCluster types.Cluster
	if err := c.ShouldBindJSON(&newCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	existingCluster, err := e.db.GetClusterByName(newCluster.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Cluster '%s' already exists", existingCluster.Name))
		return
	}
	// Automatically set default fields
	newCluster.CreatedAt = types.GetCurrentTimeMilliseconds()
	newCluster.CreatedBy = e.whoAmI()
	newCluster.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateClusterByName(&newCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newCluster)
}

// PostCluster updates an existing cluster
func (e *env) PostCluster(c *gin.Context) {
	currentCluster, err := e.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedCluster types.Cluster
	if err := c.ShouldBindJSON(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update cluster X while body says to update cluster Y
	updatedCluster.Name = currentCluster.Name
	updatedCluster.CreatedBy = currentCluster.CreatedBy
	updatedCluster.CreatedAt = currentCluster.CreatedAt
	updatedCluster.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateClusterByName(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedCluster)
}

// PostClusterAttributes updates attributes of a cluster
func (e *env) PostClusterAttributes(c *gin.Context) {
	updatedCluster, err := e.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	var receivedAttributes struct {
		Attributes []types.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	updatedCluster.Attributes = receivedAttributes.Attributes
	updatedCluster.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateClusterByName(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedCluster.Attributes})
}

// DeleteClusterAttributes delete attributes of APIProduct
func (e *env) DeleteClusterAttributes(c *gin.Context) {
	updatedCluster, err := e.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	deletedAttributes := updatedCluster.Attributes
	updatedCluster.Attributes = nil
	updatedCluster.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateClusterByName(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttributes)
}

// PostClusterAttributeByName update an attribute of APIProduct
func (e *env) PostClusterAttributeByName(c *gin.Context) {
	updatedCluster, err := e.db.GetClusterByName(c.Param("cluster"))
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
	attributeToUpdateIndex := types.FindIndexOfAttribute(
		updatedCluster.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find existing attribute, append new attribute
		updatedCluster.Attributes = append(updatedCluster.Attributes,
			types.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedCluster.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedCluster.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateClusterByName(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteClusterAttributeByName removes an attribute of cluster
func (e *env) DeleteClusterAttributeByName(c *gin.Context) {
	updatedCluster, err := e.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	attributeToDelete := c.Param("attribute")
	updatedAttributes, index, oldValue := types.DeleteAttribute(updatedCluster.Attributes, attributeToDelete)
	if index == -1 {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedCluster.Attributes = updatedAttributes
	updatedCluster.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateClusterByName(&updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteClusterByName deletes a cluster
func (e *env) DeleteClusterByName(c *gin.Context) {
	cluster, err := e.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	e.db.DeleteClusterByName(cluster.Name)
	c.IndentedJSON(http.StatusOK, cluster)
}
