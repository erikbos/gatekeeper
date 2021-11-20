package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// returns all organizations
// (GET /v1/organizations)
func (h *Handler) GetV1Organizations(c *gin.Context) {

	organizations, err := h.service.Organization.GetAll()
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseOrganizations(c, organizations)
}

// creates an organization
// (POST /v1/organizations)
func (h *Handler) PostV1Organizations(c *gin.Context) {

	var receivedOrganization Organization
	if err := c.ShouldBindJSON(&receivedOrganization); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newOrganization := fromOrganization(receivedOrganization)
	storedOrganization, err := h.service.Organization.Create(newOrganization, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseOrganizationCreated(c, storedOrganization)
}

// returns a organization
// (GET /v1/organizations/{organization_name})
func (h *Handler) GetV1OrganizationsOrganizationName(c *gin.Context, organizationName OrganizationName) {

	organization, err := h.service.Organization.Get(string(organizationName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseOrganization(c, organization)
}

// updates an existing organization
// (POST /v1/organizations/{organization_name})
func (h *Handler) PostV1OrganizationsOrganizationName(c *gin.Context, organizationName OrganizationName) {

	_, err := h.service.Organization.Get(string(organizationName))
	if err != nil {
		responseError(c, err)
		return
	}
	var receivedOrganization Organization
	if err := c.ShouldBindJSON(&receivedOrganization); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	updatedOrganization := fromOrganization(receivedOrganization)
	if updatedOrganization.Name != string(organizationName) {
		responseErrorNameValueMisMatch(c)
		return
	}
	storedOrganization, err := h.service.Organization.Update(updatedOrganization, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseOrganizationUpdated(c, storedOrganization)
}

// deletes an organization
// (DELETE /v1/organizations/{organization_name})
func (h *Handler) DeleteV1OrganizationsOrganizationName(c *gin.Context, organizationName OrganizationName) {

	organization, err := h.service.Organization.Get(string(organizationName))
	if err != nil {
		responseError(c, err)
		return
	}
	if err := h.service.Organization.Delete(string(organizationName), h.who(c)); err != nil {
		responseError(c, err)
		return
	}
	h.responseOrganization(c, organization)
}

// API responses

func (h *Handler) responseOrganizations(c *gin.Context, organizations types.Organizations) {

	all_organizations := make([]Organization, len(organizations))
	for i := range organizations {
		all_organizations[i] = h.ToOrganizationResponse(&organizations[i])
	}
	c.IndentedJSON(http.StatusOK, Organizations{
		Organization: &all_organizations,
	})
}

func (h *Handler) responseOrganization(c *gin.Context, organization *types.Organization) {

	c.IndentedJSON(http.StatusOK, h.ToOrganizationResponse(organization))
}

func (h *Handler) responseOrganizationCreated(c *gin.Context, organization *types.Organization) {

	c.IndentedJSON(http.StatusCreated, h.ToOrganizationResponse(organization))
}

func (h *Handler) responseOrganizationUpdated(c *gin.Context, organization *types.Organization) {

	c.IndentedJSON(http.StatusOK, h.ToOrganizationResponse(organization))
}

// type conversion

func (h *Handler) ToOrganizationResponse(o *types.Organization) Organization {

	return Organization{
		Attributes:     toAttributesResponse(o.Attributes),
		CreatedAt:      &o.CreatedAt,
		CreatedBy:      &o.CreatedBy,
		DisplayName:    &o.DisplayName,
		LastModifiedBy: &o.LastModifiedBy,
		LastModifiedAt: &o.LastModifiedAt,
		Name:           &o.Name,
	}
}

func fromOrganization(o Organization) types.Organization {

	organization := types.Organization{}
	if o.Attributes != nil {
		organization.Attributes = fromAttributesRequest(o.Attributes)
	}
	if o.CreatedAt != nil {
		organization.CreatedAt = *o.CreatedAt
	}
	if o.CreatedBy != nil {
		organization.CreatedBy = *o.CreatedBy
	}
	if o.DisplayName != nil {
		organization.DisplayName = *o.DisplayName
	}
	if o.LastModifiedBy != nil {
		organization.LastModifiedBy = *o.LastModifiedBy
	}
	if o.LastModifiedAt != nil {
		organization.LastModifiedAt = *o.LastModifiedAt
	}
	if o.Name != nil {
		organization.Name = *o.Name
	}
	return organization
}
