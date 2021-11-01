package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// registerDeveloperRoutes registers all routes we handle
func (h *Handler) registerDeveloperRoutes(r *gin.RouterGroup) {
	r.GET("/developers", h.handler(h.getAllDevelopers))
	r.POST("/developers", h.handler(h.createDeveloper))

	r.GET("/developers/:developer", h.handler(h.getDeveloper))
	r.POST("/developers/:developer", h.handler(h.updateDeveloper))
	r.DELETE("/developers/:developer", h.handler(h.deleteDeveloper))

	r.GET("/developers/:developer/attributes", h.handler(h.getDeveloperAttributes))
	r.POST("/developers/:developer/attributes", h.handler(h.updateDeveloperAttributes))

	r.GET("/developers/:developer/attributes/:attribute", h.handler(h.getDeveloperAttributeByName))
	r.POST("/developers/:developer/attributes/:attribute", h.handler(h.updateDeveloperAttributeByName))
	r.DELETE("/developers/:developer/attributes/:attribute", h.handler(h.deleteDeveloperAttributeByName))
}

const (
	// Name of developer parameter in the route definition
	developerParameter = "developer"

	// 	// Name of attribute parameter in the route definition
	attributeParameter = "attribute"
)

// getAllDevelopers returns all developers
// FIXME: add pagination support
func (h *Handler) getAllDevelopers(c *gin.Context) handlerResponse {

	developers, err := h.service.Developer.GetAll()
	if err != nil {
		return handleError(err)
	}
	if c.Query("expand") == "true" {
		return handleOK(StringMap{"developer": developers})
	}
	emailAddresses := make([]string, len(developers)-1)
	for _, developer := range developers {
		emailAddresses = append(emailAddresses, developer.Email)
	}
	return handleOK(emailAddresses)
}

// getDeveloper returns full details of one developer
func (h *Handler) getDeveloper(c *gin.Context) handlerResponse {

	developer, err := h.service.Developer.Get(c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(developer)
}

// getDeveloperAttributes returns attributes of a developer
func (h *Handler) getDeveloperAttributes(c *gin.Context) handlerResponse {

	developer, err := h.service.Developer.Get(c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOKAttributes(developer.Attributes)
}

// getDeveloperAttributeByName returns one particular attribute of a developer
func (h *Handler) getDeveloperAttributeByName(c *gin.Context) handlerResponse {

	developer, err := h.service.Developer.Get(c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	attributeValue, err := developer.Attributes.Get(c.Param(attributeParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(attributeValue)
}

// createDeveloper creates a new developer
func (h *Handler) createDeveloper(c *gin.Context) handlerResponse {

	var newDeveloper types.Developer
	if err := c.ShouldBindJSON(&newDeveloper); err != nil {
		return handleBadRequest(err)
	}
	storedDeveloper, err := h.service.Developer.Create(newDeveloper, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleCreated(storedDeveloper)
}

// updateDeveloper updates an existing developer
func (h *Handler) updateDeveloper(c *gin.Context) handlerResponse {

	var updatedDeveloper types.Developer
	if err := c.ShouldBindJSON(&updatedDeveloper); err != nil {
		return handleBadRequest(err)
	}
	// developer name in path must match developer name in posted body
	if updatedDeveloper.Email != c.Param(developerParameter) {
		return handleNameMismatch()
	}
	storedDeveloper, err := h.service.Developer.Update(updatedDeveloper, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedDeveloper)
}

// updateDeveloperAttributes updates attributes of developer
func (h *Handler) updateDeveloperAttributes(c *gin.Context) handlerResponse {

	var receivedAttributes struct {
		Attributes types.Attributes `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		return handleBadRequest(err)
	}
	if err := h.service.Developer.UpdateAttributes(c.Param(developerParameter),
		receivedAttributes.Attributes, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttributes(receivedAttributes.Attributes)
}

// updateDeveloperAttributeByName update an attribute of developer
func (h *Handler) updateDeveloperAttributeByName(c *gin.Context) handlerResponse {

	var receivedValue types.AttributeValue
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		return handleBadRequest(err)
	}
	newAttribute := types.Attribute{
		Name:  c.Param(attributeParameter),
		Value: receivedValue.Value,
	}
	if err := h.service.Developer.UpdateAttribute(c.Param(developerParameter),
		newAttribute, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttribute(newAttribute)
}

// deleteDeveloperAttributeByName removes an attribute of developer
func (h *Handler) deleteDeveloperAttributeByName(c *gin.Context) handlerResponse {

	attributeToDelete := c.Param(attributeParameter)
	oldValue, err := h.service.Developer.DeleteAttribute(c.Param(developerParameter),
		attributeToDelete, h.who(c))
	if err != nil {
		return handleBadRequest(err)
	}
	return handleOKAttribute(types.Attribute{
		Name:  attributeToDelete,
		Value: oldValue,
	})
}

// deleteDeveloper deletes of one developer
func (h *Handler) deleteDeveloper(c *gin.Context) handlerResponse {

	developer, err := h.service.Developer.Delete(c.Param(developerParameter), h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(developer)
}
