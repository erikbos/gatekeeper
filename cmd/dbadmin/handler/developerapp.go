package handler

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/erikbos/gatekeeper/pkg/types"
)

func (h *Handler) registerDeveloperAppRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/apps", h.handler(h.getAllDevelopersApps))
	r.GET("/v1/organizations/:organization/apps/:application", h.handler(h.getByName))

	r.GET("/v1/organizations/:organization/developers/:developer/apps", h.handler(h.getDeveloperAppsByDeveloperEmail))
	r.POST("/v1/organizations/:organization/developers/:developer/apps", h.handler(h.postCreateDeveloperApp))

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application", h.handler(h.getByName))
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application", h.handler(h.postDeveloperApp))
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application", h.handler(h.deleteDeveloperAppByName))

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes", h.handler(h.getDeveloperAppAttributes))
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/attributes", h.handler(h.updateDeveloperAppAttributes))

	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", h.handler(h.getDeveloperAppAttributeByName))
	r.POST("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", h.handler(h.postDeveloperAppAttributeByName))
	r.DELETE("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", h.handler(h.deleteDeveloperAppAttributeByName))
}

const (
	// Name of developer parameter in the route definition
	developerAppParameter = "application"
)

// getAllDevelopersApps returns all developers in organization
// FIXME: add pagination support
func (h *Handler) getAllDevelopersApps(c *gin.Context) handlerResponse {

	developerapps, err := h.service.DeveloperApp.GetByOrganization(c.Param(organizationParameter))
	if err != nil {
		return handleError(err)
	}

	// Should we return an array with names or full details?
	if c.Query("expand") == "true" {
		return handleOK(gin.H{"apps": developerapps})
	}

	var developerAppNames []string
	for _, v := range developerapps {
		developerAppNames = append(developerAppNames, v.Name)
	}
	return handleOK(developerAppNames)
}

// getDeveloperAppsByDeveloperEmail returns apps of a developer
func (h *Handler) getDeveloperAppsByDeveloperEmail(c *gin.Context) handlerResponse {

	developer, err := h.service.Developer.Get(c.Param(organizationParameter), c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(developer.Apps)
}

// getByName returns one named app of a developer
func (h *Handler) getByName(c *gin.Context) handlerResponse {

	developerApp, err := h.service.DeveloperApp.GetByName(c.Param(organizationParameter), c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(developerApp)
}

// getDeveloperAppAttributes returns attributes of a developer
func (h *Handler) getDeveloperAppAttributes(c *gin.Context) handlerResponse {

	developerApp, err := h.service.DeveloperApp.GetByName(c.Param(organizationParameter), c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOKAttributes(developerApp.Attributes)
}

// getDeveloperAppAttributeByName returns one particular attribute of a developer
func (h *Handler) getDeveloperAppAttributeByName(c *gin.Context) handlerResponse {

	developerApp, err := h.service.DeveloperApp.GetByName(c.Param(organizationParameter), c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}
	attributeValue, err := developerApp.Attributes.Get(c.Param(attributeParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(attributeValue)
}

// PostCreateDeveloperApp creates a developer application
func (h *Handler) postCreateDeveloperApp(c *gin.Context) handlerResponse {

	var newDeveloperApp types.DeveloperApp
	if err := c.ShouldBindJSON(&newDeveloperApp); err != nil {
		return handleBadRequest(err)
	}

	_, err := h.service.Developer.Get(c.Param(organizationParameter), c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	existingDeveloperApp, err := h.service.DeveloperApp.GetByName(organizationParameter, newDeveloperApp.Name)
	if err == nil {
		return handleError(types.NewBadRequestError(
			fmt.Errorf("Developer app '%s' already exists", existingDeveloperApp.Name)))
	}

	storedDeveloperApp, err := h.service.DeveloperApp.Update(c.Param(organizationParameter), newDeveloperApp)
	if err != nil {
		return handleError(err)
	}
	return handleCreated(storedDeveloperApp)
}

// postDeveloperApp updates an existing developer
func (h *Handler) postDeveloperApp(c *gin.Context) handlerResponse {

	var updateRequest types.DeveloperApp
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		return handleBadRequest(err)
	}

	_, err := h.service.Developer.Get(c.Param(organizationParameter), c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	_, err = h.service.DeveloperApp.GetByName(organizationParameter, c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}

	storedDeveloperApp, err := h.service.DeveloperApp.Update(c.Param(organizationParameter), updateRequest)
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

	if len(receivedAttributes.Attributes) == 0 {
		return handleBadRequest(errors.New("No attributes posted"))
	}

	_, err := h.service.Developer.Get(c.Param(organizationParameter), c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	developerAppToUpdate, err := h.service.DeveloperApp.GetByName(organizationParameter, c.Param(developerAppParameter))
	if err != nil {
		return handleError(err)
	}

	developerAppToUpdate.Attributes = receivedAttributes.Attributes
	developerAppToUpdate.LastmodifiedBy = h.GetSessionUser(c)

	_, err = h.service.DeveloperApp.Update(c.Param(organizationParameter), *developerAppToUpdate)
	if err != nil {
		return handleError(err)
	}
	return handleOKAttributes(developerAppToUpdate.Attributes)
}

// postDeveloperAppAttributeByName update an attribute of developer
func (h *Handler) postDeveloperAppAttributeByName(c *gin.Context) handlerResponse {

	var receivedValue struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		return handleBadRequest(err)

	}

	newAttribute := types.Attribute{
		Name:  c.Param(attributeParameter),
		Value: receivedValue.Value,
	}

	// FIXME we should set LastmodifiedBy
	if err := h.service.DeveloperApp.UpdateAttribute(c.Param(organizationParameter),
		c.Param(developerAppParameter), newAttribute); err != nil {
		return handleError(err)
	}
	return handleOKAttribute(newAttribute)
}

// deleteDeveloperAppAttributeByName removes an attribute of developer
func (h *Handler) deleteDeveloperAppAttributeByName(c *gin.Context) handlerResponse {

	_, err := h.service.Developer.Get(c.Param(organizationParameter), c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}

	attributeToDelete := c.Param(attributeParameter)
	oldValue, err := h.service.DeveloperApp.DeleteAttribute(c.Param(organizationParameter), c.Param(developerAppParameter),
		attributeToDelete)
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

	developer, err := h.service.Developer.Get(c.Param(organizationParameter), c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}

	developerApp, err := h.service.DeveloperApp.Delete(c.Param(organizationParameter),
		developer.DeveloperID, c.Param(developerParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(developerApp)

	// AppCredentialCount := h.db.Credential.GetCountByDeveloperAppID(developerApp.AppID)
	// AppCredentialCount := 1
	// if AppCredentialCount == -1 {
	// 	returnJSONMessage(c, http.StatusInternalServerError,
	// 		fmt.Errorf("Could not retrieve number of api keys of developer app '%s'",
	// 			developerApp.Name))
	// 	return
	// }
	// if AppCredentialCount > 0 {
	// 	returnJSONMessage(c, http.StatusForbidden,
	// 		fmt.Errorf("Cannot delete developer app '%s' with %d apikeys",
	// 			developerApp.Name, AppCredentialCount))
	// 	return
	// }
	// err = h.db.DeveloperApp.DeleteByID(developerApp.OrganizationName, developerApp.AppID)
	// if err != nil {
	// 	returnJSONMessage(c, http.StatusServiceUnavailable, err)
	// 	return
	// }

	// // Remove app from the apps field in developer entry as well
	// for i := 0; i < len(developer.Apps); i++ {
	// 	if developer.Apps[i] == c.Param("application") {
	// 		developer.Apps = append(developer.Apps[:i], developer.Apps[i+1:]...)
	// 		i-- // form the remove item index to start iterate next item
	// 	}
	// }

	// developerApp.LastmodifiedBy = h.GetSessionUser(c)
	// if err := h.db.Developer.UpdateByName(developer); err != nil {
	// 	returnJSONMessage(c, http.StatusBadRequest, err)
	// 	return
	// }
	// c.IndentedJSON(http.StatusOK, developerApp)
}

// generateAppID creates unique primary key for developer app row
func generateAppID() string {
	return (uuid.New().String())
}
