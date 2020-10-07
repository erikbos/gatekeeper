package service

import (
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// New sets up a new service for all entities
func New(database *db.Database, logger *zap.Logger) *Service {

	changelog := NewChangelogService(database, logger)
	return &Service{
		Organization: NewOrganizationService(database, changelog),
		Listener:     NewListenerService(database, changelog),
		Route:        NewRouteService(database, changelog),
		Cluster:      NewClusterService(database, changelog),
		Developer:    NewDeveloperService(database, changelog),
		DeveloperApp: NewDeveloperAppService(database, changelog),
		Credential:   NewCredentialService(database, changelog),
		APIProduct:   NewAPIProductService(database, changelog),
		User:         NewUserService(database, changelog),
		Role:         NewRoleService(database, changelog),
	}
}
