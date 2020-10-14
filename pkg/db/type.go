package db

import (
	"log"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// Entity types we handle
const (
	EntityTypeListener     = "listener"
	EntityTypeRoute        = "route"
	EntityTypeCluster      = "cluster"
	EntityTypeOrganization = "organization"
	EntityTypeDeveloper    = "developer"
	EntityTypeDeveloperApp = "developerapp"
	EntityTypeAPIProduct   = "apiproduct"
	EntityTypeCredential   = "credential"
	EntityTypeOAuth        = "oauth"
	EntityTypeUser         = "user"
	EntityTypeRole         = "role"
)

// Typeof returns the type of an object
func Typeof(entity interface{}) string {

	log.Printf("%T", entity)

	switch entity.(type) {
	case types.Listener:
		return EntityTypeListener
	case *types.Listener:
		return EntityTypeListener
	case types.Route:
		return EntityTypeRoute
	case *types.Route:
		return EntityTypeRoute
	case types.Cluster:
		return EntityTypeCluster
	case *types.Cluster:
		return EntityTypeCluster
	case types.Organization:
		return EntityTypeOrganization
	case *types.Organization:
		return EntityTypeOrganization
	case types.Developer:
		return EntityTypeDeveloper
	case *types.Developer:
		return EntityTypeDeveloper
	case types.DeveloperApp:
		return EntityTypeDeveloperApp
	case *types.DeveloperApp:
		return EntityTypeDeveloperApp
	case types.DeveloperAppKey:
		return EntityTypeCredential
	case *types.DeveloperAppKey:
		return EntityTypeCredential
	case types.User:
		return EntityTypeUser
	case *types.User:
		return EntityTypeUser
	case types.Role:
		return EntityTypeRole
	case *types.Role:
		return EntityTypeRole
	default:
		return "unknown"
	}
}
