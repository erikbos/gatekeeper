package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// RouteService is
type RouteService struct {
	db *db.Database
}

// NewRouteService returns a new route instance
func NewRouteService(database *db.Database) *RouteService {

	return &RouteService{db: database}
}

// GetAll returns all routes
func (rs *RouteService) GetAll() (routes types.Routes, err types.Error) {

	return rs.db.Route.GetAll()
}

// Get returns details of an route
func (rs *RouteService) Get(routeName string) (route *types.Route, err types.Error) {

	return rs.db.Route.Get(routeName)
}

// GetAttributes returns attributes of an route
func (rs *RouteService) GetAttributes(routeName string) (attributes *types.Attributes, err types.Error) {

	route, err := rs.db.Route.Get(routeName)
	if err != nil {
		return nil, err
	}
	return &route.Attributes, nil
}

// GetAttribute returns one particular attribute of an route
func (rs *RouteService) GetAttribute(routeName, attributeName string) (value string, err types.Error) {

	route, err := rs.db.Route.Get(routeName)
	if err != nil {
		return "", err
	}
	return route.Attributes.Get(attributeName)
}

// Create creates an route
func (rs *RouteService) Create(newRoute types.Route) (types.Route, types.Error) {

	existingRoute, err := rs.db.Route.Get(newRoute.Name)
	if err == nil {
		return types.NullRoute, types.NewBadRequestError(
			fmt.Errorf("Route '%s' already exists", existingRoute.Name))
	}
	// Automatically set default fields
	newRoute.CreatedAt = shared.GetCurrentTimeMilliseconds()

	err = rs.updateRoute(&newRoute)
	return newRoute, err
}

// Update updates an existing route
func (rs *RouteService) Update(updatedRoute types.Route) (types.Route, types.Error) {

	routeToUpdate, err := rs.db.Route.Get(updatedRoute.Name)
	if err != nil {
		return types.NullRoute, types.NewItemNotFoundError(err)
	}

	routeToUpdate.DisplayName = updatedRoute.DisplayName
	routeToUpdate.RouteGroup = updatedRoute.RouteGroup
	routeToUpdate.Path = updatedRoute.Path
	routeToUpdate.PathType = updatedRoute.PathType
	routeToUpdate.Attributes = updatedRoute.Attributes

	err = rs.updateRoute(routeToUpdate)
	return *routeToUpdate, err
}

// UpdateAttributes updates attributes of an route
func (rs *RouteService) UpdateAttributes(routeName string, receivedAttributes types.Attributes) types.Error {

	updatedRoute, err := rs.db.Route.Get(routeName)
	if err != nil {
		return types.NewItemNotFoundError(err)
	}
	updatedRoute.Attributes = receivedAttributes
	return rs.updateRoute(updatedRoute)
}

// UpdateAttribute update an attribute of developer
func (rs *RouteService) UpdateAttribute(routeName string, attributeValue types.Attribute) types.Error {

	updatedRoute, err := rs.db.Route.Get(routeName)
	if err != nil {
		return err
	}
	updatedRoute.Attributes.Set(attributeValue)
	return rs.updateRoute(updatedRoute)
}

// DeleteAttribute removes an attribute of an route
func (rs *RouteService) DeleteAttribute(routeName, attributeToDelete string) (string, types.Error) {

	updatedRoute, err := rs.db.Route.Get(routeName)
	if err != nil {
		return "", err
	}
	oldValue, err := updatedRoute.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}
	return oldValue, rs.updateRoute(updatedRoute)
}

// updateRoute updates last-modified field(s) and updates route in database
func (rs *RouteService) updateRoute(updatedRoute *types.Route) types.Error {

	updatedRoute.Attributes.Tidy()
	updatedRoute.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	return rs.db.Route.Update(updatedRoute)
}

// Delete deletes an route
func (rs *RouteService) Delete(routeName string) (deletedRoute types.Route, e types.Error) {

	route, err := rs.db.Route.Get(routeName)
	if err != nil {
		return types.NullRoute, err
	}
	err = rs.db.Route.Delete(routeName)
	if err != nil {
		return types.NullRoute, err
	}
	return *route, nil
}
