package service

import (
	"github.com/erikbos/gatekeeper/cmd/managementserver/audit"
	"github.com/erikbos/gatekeeper/pkg/db"
)

// New sets up services for all entities
func New(database *db.Database, auditlog *audit.Audit) *Service {

	return &Service{
		Listener:     NewListener(database, auditlog),
		Route:        NewRoute(database, auditlog),
		Cluster:      NewCluster(database, auditlog),
		Organization: NewOrganization(database, auditlog),
		Developer:    NewDeveloper(database, auditlog),
		DeveloperApp: NewDeveloperApp(database, auditlog),
		Key:          NewKey(database, auditlog),
		APIProduct:   NewAPIProduct(database, auditlog),
		User:         NewUser(database, auditlog),
		Role:         NewRole(database, auditlog),
		Audit:        NewAudit(database, auditlog),
	}
}
