package db

import "github.com/erikbos/gatekeeper/pkg/types"

type (
	// RouteStore the route information storage interface
	RouteStore interface {
		// GetAll retrieves all routes
		GetAll() (types.Routes, error)

		// GetRouteByName retrieves a route from database
		GetRouteByName(routeName string) (*types.Route, error)

		// UpdateRouteByName UPSERTs an route
		UpdateRouteByName(route *types.Route) error

		// DeleteRouteByName deletes a route
		DeleteRouteByName(routeToDelete string) error
	}
)
