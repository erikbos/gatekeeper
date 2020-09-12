package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// registerDeveloperRoutes registers all routes we handle
func (s *server) registerDeveloperRoutes(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers", s.GetAllDevelopers)
	r.POST("/v1/organizations/:organization/developers", shared.AbortIfContentTypeNotJSON, s.PostCreateDeveloper)

	r.GET("/v1/organizations/:organization/developers/:developer", s.GetDeveloperByEmail)
	r.POST("/v1/organizations/:organization/developers/:developer", shared.AbortIfContentTypeNotJSON, s.PostDeveloper)
	r.DELETE("/v1/organizations/:organization/developers/:developer", s.DeleteDeveloperByEmail)

	r.GET("/v1/organizations/:organization/developers/:developer/attributes", s.GetDeveloperAttributes)
	r.POST("/v1/organizations/:organization/developers/:developer/attributes", shared.AbortIfContentTypeNotJSON, s.PostDeveloperAttributes)

	r.GET("/v1/organizations/:organization/developers/:developer/attributes/:attribute", s.GetDeveloperAttributeByName)
	r.POST("/v1/organizations/:organization/developers/:developer/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostDeveloperAttributeByName)
	r.DELETE("/v1/organizations/:organization/developers/:developer/attributes/:attribute", s.DeleteDeveloperAttributeByName)
}

// GetAllDevelopers returns all developers in organization
// FIXME: add pagination support
func (s *server) GetAllDevelopers(c *gin.Context) {

	developers, err := s.db.Developer.GetByOrganization(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"developers": developers})
}

// GetDeveloperByEmail returns full details of one developer
func (s *server) GetDeveloperByEmail(c *gin.Context) {

	developer, err := s.db.Developer.GetByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, developer.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, developer)
}

// GetDeveloperAttributes returns attributes of a developer
func (s *server) GetDeveloperAttributes(c *gin.Context) {

	developer, err := s.db.Developer.GetByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	setLastModifiedHeader(c, developer.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developer.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of a developer
func (s *server) GetDeveloperAttributeByName(c *gin.Context) {

	developer, err := s.db.Developer.GetByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	value, err := developer.Attributes.Get(c.Param("attribute"))
	if err != nil {
		returnCanNotFindAttribute(c, c.Param("attribute"))
		return
	}

	setLastModifiedHeader(c, developer.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, value)
}

// PostCreateDeveloper creates a new developer
func (s *server) PostCreateDeveloper(c *gin.Context) {

	var newDeveloper types.Developer
	if err := c.ShouldBindJSON(&newDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	// we don't allow recreation of existing developer
	existingDeveloper, err := s.db.Developer.GetByEmail(c.Param("organization"), newDeveloper.Email)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Developer '%s' already exists", existingDeveloper.Email))
		return
	}

	// Automatically assign new developer to organization
	newDeveloper.OrganizationName = c.Param("organization")
	// Generate primary key for new row
	newDeveloper.DeveloperID = generatePrimaryKeyOfDeveloper(newDeveloper.OrganizationName, newDeveloper.Email)
	// starts without any apps created
	newDeveloper.Apps = nil
	// starts active
	newDeveloper.Status = "active"
	newDeveloper.SuspendedTill = -1
	newDeveloper.CreatedBy = s.whoAmI()
	newDeveloper.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newDeveloper.LastmodifiedBy = s.whoAmI()

	if err := s.db.Developer.UpdateByName(&newDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, newDeveloper)
}

// PostDeveloper updates an existing developer
func (s *server) PostDeveloper(c *gin.Context) {

	developerToUpdate, err := s.db.Developer.GetByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	var updateRequest types.Developer
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	// Copy over the fields we allow to be updated
	developerToUpdate.Email = updateRequest.Email
	developerToUpdate.FirstName = updateRequest.FirstName
	developerToUpdate.LastName = updateRequest.LastName
	developerToUpdate.Attributes = updateRequest.Attributes
	developerToUpdate.Status = updateRequest.Status

	developerToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Developer.UpdateByName(developerToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, developerToUpdate)
}

// PostDeveloperAttributes updates attributes of developer
func (s *server) PostDeveloperAttributes(c *gin.Context) {

	developerToUpdate, err := s.db.Developer.GetByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	var body struct {
		Attributes types.Attributes `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	if len(body.Attributes) == 0 {
		returnJSONMessage(c, http.StatusBadRequest, errors.New("No attributes posted"))
		return
	}

	developerToUpdate.Attributes = body.Attributes

	developerToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Developer.UpdateByName(developerToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developerToUpdate.Attributes})
}

// PostDeveloperAttributeByName update an attribute of developer
func (s *server) PostDeveloperAttributeByName(c *gin.Context) {

	developerToUpdate, err := s.db.Developer.GetByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	var body struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	attributeToUpdate := c.Param("attribute")
	developerToUpdate.Attributes.Set(attributeToUpdate, body.Value)
	developerToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Developer.UpdateByName(developerToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"name": attributeToUpdate, "value": body.Value})
}

// DeleteDeveloperAttributeByName removes an attribute of developer
func (s *server) DeleteDeveloperAttributeByName(c *gin.Context) {

	developerToUpdate, err := s.db.Developer.GetByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	attributeToDelete := c.Param("attribute")
	deleted, oldValue := developerToUpdate.Attributes.Delete(attributeToDelete)
	if !deleted {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	developerToUpdate.LastmodifiedBy = s.whoAmI()

	if err := s.db.Developer.UpdateByName(developerToUpdate); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteDeveloperByEmail deletes of one developer
func (s *server) DeleteDeveloperByEmail(c *gin.Context) {

	developer, err := s.db.Developer.GetByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}

	developerAppCount := s.db.DeveloperApp.GetCountByDeveloperID(developer.DeveloperID)
	switch developerAppCount {
	case -1:
		returnJSONMessage(c, http.StatusInternalServerError,
			fmt.Errorf("Could not retrieve number of developerapps of developer (%s)",
				developer.Email))
	case 0:
		if err := s.db.Developer.DeleteByEmail(developer.OrganizationName, developer.Email); err != nil {
			returnJSONMessage(c, http.StatusServiceUnavailable, err)
			return
		}
		c.IndentedJSON(http.StatusOK, developer)
	default:
		returnJSONMessage(c, http.StatusForbidden,
			fmt.Errorf("Cannot delete developer '%s' with %d active developer apps",
				developer.Email, developerAppCount))
	}
}

// FIXME this should live in /pkg/db
// GeneratePrimaryKeyOfDeveloper creates primary key for developer db row
func generatePrimaryKeyOfDeveloper(organization, developer string) string {
	return (uniuri.New())
}
