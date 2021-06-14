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

// UpdateAttributes updates attributes of an listener
func (ls *ListenerService) UpdateAttributes(listenerName string,
	receivedAttributes types.Attributes, who Requester) types.Error {

	currentListener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return err
	}
	updatedListener := currentListener
	if err = updatedListener.Attributes.SetMultiple(receivedAttributes); err != nil {
		return err
	}

	if err = ls.updateListener(updatedListener, who); err != nil {
		return err
	}
	ls.changelog.Update(currentListener, updatedListener, who)
	return nil
}

// UpdateAttribute update an attribute of developer
func (ls *ListenerService) UpdateAttribute(listenerName string,
	attributeValue types.Attribute, who Requester) types.Error {

	currentListener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return err
	}
	updatedListener := currentListener
	if err := updatedListener.Attributes.Set(attributeValue); err != nil {
		return err
	}

	if err = ls.updateListener(updatedListener, who); err != nil {
		return err
	}
	ls.changelog.Update(currentListener, updatedListener, who)
	return nil
}

// DeleteAttribute removes an attribute of an listener
func (ls *ListenerService) DeleteAttribute(listenerName, attributeToDelete string,
	who Requester) (string, types.Error) {

	currentListener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return "", err
	}
	updatedListener := currentListener
	oldValue, err := updatedListener.Attributes.Delete(attributeToDelete)
	if err != nil {
		return "", err
	}

	if err = ls.updateListener(updatedListener, who); err != nil {
		return "", err
	}
	ls.changelog.Update(currentListener, updatedListener, who)
	return oldValue, nil
}

// updateListener updates last-modified field(s) and updates cluster in database
func (ls *ListenerService) updateListener(updatedListener *types.Listener, who Requester) types.Error {

	updatedListener.Attributes.Tidy()
	updatedListener.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedListener.LastmodifiedBy = who.User
	return ls.db.Listener.Update(updatedListener)
}

// Delete deletes an listener
func (ls *ListenerService) Delete(listenerName string, who Requester) (
	deletedListener types.Listener, e types.Error) {

	listener, err := ls.db.Listener.Get(listenerName)
	if err != nil {
		return types.NullListener, err
	}
	if err = ls.db.Listener.Delete(listenerName); err != nil {
		return types.NullListener, err
	}
	ls.changelog.Delete(listener, who)
	return *listener, nil
}
