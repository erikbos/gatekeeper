package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/managementserver/service"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// GetV1AuditOrganizationsOrganizationName retrieves audit records of organization.
// (GET /v1/audit/organizations/{organization_name})
func (h *Handler) GetV1AuditOrganizationsOrganizationName(c *gin.Context, organizationName OrganizationName, params GetV1AuditOrganizationsOrganizationNameParams) {

	audits, err := h.service.Audit.GetOrganization(string(organizationName), parseQueryParams(params))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAudits(c, audits)
}

// GetV1AuditOrganizationsOrganizationNameApiproductsApiproductName retrieves audit records of apiproduct.
// (GET /v1/audit/organizations/{organization_name}/apiproducts/{apiproduct_name})
func (h *Handler) GetV1AuditOrganizationsOrganizationNameApiproductsApiproductName(c *gin.Context, organizationName OrganizationName, apiproductName ApiproductName, params GetV1AuditOrganizationsOrganizationNameApiproductsApiproductNameParams) {

	audits, err := h.service.Audit.GetAPIProduct(string(organizationName), string(apiproductName), parseQueryParams(params))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAudits(c, audits)
}

// GetV1AuditOrganizationsOrganizationNameDevelopersDeveloperId retrieve audit records of developer.
// (GET /v1/audit/organizations/{organization_name}/developers/{developer_id})
func (h *Handler) GetV1AuditOrganizationsOrganizationNameDevelopersDeveloperId(c *gin.Context, organizationName OrganizationName, developerId DeveloperId, params GetV1AuditOrganizationsOrganizationNameDevelopersDeveloperIdParams) {

	audits, err := h.service.Audit.GetDeveloper(string(organizationName), string(developerId), parseQueryParams(params))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseAudits(c, audits)
}

// Retrieve audit records of application.
// (GET /v1/audit/organizations/{organization_name}/developers/{developer_id}/apps/{app_id})
func (h *Handler) GetV1AuditOrganizationsOrganizationNameDevelopersDeveloperIdAppsAppId(c *gin.Context, organizationName OrganizationName, developerId DeveloperId, appId AppId, params GetV1AuditOrganizationsOrganizationNameDevelopersDeveloperIdAppsAppIdParams) {

	audits, err := h.service.Audit.GetApplication(string(organizationName), string(developerId), string(appId), parseQueryParams(params))
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
	case GetV1AuditOrganizationsOrganizationNameParams:
		if p.StartTime != nil {
			auditQuery.StartTime = int64(*p.StartTime)
		}
		if p.EndTime != nil {
			auditQuery.EndTime = int64(*p.EndTime)
		}
		if p.Count != nil {
			auditQuery.Count = int64(*p.Count)
		}
	case GetV1AuditOrganizationsOrganizationNameApiproductsApiproductNameParams:
		if p.StartTime != nil {
			auditQuery.StartTime = int64(*p.StartTime)
		}
		if p.EndTime != nil {
			auditQuery.EndTime = int64(*p.EndTime)
		}
		if p.Count != nil {
			auditQuery.Count = int64(*p.Count)
		}

	case GetV1AuditOrganizationsOrganizationNameDevelopersDeveloperIdParams:
		if p.StartTime != nil {
			auditQuery.StartTime = int64(*p.StartTime)
		}
		if p.EndTime != nil {
			auditQuery.EndTime = int64(*p.EndTime)
		}
		if p.Count != nil {
			auditQuery.Count = int64(*p.Count)
		}
	case GetV1AuditOrganizationsOrganizationNameDevelopersDeveloperIdAppsAppIdParams:
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
			OldValue: h.convertInterfaceMapString(&a.OldValue),
			NewValue: h.convertInterfaceMapString(&a.NewValue),
		},
		Organization: &a.Organization,
		DeveloperId:  &a.DeveloperID,
		AppId:        &a.AppID,
	}
	return audit
}

// convertInterfaceMapString converts interface{} to *map[string]interface{}
func (h *Handler) convertInterfaceMapString(m interface{}) *map[string]interface{} {

	var data []byte
	var mapString map[string]interface{}
	var err error

	data, err = json.Marshal(m)
	if err != nil {
		h.logger.Fatal("Cannot marshal", zap.Any("InterfaceStringMap", m))
	}
	if err = json.Unmarshal(data, &mapString); err != nil {
		h.logger.Fatal("Cannot unmarshal", zap.Binary("InterfaceStringMap", data))
	}

	return &mapString
}
