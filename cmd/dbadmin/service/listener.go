package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// ListenerService is
type ListenerService struct {
	db *db.Database
}

// NewListenerService returns a new listener instance
func NewListenerService(database *db.Database) *ListenerService {

	return &ListenerService{db: database}
}

// GetAll returns all listeners
func (ls *ListenerService) GetAll() (listeners types.Listeners, err types.Error) {

	return ls.db.Listener.GetAll()
}

// Get returns details of an listener
func (ls *ListenerService) Get(listenerName string) (listener *types.Listener, err types.Error) {

	return ls.db.Listener.Get(listenerName)
}

// GetAttributes returns attributes of an listener
func (ls *ListenerService) GetAttributes(listenerName string) (attributes *types.Attributes, err types.Error) {

	listener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return nil, err
	}
	return &listener.Attributes, nil
}

// GetAttribute returns one particular attribute of an listener
func (ls *ListenerService) GetAttribute(listenerName, attributeName string) (value string, err types.Error) {

	listener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return "", err
	}
	return listener.Attributes.Get(attributeName)
}

// Create creates an listener
func (ls *ListenerService) Create(newListener types.Listener) (types.Listener, types.Error) {

	existingListener, err := ls.db.Listener.Get(newListener.Name)
	if err == nil {
		return types.NullListener, types.NewBadRequestError(
			fmt.Errorf("Listener '%s' already exists", existingListener.Name))
	}
	// Automatically set default fields
	newListener.CreatedAt = shared.GetCurrentTimeMilliseconds()
	err = ls.updateListener(&newListener)
	return newListener, err
}

// Update updates an existing listener
func (ls *ListenerService) Update(updatedListener types.Listener) (types.Listener, types.Error) {

	listenerToUpdate, err := ls.db.Listener.Get(updatedListener.Name)
	if err != nil {
		return types.NullListener, types.NewItemNotFoundError(err)
	}

	// Copy over fields we allow to be updated
	listenerToUpdate.VirtualHosts = updatedListener.VirtualHosts
	listenerToUpdate.Port = updatedListener.Port
	listenerToUpdate.DisplayName = updatedListener.DisplayName
	listenerToUpdate.Attributes = updatedListener.Attributes
	listenerToUpdate.RouteGroup = updatedListener.RouteGroup
	listenerToUpdate.Policies = updatedListener.Policies
	listenerToUpdate.OrganizationName = updatedListener.OrganizationName

	err = ls.updateListener(listenerToUpdate)
	return *listenerToUpdate, err
}

// UpdateAttributes updates attributes of an listener
func (ls *ListenerService) UpdateAttributes(listenerName string, receivedAttributes types.Attributes) types.Error {

	updatedListener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return types.NewItemNotFoundError(err)
	}
	updatedListener.Attributes = receivedAttributes
	return ls.updateListener(updatedListener)
}

// UpdateAttribute update an attribute of developer
func (ls *ListenerService) UpdateAttribute(listenerName string,
	attributeValue types.Attribute) types.Error {

	updatedListener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return err
	}
	updatedListener.Attributes.Set(attributeValue)
	return ls.updateListener(updatedListener)
}

// DeleteAttribute removes an attribute of an listener
func (ls *ListenerService) DeleteAttribute(listenerName, attributeToDelete string) (string, types.Error) {

	updatedListener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return "", err
	}
	oldValue, err := updatedListener.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}
	return oldValue, ls.updateListener(updatedListener)
}

// updateListener updates last-modified field(s) and updates cluster in database
func (ls *ListenerService) updateListener(updatedListener *types.Listener) types.Error {

	updatedListener.Attributes.Tidy()
	updatedListener.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	return ls.db.Listener.Update(updatedListener)
}

// Delete deletes an listener
func (ls *ListenerService) Delete(listenerName string) (deletedListener types.Listener, e types.Error) {

	listener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return types.NullListener, err
	}
	attachedRouteCount := ls.countRoutesOfRouteGroup(listener.RouteGroup)
	if attachedRouteCount > 0 {
		return types.NullListener, types.NewBadRequestError(
			fmt.Errorf("Cannot delete listener '%s' with %d routes attached",
				listener.Name, attachedRouteCount))
	}

	err = ls.db.Listener.Delete(listenerName)
	if err != nil {
		return types.NullListener, err
	}
	return *listener, nil
}

// counts number of routes with a specific routegroup
func (ls *ListenerService) countRoutesOfRouteGroup(routeGroup string) int {

	routes, err := ls.db.Route.GetAll()
	if err != nil {
		return 0
	}
	var count int
	for _, route := range routes {
		if route.RouteGroup == routeGroup {
			count++
		}
	}
	return count
}
