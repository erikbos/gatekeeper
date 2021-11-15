package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// returns all clusters
// (GET /v1/clusters)
func (h *Handler) GetV1Clusters(c *gin.Context) {

	clusters, err := h.service.Cluster.GetAll()
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseClusterss(c, clusters)
}

// creates a new cluster
// (POST /v1/clusters)
func (h *Handler) PostV1Clusters(c *gin.Context) {

	var receivedCluster Cluster
	if err := c.ShouldBindJSON(&receivedCluster); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newCluster := fromCluster(receivedCluster)
	createdDeveloper, err := h.service.Cluster.Create(newCluster, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseClusterCreated(c, &createdDeveloper)
}

// deletes an cluster
// (DELETE /v1/clusters/{cluster_name})
func (h *Handler) DeleteV1ClustersClusterName(c *gin.Context, clusterName ClusterName) {

	cluster, err := h.service.Cluster.Delete(string(clusterName), h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseClusters(c, &cluster)
}

// returns full details of one cluster
// (GET /v1/clusters/{cluster_name})
func (h *Handler) GetV1ClustersClusterName(c *gin.Context, clusterName ClusterName) {

	cluster, err := h.service.Cluster.Get(string(clusterName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseClusters(c, cluster)
}

// (POST /v1/clusters/{cluster_name})
func (h *Handler) PostV1ClustersClusterName(c *gin.Context, clusterName ClusterName) {

	var receivedCluster Cluster
	if err := c.ShouldBindJSON(&receivedCluster); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	updatedCluster := fromCluster(receivedCluster)
	storedCluster, err := h.service.Cluster.Update(updatedCluster, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseClustersUpdated(c, &storedCluster)
}

// returns attributes of a cluster
// (GET /v1/clusters/{cluster_name}/attributes)
func (h *Handler) GetV1ClustersClusterNameAttributes(c *gin.Context, clusterName ClusterName) {

	cluster, err := h.service.Cluster.Get(string(clusterName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, cluster.Attributes)
}

// replaces attributes of an cluster
// (POST /v1/clusters/{cluster_name}/attributes)
func (h *Handler) PostV1ClustersClusterNameAttributes(c *gin.Context, clusterName ClusterName) {

	var receivedAttributes Attributes
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	attributes := fromAttributesRequest(receivedAttributes.Attribute)
	if err := h.service.Cluster.UpdateAttributes(
		string(clusterName), attributes, h.who(c)); err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, attributes)
}

// deletes one attribute of an cluster
// (DELETE /v1/clusters/{cluster_name}/attributes/{attribute_name})
func (h *Handler) DeleteV1ClustersClusterNameAttributesAttributeName(c *gin.Context, clusterName ClusterName, attributeName AttributeName) {

	oldValue, err := h.service.Cluster.DeleteAttribute(
		string(clusterName), string(attributeName), h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeDeleted(c, types.NewAttribute(string(attributeName), oldValue))
}

// returns one attribute of an cluster
// (GET /v1/clusters/{cluster_name}/attributes/{attribute_name})
func (h *Handler) GetV1ClustersClusterNameAttributesAttributeName(c *gin.Context, clusterName ClusterName, attributeName AttributeName) {

	cluster, err := h.service.Cluster.Get(string(clusterName))
	if err != nil {
		responseError(c, err)
		return
	}
	attributeValue, err := cluster.Attributes.Get(string(attributeName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeRetrieved(c, types.NewAttribute(string(attributeName), attributeValue))
}

// updates an attribute of an cluster
// (POST /v1/clusters/{cluster_name}/attributes/{attribute_name})
func (h *Handler) PostV1ClustersClusterNameAttributesAttributeName(c *gin.Context, clusterName ClusterName, attributeName AttributeName) {

	var receivedValue Attribute
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newAttribute := types.NewAttribute(string(attributeName), *receivedValue.Value)
	if err := h.service.Cluster.UpdateAttribute(
		string(clusterName), *newAttribute, h.who(c)); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	h.responseAttributeUpdated(c, newAttribute)
}

// API responses

func (h *Handler) responseClusterss(c *gin.Context, clusters types.Clusters) {

	all_clusters := make([]Cluster, len(clusters))
	for i := range clusters {
		all_clusters[i] = h.ToClusterResponse(&clusters[i])
	}
	c.IndentedJSON(http.StatusOK, Clusters{
		Clusters: &all_clusters,
	})
}

func (h *Handler) responseClusters(c *gin.Context, cluster *types.Cluster) {

	c.IndentedJSON(http.StatusOK, h.ToClusterResponse(cluster))
}

func (h *Handler) responseClusterCreated(c *gin.Context, cluster *types.Cluster) {

	c.IndentedJSON(http.StatusCreated, h.ToClusterResponse(cluster))
}

func (h *Handler) responseClustersUpdated(c *gin.Context, cluster *types.Cluster) {

	c.IndentedJSON(http.StatusOK, h.ToClusterResponse(cluster))
}

// type conversion

func (h *Handler) ToClusterResponse(c *types.Cluster) Cluster {

	cluster := Cluster{
		Attributes:     toAttributesResponse(c.Attributes),
		CreatedAt:      &c.CreatedAt,
		CreatedBy:      &c.CreatedBy,
		DisplayName:    &c.DisplayName,
		LastModifiedBy: &c.LastModifiedBy,
		LastModifiedAt: &c.LastModifiedAt,
		Name:           c.Name,
	}
	return cluster
}

func fromCluster(c Cluster) types.Cluster {

	cluster := types.Cluster{}
	if c.Attributes != nil {
		cluster.Attributes = fromAttributesRequest(c.Attributes)
	}
	if c.CreatedAt != nil {
		cluster.CreatedAt = *c.CreatedAt
	}
	if c.CreatedBy != nil {
		cluster.CreatedBy = *c.CreatedBy
	}
	if c.DisplayName != nil {
		cluster.DisplayName = *c.DisplayName
	}
	if c.LastModifiedBy != nil {
		cluster.LastModifiedBy = *c.LastModifiedBy
	}
	if c.LastModifiedAt != nil {
		cluster.LastModifiedAt = *c.LastModifiedAt
	}
	if c.Name != "" {
		cluster.Name = c.Name
	}
	return cluster
}
