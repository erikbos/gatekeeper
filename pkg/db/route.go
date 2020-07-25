package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// RouteStore the route information storage interface
	RouteStore interface {
		// GetAll retrieves all routes
		GetAll() ([]shared.Route, error)

		// GetRouteByName retrieves a route from database
		GetRouteByName(routeName string) (*shared.Route, error)

		// UpdateRouteByName UPSERTs an route
		UpdateRouteByName(route *shared.Route) error

		// DeleteRouteByName deletes a route
		DeleteRouteByName(routeToDelete string) error
	}
)
