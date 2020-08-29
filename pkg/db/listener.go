package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// ListenerStore is the listener information storage interface
	ListenerStore interface {
		// GetAll retrieves all listeners
		GetAll() ([]shared.Listener, error)

		// GetByName retrieves a listener
		GetByName(listener string) (*shared.Listener, error)

		// UpdateByName updates a listener
		UpdateByName(vhost *shared.Listener) error

		// DeleteByName deletes a listener
		DeleteByName(listenerToDelete string) error
	}
)
