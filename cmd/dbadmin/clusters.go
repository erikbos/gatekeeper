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
	r.POST("/v1/clusters", e.CheckForJSONContentType, e.PostCreateCluster)

	r.GET("/v1/clusters/:cluster", e.GetClusterByName)
	r.POST("/v1/clusters/:cluster", e.CheckForJSONContentType, e.PostCluster)
	r.DELETE("/v1/clusters/:cluster", e.DeleteClusterByName)
}

// GetClusters returns all clusters
func (e *env) GetClusters(c *gin.Context) {
	clusters, err := e.db.GetClusters()
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"clusters": clusters})
}

// GetClusterByName returns details of an cluster
func (e *env) GetClusterByName(c *gin.Context) {
	cluster, err := e.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	e.SetLastModifiedHeader(c, cluster.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, cluster)
}

// PostCreateCluster creates an cluster
func (e *env) PostCreateCluster(c *gin.Context) {
	var newCluster types.Cluster
	if err := c.ShouldBindJSON(&newCluster); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	existingCluster, err := e.db.GetClusterByName(newCluster.Name)
	if err == nil {
		e.returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Cluster '%s' already exists", existingCluster.Name))
		return
	}
	// Automatically set default fields
	newCluster.Key = newCluster.Name
	newCluster.CreatedBy = e.whoAmI()
	newCluster.CreatedAt = e.getCurrentTimeMilliseconds()
	newCluster.LastmodifiedAt = newCluster.CreatedAt
	newCluster.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateClusterByName(newCluster); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newCluster)
}

// PostCluster updates an existing cluster
func (e *env) PostCluster(c *gin.Context) {
	currentCluster, err := e.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedCluster types.Cluster
	if err := c.ShouldBindJSON(&updatedCluster); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// We don't allow POSTing to update cluster X while body says to update cluster Y
	updatedCluster.Key = currentCluster.Key
	updatedCluster.CreatedBy = currentCluster.CreatedBy
	updatedCluster.CreatedAt = currentCluster.CreatedAt
	updatedCluster.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedCluster.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateClusterByName(updatedCluster); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedCluster)
}

// DeleteClusterByName deletes a cluster
func (e *env) DeleteClusterByName(c *gin.Context) {
	cluster, err := e.db.GetClusterByName(c.Param("cluster"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	e.db.DeleteClusterByName(cluster.Name)
	c.IndentedJSON(http.StatusOK, cluster)
}
