package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// returns all developer apps
// (GET /v1/organizations/{organization_name}/apps)
func (h *Handler) GetV1OrganizationsOrganizationNameApps(c *gin.Context, organizationName OrganizationName, params GetV1OrganizationsOrganizationNameAppsParams) {

	apps, err := h.service.DeveloperApp.GetAll()
	if err != nil {
		responseError(c, err)
		return
	}
	// Do we have to return full developer details?
	if params.Expand != nil && *params.Expand {
		h.responseDeveloperAllApps(c, apps)
		return
	}
	h.responseDeveloperAppIDs(c, apps)
}

// returns one app identified by appId of a developer
// (GET /v1/organizations/{organization_name}/apps/{app_id})
func (h *Handler) GetV1OrganizationsOrganizationNameAppsAppId(c *gin.Context, organizationName OrganizationName, appId AppId) {

	app, err := h.service.DeveloperApp.GetByID(string(appId))
	if err != nil {
		responseError(c, err)
		return
	}
	keys, err := h.service.Key.GetByDeveloperAppID(app.AppID)
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseApplication(c, app, &keys)
}

// returns names of all application of a developer
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps)
func (h *Handler) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressApps(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, params GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsParams) {

	apps, err := h.service.DeveloperApp.GetAllByEmail(string(developerEmailaddress))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseDeveloperAppNames(c, apps)
}

// creates a developer application
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps)
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressApps(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress) {

	var receivedApplication Application
	if err := c.ShouldBindJSON(&receivedApplication); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	_, err := h.service.Developer.Get(string(developerEmailaddress))
	if err != nil {
		responseError(c, err)
		return
	}
	newApp := fromApplication(receivedApplication)
	createdApp, err := h.service.DeveloperApp.Create(string(developerEmailaddress), newApp, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}

	createdKey, err := h.service.Key.Create(types.NullDeveloperAppKey, &createdApp, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	createdKeys := types.Keys{createdKey}
	h.responseApplicationCreated(c, &createdApp, &createdKeys)
}

// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	developer, err := h.service.Developer.Get(string(developerEmailaddress))
	if err != nil {
		responseError(c, err)
		return
	}
	app, err := h.service.DeveloperApp.Delete(
		developer.DeveloperID, string(appName), h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseApplication(c, &app, nil)
}

// returns one app identified by name of a developer
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	app, err := h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	keys, err := h.service.Key.GetByDeveloperAppID(app.AppID)
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseApplication(c, app, &keys)
}

// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, params PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameParams) {

	if params.Action != nil && c.ContentType() == "application/octet-stream" {
		h.changeDeveloperAppStatus(c, string(developerEmailaddress), string(appName), string(*params.Action))
		return
	}
	var receivedApplication Application
	if err := c.ShouldBindJSON(&receivedApplication); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	updatedApp := fromApplication(receivedApplication)
	storedApp, err := h.service.DeveloperApp.Update(updatedApp, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseApplicationUpdated(c, &storedApp)
}

// change status of application
func (h *Handler) changeDeveloperAppStatus(c *gin.Context, developerEmailaddress, appName, requestedStatus string) {

	app, err := h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	switch requestedStatus {
	case "approve":
		app.Approve()
	case "revoke":
		app.Revoke()
	default:
		responseErrorBadRequest(c, errors.New("unknown status requested"))
		return
	}
	_, err = h.service.DeveloperApp.Update(*app, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// returns attributes of an application
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes)
func (h *Handler) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributes(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	app, err := h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, app.Attributes)
}

// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes)
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributes(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	var receivedAttributes Attributes
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	attributes := fromAttributesRequest(receivedAttributes.Attribute)
	if err := h.service.DeveloperApp.UpdateAttributes(
		string(appName), attributes, h.who(c)); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	h.responseAttributes(c, attributes)
}

// deletes one attriubte of an application
// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes/{attribute_name})´
func (h *Handler) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, attributeName AttributeName) {

	oldValue, err := h.service.DeveloperApp.DeleteAttribute(
		string(appName), string(attributeName), h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeDeleted(c, &types.Attribute{
		Name:  string(attributeName),
		Value: oldValue,
	})
}

// returns one attribute of an application
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes/{attribute_name})
func (h *Handler) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, attributeName AttributeName) {

	app, err := h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	attributeValue, err := app.Attributes.Get(string(attributeName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeRetrieved(c, &types.Attribute{
		Name:  string(attributeName),
		Value: attributeValue,
	})
}

// updates an attribute of an application
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes/{attribute_name})
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, attributeName AttributeName) {

	var receivedValue Attribute
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newAttribute := types.Attribute{
		Name:  string(attributeName),
		Value: *receivedValue.Value,
	}
	if err := h.service.DeveloperApp.UpdateAttribute(
		string(appName), newAttribute, h.who(c)); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	h.responseAttributeUpdated(c, &newAttribute)
}

// Returns API response list ALL developer application ids
func (h *Handler) responseDeveloperAppIDs(c *gin.Context, developerapps types.DeveloperApps) {

	ApplicationNames := make([]string, len(developerapps))
	for i := range developerapps {
		ApplicationNames[i] = developerapps[i].AppID
	}
	c.IndentedJSON(http.StatusOK, ApplicationNames)
}

// Returns API response list application names of a developer
func (h *Handler) responseDeveloperAppNames(c *gin.Context, developerapps types.DeveloperApps) {

	ApplicationNames := make([]string, len(developerapps))
	for i := range developerapps {
		ApplicationNames[i] = developerapps[i].Name
	}
	c.IndentedJSON(http.StatusOK, ApplicationNames)
}

// Returns API response list ALL applcations
func (h *Handler) responseDeveloperAllApps(c *gin.Context, developerapps types.DeveloperApps) {

	all_apps := make([]Application, len(developerapps))
	for i := range developerapps {
		all_apps[i] = ToApplicationResponse(&developerapps[i], nil)
	}
	c.IndentedJSON(http.StatusOK, Applications{
		Application: &all_apps,
	})
}

func (h *Handler) responseApplication(c *gin.Context, app *types.DeveloperApp, keys *types.Keys) {

	c.IndentedJSON(http.StatusOK, ToApplicationResponse(app, keys))
}

func (h *Handler) responseApplicationCreated(c *gin.Context, app *types.DeveloperApp, keys *types.Keys) {

	c.IndentedJSON(http.StatusCreated, ToApplicationResponse(app, keys))
}

func (h *Handler) responseApplicationUpdated(c *gin.Context, app *types.DeveloperApp) {

	c.IndentedJSON(http.StatusOK, ToApplicationResponse(app, nil))
}

// type conversion
func ToApplicationResponse(d *types.DeveloperApp, k *types.Keys) Application {

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
	if k != nil {
		app.Credentials = ToKeySlice(*k)
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
