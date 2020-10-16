package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// RouteService is
type RouteService struct {
	db        *db.Database
	changelog *Changelog
}

// NewRoute returns a new route instance
func NewRoute(database *db.Database, c *Changelog) *RouteService {

	return &RouteService{db: database, changelog: c}
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
func (rs *RouteService) Create(newRoute types.Route, who Requester) (types.Route, types.Error) {

	existingRoute, err := rs.db.Route.Get(newRoute.Name)
	if err == nil {
		return types.NullRoute, types.NewBadRequestError(
			fmt.Errorf("Route '%s' already exists", existingRoute.Name))
	}
	// Automatically set default fields
	newRoute.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newRoute.CreatedBy = who.User

	err = rs.updateRoute(&newRoute, who)
	rs.changelog.Create(newRoute, who)
	return newRoute, err
}

// Update updates an existing route
func (rs *RouteService) Update(updatedRoute types.Route,
	who Requester) (types.Route, types.Error) {

	currentRoute, err := rs.db.Route.Get(updatedRoute.Name)
	if err != nil {
		return types.NullRoute, types.NewItemNotFoundError(err)
	}
	// Copy over fields we do not allow to be updated
	updatedRoute.Name = currentRoute.Name
	updatedRoute.CreatedAt = currentRoute.CreatedAt
	updatedRoute.CreatedBy = currentRoute.CreatedBy

	err = rs.updateRoute(&updatedRoute, who)
	rs.changelog.Update(currentRoute, updatedRoute, who)
	return updatedRoute, err
}

// UpdateAttributes updates attributes of an route
func (rs *RouteService) UpdateAttributes(routeName string,
	receivedAttributes types.Attributes, who Requester) types.Error {

	currentRoute, err := rs.db.Route.Get(routeName)
	if err != nil {
		return types.NewItemNotFoundError(err)
	}
	updatedRoute := currentRoute
	if err = updatedRoute.Attributes.SetMultiple(receivedAttributes); err != nil {
		return err
	}

	err = rs.updateRoute(updatedRoute, who)
	rs.changelog.Update(currentRoute, updatedRoute, who)
	return err
}

// UpdateAttribute update an attribute of developer
func (rs *RouteService) UpdateAttribute(routeName string,
	attributeValue types.Attribute, who Requester) types.Error {

	currentRoute, err := rs.db.Route.Get(routeName)
	if err != nil {
		return err
	}
	updatedRoute := currentRoute
	updatedRoute.Attributes.Set(attributeValue)

	err = rs.updateRoute(updatedRoute, who)
	rs.changelog.Update(currentRoute, updatedRoute, who)
	return err
}

// DeleteAttribute removes an attribute of an route
func (rs *RouteService) DeleteAttribute(routeName, attributeToDelete string,
	who Requester) (string, types.Error) {

	currentRoute, err := rs.db.Route.Get(routeName)
	if err != nil {
		return "", err
	}
	updatedRoute := currentRoute
	oldValue, err := updatedRoute.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}

	err = rs.updateRoute(updatedRoute, who)
	rs.changelog.Update(currentRoute, updatedRoute, who)
	return oldValue, err
}

// updateRoute updates last-modified field(s) and updates route in database
func (rs *RouteService) updateRoute(updatedRoute *types.Route, who Requester) types.Error {

	updatedRoute.Attributes.Tidy()
	updatedRoute.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedRoute.LastmodifiedBy = who.User
	return rs.db.Route.Update(updatedRoute)
}

// Delete deletes an route
func (rs *RouteService) Delete(routeName string, who Requester) (
	deletedRoute types.Route, e types.Error) {

	route, err := rs.db.Route.Get(routeName)
	if err != nil {
		return types.NullRoute, err
	}
	err = rs.db.Route.Delete(routeName)
	if err != nil {
		return types.NullRoute, err
	}
	rs.changelog.Delete(route, who)
	return *route, nil
}
