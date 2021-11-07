package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// returns all developer apps
// (GET /v1/organizations/{organization_name}/apps)
func (h *Handler2) GetV1OrganizationsOrganizationNameApps(c *gin.Context, organizationName OrganizationName, params GetV1OrganizationsOrganizationNameAppsParams) {

	apps, err := h.service.DeveloperApp.GetAll()
	if err != nil {
		h.responseError(c, err)
		return
	}
	// Do we have to return full developer details?
	if params.Expand != nil && *params.Expand {
		h.responseDeveloperAllApps(c, apps)
		return
	}
	h.responseDeveloperAppIDs(c, apps)
}

// (GET /v1/organizations/{organization_name}/apps/{app_id})
func (h *Handler2) GetV1OrganizationsOrganizationNameAppsAppId(c *gin.Context, organizationName OrganizationName, appId AppId) {

	app, err := h.service.DeveloperApp.GetByID(string(appId))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseApplication(c, app)
}

// returns one named app of a developer
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps)
func (h *Handler2) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressApps(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, params GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsParams) {

	apps, err := h.service.DeveloperApp.GetAllByEmail(string(developerEmailaddress))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseDeveloperAppNames(c, apps)
}

// creates a developer application
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps)
func (h *Handler2) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressApps(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress) {

	var receivedApplication Application
	if err := c.ShouldBindJSON(&receivedApplication); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	_, err := h.service.Developer.Get(c.Param("developer"))
	if err != nil {
		h.responseError(c, err)
		return
	}
	newApp := fromApplication(receivedApplication)
	// FIXME there should be a service Create method
	createdApp, err := h.service.DeveloperApp.Create(string(developerEmailaddress), newApp, h.who(c))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseApplicationCreated(c, &createdApp)
}

// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler2) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	developer, err := h.service.Developer.Get(string(developerEmailaddress))
	if err != nil {
		h.responseError(c, err)
		return
	}
	app, err := h.service.DeveloperApp.Delete(
		developer.DeveloperID, string(appName), h.who(c))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseApplication(c, &app)
}

// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler2) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	app, err := h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseApplication(c, app)
}

// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler2) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, params PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameParams) {

	if params.Action != nil && c.ContentType() == "application/octet-stream" {
		h.changeDeveloperAppStatus(c, string(developerEmailaddress), string(appName), string(*params.Action))
		return
	}
	var receivedApplication Application
	if err := c.ShouldBindJSON(&receivedApplication); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	updatedApp := fromApplication(receivedApplication)
	storedApp, err := h.service.DeveloperApp.Update(updatedApp, h.who(c))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseApplicationUpdated(c, &storedApp)
}

// change status of application
func (h *Handler2) changeDeveloperAppStatus(c *gin.Context, developerEmailaddress, appName, requestedStatus string) {

	app, err := h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		h.responseError(c, err)
		return
	}
	switch requestedStatus {
	case "approve":
		app.Approve()
	case "revoke":
		app.Revoke()
	default:
		h.responseErrorBadRequest(c, errors.New("unknown status requested"))
		return
	}
	_, err = h.service.DeveloperApp.Update(*app, h.who(c))
	if err != nil {
		h.responseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// returns attributes of an application
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes)
func (h *Handler2) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributes(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	app, err := h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseAttributes(c, app.Attributes)
}

// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes)
func (h *Handler2) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributes(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	var receivedAttributes Attributes
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	attributes := fromAttributesRequest(receivedAttributes.Attribute)
	if err := h.service.DeveloperApp.UpdateAttributes(
		string(appName), attributes, h.who(c)); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	h.responseAttributes(c, attributes)
}

// deletes one attriubte of an application
// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes/{attribute_name})´
func (h *Handler2) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, attributeName AttributeName) {

	oldValue, err := h.service.DeveloperApp.DeleteAttribute(
		string(appName), string(attributeName), h.who(c))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseAttributeDeleted(c, &types.Attribute{
		Name:  string(attributeName),
		Value: oldValue,
	})
}

// returns one attribute of an application
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes/{attribute_name})
func (h *Handler2) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, attributeName AttributeName) {

	app, err := h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		h.responseError(c, err)
		return
	}
	attributeValue, err := app.Attributes.Get(string(attributeName))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseAttributeRetrieved(c, &types.Attribute{
		Name:  string(attributeName),
		Value: attributeValue,
	})
}

// updates an attribute of an application
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes/{attribute_name})
func (h *Handler2) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, attributeName AttributeName) {

	var receivedValue Attribute
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	newAttribute := types.Attribute{
		Name:  string(attributeName),
		Value: *receivedValue.Value,
	}
	if err := h.service.DeveloperApp.UpdateAttribute(
		string(appName), newAttribute, h.who(c)); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	h.responseAttributeUpdated(c, &newAttribute)
}

// Returns API response list ALL developer application names
func (h *Handler2) responseDeveloperAppIDs(c *gin.Context, developerapps types.DeveloperApps) {

	ApplicationNames := make([]string, len(developerapps))
	for i, d := range developerapps {
		ApplicationNames[i] = d.AppID
	}
	c.IndentedJSON(http.StatusOK, ApplicationNames)
}

// Returns API response list ALL developer application names
func (h *Handler2) responseDeveloperAppNames(c *gin.Context, developerapps types.DeveloperApps) {

	ApplicationNames := make([]string, len(developerapps))
	for i, d := range developerapps {
		ApplicationNames[i] = d.Name
	}
	c.IndentedJSON(http.StatusOK, ApplicationNames)
}

// Returns API response list ALL applcations
func (h *Handler2) responseDeveloperAllApps(c *gin.Context, developerapps types.DeveloperApps) {

	all_apps := make([]Application, len(developerapps))
	for i, d := range developerapps {
		all_apps[i] = Application(ToApplicationResponse(&d))
	}
	c.IndentedJSON(http.StatusOK, Applications{
		Application: &all_apps,
	})
}

func (h *Handler2) responseApplication(c *gin.Context, app *types.DeveloperApp) {
	c.IndentedJSON(http.StatusOK, ToApplicationResponse(app))
}

func (h *Handler2) responseApplicationCreated(c *gin.Context, app *types.DeveloperApp) {
	c.IndentedJSON(http.StatusCreated, ToApplicationResponse(app))
}

func (h *Handler2) responseApplicationUpdated(c *gin.Context, app *types.DeveloperApp) {
	c.IndentedJSON(http.StatusOK, ToApplicationResponse(app))
}

// type conversion

func ToApplicationResponse(d *types.DeveloperApp) Application {

	app := Application{
		AppId:          &d.AppID,
		CallbackUrl:    &d.CallbackUrl,
		Attributes:     toAttributesResponse(d.Attributes),
		CreatedAt:      &d.CreatedAt,
		CreatedBy:      &d.CreatedBy,
		DeveloperId:    &d.DeveloperID,
		DisplayName:    &d.DisplayName,
		LastModifiedBy: &d.LastModifiedBy,
		LastModifiedAt: &d.LastModifiedAt,
		Name:           &d.Name,
		Status:         &d.Status,
	}
	if d.Scopes != nil {
		app.Scopes = &d.Scopes
	} else {
		app.Scopes = &[]string{}
	}
	return app
}

func fromApplication(a Application) types.DeveloperApp {

	app := types.DeveloperApp{}
	if a.AppId != nil {
		app.AppID = *a.AppId
	}
	if a.Attributes != nil {
		app.Attributes = fromAttributesRequest(a.Attributes)
	}
	if a.CallbackUrl != nil {
		app.CallbackUrl = *a.CallbackUrl
	}
	if a.CreatedAt != nil {
		app.CreatedAt = *a.CreatedAt
	}
	if a.CreatedBy != nil {
		app.CreatedBy = *a.CreatedBy
	}
	if a.DeveloperId != nil {
		app.DeveloperID = *a.DeveloperId
	}
	if a.DisplayName != nil {
		app.DisplayName = *a.DisplayName
	}
	if a.LastModifiedBy != nil {
		app.LastModifiedBy = *a.LastModifiedBy
	}
	if a.LastModifiedAt != nil {
		app.LastModifiedAt = *a.LastModifiedAt
	}
	if a.Name != nil {
		app.Name = *a.Name
	}
	if a.Scopes != nil {
		app.Scopes = *a.Scopes
	} else {
		app.Scopes = []string{}
	}
	if a.Status != nil {
		app.Status = *a.Status
	}
	return app
}
