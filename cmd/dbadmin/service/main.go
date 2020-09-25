package service

import (
	"github.com/erikbos/gatekeeper/pkg/db"
)

// Service can manipulate all our entities
type Service struct {
	Organization *OrganizationService
	Listener     *ListenerService
	Route        *RouteService
	Cluster      *ClusterService
	Developer    *DeveloperService
	DeveloperApp *DeveloperAppService
	Credential   *CredentialService
	APIProduct   *APIProductService
}

// NewService sets up a new service for all entities
func NewService(db *db.Database) *Service {

	return &Service{
		Organization: NewOrganizationService(db),
		Listener:     NewListenerService(db),
		Route:        NewRouteService(db),
		Cluster:      NewClusterService(db),
		Developer:    NewDeveloperService(db),
		DeveloperApp: NewDeveloperAppService(db),
		Credential:   NewCredentialService(db),
		APIProduct:   NewAPIProductService(db),
	}
}

type Identity struct {
}
