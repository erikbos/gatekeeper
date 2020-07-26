package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// registerOrganizationRoutes registers all routes we handle
func (s *server) registerOrganizationRoutes(r *gin.Engine) {
	r.GET("/v1/organizations", s.GetAllOrganizations)
	r.POST("/v1/organizations", shared.AbortIfContentTypeNotJSON, s.PostCreateOrganization)

	r.GET("/v1/organizations/:organization", s.GetOrganizationByName)
	r.POST("/v1/organizations/:organization", shared.AbortIfContentTypeNotJSON, s.PostOrganization)
	r.DELETE("/v1/organizations/:organization", s.DeleteByName)

	r.GET("/v1/organizations/:organization/attributes", s.GetOrganizationAttributes)
	r.POST("/v1/organizations/:organization/attributes", shared.AbortIfContentTypeNotJSON, s.PostOrganizationAttributes)
	r.DELETE("/v1/organizations/:organization/attributes", s.DeleteOrganizationAttributes)

	r.GET("/v1/organizations/:organization/attributes/:attribute", s.GetOrganizationAttributeByName)
	r.POST("/v1/organizations/:organization/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostOrganizationAttributeByName)
	r.DELETE("/v1/organizations/:organization/attributes/:attribute", s.DeleteOrganizationAttributeByName)
}

// GetAllOrganizations returns all organizations
func (s *server) GetAllOrganizations(c *gin.Context) {
	organizations, err := s.db.Organization.GetAll()

	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"organizations": organizations})
}

// GetByName returns details of an organization
func (s *server) GetOrganizationByName(c *gin.Context) {

	organization, err := s.db.Organization.GetByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, organization.LastmodifiedAt)

	c.IndentedJSON(http.StatusOK, organization)
}

// GetOrganizationAttributes returns attributes of an organization
func (s *server) GetOrganizationAttributes(c *gin.Context) {

	organization, err := s.db.Organization.GetByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, organization.LastmodifiedAt)

	c.IndentedJSON(http.StatusOK, gin.H{"attribute": organization.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of an organization
func (s *server) GetOrganizationAttributeByName(c *gin.Context) {

	organization, err := s.db.Organization.GetByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	// lets find the attribute requested
	for i := 0; i < len(organization.Attributes); i++ {
		if organization.Attributes[i].Name == c.Param("attribute") {
			setLastModifiedHeader(c, organization.LastmodifiedAt)
			c.IndentedJSON(http.StatusOK, organization.Attributes[i])
			return
		}
	}

	returnJSONMessage(c, http.StatusNotFound,
		fmt.Errorf("Could not retrieve attribute '%s'", c.Param("attribute")))
}

// PostCreateOrganization creates an organization
func (s *server) PostCreateOrganization(c *gin.Context) {

	var newOrganization shared.Organization
	if err := c.ShouldBindJSON(&newOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	existingOrganization, err := s.db.Organization.GetByName(newOrganization.Name)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Organization '%s' already exists", existingOrganization.Name))
		return
	}

	// Automatically set default fields
	newOrganization.CreatedBy = s.whoAmI()
	newOrganization.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newOrganization.LastmodifiedBy = s.whoAmI()

	if err := s.db.Organization.UpdateByName(&newOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, newOrganization)
}

// PostOrganization updates an existing organization
func (s *server) PostOrganization(c *gin.Context) {

	currentOrganization, err := s.db.Organization.GetByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	var updatedOrganization shared.Organization
	if err := c.ShouldBindJSON(&updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	// We don't allow POSTing to update organization X while body says to update organization Y
	updatedOrganization.Name = currentOrganization.Name
	updatedOrganization.CreatedBy = currentOrganization.CreatedBy
	updatedOrganization.CreatedAt = currentOrganization.CreatedAt
	updatedOrganization.LastmodifiedBy = s.whoAmI()

	if err := s.db.Organization.UpdateByName(&updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, updatedOrganization)
}

// PostOrganizationAttributes updates attributes of an organization
func (s *server) PostOrganizationAttributes(c *gin.Context) {
	updatedOrganization, err := s.db.Organization.GetByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var receivedAttributes struct {
		Attributes []shared.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	updatedOrganization.Attributes = receivedAttributes.Attributes
	updatedOrganization.LastmodifiedBy = s.whoAmI()
	if err := s.db.Organization.UpdateByName(updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedOrganization.Attributes})
}

// PostOrganizationAttributeByName update an attribute of developer
func (s *server) PostOrganizationAttributeByName(c *gin.Context) {
	updatedOrganization, err := s.db.Organization.GetByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	attributeToUpdate := c.Param("attribute")
	var receivedValue struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// Find & update existing attribute in array
	attributeToUpdateIndex := shared.FindIndexOfAttribute(
		updatedOrganization.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find exist attribute, append new attribute
		updatedOrganization.Attributes = append(updatedOrganization.Attributes,
			shared.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedOrganization.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedOrganization.LastmodifiedBy = s.whoAmI()
	if err := s.db.Organization.UpdateByName(updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteOrganizationAttributeByName removes an attribute of an organization
func (s *server) DeleteOrganizationAttributeByName(c *gin.Context) {
	updatedOrganization, err := s.db.Organization.GetByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	attributeToDelete := c.Param("attribute")
	updatedAttributes, index, oldValue := shared.DeleteAttribute(updatedOrganization.Attributes, attributeToDelete)
	if index == -1 {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedOrganization.Attributes = updatedAttributes
	updatedOrganization.LastmodifiedBy = s.whoAmI()
	if err := s.db.Organization.UpdateByName(updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteOrganizationAttributes removes all attribute of an organization
func (s *server) DeleteOrganizationAttributes(c *gin.Context) {
	updatedOrganization, err := s.db.Organization.GetByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	DeleteDeveloperAttributes := updatedOrganization.Attributes
	updatedOrganization.Attributes = nil
	updatedOrganization.LastmodifiedBy = s.whoAmI()
	if err := s.db.Organization.UpdateByName(updatedOrganization); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": DeleteDeveloperAttributes})
}

// DeleteByName deletes an organization
func (s *server) DeleteByName(c *gin.Context) {
	organization, err := s.db.Organization.GetByName(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerCount := s.db.Developer.GetCountByOrganization(organization.Name)
	switch developerCount {
	case -1:
		returnJSONMessage(c, http.StatusInternalServerError,
			fmt.Errorf("Could not retrieve number of developers in organization"))
	case 0:
		if err := s.db.Organization.DeleteByName(organization.Name); err != nil {
			returnJSONMessage(c, http.StatusServiceUnavailable, err)
		}
		c.IndentedJSON(http.StatusOK, organization)
	default:
		returnJSONMessage(c, http.StatusForbidden,
			fmt.Errorf("Cannot delete organization '%s' with %d active developers",
				organization.Name, developerCount))
	}
}
