package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// RouteStore the route information storage interface
	RouteStore interface {
		// GetAll retrieves all routes
		GetAll() (types.Routes, types.Error)

		// Get retrieves a route from database
		Get(routeName string) (*types.Route, types.Error)

		// Update UPSERTs an route
		Update(route *types.Route) types.Error

		// Delete deletes a route
		Delete(routeToDelete string) types.Error
	}
)
