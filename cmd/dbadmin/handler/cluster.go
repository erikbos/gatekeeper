package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// registerClusterRoutes registers all routes we handle
func (h *Handler) registerClusterRoutes(r *gin.RouterGroup) {
	r.GET("/clusters", h.handler(h.getAllClusters))
	r.POST("/clusters", h.handler(h.createCluster))

	r.GET("/clusters/:cluster", h.handler(h.getCluster))
	r.POST("/clusters/:cluster", h.handler(h.updateCluster))
	r.DELETE("/clusters/:cluster", h.handler(h.deleteCluster))

	r.GET("/clusters/:cluster/attributes", h.handler(h.getClusterAttributes))
	r.POST("/clusters/:cluster/attributes", h.handler(h.updateClusterAttributes))

	r.GET("/clusters/:cluster/attributes/:attribute", h.handler(h.getClusterAttribute))
	r.POST("/clusters/:cluster/attributes/:attribute", h.handler(h.updateClusterAttribute))
	r.DELETE("/clusters/:cluster/attributes/:attribute", h.handler(h.deleteClusterAttribute))
}

const (
	// Name of cluster parameter in the route definition
	clusterParameter = "cluster"
)

// getAllClusters returns all clusters
func (h *Handler) getAllClusters(c *gin.Context) handlerResponse {

	clusters, err := h.service.Cluster.GetAll()
	if err != nil {
		return handleError(err)
	}
	return handleOK(StringMap{"clusters": clusters})
}

// getCluster returns details of an cluster
func (h *Handler) getCluster(c *gin.Context) handlerResponse {

	cluster, err := h.service.Cluster.Get(c.Param(clusterParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(cluster)
}

// getClusterAttributes returns attributes of an cluster
func (h *Handler) getClusterAttributes(c *gin.Context) handlerResponse {

	cluster, err := h.service.Cluster.Get(c.Param(clusterParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOKAttributes(cluster.Attributes)
}

// getClusterAttribute returns one particular attribute of an cluster
func (h *Handler) getClusterAttribute(c *gin.Context) handlerResponse {

	cluster, err := h.service.Cluster.Get(c.Param(clusterParameter))
	if err != nil {
		return handleError(err)
	}
	attributeValue, err := cluster.Attributes.Get(c.Param("attribute"))
	if err != nil {
		return handleError(err)
	}
	return handleOK(attributeValue)
}

// createCluster creates an cluster
func (h *Handler) createCluster(c *gin.Context) handlerResponse {

	var newCluster types.Cluster
	if err := c.ShouldBindJSON(&newCluster); err != nil {
		return handleBadRequest(err)
	}
	storedCluster, err := h.service.Cluster.Create(newCluster, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleCreated(storedCluster)
}

// updateCluster updates an existing cluster
func (h *Handler) updateCluster(c *gin.Context) handlerResponse {

	var updatedCluster types.Cluster
	if err := c.ShouldBindJSON(&updatedCluster); err != nil {
		return handleBadRequest(err)
	}
	// clustername in path must match clustername in posted body
	if updatedCluster.Name != c.Param(clusterParameter) {
		return handleNameMismatch()
	}
	storedCluster, err := h.service.Cluster.Update(updatedCluster, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedCluster)
}

// updateClusterAttributes updates attributes of an cluster
func (h *Handler) updateClusterAttributes(c *gin.Context) handlerResponse {

	var receivedAttributes struct {
		Attributes types.Attributes `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		return handleBadRequest(err)
	}
	if err := h.service.Cluster.UpdateAttributes(c.Param(clusterParameter),
		receivedAttributes.Attributes, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttributes(receivedAttributes.Attributes)
}

// updateClusterAttribute update an attribute of developer
func (h *Handler) updateClusterAttribute(c *gin.Context) handlerResponse {

	var receivedValue AttributeValue
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		return handleBadRequest(err)
	}
	newAttribute := types.Attribute{
		Name:  c.Param("attribute"),
		Value: receivedValue.Value,
	}
	if err := h.service.Cluster.UpdateAttribute(c.Param(clusterParameter),
		newAttribute, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttribute(newAttribute)
}

// deleteClusterAttribute removes an attribute of an cluster
func (h *Handler) deleteClusterAttribute(c *gin.Context) handlerResponse {

	attributeToDelete := c.Param("attribute")
	oldValue, err := h.service.Cluster.DeleteAttribute(c.Param(clusterParameter),
		attributeToDelete, h.who(c))
	if err != nil {
		return handleBadRequest(err)
	}
	return handleOKAttribute(types.Attribute{
		Name:  attributeToDelete,
		Value: oldValue,
	})
}

// deleteCluster deletes an cluster
func (h *Handler) deleteCluster(c *gin.Context) handlerResponse {

	deletedCluster, err := h.service.Cluster.Delete(c.Param(clusterParameter), h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(deletedCluster)
}
