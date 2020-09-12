package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// registerClusterRoutes registers all routes we handle
func (s *server) registerClusterRoutes(r *gin.Engine) {
	r.GET("/v1/clusters", s.GetAllClusters)
	r.POST("/v1/clusters", shared.AbortIfContentTypeNotJSON, s.PostCreateCluster)

	r.GET("/v1/clusters/:cluster", s.GetClusterByName)
	r.POST("/v1/clusters/:cluster", shared.AbortIfContentTypeNotJSON, s.PostCluster)
	r.DELETE("/v1/clusters/:cluster", s.DeleteClusterByName)

	r.GET("/v1/clusters/:cluster/attributes", s.GetClusterAttributes)
	r.POST("/v1/clusters/:cluster/attributes", shared.AbortIfContentTypeNotJSON, s.PostClusterAttributes)

	r.GET("/v1/clusters/:cluster/attributes/:attribute", s.GetClusterAttributeByName)
	r.POST("/v1/clusters/:cluster/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostClusterAttributeByName)
	r.DELETE("/v1/clusters/:cluster/attributes/:attribute", s.DeleteClusterAttributeByName)
}

// GetAllClusters returns all clusters
func (s *server) GetAllClusters(c *gin.Context) {

	clusters, err := s.db.Cluster.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"clusters": clusters})
}

// GetClusterByName returns details of an cluster
func (s *server) GetClusterByName(c *gin.Context) {

	cluster, err := s.db.Cluster.GetByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, cluster.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, cluster)
}

// GetClusterAttributes returns attributes of a cluster
func (s *server) GetClusterAttributes(c *gin.Context) {

	cluster, err := s.db.Cluster.GetByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, cluster.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": cluster.Attributes})
}

// GetClusterAttributeByName returns one particular attribute of a cluster
func (s *server) GetClusterAttributeByName(c *gin.Context) {

	cluster, err := s.db.Cluster.GetByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	value, err := cluster.Attributes.Get(c.Param("attribute"))
	if err != nil {
		returnCanNotFindAttribute(c, c.Param("attribute"))
		return
	}

	setLastModifiedHeader(c, cluster.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, value)
}

// PostCreateCluster creates a cluster
func (s *server) PostCreateCluster(c *gin.Context) {

	var newCluster types.Cluster
	if err := c.ShouldBindJSON(&newCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	existingCluster, err := s.db.Cluster.GetByName(newCluster.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Cluster '%s' already exists", existingCluster.Name))
		return
	}

	// Automatically set default fields
	newCluster.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newCluster.CreatedBy = s.whoAmI()
	newCluster.LastmodifiedBy = s.whoAmI()

	if err := s.db.Cluster.UpdateByName(&newCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newCluster)
}

// PostCluster updates an existing cluster
func (s *server) PostCluster(c *gin.Context) {

	clusterToUpdate, err := s.db.Cluster.GetByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updateRequest types.Cluster
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return

	}

	// Copy over the fields we allow to be updated
	clusterToUpdate.HostName = updateRequest.HostName
	clusterToUpdate.Port = updateRequest.Port
	clusterToUpdate.DisplayName = updateRequest.DisplayName
	clusterToUpdate.Attributes = updateRequest.Attributes

	clusterToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Cluster.UpdateByName(clusterToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, clusterToUpdate)
}

// PostClusterAttributes updates attributes of a cluster
func (s *server) PostClusterAttributes(c *gin.Context) {

	clusterToUpdate, err := s.db.Cluster.GetByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	var body struct {
		Attributes types.Attributes `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	if len(body.Attributes) == 0 {
		returnJSONMessage(c, http.StatusBadRequest, errors.New("No attributes posted"))
		return
	}

	clusterToUpdate.Attributes = body.Attributes
	clusterToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Cluster.UpdateByName(clusterToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": clusterToUpdate.Attributes})
}

// PostClusterAttributeByName update an attribute of APIProduct
func (s *server) PostClusterAttributeByName(c *gin.Context) {

	clusterToUpdate, err := s.db.Cluster.GetByName(c.Param("cluster"))
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
	clusterToUpdate.Attributes.Set(attributeToUpdate, body.Value)
	clusterToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Cluster.UpdateByName(clusterToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": body.Value})
}

// DeleteClusterAttributeByName removes an attribute of cluster
func (s *server) DeleteClusterAttributeByName(c *gin.Context) {

	updatedCluster, err := s.db.Cluster.GetByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	attributeToDelete := c.Param("attribute")
	deleted, oldValue := updatedCluster.Attributes.Delete(attributeToDelete)
	if !deleted {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedCluster.LastmodifiedBy = s.whoAmI()

	if err := s.db.Cluster.UpdateByName(updatedCluster); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteClusterByName deletes a cluster
func (s *server) DeleteClusterByName(c *gin.Context) {

	cluster, err := s.db.Cluster.GetByName(c.Param("cluster"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	if err := s.db.Cluster.DeleteByName(cluster.Name); err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
	}
	c.IndentedJSON(http.StatusOK, cluster)
}
