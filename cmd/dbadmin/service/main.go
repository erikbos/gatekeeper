package service

import (
	"github.com/erikbos/gatekeeper/pkg/db"
)

// NewService sets up a new service for all entities
func NewService(database *db.Database) *Service {

	changelog := NewChangelogService(database)
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
