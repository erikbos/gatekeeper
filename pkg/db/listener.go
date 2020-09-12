package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// ListenerStore is the listener information storage interface
	ListenerStore interface {
		// GetAll retrieves all listeners
		GetAll() (types.Listeners, error)

		// GetByName retrieves a listener
		GetByName(listener string) (*types.Listener, error)

		// UpdateByName updates a listener
		UpdateByName(vhost *types.Listener) error

		// DeleteByName deletes a listener
		DeleteByName(listenerToDelete string) error
	}
)
