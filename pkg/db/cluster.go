package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// ClusterStore the cluster information storage interface
	ClusterStore interface {
		// GetAll retrieves all clusters
		GetAll() (types.Clusters, types.Error)

		// Get retrieves a cluster from database
		Get(clusterName string) (*types.Cluster, types.Error)

		// Update UPSERTs an cluster in database
		Update(c *types.Cluster) types.Error

		// Update UPSERTs an cluster in database
		Delete(clusterToDelete string) types.Error
	}
)
