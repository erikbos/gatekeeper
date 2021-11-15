package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// returns all listeners
// (GET /v1/listeners)
func (h *Handler) GetV1Listeners(c *gin.Context) {

	listeners, err := h.service.Listener.GetAll()
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseListenerss(c, listeners)
}

// creates a new listener
// (POST /v1/listeners)
func (h *Handler) PostV1Listeners(c *gin.Context) {

	var receivedListener Listener
	if err := c.ShouldBindJSON(&receivedListener); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newListener := fromListener(receivedListener)
	createdDeveloper, err := h.service.Listener.Create(newListener, h.who(c))
	if err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	h.responseListenerCreated(c, &createdDeveloper)
}

// deletes an listener
// (DELETE /v1/listeners/{listener_name})
func (h *Handler) DeleteV1ListenersListenerName(c *gin.Context, listenerName ListenerName) {

	listener, err := h.service.Listener.Delete(string(listenerName), h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseListeners(c, &listener)
}

// returns full details of one listener
// (GET /v1/listeners/{listener_name})
func (h *Handler) GetV1ListenersListenerName(c *gin.Context, listenerName ListenerName) {

	listener, err := h.service.Listener.Get(string(listenerName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseListeners(c, listener)
}

// (POST /v1/listeners/{listener_name})
func (h *Handler) PostV1ListenersListenerName(c *gin.Context, listenerName ListenerName) {

	var receivedListener Listener
	if err := c.ShouldBindJSON(&receivedListener); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	updatedListener := fromListener(receivedListener)
	storedListener, err := h.service.Listener.Update(updatedListener, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseListenersUpdated(c, &storedListener)
}

// returns attributes of a listener
// (GET /v1/listeners/{listener_name}/attributes)
func (h *Handler) GetV1ListenersListenerNameAttributes(c *gin.Context, listenerName ListenerName) {

	listener, err := h.service.Listener.Get(string(listenerName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, listener.Attributes)
}

// replaces attributes of an listener
// (POST /v1/listeners/{listener_name}/attributes)
func (h *Handler) PostV1ListenersListenerNameAttributes(c *gin.Context, listenerName ListenerName) {

	var receivedAttributes Attributes
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	attributes := fromAttributesRequest(receivedAttributes.Attribute)
	if err := h.service.Listener.UpdateAttributes(
		string(listenerName), attributes, h.who(c)); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	h.responseAttributes(c, attributes)
}

// deletes one attribute of an listener
// (DELETE /v1/listeners/{listener_name}/attributes/{attribute_name})
func (h *Handler) DeleteV1ListenersListenerNameAttributesAttributeName(c *gin.Context, listenerName ListenerName, attributeName AttributeName) {

	oldValue, err := h.service.Listener.DeleteAttribute(
		string(listenerName), string(attributeName), h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeDeleted(c, types.NewAttribute(string(attributeName), oldValue))
}

// returns one attribute of an listener
// (GET /v1/listeners/{listener_name}/attributes/{attribute_name})
func (h *Handler) GetV1ListenersListenerNameAttributesAttributeName(c *gin.Context, listenerName ListenerName, attributeName AttributeName) {

	listener, err := h.service.Listener.Get(string(listenerName))
	if err != nil {
		responseError(c, err)
		return
	}
	attributeValue, err := listener.Attributes.Get(string(attributeName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeRetrieved(c, types.NewAttribute(string(attributeName), attributeValue))
}

// updates an attribute of an listener
// (POST /v1/listeners/{listener_name}/attributes/{attribute_name})
func (h *Handler) PostV1ListenersListenerNameAttributesAttributeName(c *gin.Context, listenerName ListenerName, attributeName AttributeName) {

	var receivedValue Attribute
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newAttribute := types.NewAttribute(string(attributeName), *receivedValue.Value)

	if err := h.service.Listener.UpdateAttribute(
		string(listenerName), *newAttribute, h.who(c)); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	h.responseAttributeUpdated(c, newAttribute)
}

// API responses

func (h *Handler) responseListenerss(c *gin.Context, listeners types.Listeners) {

	all_listeners := make([]Listener, len(listeners))
	for i := range listeners {
		all_listeners[i] = h.ToListenerResponse(&listeners[i])
	}
	c.IndentedJSON(http.StatusOK, Listeners{
		Listeners: &all_listeners,
	})
}

func (h *Handler) responseListeners(c *gin.Context, listener *types.Listener) {

	c.IndentedJSON(http.StatusOK, h.ToListenerResponse(listener))
}

func (h *Handler) responseListenerCreated(c *gin.Context, listener *types.Listener) {

	c.IndentedJSON(http.StatusCreated, h.ToListenerResponse(listener))
}

func (h *Handler) responseListenersUpdated(c *gin.Context, listener *types.Listener) {

	c.IndentedJSON(http.StatusOK, h.ToListenerResponse(listener))
}

// type conversion

func (h *Handler) ToListenerResponse(l *types.Listener) Listener {

	lis := Listener{
		Attributes:     toAttributesResponse(l.Attributes),
		CreatedAt:      &l.CreatedAt,
		CreatedBy:      &l.CreatedBy,
		DisplayName:    &l.DisplayName,
		LastModifiedBy: &l.LastModifiedBy,
		LastModifiedAt: &l.LastModifiedAt,
		Name:           l.Name,
		Policies:       &l.Policies,
		Port:           &l.Port,
		RouteGroup:     &l.RouteGroup,
	}
	if l.VirtualHosts != nil {
		lis.VirtualHosts = &l.VirtualHosts
	} else {
		lis.VirtualHosts = &[]string{}
	}
	return lis
}

func fromListener(l Listener) types.Listener {

	listener := types.Listener{}
	if l.Attributes != nil {
		listener.Attributes = fromAttributesRequest(l.Attributes)
	}
	if l.CreatedAt != nil {
		listener.CreatedAt = *l.CreatedAt
	}
	if l.CreatedBy != nil {
		listener.CreatedBy = *l.CreatedBy
	}
	if l.DisplayName != nil {
		listener.DisplayName = *l.DisplayName
	}
	if l.LastModifiedBy != nil {
		listener.LastModifiedBy = *l.LastModifiedBy
	}
	if l.LastModifiedAt != nil {
		listener.LastModifiedAt = *l.LastModifiedAt
	}
	if l.Name != "" {
		listener.Name = l.Name
	}
	if l.Policies != nil {
		listener.Policies = *l.Policies
	}
	if l.Port != nil {
		listener.Port = *l.Port
	}
	if l.RouteGroup != nil {
		listener.RouteGroup = *l.RouteGroup
	}
	if l.VirtualHosts != nil {
		listener.VirtualHosts = *l.VirtualHosts
	} else {
		listener.VirtualHosts = []string{}
	}
	return listener
}
