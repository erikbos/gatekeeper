package service

import (
	"github.com/erikbos/gatekeeper/pkg/audit"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// AuditService is
type AuditService struct {
	db *db.Database
	// audit *audit.Auditlog
}

// NewAuditlog returns a new Auditlog instance
func NewAudit(database *db.Database, a *audit.Audit) *AuditService {

	return &AuditService{
		db: database,
	}
}

// GetAll returns all audit record entries
func (as *AuditService) GetAll(organizationName string) (audit types.Audits, err types.Error) {

	return as.db.Audit.Get(organizationName)
}
