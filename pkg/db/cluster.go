package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// ClusterStore the cluster information storage interface
	ClusterStore interface {
		// GetAll retrieves all clusters
		GetAll() ([]shared.Cluster, error)

		// GetByName retrieves a cluster from database
		GetByName(clusterName string) (*shared.Cluster, error)

		// UpdateByName UPSERTs an cluster in database
		UpdateByName(c *shared.Cluster) error

		// UpdateByName UPSERTs an cluster in database
		DeleteByName(clusterToDelete string) error
	}
)
