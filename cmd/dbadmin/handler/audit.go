package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/service"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// GetV1AuditOrganizationName retrieves audit records of organization.
// (GET /v1/audit/{organization_name})
func (h *Handler) GetV1AuditOrganizationName(c *gin.Context, organizationName OrganizationName, params GetV1AuditOrganizationNameParams) {

	audits, err := h.service.Audit.GetOrganization(string(organizationName), parseQueryParams(params))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAudits(c, audits)
}

// GetV1AuditOrganizationNameApiproductsApiproductName retrieves audit records of apiproduct.
// (GET /v1/audit/{organization_name}/apiproducts/{apiproduct_name})
func (h *Handler) GetV1AuditOrganizationNameApiproductsApiproductName(c *gin.Context, organizationName OrganizationName, apiproductName ApiproductName, params GetV1AuditOrganizationNameApiproductsApiproductNameParams) {

	audits, err := h.service.Audit.GetAPIProduct(string(organizationName), string(apiproductName), parseQueryParams(params))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAudits(c, audits)
}

// GetV1AuditOrganizationNameDevelopersDeveloperEmailaddress retrieves audit records of developer
// (GET /v1/audit/{organization_name}/developers/{developer_emailaddress})
func (h *Handler) GetV1AuditOrganizationNameDevelopersDeveloperEmailaddress(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, params GetV1AuditOrganizationNameDevelopersDeveloperEmailaddressParams) {

	audits, err := h.service.Audit.GetDeveloper(string(organizationName), string(developerEmailaddress), parseQueryParams(params))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAudits(c, audits)
}

// GetV1AuditOrganizationNameDevelopersDeveloperEmailaddressAppsAppName retrieves audit records of application.
// (GET /v1/audit/{organization_name}/developers/{developer_emailaddress}/apps/{app_name})
func (h *Handler) GetV1AuditOrganizationNameDevelopersDeveloperEmailaddressAppsAppName(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, params GetV1AuditOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameParams) {

	audits, err := h.service.Audit.GetApplication(string(organizationName), string(developerEmailaddress), string(appName), parseQueryParams(params))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAudits(c, audits)
}

// GetV1AuditUsersUserName retrieves audit records of user
// (GET /v1/audit/users/{user_name})
func (h *Handler) GetV1AuditUsersUserName(c *gin.Context, userName UserName, params GetV1AuditUsersUserNameParams) {

	audits, err := h.service.Audit.GetUser(string(userName), parseQueryParams(params))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAudits(c, audits)
}

// parseQueryParams converts API query parameters into audit query parameters
func parseQueryParams(queryParams interface{}) service.AuditQueryParams {

	var auditQuery service.AuditQueryParams

	switch p := queryParams.(type) {
	case GetV1AuditOrganizationNameParams:
		if p.StartTime != nil {
			auditQuery.StartTime = int64(*p.StartTime)
		}
		if p.EndTime != nil {
			auditQuery.EndTime = int64(*p.EndTime)
		}
		if p.Count != nil {
			auditQuery.Count = int64(*p.Count)
		}
	case GetV1AuditOrganizationNameApiproductsApiproductNameParams:
		if p.StartTime != nil {
			auditQuery.StartTime = int64(*p.StartTime)
		}
		if p.EndTime != nil {
			auditQuery.EndTime = int64(*p.EndTime)
		}
		if p.Count != nil {
			auditQuery.Count = int64(*p.Count)
		}

	case GetV1AuditOrganizationNameDevelopersDeveloperEmailaddressParams:
		if p.StartTime != nil {
			auditQuery.StartTime = int64(*p.StartTime)
		}
		if p.EndTime != nil {
			auditQuery.EndTime = int64(*p.EndTime)
		}
		if p.Count != nil {
			auditQuery.Count = int64(*p.Count)
		}
	case GetV1AuditOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameParams:
		if p.StartTime != nil {
			auditQuery.StartTime = int64(*p.StartTime)
		}
		if p.EndTime != nil {
			auditQuery.EndTime = int64(*p.EndTime)
		}
		if p.Count != nil {
			auditQuery.Count = int64(*p.Count)
		}
	case GetV1AuditUsersUserNameParams:
		if p.StartTime != nil {
			auditQuery.StartTime = int64(*p.StartTime)
		}
		if p.EndTime != nil {
			auditQuery.EndTime = int64(*p.EndTime)
		}
		if p.Count != nil {
			auditQuery.Count = int64(*p.Count)
		}
	}

	// Set default values in case not provided.
	// Default start time is one day ago.
	if auditQuery.StartTime == 0 {
		auditQuery.StartTime = shared.GetCurrentTimeMilliseconds() - (86400 * 1000)
	}
	if auditQuery.EndTime == 0 {
		auditQuery.EndTime = shared.GetCurrentTimeMilliseconds()
	}
	if auditQuery.Count == 0 {
		auditQuery.Count = 100
	}

	return auditQuery
}

// API responses

func (h *Handler) responseAudits(c *gin.Context, audits types.Audits) {

	allAudits := make([]Audit, len(audits))
	for i := range audits {
		allAudits[i] = h.ToAuditResponse(&audits[i])
	}
	c.IndentedJSON(http.StatusOK, Audits{
		Audit: &allAudits,
	})
}

// type conversion

func (h *Handler) ToAuditResponse(a *types.Audit) Audit {

	audit := Audit{
		AuditId:   &a.ID,
		AuditType: &a.AuditType,
		Timestamp: &a.Timestamp,
		Requestor: &AuditRequestor{
			Ipaddress: &a.IPaddress,
			RequestId: &a.RequestID,
			User:      &a.User,
			Role:      &a.Role,
			UserAgent: &a.UserAgent,
		},
		Entity: &AuditEntity{
			Type:     &a.EntityType,
			Id:       &a.EntityID,
			OldValue: convertInterfaceMapString(&a.OldValue),
			NewValue: convertInterfaceMapString(&a.NewValue),
		},
		// Organization: &a.Organization,
		// DeveloperId:  &a.DeveloperID,
		// AppId:        &a.AppID,
	}
	return audit
}

// convertInterfaceMapString converts interface{} to *map[string]interface{}
func convertInterfaceMapString(m interface{}) *map[string]interface{} {

	var data []byte
	var mapString map[string]interface{}

	data, _ = json.Marshal(m)
	json.Unmarshal(data, &mapString)

	return &mapString
}
