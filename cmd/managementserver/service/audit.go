package service

import (
	"github.com/erikbos/gatekeeper/cmd/managementserver/audit"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// AuditService is
type AuditService struct {
	db *db.Database
}

// AuditQueryParams holds query filter parameters for audit record retrieval
type AuditQueryParams struct {
	// Start timestamp in epoch milliseconds
	StartTime int64
	// End timestamp in epoch milliseconds
	EndTime int64
	// Maximum number of entities to return
	Count int64
}

// NewAudit returns a new AuditService instance
func NewAudit(database *db.Database, a *audit.Audit) *AuditService {

	return &AuditService{
		db: database,
	}
}

func (as *AuditService) GetOrganization(organizationName string, params AuditQueryParams) (audits types.Audits, err types.Error) {

	return as.db.Audit.GetOrganization(organizationName, filterParams(params))
}

func (as *AuditService) GetAPIProduct(organizationName, apiproductName string, params AuditQueryParams) (audits types.Audits, err types.Error) {

	return as.db.Audit.GetAPIProduct(organizationName, apiproductName, filterParams(params))
}

func (as *AuditService) GetDeveloper(organizationName, developerEmailaddress string, params AuditQueryParams) (audits types.Audits, err types.Error) {

	return as.db.Audit.GetDeveloper(organizationName, developerEmailaddress, filterParams(params))
}

func (as *AuditService) GetApplication(organizationName, developerEmailaddress, appName string, params AuditQueryParams) (audits types.Audits, err types.Error) {

	return as.db.Audit.GetApplication(organizationName, developerEmailaddress, appName, filterParams(params))
}

func (as *AuditService) GetUser(userName string, params AuditQueryParams) (audits types.Audits, err types.Error) {

	return as.db.Audit.GetUser(userName, filterParams(params))
}

func filterParams(params AuditQueryParams) db.AuditFilterParams {

	return db.AuditFilterParams{
		StartTime: params.StartTime,
		EndTime:   params.EndTime,
		Count:     params.Count,
	}
}
