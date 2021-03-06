package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

func (h *Handler) registerDeveloperAppRoutes(r *gin.RouterGroup) {
	r.GET("/apps", h.handler(h.getAllDevelopersApps))
	r.GET("/apps/:application", h.handler(h.getDeveloperAppByName))

	r.GET("/developers/:developer/apps", h.handler(h.getDeveloperAppsByDeveloperEmail))
	r.POST("/developers/:developer/apps", h.handler(h.createDeveloperApp))

	r.GET("/developers/:developer/apps/:application", h.handler(h.getDeveloperAppByName))
	r.POST("/developers/:developer/apps/:application", h.handler(h.updateDeveloperApp))
	r.DELETE("/developers/:developer/apps/:application", h.handler(h.deleteDeveloperAppByName))

	r.GET("/developers/:developer/apps/:application/attributes", h.handler(h.getDeveloperAppAttributes))
	r.POST("/developers/:developer/apps/:application/attributes", h.handler(h.updateDeveloperAppAttributes))

	r.GET("/developers/:developer/apps/:application/attributes/:attribute", h.handler(h.getDeveloperAppAttributeByName))
	r.POST("/developers/:developer/apps/:application/attributes/:attribute", h.handler(h.updateDeveloperAppAttributeByName))
	r.DELETE("/developers/:developer/apps/:application/attributes/:attribute", h.handler(h.deleteDeveloperAppAttributeByName))
}

const (
	// Name of developer parameter in the route definition
	developerAppParameter = "application"
)

// getAllDevelopersApps returns all developer apps
// FIXME: add pagination support
func (h *Handler) getAllDevelopersApps(c *gin.Context) handlerResponse {

	developerapps, err := h.service.DeveloperApp.GetAll()
	if err != nil {
		return handleError(err)
	}

	// Should we return an array with names or full details?
	if c.Query("expand") == "true" {
		return handleOK(StringMap{"apps": developerapps})
	}

	var developerAppNames []string
	for _, v := range developerapps {
		developerAppNames = append(developerAppNames, v.Name)
	}
	return handleOK(developerAppNames)
}

// getDeveloperAppsByDeveloperEmail returns apps of a developer
func (h *Handler) getDeveloperAppsByDeveloperEmail(c *gin.Context) handlerResponse {

	developerapps, err := h.service.DeveloperApp.GetAllByEmail(c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(StringMap{"apps": developerapps})
}

// getDeveloperAppByName returns one named app of a developer
func (h *Handler) getDeveloperAppByName(c *gin.Context) handlerResponse {

	developerApp, err := h.service.DeveloperApp.GetByName(c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(developerApp)
}

// getDeveloperAppAttributes returns attributes of a developer
func (h *Handler) getDeveloperAppAttributes(c *gin.Context) handlerResponse {

	developerApp, err := h.service.DeveloperApp.GetByName(c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOKAttributes(developerApp.Attributes)
}

// getDeveloperAppAttributeByName returns one particular attribute of a developer
func (h *Handler) getDeveloperAppAttributeByName(c *gin.Context) handlerResponse {

	developerApp, err := h.service.DeveloperApp.GetByName(c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}
	attributeValue, err := developerApp.Attributes.Get(c.Param(attributeParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(attributeValue)
}

// createDeveloperApp creates a developer application
func (h *Handler) createDeveloperApp(c *gin.Context) handlerResponse {

	var newDeveloperApp types.DeveloperApp
	if err := c.ShouldBindJSON(&newDeveloperApp); err != nil {
		return handleBadRequest(err)
	}
	developer, err := h.service.Developer.Get(c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	if _, err := h.service.DeveloperApp.GetByName(newDeveloperApp.Name); err == nil {
		return handleError(types.NewBadRequestError(
			fmt.Errorf("developer app '%s' already exists", newDeveloperApp.Name)))
	}
	storedDeveloperApp, err := h.service.DeveloperApp.Create(developer.Email, newDeveloperApp, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleCreated(storedDeveloperApp)
}

// updateDeveloperApp updates an existing developer
func (h *Handler) updateDeveloperApp(c *gin.Context) handlerResponse {

	var updateRequest types.DeveloperApp
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		return handleBadRequest(err)
	}
	// developer name in path must match developer name in posted body
	if updateRequest.Name != c.Param(developerAppParameter) {
		return handleNameMismatch()
	}
	_, err := h.service.Developer.Get(c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	storedDeveloperApp, err := h.service.DeveloperApp.Update(updateRequest, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleCreated(storedDeveloperApp)
}

// updateDeveloperAppAttributes updates attribute of one particular app
func (h *Handler) updateDeveloperAppAttributes(c *gin.Context) handlerResponse {

	var receivedAttributes struct {
		Attributes types.Attributes `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		return handleBadRequest(err)
	}
	_, err := h.service.Developer.Get(c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	developerAppToUpdate, err := h.service.DeveloperApp.GetByName(c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}
	if err := h.service.DeveloperApp.UpdateAttributes(developerAppToUpdate.Name,
		receivedAttributes.Attributes, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttributes(developerAppToUpdate.Attributes)
}

// updateDeveloperAppAttributeByName update an attribute of developer
func (h *Handler) updateDeveloperAppAttributeByName(c *gin.Context) handlerResponse {

	var receivedValue types.AttributeValue
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		return handleBadRequest(err)
	}
	newAttribute := types.Attribute{
		Name:  c.Param(attributeParameter),
		Value: receivedValue.Value,
	}
	if err := h.service.DeveloperApp.UpdateAttribute(c.Param(developerAppParameter),
		newAttribute, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttribute(newAttribute)
}

// deleteDeveloperAppAttributeByName removes an attribute of developer
func (h *Handler) deleteDeveloperAppAttributeByName(c *gin.Context) handlerResponse {

	_, err := h.service.Developer.Get(c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	attributeToDelete := c.Param(attributeParameter)
	oldValue, err := h.service.DeveloperApp.DeleteAttribute(c.Param(developerAppParameter),
		attributeToDelete, h.who(c))
	if err != nil {
		return handleBadRequest(err)
	}
	return handleOKAttribute(types.Attribute{
		Name:  attributeToDelete,
		Value: oldValue,
	})
}

// deleteDeveloperAppByName deletes a developer app
func (h *Handler) deleteDeveloperAppByName(c *gin.Context) handlerResponse {

	developer, err := h.service.Developer.Get(c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	developerApp, err := h.service.DeveloperApp.Delete(developer.DeveloperID,
		c.Param(developerAppParameter), h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(developerApp)
}
