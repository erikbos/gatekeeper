package service

import (
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// New sets up services for all entities
func New(database *db.Database, auditlogLogger *zap.Logger) *Service {

	auditlog := NewAuditlog(database, auditlogLogger)
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
	}
}
