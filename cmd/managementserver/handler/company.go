package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// Retrieve companies
// (GET /v1/organizations/{organization_name}/companies)
func (h *Handler) GetV1OrganizationsOrganizationNameCompanies(c *gin.Context,
	organizationName OrganizationName, params GetV1OrganizationsOrganizationNameCompaniesParams) {

	companies, err := h.service.Company.GetAll(string(organizationName))
	if err != nil {
		responseError(c, err)
		return
	}
	// Do we have to return full company details?
	if params.Expand != nil && *params.Expand {
		h.responseCompanies(c, companies)
		return
	}
	h.responseCompanyNames(c, companies)
}

// creates a new company
// (POST /v1/organizations/{organization_name}/companies)
func (h *Handler) PostV1OrganizationsOrganizationNameCompanies(c *gin.Context,
	organizationName OrganizationName) {

	var receivedCompany Company
	if err := c.ShouldBindJSON(&receivedCompany); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newCompany := fromCompany(receivedCompany)
	createdDeveloper, err := h.service.Company.Create(string(organizationName), newCompany, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseCompanyCreated(c, createdDeveloper)
}

// deletes an company
// (DELETE /v1/organizations/{organization_name}/companies/{company_name})
func (h *Handler) DeleteV1OrganizationsOrganizationNameCompaniesCompanyName(
	c *gin.Context, organizationName OrganizationName, companyName CompanyName) {

	deletedCompany, err := h.service.Company.Get(string(organizationName), string(companyName))
	if err != nil {
		responseError(c, err)
		return
	}
	if err := h.service.Company.Delete(
		string(organizationName), string(companyName), h.who(c)); err != nil {
		responseError(c, err)
		return
	}
	h.responseCompany(c, deletedCompany)
}

// returns full details of one company
// (GET /v1/organizations/{organization_name}/companies/{company_name})
func (h *Handler) GetV1OrganizationsOrganizationNameCompaniesCompanyName(
	c *gin.Context, organizationName OrganizationName, companyName CompanyName) {

	company, err := h.service.Company.Get(string(organizationName), string(companyName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseCompany(c, company)
}

// updates existing company
// (POST /v1/organizations/{organization_name}/companies/{company_name})
func (h *Handler) PostV1OrganizationsOrganizationNameCompaniesCompanyName(
	c *gin.Context, organizationName OrganizationName, companyName CompanyName,
	params PostV1OrganizationsOrganizationNameCompaniesCompanyNameParams) {

	_, err := h.service.Company.Get(string(organizationName), string(companyName))
	if err != nil {
		responseError(c, err)
		return
	}
	if params.Action != nil && c.ContentType() == "application/octet-stream" {
		h.changeCompanyStatus(c, string(organizationName), string(companyName), string(*params.Action))
		return
	}
	var receivedCompany Company
	if err := c.ShouldBindJSON(&receivedCompany); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	storedCompany, err := h.service.Company.Update(string(organizationName), fromCompany(receivedCompany), h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseCompanyUpdated(c, storedCompany)
}

// change status of devecloper
func (h *Handler) changeCompanyStatus(c *gin.Context, organizationName, companyName, requestedStatus string) {

	company, err := h.service.Company.Get(organizationName, companyName)
	if err != nil {
		responseError(c, err)
		return
	}
	switch requestedStatus {
	case "active":
		company.Activate()
	case "inactive":
		company.Deactivate()
	default:
		responseErrorBadRequest(c, errUnknownDeveloperStatus)
		return
	}
	_, err = h.service.Company.Update(organizationName, *company, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// returns attributes of a company
// (GET /v1/organizations/{organization_name}/companies/{company_name}/attributes)
func (h *Handler) GetV1OrganizationsOrganizationNameCompaniesCompanyNameAttributes(
	c *gin.Context, organizationName OrganizationName, companyName CompanyName) {

	company, err := h.service.Company.Get(string(organizationName), string(companyName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, company.Attributes)
}

// replaces attributes of an company
// (POST /v1/organizations/{organization_name}/companies/{company_name}/attributes)
func (h *Handler) PostV1OrganizationsOrganizationNameCompaniesCompanyNameAttributes(
	c *gin.Context, organizationName OrganizationName, companyName CompanyName) {

	var receivedAttributes Attributes
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	company, err := h.service.Company.Get(string(organizationName), string(companyName))
	if err != nil {
		responseError(c, err)
		return
	}
	company.Attributes = fromAttributesRequest(receivedAttributes.Attribute)
	storedCompany, err := h.service.Company.Update(string(organizationName), *company, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributes(c, storedCompany.Attributes)
}

// deletes one attribute of an company
// (DELETE /v1/organizations/{organization_name}/companies/{company_name}/attributes/{attribute_name})
func (h *Handler) DeleteV1OrganizationsOrganizationNameCompaniesCompanyNameAttributesAttributeName(
	c *gin.Context, organizationName OrganizationName, companyName CompanyName, attributeName AttributeName) {

	company, err := h.service.Company.Get(string(organizationName), string(companyName))
	if err != nil {
		responseError(c, err)
		return
	}
	oldValue, err := company.Attributes.Delete(string(attributeName))
	if err != nil {
		responseError(c, err)
	}
	_, err = h.service.Company.Update(string(organizationName), *company, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeDeleted(c, types.NewAttribute(string(attributeName), oldValue))
}

// returns one attribute of an company
// (GET /v1/organizations/{organization_name}/companies/{company_name}/attributes/{attribute_name})
func (h *Handler) GetV1OrganizationsOrganizationNameCompaniesCompanyNameAttributesAttributeName(
	c *gin.Context, organizationName OrganizationName, companyName CompanyName, attributeName AttributeName) {

	company, err := h.service.Company.Get(string(organizationName), string(companyName))
	if err != nil {
		responseError(c, err)
		return
	}
	attributeValue, err := company.Attributes.Get(string(attributeName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeRetrieved(c, types.NewAttribute(string(attributeName), attributeValue))
}

// updates an attribute of an company
// (POST /v1/organizations/{organization_name}/companies/{company_name}/attributes/{attribute_name})
func (h *Handler) PostV1OrganizationsOrganizationNameCompaniesCompanyNameAttributesAttributeName(
	c *gin.Context, organizationName OrganizationName, companyName CompanyName, attributeName AttributeName) {

	var receivedValue Attribute
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	company, err := h.service.Company.Get(string(organizationName), string(companyName))
	if err != nil {
		responseError(c, err)
		return
	}
	newAttribute := types.NewAttribute(string(attributeName), *receivedValue.Value)
	if err := company.Attributes.Set(newAttribute); err != nil {
		responseError(c, err)
	}
	_, err = h.service.Company.Update(string(organizationName), *company, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAttributeUpdated(c, newAttribute)
}

// API responses

func (h *Handler) responseCompanyNames(c *gin.Context, companys types.Companies) {

	CompanyNames := make([]string, len(companys))
	for i := range companys {
		CompanyNames[i] = companys[i].Name
	}
	c.IndentedJSON(http.StatusOK, CompanyNames)
}

func (h *Handler) responseCompanies(c *gin.Context, companys types.Companies) {

	allCompanys := make([]Company, len(companys))
	for i := range companys {
		allCompanys[i] = h.ToCompanyResponse(&companys[i])
	}
	c.IndentedJSON(http.StatusOK, Companies{
		Company: &allCompanys,
	})
}

func (h *Handler) responseCompany(c *gin.Context, company *types.Company) {

	c.IndentedJSON(http.StatusOK, h.ToCompanyResponse(company))
}

func (h *Handler) responseCompanyCreated(c *gin.Context, company *types.Company) {

	c.IndentedJSON(http.StatusCreated, h.ToCompanyResponse(company))
}

func (h *Handler) responseCompanyUpdated(c *gin.Context, company *types.Company) {

	c.IndentedJSON(http.StatusOK, h.ToCompanyResponse(company))
}

// type conversion

func (h *Handler) ToCompanyResponse(c *types.Company) Company {

	company := Company{
		Attributes:     toAttributesResponse(c.Attributes),
		CreatedAt:      &c.CreatedAt,
		CreatedBy:      &c.CreatedBy,
		DisplayName:    &c.DisplayName,
		LastModifiedBy: &c.LastModifiedBy,
		LastModifiedAt: &c.LastModifiedAt,
		Name:           &c.Name,
		Status:         &c.Status,
	}
	if c.Apps != nil {
		company.Apps = &c.Apps
	} else {
		company.Apps = &[]string{}
	}
	return company
}

func fromCompany(p Company) types.Company {

	product := types.Company{}
	if p.Attributes != nil {
		product.Attributes = fromAttributesRequest(p.Attributes)
	}
	if p.CreatedAt != nil {
		product.CreatedAt = *p.CreatedAt
	}
	if p.CreatedBy != nil {
		product.CreatedBy = *p.CreatedBy
	}
	if p.DisplayName != nil {
		product.DisplayName = *p.DisplayName
	}
	if p.Name != nil {
		product.Name = *p.Name
	}
	if p.LastModifiedBy != nil {
		product.LastModifiedBy = *p.LastModifiedBy
	}
	if p.LastModifiedAt != nil {
		product.LastModifiedAt = *p.LastModifiedAt
	}
	if p.Status != nil {
		product.Status = *p.Status
	}
	return product
}
