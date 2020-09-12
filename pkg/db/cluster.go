package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// ClusterStore the cluster information storage interface
	ClusterStore interface {
		// GetAll retrieves all clusters
		GetAll() (types.Clusters, error)

		// GetByName retrieves a cluster from database
		GetByName(clusterName string) (*types.Cluster, error)

		// UpdateByName UPSERTs an cluster in database
		UpdateByName(c *types.Cluster) error

		// UpdateByName UPSERTs an cluster in database
		DeleteByName(clusterToDelete string) error
	}
)
