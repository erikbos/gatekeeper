// Package handler provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.0 DO NOT EDIT.
package handler

import (
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/gin-gonic/gin"
)

// Attribute defines model for Attribute.
type Attribute struct {
	// Name of attribute
	Name *string `json:"name,omitempty"`

	// Value of attribute
	Value *string `json:"value,omitempty"`
}

// All attributes
type Attributes struct {
	Attribute *[]Attribute `json:"attribute,omitempty"`
}

// Developer defines model for Developer.
type Developer struct {
	// User who last updated this developer
	LastModifiedBy *string `json:"LastModifiedBy,omitempty"`

	// Name of developer applications
	Apps       *[]string    `json:"apps,omitempty"`
	Attributes *[]Attribute `json:"attributes,omitempty"`

	// Create timestamp in milliseconds since epoch.
	CreatedAt *int64 `json:"createdAt,omitempty"`

	// User who created this developer
	CreatedBy *string `json:"createdBy,omitempty"`

	// Internal id of developer
	DeveloperId *string `json:"developerId,omitempty"`

	// Email address of developer
	Email *string `json:"email,omitempty"`

	// First name
	FirstName *string `json:"firstName,omitempty"`

	// Last modified timestamp in milliseconds since epoch.
	LastModifiedAt *int64 `json:"lastModifiedAt,omitempty"`

	// Last name
	LastName *string `json:"lastName,omitempty"`

	// Status of develoepr
	Status *string `json:"status,omitempty"`

	// Username
	UserName *string `json:"userName,omitempty"`
}

// All developer details
type Developers struct {
	Developer *[]Developer `json:"developer,omitempty"`
}

// All developer email addresses
type DevelopersEmailAddresses []string

// ErrorMessage defines model for ErrorMessage.
type ErrorMessage struct {
	Code    *int    `json:"code,omitempty"`
	Message *string `json:"message,omitempty"`
}

// Action defines model for action.
type Action string

// AttributeName defines model for attribute_name.
type AttributeName string

// DeveloperEmailaddress defines model for developer_emailaddress.
type DeveloperEmailaddress string

// OrganizationName defines model for organization_name.
type OrganizationName string

// AttributeCreated defines model for AttributeCreated.
type AttributeCreated Attribute

// AttributeDeleted defines model for AttributeDeleted.
type AttributeDeleted Attribute

// AttributeDoesNotExist defines model for AttributeDoesNotExist.
type AttributeDoesNotExist ErrorMessage

// AttributeRetrieved defines model for AttributeRetrieved.
type AttributeRetrieved Attribute

// AttributeUpdated defines model for AttributeUpdated.
type AttributeUpdated Attribute

// AttributesRetrieved defines model for AttributesRetrieved.
type AttributesRetrieved struct {
	Attributes *[]Attribute `json:"attributes,omitempty"`
}

// AttributesUpdated defines model for AttributesUpdated.
type AttributesUpdated struct {
	Attributes *[]Attribute `json:"attributes,omitempty"`
}

// BadRequest defines model for BadRequest.
type BadRequest ErrorMessage

// GetV1OrganizationsOrganizationNameDevelopersParams defines parameters for GetV1OrganizationsOrganizationNameDevelopers.
type GetV1OrganizationsOrganizationNameDevelopersParams struct {
	// Return full developer details
	Expand *bool `json:"expand,omitempty"`

	// maximum number of developers to return
	Count *int32 `json:"count,omitempty"`
}

// PostV1OrganizationsOrganizationNameDevelopersJSONBody defines parameters for PostV1OrganizationsOrganizationNameDevelopers.
type PostV1OrganizationsOrganizationNameDevelopersJSONBody Developer

// PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressJSONBody defines parameters for PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress.
type PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressJSONBody Developer

// PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressParams defines parameters for PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress.
type PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressParams struct {
	// Optional, request status change of developer to 'active' or 'inactive', requires Content-type to be set to 'application/octet-stream'.
	Action *Action `json:"action,omitempty"`
}

// PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesJSONBody defines parameters for PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes.
type PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesJSONBody Attributes

// PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeNameJSONBody defines parameters for PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName.
type PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeNameJSONBody Attribute

// PostV1OrganizationsOrganizationNameDevelopersJSONRequestBody defines body for PostV1OrganizationsOrganizationNameDevelopers for application/json ContentType.
type PostV1OrganizationsOrganizationNameDevelopersJSONRequestBody PostV1OrganizationsOrganizationNameDevelopersJSONBody

// PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressJSONRequestBody defines body for PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress for application/json ContentType.
type PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressJSONRequestBody PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressJSONBody

// PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesJSONRequestBody defines body for PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes for application/json ContentType.
type PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesJSONRequestBody PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesJSONBody

// PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeNameJSONRequestBody defines body for PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName for application/json ContentType.
type PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeNameJSONRequestBody PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeNameJSONBody

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /v1/organizations/{organization_name}/developers)
	GetV1OrganizationsOrganizationNameDevelopers(c *gin.Context, organizationName OrganizationName, params GetV1OrganizationsOrganizationNameDevelopersParams)

	// (POST /v1/organizations/{organization_name}/developers)
	PostV1OrganizationsOrganizationNameDevelopers(c *gin.Context, organizationName OrganizationName)

	// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress})
	DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress)

	// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress})
	GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress)

	// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress})
	PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, params PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressParams)

	// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/attributes)
	GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress)

	// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/attributes)
	PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress)

	// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress}/attributes/{attribute_name})
	DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, attributeName AttributeName)

	// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/attributes/{attribute_name})
	GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, attributeName AttributeName)

	// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/attributes/{attribute_name})
	PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, attributeName AttributeName)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
}

type MiddlewareFunc func(c *gin.Context)

// GetV1OrganizationsOrganizationNameDevelopers operation middleware
func (siw *ServerInterfaceWrapper) GetV1OrganizationsOrganizationNameDevelopers(c *gin.Context) {

	var err error

	// ------------- Path parameter "organization_name" -------------
	var organizationName OrganizationName

	err = runtime.BindStyledParameter("simple", false, "organization_name", c.Param("organization_name"), &organizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter organization_name: %s", err)})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params GetV1OrganizationsOrganizationNameDevelopersParams

	// ------------- Optional query parameter "expand" -------------
	if paramValue := c.Query("expand"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "expand", c.Request.URL.Query(), &params.Expand)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter expand: %s", err)})
		return
	}

	// ------------- Optional query parameter "count" -------------
	if paramValue := c.Query("count"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "count", c.Request.URL.Query(), &params.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter count: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.GetV1OrganizationsOrganizationNameDevelopers(c, organizationName, params)
}

// PostV1OrganizationsOrganizationNameDevelopers operation middleware
func (siw *ServerInterfaceWrapper) PostV1OrganizationsOrganizationNameDevelopers(c *gin.Context) {

	var err error

	// ------------- Path parameter "organization_name" -------------
	var organizationName OrganizationName

	err = runtime.BindStyledParameter("simple", false, "organization_name", c.Param("organization_name"), &organizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter organization_name: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.PostV1OrganizationsOrganizationNameDevelopers(c, organizationName)
}

// DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress operation middleware
func (siw *ServerInterfaceWrapper) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context) {

	var err error

	// ------------- Path parameter "organization_name" -------------
	var organizationName OrganizationName

	err = runtime.BindStyledParameter("simple", false, "organization_name", c.Param("organization_name"), &organizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter organization_name: %s", err)})
		return
	}

	// ------------- Path parameter "developer_emailaddress" -------------
	var developerEmailaddress DeveloperEmailaddress

	err = runtime.BindStyledParameter("simple", false, "developer_emailaddress", c.Param("developer_emailaddress"), &developerEmailaddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter developer_emailaddress: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c, organizationName, developerEmailaddress)
}

// GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress operation middleware
func (siw *ServerInterfaceWrapper) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context) {

	var err error

	// ------------- Path parameter "organization_name" -------------
	var organizationName OrganizationName

	err = runtime.BindStyledParameter("simple", false, "organization_name", c.Param("organization_name"), &organizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter organization_name: %s", err)})
		return
	}

	// ------------- Path parameter "developer_emailaddress" -------------
	var developerEmailaddress DeveloperEmailaddress

	err = runtime.BindStyledParameter("simple", false, "developer_emailaddress", c.Param("developer_emailaddress"), &developerEmailaddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter developer_emailaddress: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c, organizationName, developerEmailaddress)
}

// PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress operation middleware
func (siw *ServerInterfaceWrapper) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context) {

	var err error

	// ------------- Path parameter "organization_name" -------------
	var organizationName OrganizationName

	err = runtime.BindStyledParameter("simple", false, "organization_name", c.Param("organization_name"), &organizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter organization_name: %s", err)})
		return
	}

	// ------------- Path parameter "developer_emailaddress" -------------
	var developerEmailaddress DeveloperEmailaddress

	err = runtime.BindStyledParameter("simple", false, "developer_emailaddress", c.Param("developer_emailaddress"), &developerEmailaddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter developer_emailaddress: %s", err)})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressParams

	// ------------- Optional query parameter "action" -------------
	if paramValue := c.Query("action"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "action", c.Request.URL.Query(), &params.Action)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter action: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress(c, organizationName, developerEmailaddress, params)
}

// GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes operation middleware
func (siw *ServerInterfaceWrapper) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes(c *gin.Context) {

	var err error

	// ------------- Path parameter "organization_name" -------------
	var organizationName OrganizationName

	err = runtime.BindStyledParameter("simple", false, "organization_name", c.Param("organization_name"), &organizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter organization_name: %s", err)})
		return
	}

	// ------------- Path parameter "developer_emailaddress" -------------
	var developerEmailaddress DeveloperEmailaddress

	err = runtime.BindStyledParameter("simple", false, "developer_emailaddress", c.Param("developer_emailaddress"), &developerEmailaddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter developer_emailaddress: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes(c, organizationName, developerEmailaddress)
}

// PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes operation middleware
func (siw *ServerInterfaceWrapper) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes(c *gin.Context) {

	var err error

	// ------------- Path parameter "organization_name" -------------
	var organizationName OrganizationName

	err = runtime.BindStyledParameter("simple", false, "organization_name", c.Param("organization_name"), &organizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter organization_name: %s", err)})
		return
	}

	// ------------- Path parameter "developer_emailaddress" -------------
	var developerEmailaddress DeveloperEmailaddress

	err = runtime.BindStyledParameter("simple", false, "developer_emailaddress", c.Param("developer_emailaddress"), &developerEmailaddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter developer_emailaddress: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes(c, organizationName, developerEmailaddress)
}

// DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName operation middleware
func (siw *ServerInterfaceWrapper) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c *gin.Context) {

	var err error

	// ------------- Path parameter "organization_name" -------------
	var organizationName OrganizationName

	err = runtime.BindStyledParameter("simple", false, "organization_name", c.Param("organization_name"), &organizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter organization_name: %s", err)})
		return
	}

	// ------------- Path parameter "developer_emailaddress" -------------
	var developerEmailaddress DeveloperEmailaddress

	err = runtime.BindStyledParameter("simple", false, "developer_emailaddress", c.Param("developer_emailaddress"), &developerEmailaddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter developer_emailaddress: %s", err)})
		return
	}

	// ------------- Path parameter "attribute_name" -------------
	var attributeName AttributeName

	err = runtime.BindStyledParameter("simple", false, "attribute_name", c.Param("attribute_name"), &attributeName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter attribute_name: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c, organizationName, developerEmailaddress, attributeName)
}

// GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName operation middleware
func (siw *ServerInterfaceWrapper) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c *gin.Context) {

	var err error

	// ------------- Path parameter "organization_name" -------------
	var organizationName OrganizationName

	err = runtime.BindStyledParameter("simple", false, "organization_name", c.Param("organization_name"), &organizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter organization_name: %s", err)})
		return
	}

	// ------------- Path parameter "developer_emailaddress" -------------
	var developerEmailaddress DeveloperEmailaddress

	err = runtime.BindStyledParameter("simple", false, "developer_emailaddress", c.Param("developer_emailaddress"), &developerEmailaddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter developer_emailaddress: %s", err)})
		return
	}

	// ------------- Path parameter "attribute_name" -------------
	var attributeName AttributeName

	err = runtime.BindStyledParameter("simple", false, "attribute_name", c.Param("attribute_name"), &attributeName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter attribute_name: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c, organizationName, developerEmailaddress, attributeName)
}

// PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName operation middleware
func (siw *ServerInterfaceWrapper) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c *gin.Context) {

	var err error

	// ------------- Path parameter "organization_name" -------------
	var organizationName OrganizationName

	err = runtime.BindStyledParameter("simple", false, "organization_name", c.Param("organization_name"), &organizationName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter organization_name: %s", err)})
		return
	}

	// ------------- Path parameter "developer_emailaddress" -------------
	var developerEmailaddress DeveloperEmailaddress

	err = runtime.BindStyledParameter("simple", false, "developer_emailaddress", c.Param("developer_emailaddress"), &developerEmailaddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter developer_emailaddress: %s", err)})
		return
	}

	// ------------- Path parameter "attribute_name" -------------
	var attributeName AttributeName

	err = runtime.BindStyledParameter("simple", false, "attribute_name", c.Param("attribute_name"), &attributeName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("Invalid format for parameter attribute_name: %s", err)})
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
	}

	siw.Handler.PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName(c, organizationName, developerEmailaddress, attributeName)
}

// GinServerOptions provides options for the Gin server.
type GinServerOptions struct {
	BaseURL     string
	Middlewares []MiddlewareFunc
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router *gin.Engine, si ServerInterface) *gin.Engine {
	return RegisterHandlersWithOptions(router, si, GinServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router *gin.Engine, si ServerInterface, options GinServerOptions) *gin.Engine {
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
	}

	router.GET(options.BaseURL+"/v1/organizations/:organization_name/developers", wrapper.GetV1OrganizationsOrganizationNameDevelopers)

	router.POST(options.BaseURL+"/v1/organizations/:organization_name/developers", wrapper.PostV1OrganizationsOrganizationNameDevelopers)

	router.DELETE(options.BaseURL+"/v1/organizations/:organization_name/developers/:developer_emailaddress", wrapper.DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress)

	router.GET(options.BaseURL+"/v1/organizations/:organization_name/developers/:developer_emailaddress", wrapper.GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress)

	router.POST(options.BaseURL+"/v1/organizations/:organization_name/developers/:developer_emailaddress", wrapper.PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddress)

	router.GET(options.BaseURL+"/v1/organizations/:organization_name/developers/:developer_emailaddress/attributes", wrapper.GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes)

	router.POST(options.BaseURL+"/v1/organizations/:organization_name/developers/:developer_emailaddress/attributes", wrapper.PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributes)

	router.DELETE(options.BaseURL+"/v1/organizations/:organization_name/developers/:developer_emailaddress/attributes/:attribute_name", wrapper.DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName)

	router.GET(options.BaseURL+"/v1/organizations/:organization_name/developers/:developer_emailaddress/attributes/:attribute_name", wrapper.GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName)

	router.POST(options.BaseURL+"/v1/organizations/:organization_name/developers/:developer_emailaddress/attributes/:attribute_name", wrapper.PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAttributesAttributeName)

	return router
}
