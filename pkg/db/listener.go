package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// ListenerStore is the listener information storage interface
	ListenerStore interface {
		// GetAll retrieves all listeners
		GetAll() (types.Listeners, types.Error)

		// Get retrieves a listener
		Get(listener string) (*types.Listener, types.Error)

		// Update updates a listener
		Update(listener *types.Listener) types.Error

		// Delete deletes a listener
		Delete(listenerToDelete string) types.Error
	}
)
