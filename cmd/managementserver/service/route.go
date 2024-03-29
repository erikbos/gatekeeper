package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/cmd/managementserver/audit"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// RouteService is
type RouteService struct {
	db    *db.Database
	audit *audit.Audit
}

// NewRoute returns a new route instance
func NewRoute(database *db.Database, a *audit.Audit) *RouteService {

	return &RouteService{
		db:    database,
		audit: a,
	}
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
func (rs *RouteService) Create(newRoute types.Route, who audit.Requester) (*types.Route, types.Error) {

	if _, err := rs.db.Route.Get(newRoute.Name); err == nil {
		return nil, types.NewBadRequestError(
			fmt.Errorf("route '%s' already exists", newRoute.Name))
	}
	// Automatically set default fields
	newRoute.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newRoute.CreatedBy = who.User

	if err := rs.updateRoute(&newRoute, who); err != nil {
		return nil, err
	}
	rs.audit.Create(newRoute, nil, who)
	return &newRoute, nil
}

// Update updates an existing route
func (rs *RouteService) Update(updatedRoute types.Route,
	who audit.Requester) (*types.Route, types.Error) {

	currentRoute, err := rs.db.Route.Get(updatedRoute.Name)
	if err != nil {
		return nil, err
	}
	// Copy over fields we do not allow to be updated
	updatedRoute.Name = currentRoute.Name
	updatedRoute.CreatedAt = currentRoute.CreatedAt
	updatedRoute.CreatedBy = currentRoute.CreatedBy

	if err = rs.updateRoute(&updatedRoute, who); err != nil {
		return nil, err
	}
	rs.audit.Update(currentRoute, updatedRoute, nil, who)
	return &updatedRoute, nil
}

// updateRoute updates last-modified field(s) and updates route in database
func (rs *RouteService) updateRoute(updatedRoute *types.Route, who audit.Requester) types.Error {

	updatedRoute.Attributes.Tidy()
	updatedRoute.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedRoute.LastModifiedBy = who.User

	if err := updatedRoute.Validate(); err != nil {
		return types.NewBadRequestError(err)
	}
	return rs.db.Route.Update(updatedRoute)
}

// Delete deletes an route
func (rs *RouteService) Delete(routeName string, who audit.Requester) (e types.Error) {

	route, err := rs.db.Route.Get(routeName)
	if err != nil {
		return err
	}
	err = rs.db.Route.Delete(routeName)
	if err != nil {
		return err
	}
	rs.audit.Delete(route, nil, who)
	return nil
}
