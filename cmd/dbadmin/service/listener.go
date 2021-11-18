package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// ListenerService is
type ListenerService struct {
	db        *db.Database
	changelog *Changelog
}

// NewListener returns a new listener instance
func NewListener(database *db.Database, c *Changelog) *ListenerService {

	return &ListenerService{
		db:        database,
		changelog: c,
	}
}

// GetAll returns all listeners
func (ls *ListenerService) GetAll() (listeners types.Listeners, err types.Error) {

	return ls.db.Listener.GetAll()
}

// Get returns details of an listener
func (ls *ListenerService) Get(listenerName string) (listener *types.Listener, err types.Error) {

	return ls.db.Listener.Get(listenerName)
}

// Create creates an listener
func (ls *ListenerService) Create(newListener types.Listener, who Requester) (types.Listener, types.Error) {

	if _, err := ls.db.Listener.Get(newListener.Name); err == nil {
		return types.NullListener, types.NewBadRequestError(
			fmt.Errorf("listener '%s' already exists", newListener.Name))
	}
	// Automatically set default fields
	newListener.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newListener.CreatedBy = who.User

	if err := ls.updateListener(&newListener, who); err != nil {
		return types.NullListener, err
	}
	ls.changelog.Create(newListener, who)
	return newListener, nil
}

// Update updates an existing listener
func (ls *ListenerService) Update(updatedListener types.Listener, who Requester) (types.Listener, types.Error) {

	currentListener, err := ls.db.Listener.Get(updatedListener.Name)
	if err != nil {
		return types.NullListener, err
	}
	// Copy over fields we do not allow to be updated
	updatedListener.Name = currentListener.Name
	updatedListener.CreatedAt = currentListener.CreatedAt
	updatedListener.CreatedBy = currentListener.CreatedBy

	if err = ls.updateListener(&updatedListener, who); err != nil {
		return types.NullListener, err
	}
	ls.changelog.Update(currentListener, updatedListener, who)
	return updatedListener, nil
}

// updateListener updates last-modified field(s) and updates cluster in database
func (ls *ListenerService) updateListener(updatedListener *types.Listener, who Requester) types.Error {

	updatedListener.Attributes.Tidy()
	updatedListener.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedListener.LastModifiedBy = who.User

	if err := updatedListener.Validate(); err != nil {
		return types.NewBadRequestError(err)
	}
	return ls.db.Listener.Update(updatedListener)
}

// Delete deletes an listener
func (ls *ListenerService) Delete(listenerName string, who Requester) (e types.Error) {

	listener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return err
	}
	if err = ls.db.Listener.Delete(listenerName); err != nil {
		return err
	}
	ls.changelog.Delete(listener, who)
	return nil
}
