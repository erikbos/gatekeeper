package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// ClusterService is
type ClusterService struct {
	db *db.Database
}

// NewClusterService returns a new cluster instance
func NewClusterService(database *db.Database) *ClusterService {

	return &ClusterService{db: database}
}

// GetAll returns all clusters
func (cs *ClusterService) GetAll() (clusters types.Clusters, err types.Error) {

	return cs.db.Cluster.GetAll()
}

// Get returns details of an cluster
func (cs *ClusterService) Get(clusterName string) (cluster *types.Cluster, err types.Error) {

	return cs.db.Cluster.Get(clusterName)
}

// GetAttributes returns attributes of an cluster
func (cs *ClusterService) GetAttributes(clusterName string) (attributes *types.Attributes, err types.Error) {

	cluster, err := cs.db.Cluster.Get(clusterName)
	if err != nil {
		return nil, err
	}
	return &cluster.Attributes, nil
}

// GetAttribute returns one particular attribute of an cluster
func (cs *ClusterService) GetAttribute(clusterName, attributeName string) (value string, err types.Error) {

	cluster, err := cs.db.Cluster.Get(clusterName)
	if err != nil {
		return "", err
	}
	return cluster.Attributes.Get(attributeName)
}

// Create creates an cluster
func (cs *ClusterService) Create(newCluster types.Cluster) (types.Cluster, types.Error) {

	existingCluster, err := cs.db.Cluster.Get(newCluster.Name)
	if err == nil {
		return types.NullCluster, types.NewBadRequestError(
			fmt.Errorf("Cluster '%s' already exists", existingCluster.Name))
	}
	// Automatically set default fields
	newCluster.CreatedAt = shared.GetCurrentTimeMilliseconds()

	err = cs.updateCluster(&newCluster)
	return newCluster, err
}

// Update updates an existing cluster
func (cs *ClusterService) Update(updatedCluster types.Cluster) (types.Cluster, types.Error) {

	clusterToUpdate, err := cs.db.Cluster.Get(updatedCluster.Name)
	if err != nil {
		return types.NullCluster, types.NewItemNotFoundError(err)
	}
	// Copy over the fields we allow to be updated
	clusterToUpdate.HostName = updatedCluster.HostName
	clusterToUpdate.Port = updatedCluster.Port
	clusterToUpdate.DisplayName = updatedCluster.DisplayName
	clusterToUpdate.Attributes = updatedCluster.Attributes

	err = cs.updateCluster(&updatedCluster)
	return *clusterToUpdate, err
}

// UpdateAttributes updates attributes of an cluster
func (cs *ClusterService) UpdateAttributes(clusterName string, receivedAttributes types.Attributes) types.Error {

	updatedCluster, err := cs.db.Cluster.Get(clusterName)
	if err != nil {
		return types.NewItemNotFoundError(err)
	}
	updatedCluster.Attributes = receivedAttributes

	return cs.updateCluster(updatedCluster)
}

// UpdateAttribute update an attribute of developer
func (cs *ClusterService) UpdateAttribute(clusterName string, attributeValue types.Attribute) types.Error {

	updatedCluster, err := cs.db.Cluster.Get(clusterName)
	if err != nil {
		return err
	}
	updatedCluster.Attributes.Set(attributeValue)
	return cs.updateCluster(updatedCluster)
}

// DeleteAttribute removes an attribute of an cluster and return its former value
func (cs *ClusterService) DeleteAttribute(clusterName, attributeToDelete string) (string, types.Error) {

	updatedCluster, err := cs.db.Cluster.Get(clusterName)
	if err != nil {
		return "", err
	}
	oldValue, err := updatedCluster.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}
	return oldValue, cs.updateCluster(updatedCluster)
}

// updateCluster updates last-modified field(s) and updates cluster in database
func (cs *ClusterService) updateCluster(updatedCluster *types.Cluster) types.Error {

	updatedCluster.Attributes.Tidy()
	updatedCluster.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	return cs.db.Cluster.Update(updatedCluster)
}

// Delete deletes an cluster
func (cs *ClusterService) Delete(clusterName string) (deletedCluster types.Cluster, e types.Error) {

	cluster, err := cs.db.Cluster.Get(clusterName)
	if err != nil {
		return types.NullCluster, err
	}
	err = cs.db.Route.Delete(clusterName)
	if err != nil {
		return types.NullCluster, err
	}
	return *cluster, nil
}
