package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

var (
	errUnknownApplicationStatus = errors.New("unknown status requested")
)

// returns all developer apps
// (GET /v1/organizations/{organization_name}/apps)
func (h *Handler) GetV1OrganizationsOrganizationNameApps(c *gin.Context,
	organizationName OrganizationName, params GetV1OrganizationsOrganizationNameAppsParams) {

	apps, err := h.service.DeveloperApp.GetAll(string(organizationName))
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
func (h *Handler) GetV1OrganizationsOrganizationNameAppsAppId(c *gin.Context,
	organizationName OrganizationName, appId AppId) {

	app, err := h.service.DeveloperApp.GetByID(string(organizationName), string(appId))
	if err != nil {
		responseError(c, err)
		return
	}
	keys, err := h.service.Key.GetByDeveloperAppID(string(organizationName), app.AppID)
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseApplication(c, app, &keys)
}

// returns names of all application of a developer
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps)
func (h *Handler) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressApps(c *gin.Context,
	organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, params GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsParams) {

	apps, err := h.service.DeveloperApp.GetAllByEmail(string(organizationName), string(developerEmailaddress))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseDeveloperAppNames(c, apps)
}

// creates a developer application
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps)
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressApps(c *gin.Context,
	organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress) {

	var receivedApplication Application
	if err := c.ShouldBindJSON(&receivedApplication); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newApp := fromApplication(receivedApplication)
	createdApp, err := h.service.DeveloperApp.Create(string(organizationName), string(developerEmailaddress), newApp, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}

	createdKey, err := h.service.Key.Create(string(organizationName), string(developerEmailaddress), createdApp.Name, types.NullDeveloperAppKey, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	createdKeys := types.Keys{*createdKey}
	h.responseApplicationCreated(c, createdApp, &createdKeys)
}

// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(
	c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	app, err := h.service.DeveloperApp.GetByName(string(organizationName), string(developerEmailaddress), string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	if err := h.service.DeveloperApp.Delete(
		string(organizationName), string(developerEmailaddress), string(appName), h.who(c)); err != nil {
		responseError(c, err)
		return
	}
	h.responseApplication(c, app, nil)
}

// returns one app identified by name of a developer
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(
	c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	app, err := h.service.DeveloperApp.GetByName(string(organizationName), string(developerEmailaddress), string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	keys, err := h.service.Key.GetByDeveloperAppID(string(organizationName), app.AppID)
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseApplication(c, app, &keys)
}

// Updates an application
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(
	c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, params PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameParams) {

	_, err := h.service.DeveloperApp.GetByName(string(organizationName), string(developerEmailaddress), string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	if params.Action != nil && c.ContentType() == "application/octet-stream" {
		h.changeDeveloperAppStatus(c,
			string(organizationName), string(developerEmailaddress), string(appName),
			string(*params.Action))
		return
	}
	var receivedApplication Application
	if err := c.ShouldBindJSON(&receivedApplication); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	updatedApp := fromApplication(receivedApplication)
	storedApp, err := h.service.DeveloperApp.Update(string(organizationName), string(developerEmailaddress), updatedApp, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseApplicationUpdated(c, storedApp)
}

// change status of application
func (h *Handler) changeDeveloperAppStatus(c *gin.Context, organizationName, developerEmailaddress, appName, requestedStatus string) {

	app, err := h.service.DeveloperApp.GetByName(organizationName, developerEmailaddress, appName)
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
		responseErrorBadRequest(c, errUnknownApplicationStatus)
		return
	}
	_, err = h.service.DeveloperApp.Update(organizationName, developerEmailaddress, *app, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// updates an application and creates a new key if apiproducts provided
// (PUT /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler) PutV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(
	c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, params PutV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameParams) {

	var receivedApplicationUpdate ApplicationUpdate
	if err := c.ShouldBindJSON(&receivedApplicationUpdate); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	app, err := h.service.DeveloperApp.GetByName(string(organizationName), string(developerEmailaddress), string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	// In case apiproduct(s) were provided we create a new key with all apiproducts assigned
	if receivedApplicationUpdate.ApiProducts != nil {
		createdKey, err := h.service.Key.Create(string(organizationName), string(developerEmailaddress), string(appName), types.NullDeveloperAppKey, h.who(c))
		if err != nil {
			responseError(c, err)
			return
		}
		createdKey.APIProducts = createdKey.APIProducts.AddProducts(receivedApplicationUpdate.ApiProducts)
		_, err = h.service.Key.Update(
			string(organizationName), string(developerEmailaddress), string(appName), string(createdKey.ConsumerKey), *createdKey, h.who(c))
		if err != nil {
			responseError(c, err)
			return
		}
	}
	var applicationChanged bool
	if receivedApplicationUpdate.Attributes != nil {
		app.Attributes = fromAttributesRequest(receivedApplicationUpdate.Attributes)
		applicationChanged = true
	}
	if receivedApplicationUpdate.DisplayName != nil {
		app.DisplayName = *receivedApplicationUpdate.DisplayName
		applicationChanged = true
	}
	if receivedApplicationUpdate.CallbackUrl != nil {
		app.DisplayName = *receivedApplicationUpdate.CallbackUrl
		applicationChanged = true
	}
	if applicationChanged {
		_, err = h.service.DeveloperApp.Update(string(organizationName), string(developerEmailaddress), *app, h.who(c))
		if err != nil {
			responseError(c, err)
			return
		}
	}
	keys, err := h.service.Key.GetByDeveloperAppID(string(organizationName), app.AppID)
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseApplication(c, app, &keys)
}

// returns attributes of an application
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes)
func (h *Handler) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributes(
	c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	app, err := h.service.DeveloperApp.GetByName(string(organizationName), string(developerEmailaddress), string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, app.Attributes)
}

// replaces attributes of an application
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes)
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributes(
	c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	var receivedAttributes Attributes
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	app, err := h.service.DeveloperApp.GetByName(string(organizationName), string(developerEmailaddress), string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	app.Attributes = fromAttributesRequest(receivedAttributes.Attribute)
	storedDeveloper, err := h.service.DeveloperApp.Update(string(organizationName), string(developerEmailaddress), *app, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, storedDeveloper.Attributes)
}

// deletes one attriubte of an application
// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes/{attribute_name})´
func (h *Handler) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributesAttributeName(
	c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, attributeName AttributeName) {

	app, err := h.service.DeveloperApp.GetByName(string(organizationName), string(developerEmailaddress), string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	oldValue, err := app.Attributes.Delete(string(attributeName))
	if err != nil {
		responseError(c, err)
	}
	_, err = h.service.DeveloperApp.Update(string(organizationName), string(developerEmailaddress), *app, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeDeleted(c, types.NewAttribute(string(attributeName), oldValue))
}

// returns one attribute of an application
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes/{attribute_name})
func (h *Handler) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributesAttributeName(
	c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, attributeName AttributeName) {

	app, err := h.service.DeveloperApp.GetByName(string(organizationName), string(developerEmailaddress), string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	attributeValue, err := app.Attributes.Get(string(attributeName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeRetrieved(c, types.NewAttribute(string(attributeName), attributeValue))
}

// updates an attribute of an application
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/attributes/{attribute_name})
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameAttributesAttributeName(
	c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, attributeName AttributeName) {

	var receivedValue Attribute
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	app, err := h.service.DeveloperApp.GetByName(string(organizationName), string(developerEmailaddress), string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	newAttribute := types.NewAttribute(string(attributeName), *receivedValue.Value)
	if err := app.Attributes.Set(newAttribute); err != nil {
		responseError(c, err)
	}
	_, err = h.service.DeveloperApp.Update(string(organizationName), string(developerEmailaddress), *app, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeUpdated(c, newAttribute)
}

// Returns API response list ALL developer application ids
func (h *Handler) responseDeveloperAppIDs(c *gin.Context, developerapps types.DeveloperApps) {

	ApplicationNames := make([]string, len(developerapps))
	for i := range developerapps {
		ApplicationNames[i] = developerapps[i].AppID
	}
	c.JSON(http.StatusOK, ApplicationNames)
}

// Returns API response list application names of a developer
func (h *Handler) responseDeveloperAppNames(c *gin.Context, developerapps types.DeveloperApps) {

	ApplicationNames := make([]string, len(developerapps))
	for i := range developerapps {
		ApplicationNames[i] = developerapps[i].Name
	}
	c.JSON(http.StatusOK, ApplicationNames)
}

// API responses

func (h *Handler) responseDeveloperAllApps(c *gin.Context, developerapps types.DeveloperApps) {

	allApps := make([]Application, len(developerapps))
	for i := range developerapps {
		allApps[i] = ToApplicationResponse(&developerapps[i], nil)
	}
	c.JSON(http.StatusOK, Applications{
		Application: &allApps,
	})
}

func (h *Handler) responseApplication(c *gin.Context, app *types.DeveloperApp, keys *types.Keys) {

	c.JSON(http.StatusOK, ToApplicationResponse(app, keys))
}

func (h *Handler) responseApplicationCreated(c *gin.Context, app *types.DeveloperApp, keys *types.Keys) {

	c.JSON(http.StatusCreated, ToApplicationResponse(app, keys))
}

func (h *Handler) responseApplicationUpdated(c *gin.Context, app *types.DeveloperApp) {

	c.JSON(http.StatusOK, ToApplicationResponse(app, nil))
}

// type conversion
func ToApplicationResponse(d *types.DeveloperApp, k *types.Keys) Application {

	app := Application{
		AppId:          &d.AppID,
		CallbackUrl:    &d.CallbackURL,
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
		app.CallbackURL = *a.CallbackUrl
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
