package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Retrieve audit records of organization.
// (GET /v1/audit/{organization_name})
func (h *Handler) GetV1AuditOrganizationName(c *gin.Context, organizationName OrganizationName) {

	audits, err := h.service.Audit.GetAll(string(organizationName))
	if err != nil {
		responseError(c, err)
		return
	}
	c.JSON(http.StatusOK, audits)
}

// Retrieve audit records of developer
// (GET /v1/audit/{organization_name}/developers/{developer_emailaddress})
func (h *Handler) GetV1AuditOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress) {

	audits, err := h.service.Audit.GetAll(string(organizationName))
	if err != nil {
		responseError(c, err)
		return
	}
	c.JSON(http.StatusOK, audits)
}

// Retrieve audit records of user
// (GET /v1/audit/{organization_name}/users/{user_name})
func (h *Handler) GetV1AuditOrganizationNameUsersUserName(c *gin.Context, organizationName OrganizationName, userName UserName) {

	audits, err := h.service.Audit.GetAll(string(organizationName))
	if err != nil {
		responseError(c, err)
		return
	}
	c.JSON(http.StatusOK, audits)
}
