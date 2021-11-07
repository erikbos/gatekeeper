package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// return all developers
// (GET /v1/organizations/{organization_name}/developers)
func (h *Handler2) GetV1OrganizationsOrganizationNameDevelopers(c *gin.Context, organizationName OrganizationName, params GetV1OrganizationsOrganizationNameDevelopersParams) {

	developers, err := h.service.Developer.GetAll()
	if err != nil {
		h.responseError(c, err)
		return
	}
	// Do we have to return full developer details?
	if params.Expand != nil && *params.Expand {
		h.responseDevelopers(c, developers)
		return
	}
	h.responseDeveloperEmailAddresses(c, developers)
}

// creates a new developer
// (POST /v1/organizations/{organization_name}/developers)
func (h *Handler2) PostV1OrganizationsOrganizationNameDevelopers(c *gin.Context, organizationName OrganizationName) {

	var receivedDeveloper Developer
	if err := c.ShouldBindJSON(&receivedDeveloper); err != nil {
		h.responseError(c, types.NewBadRequestError(err))
		return
	}
	newDeveloper := fromDeveloper(receivedDeveloper)
	createdDeveloper, err := h.service.Developer.Create(newDeveloper, h.who(c))
	if err != nil {
		h.responseError(c, types.NewBadRequestError(err))
		return
	}
	h.responseDeveloperCreated(c, &createdDeveloper)
}

// deletes a developer
// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress})
func (h *Handler2) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress) {

	developer, err := h.service.Developer.Delete(string(developerEmailaddress), h.who(c))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseDeveloper(c, &developer)
}

// returns full details of one developer
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress})
func (h *Handler2) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress) {

	developer, err := h.service.Developer.Get(string(developerEmailaddress))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseDeveloper(c, developer)
}

// updates existing developer
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress})
func (h *Handler2) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, params PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressParams) {

	if params.Action != nil && c.ContentType() == "application/octet-stream" {
		h.changeDeveloperStatus(c, string(developerEmailaddress), string(*params.Action))
		return
	}
	var receivedDeveloper Developer
	if err := c.ShouldBindJSON(&receivedDeveloper); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	updatedDeveloper := fromDeveloper(receivedDeveloper)
	storedDeveloper, err := h.service.Developer.Update(string(developerEmailaddress), updatedDeveloper, h.who(c))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseDeveloperUpdated(c, &storedDeveloper)
}

// change status of developer
func (h *Handler2) changeDeveloperStatus(c *gin.Context, developerEmailaddress, requestedStatus string) {

	developer, err := h.service.Developer.Get(string(developerEmailaddress))
	if err != nil {
		h.responseError(c, err)
		return
	}
	switch requestedStatus {
	case "active":
		developer.Activate()
	case "inactive":
		developer.Deactivate()
	default:
		h.responseErrorBadRequest(c, errors.New("unknown status requested"))
		return
	}
	_, err = h.service.Developer.Update(string(developerEmailaddress), *developer, h.who(c))
	if err != nil {
		h.responseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// returns attributes of a developer
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/attributes)
func (h *Handler2) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress) {

	developer, err := h.service.Developer.Get(string(developerEmailaddress))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseAttributes(c, developer.Attributes)
}

// replaces attributes of a developer
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/attributes)
func (h *Handler2) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress) {

	var receivedAttributes Attributes
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	attributes := fromAttributesRequest(receivedAttributes.Attribute)
	if err := h.service.Developer.UpdateAttributes(
		string(developerEmailaddress), attributes, h.who(c)); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	h.responseAttributes(c, attributes)
}

// deletes one attriubte of a developer
// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress}/attributes/{attribute_name})
func (h *Handler2) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, attributeName AttributeName) {

	oldValue, err := h.service.Developer.DeleteAttribute(
		string(developerEmailaddress), string(attributeName), h.who(c))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseAttributeDeleted(c, &types.Attribute{
		Name:  string(attributeName),
		Value: oldValue,
	})
}

// returns one attribute of a developer
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/attributes/{attribute_name})
func (h *Handler2) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, attributeName AttributeName) {

	developer, err := h.service.Developer.Get(string(developerEmailaddress))
	if err != nil {
		h.responseError(c, err)
		return
	}
	attributeValue, err := developer.Attributes.Get(string(attributeName))
	if err != nil {
		h.responseError(c, err)
		return
	}
	h.responseAttributeRetrieved(c, &types.Attribute{
		Name:  string(attributeName),
		Value: attributeValue,
	})
}

// updates an attribute of a developer
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/attributes/{attribute_name})
func (h *Handler2) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, attributeName AttributeName) {

	var receivedValue Attribute
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	newAttribute := types.Attribute{
		Name:  string(attributeName),
		Value: *receivedValue.Value,
	}
	if err := h.service.Developer.UpdateAttribute(
		string(developerEmailaddress), newAttribute, h.who(c)); err != nil {
		h.responseErrorBadRequest(c, err)
		return
	}
	h.responseAttributeUpdated(c, &newAttribute)
}

// Responses

// Returns API response list of developer email addresses
func (h *Handler2) responseDeveloperEmailAddresses(c *gin.Context, developers types.Developers) {

	DevelopersEmailAddresses := make([]string, len(developers))
	for i := range developers {
		DevelopersEmailAddresses[i] = developers[i].Email
	}
	c.IndentedJSON(http.StatusOK, DevelopersEmailAddresses)
}

// Returns API response all developer details
func (h *Handler2) responseDevelopers(c *gin.Context, developers types.Developers) {

	all_developers := make([]Developer, len(developers))
	for i := range developers {
		all_developers[i] = h.ToDeveloperResponse(&developers[i])
	}
	c.IndentedJSON(http.StatusOK, Developers{
		Developer: &all_developers,
	})
}

func (h *Handler2) responseDeveloper(c *gin.Context, developer *types.Developer) {
	c.IndentedJSON(http.StatusOK, h.ToDeveloperResponse(developer))
}

func (h *Handler2) responseDeveloperCreated(c *gin.Context, developer *types.Developer) {
	c.IndentedJSON(http.StatusCreated, h.ToDeveloperResponse(developer))
}

func (h *Handler2) responseDeveloperUpdated(c *gin.Context, developer *types.Developer) {
	c.IndentedJSON(http.StatusOK, h.ToDeveloperResponse(developer))
}

// type conversion

func (h *Handler2) ToDeveloperResponse(d *types.Developer) Developer {

	dev := Developer{
		Attributes:     toAttributesResponse(d.Attributes),
		CreatedAt:      &d.CreatedAt,
		CreatedBy:      &d.CreatedBy,
		DeveloperId:    &d.DeveloperID,
		Email:          &d.Email,
		FirstName:      &d.FirstName,
		LastModifiedBy: &d.LastModifiedBy,
		LastModifiedAt: &d.LastModifiedAt,
		LastName:       &d.LastName,
		Status:         &d.Status,
		UserName:       &d.UserName,
	}
	if d.Apps != nil {
		dev.Apps = &d.Apps
	} else {
		dev.Apps = &[]string{}
	}
	return dev
}

func fromDeveloper(d Developer) types.Developer {

	dev := types.Developer{}
	if d.Apps != nil {
		dev.Apps = *d.Apps
	}
	if d.Attributes != nil {
		dev.Attributes = fromAttributesRequest(d.Attributes)
	}
	if d.CreatedAt != nil {
		dev.CreatedAt = *d.CreatedAt
	}
	if d.CreatedBy != nil {
		dev.CreatedBy = *d.CreatedBy
	}
	if d.DeveloperId != nil {
		dev.DeveloperID = *d.DeveloperId
	}
	if d.Email != nil {
		dev.Email = *d.Email
	}
	if d.FirstName != nil {
		dev.FirstName = *d.FirstName
	}
	if d.LastModifiedBy != nil {
		dev.LastModifiedBy = *d.LastModifiedBy
	}
	if d.LastModifiedAt != nil {
		dev.LastModifiedAt = *d.LastModifiedAt
	}
	if d.LastName != nil {
		dev.LastName = *d.LastName
	}
	if d.Status != nil {
		dev.Status = *d.Status
	}
	if d.UserName != nil {
		dev.UserName = *d.UserName
	}
	return dev
}
