package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// VirtualhostStore the virtual host information storage interface
	VirtualhostStore interface {
		// GetAll retrieves all virtualhosts
		GetAll() ([]shared.VirtualHost, error)

		// GetByName retrieves a virtualhost
		GetByName(virtualHost string) (*shared.VirtualHost, error)

		// UpdateByName updates a virtualhost
		UpdateByName(vhost *shared.VirtualHost) error

		// DeleteByName deletes a virtualhost
		DeleteByName(virtualHostToDelete string) error
	}
)
