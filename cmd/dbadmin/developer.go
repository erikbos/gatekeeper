package main

import (
	"fmt"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"

	"github.com/erikbos/apiauth/pkg/shared"
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
	r.DELETE("/v1/organizations/:organization/developers/:developer/attributes", s.DeleteDeveloperAttributes)

	r.GET("/v1/organizations/:organization/developers/:developer/attributes/:attribute", s.GetDeveloperAttributeByName)
	r.POST("/v1/organizations/:organization/developers/:developer/attributes/:attribute", shared.AbortIfContentTypeNotJSON, s.PostDeveloperAttributeByName)
	r.DELETE("/v1/organizations/:organization/developers/:developer/attributes/:attribute", s.DeleteDeveloperAttributeByName)
}

// GetAllDevelopers returns all developers in organization
// FIXME: add pagination support
func (s *server) GetAllDevelopers(c *gin.Context) {
	developers, err := s.db.GetDevelopersByOrganization(c.Param("organization"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"developers": developers})
}

// GetDeveloperByEmail returns full details of one developer
func (s *server) GetDeveloperByEmail(c *gin.Context) {
	developer, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, developer.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, developer)
}

// GetDeveloperAttributes returns attributes of a developer
func (s *server) GetDeveloperAttributes(c *gin.Context) {
	developer, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	setLastModifiedHeader(c, developer.LastmodifiedAt)
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developer.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of a developer
func (s *server) GetDeveloperAttributeByName(c *gin.Context) {
	developer, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(developer.Attributes); i++ {
		if developer.Attributes[i].Name == c.Param("attribute") {
			setLastModifiedHeader(c, developer.LastmodifiedAt)
			c.IndentedJSON(http.StatusOK, developer.Attributes[i])
			return
		}
	}
	returnJSONMessage(c, http.StatusNotFound,
		fmt.Errorf("Could not retrieve attribute '%s'", c.Param("attribute")))
}

// PostCreateDeveloper creates a new developer
func (s *server) PostCreateDeveloper(c *gin.Context) {
	var newDeveloper shared.Developer
	if err := c.ShouldBindJSON(&newDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	// we don't allow recreation of existing developer
	existingDeveloper, err := s.db.GetDeveloperByEmail(c.Param("organization"), newDeveloper.Email)
	if err == nil {
		returnJSONMessage(c, http.StatusBadRequest,
			fmt.Errorf("Developer '%s' already exists", existingDeveloper.Email))
		return
	}
	// Automatically assign new developer to organization
	newDeveloper.OrganizationName = c.Param("organization")
	// Generate primary key for new row
	newDeveloper.DeveloperID = generatePrimaryKeyOfDeveloper(newDeveloper.OrganizationName,
		newDeveloper.Email)
	// Developer starts without any apps
	newDeveloper.Apps = nil
	// New developers starts actived
	newDeveloper.Status = "active"
	newDeveloper.CreatedBy = s.whoAmI()
	newDeveloper.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newDeveloper.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateDeveloperByName(&newDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, newDeveloper)
}

// PostDeveloper updates an existing developer
func (s *server) PostDeveloper(c *gin.Context) {
	// Developer to update should exist
	currentDeveloper, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	var updatedDeveloper shared.Developer
	if err := c.ShouldBindJSON(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}

	// We don't allow POSTing to update developer X while body says to update developer Y ;-)
	updatedDeveloper.Email = currentDeveloper.Email
	updatedDeveloper.DeveloperID = currentDeveloper.DeveloperID
	updatedDeveloper.OrganizationName = currentDeveloper.OrganizationName
	updatedDeveloper.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateDeveloperByName(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedDeveloper)
}

// PostDeveloperAttributes updates attributes of developer
func (s *server) PostDeveloperAttributes(c *gin.Context) {
	updatedDeveloper, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	var receivedAttributes struct {
		Attributes []shared.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	updatedDeveloper.Attributes = receivedAttributes.Attributes
	updatedDeveloper.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateDeveloperByName(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedDeveloper.Attributes})
}

// DeleteDeveloperAttributes delete attributes of developer
func (s *server) DeleteDeveloperAttributes(c *gin.Context) {
	updatedDeveloper, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	deletedAttributes := updatedDeveloper.Attributes
	updatedDeveloper.Attributes = nil
	updatedDeveloper.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateDeveloperByName(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK, deletedAttributes)
}

// PostDeveloperAttributeByName update an attribute of developer
func (s *server) PostDeveloperAttributeByName(c *gin.Context) {
	updatedDeveloper, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
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
		updatedDeveloper.Attributes, attributeToUpdate)
	if attributeToUpdateIndex == -1 {
		// We did not find existing attribute, append new attribute
		updatedDeveloper.Attributes = append(updatedDeveloper.Attributes,
			shared.AttributeKeyValues{Name: attributeToUpdate, Value: receivedValue.Value})
	} else {
		updatedDeveloper.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedDeveloper.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateDeveloperByName(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToUpdate, "value": receivedValue.Value})
}

// DeleteDeveloperAttributeByName removes an attribute of developer
func (s *server) DeleteDeveloperAttributeByName(c *gin.Context) {
	updatedDeveloper, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	attributeToDelete := c.Param("attribute")
	updatedAttributes, index, oldValue := shared.DeleteAttribute(updatedDeveloper.Attributes, attributeToDelete)
	if index == -1 {
		returnJSONMessage(c, http.StatusNotFound,
			fmt.Errorf("Could not find attribute '%s'", attributeToDelete))
		return
	}
	updatedDeveloper.Attributes = updatedAttributes
	updatedDeveloper.LastmodifiedBy = s.whoAmI()
	if err := s.db.UpdateDeveloperByName(&updatedDeveloper); err != nil {
		returnJSONMessage(c, http.StatusBadRequest, err)
		return
	}
	c.IndentedJSON(http.StatusOK,
		gin.H{"name": attributeToDelete, "value": oldValue})
}

// DeleteDeveloperByEmail deletes of one developer
func (s *server) DeleteDeveloperByEmail(c *gin.Context) {
	developer, err := s.db.GetDeveloperByEmail(c.Param("organization"), c.Param("developer"))
	if err != nil {
		returnJSONMessage(c, http.StatusNotFound, err)
		return
	}
	developerAppCount := s.db.GetDeveloperAppCountByDeveloperID(developer.DeveloperID)
	switch developerAppCount {
	case -1:
		returnJSONMessage(c, http.StatusInternalServerError,
			fmt.Errorf("Could not retrieve number of developerapps of developer (%s)",
				developer.Email))
	case 0:
		if err := s.db.DeleteDeveloperByEmail(developer.OrganizationName, developer.Email); err != nil {
			returnJSONMessage(c, http.StatusServiceUnavailable, err)
			return
		}
		c.IndentedJSON(http.StatusOK, developer)
	default:
		returnJSONMessage(c, http.StatusForbidden,
			fmt.Errorf("Cannot delete developer '%s' with %d developer apps",
				developer.Email, developerAppCount))
	}
}

// GeneratePrimaryKeyOfDeveloper creates unique primary key for developer db row
func generatePrimaryKeyOfDeveloper(organization, developer string) string {
	return (fmt.Sprintf("%s@@@%s", organization, uniuri.New()))
}
