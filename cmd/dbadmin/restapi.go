package main

import (
	"net/http"

	"github.com/erikbos/apiauth/pkg/db"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// registerDeveloperRouters registers all routes we handle
func (e *env) registerDeveloperRouters(r *gin.Engine) {
	r.GET("/v1/organizations/:organization/developers", e.GetAllDevelopers)
	r.GET("/v1/organizations/:organization/developers/:developer", e.GetDeveloperByEmail)
	r.GET("/v1/organizations/:organization/developers/:developer/attributes", e.GetDeveloperAttributes)
	r.GET("/v1/organizations/:organization/developers/:developer/attributes/:attribute", e.GetDeveloperAttributeByName)
	r.POST("/v1/organizations/:organization/developers", e.PostCreateDeveloper)
	r.POST("/v1/organizations/:organization/developers/:developer", e.PostUpdateDeveloper)
	r.POST("/v1/organizations/:organization/developers/:developer/attributes", e.PostUpdateDeveloperAttributes)
	r.POST("/v1/organizations/:organization/developers/:developer/attributes/:attribute", e.PostUpdateDeveloperAttributeByName)
	r.DELETE("/v1/organizations/:organization/developers/:developer", e.DeleteDeveloperByEmail)

	r.GET("/v1/organizations/:organization/developers/:developer/apps", e.GetDeveloperApps)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application", e.GetDeveloperAppByName)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes", e.GetDeveloperAppAttributes)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/attributes/:attribute", e.GetDeveloperAppAttributeByName)
	r.GET("/v1/organizations/:organization/developers/:developer/apps/:application/keys/:key", e.GetDeveloperAppByKey)

	// r.POST("/v1/organizations/:organization/developers/:developer/apps", e.PostAllDeveloperApps)
	// r.POST("/v1/organizations/:organization/developers/:developer/apps/:application", e.PostOneDeveloperApp)
}

// GetAllDevelopers returns all developers in organization
// FIXME: add pagination support
func (e *env) GetAllDevelopers(c *gin.Context) {
	developers, err := e.db.GetDevelopersByOrganization(c.Param("organization"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"developers": developers})
}

// GetDeveloperByEmail returns full details of one developer
func (e *env) GetDeveloperByEmail(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, developer)
}

// GetDeveloperApps returns apps of a developer
func (e *env) GetDeveloperApps(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, developer.Apps)
}

// GetDeveloperAppByName returns one named app of a developer
func (e *env) GetDeveloperAppByName(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound, "Developer id wrong?")
		return
	}
	// All apikeys belonging to this developer app
	AppCredentials, err := e.db.GetAppCredentialByDeveloperAppID(developerApp.Key)
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	developerApp.Credentials = AppCredentials
	c.IndentedJSON(http.StatusOK, developerApp)
}

// GetDeveloperAppByKey returns keys of one particular developer application
func (e *env) GetDeveloperAppByKey(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	AppCredential, err := e.db.GetAppCredentialByKey(c.Param("key"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, AppCredential)
}

// GetDeveloperAttributes returns attributes of a developer
func (e *env) GetDeveloperAttributes(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developer.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of a developer
func (e *env) GetDeveloperAttributeByName(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(developer.Attributes); i++ {
		if developer.Attributes[i].Name == c.Param("attribute") {
			c.IndentedJSON(http.StatusOK, gin.H{"value": developer.Attributes[i].Value})
			return
		}
	}
	e.returnJSONMessage(c, http.StatusNotFound, "Could not retrieve attribute")
}

// GetDeveloperAppAttributes returns attributes of a developer
func (e *env) GetDeveloperAppAttributes(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": developerApp.Attributes})
}

// GetDeveloperAttributeByName returns one particular attribute of a developer
func (e *env) GetDeveloperAppAttributeByName(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	developerApp, err := e.db.GetDeveloperAppByName(c.Param("organization"), c.Param("application"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// Developer and DeveloperApp need to be linked to eachother #nosqlintegritycheck
	if developer.DeveloperID != developerApp.ParentID {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// lets find the attribute requested
	for i := 0; i < len(developerApp.Attributes); i++ {
		if developerApp.Attributes[i].Name == c.Param("attribute") {
			c.IndentedJSON(http.StatusOK, developerApp.Attributes[i])
			return
		}
	}
	e.returnJSONMessage(c, http.StatusNotFound, "Could not retrieve attributes")
}

// PostCreateDeveloperDetails creates a new developer
func (e *env) PostCreateDeveloper(c *gin.Context) {
	var newDeveloper db.Developer
	if err := c.ShouldBindJSON(&newDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	// Automatically assign new developer to organization
	newDeveloper.OrganizationName = c.Param("organization")
	// Dedup provided attributes
	newDeveloper.Attributes = e.removeDuplicateAttributes(newDeveloper.Attributes)
	// New developers starts actived
	newDeveloper.Status = "active"
	newDeveloper.CreatedBy = e.whoAmI()
	newDeveloper.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	newDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.CreateDeveloper(newDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, newDeveloper)
}

// PostUpdateDeveloper updates an existing developer
func (e *env) PostUpdateDeveloper(c *gin.Context) {
	var updatedDeveloper db.Developer
	if err := c.ShouldBindJSON(&updatedDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	// Developer to update should exist
	currentDeveloper, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	// We don't allow POSTing to update developer X while body says to update developer Y
	updatedDeveloper.Email = currentDeveloper.Email
	updatedDeveloper.DeveloperID = currentDeveloper.DeveloperID
	updatedDeveloper.OrganizationName = currentDeveloper.OrganizationName
	updatedDeveloper.Attributes = e.removeDuplicateAttributes(updatedDeveloper.Attributes)
	updatedDeveloper.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByEmail(updatedDeveloper.Email, updatedDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, updatedDeveloper)
}

// PostUpdateDeveloperAttributes updates attributes of developer
func (e *env) PostUpdateDeveloperAttributes(c *gin.Context) {
	var receivedAttributes struct {
		Attributes []db.AttributeKeyValues `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("received array2: %+v", receivedAttributes)
	log.Printf("received array len: %d", len(receivedAttributes.Attributes))
	updatedDeveloper, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// We don't allow updating developer who is member of different organization
	if c.Param("organization") != updatedDeveloper.OrganizationName {
		e.returnJSONMessage(c, http.StatusBadRequest, "Organisation mismatch")
		return
	}
	updatedDeveloper.Attributes = e.removeDuplicateAttributes(receivedAttributes.Attributes)
	updatedDeveloper.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	if err := e.db.UpdateDeveloperByEmail(updatedDeveloper.Email, updatedDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"attribute": updatedDeveloper.Attributes})
}

func findAttributePositionInAttributeArray(attributes []db.AttributeKeyValues, name string) int {
	for index, element := range attributes {
		if element.Name == name {
			return index
		}
	}
	return -1
}

// PostUpdateDeveloperAttributeByName update an attribute of developer
func (e *env) PostUpdateDeveloperAttributeByName(c *gin.Context) {
	var receivedValue struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	updatedDeveloper, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// We don't allow updating developer who is member of different organization
	if c.Param("organization") != updatedDeveloper.OrganizationName {
		e.returnJSONMessage(c, http.StatusBadRequest, "Organization mismatch")
		return
	}
	// Find & update existing attribute in array
	attributeToUpdateIndex := findAttributePositionInAttributeArray(
		updatedDeveloper.Attributes, c.Param("attribute"))
	if attributeToUpdateIndex == -1 {
		// We did not find exist attribute, so we cannot update its value
		updatedDeveloper.Attributes = append(updatedDeveloper.Attributes,
			db.AttributeKeyValues{Name: c.Param("attribute"), Value: receivedValue.Value})
	} else {
		updatedDeveloper.Attributes[attributeToUpdateIndex].Value = receivedValue.Value
	}
	updatedDeveloper.LastmodifiedBy = e.whoAmI()
	updatedDeveloper.LastmodifiedAt = e.getCurrentTimeMilliseconds()
	if err := e.db.UpdateDeveloperByEmail(updatedDeveloper.Email, updatedDeveloper); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, receivedValue)
}

// DeleteDeveloperByEmail deletes of one developer
func (e *env) DeleteDeveloperByEmail(c *gin.Context) {
	developer, err := e.db.GetDeveloperByEmail(c.Param("developer"))
	if err != nil {
		e.returnJSONMessage(c, http.StatusNotFound, err.Error())
		return
	}
	// We don't allow updating developer who is member of different organization
	if c.Param("organization") != developer.OrganizationName {
		e.returnJSONMessage(c, http.StatusBadRequest, "Organization mismatch")
		return
	}
	if err := e.db.DeleteDeveloperByEmail(developer.Email); err != nil {
		e.returnJSONMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
