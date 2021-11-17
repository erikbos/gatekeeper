package service

import (
	"github.com/erikbos/gatekeeper/pkg/types"
)

// Service can manipulate all our entities
type Service struct {
	Listener
	Route
	Cluster
	Organization
	Developer
	DeveloperApp
	Key
	APIProduct
	User
	Role
}

// All interface of service layer
type (
	// Listener is the service interface to manipulate Listener entities
	Listener interface {
		GetAll() (listeners types.Listeners, err types.Error)

		Get(listenerName string) (listener *types.Listener, err types.Error)

		Create(newListener types.Listener, who Requester) (types.Listener, types.Error)

		Update(updatedListener types.Listener, who Requester) (types.Listener, types.Error)

		Delete(listenerName string, who Requester) (e types.Error)
	}

	// Route is the service interface to manipulate Route entities
	Route interface {
		GetAll() (routes types.Routes, err types.Error)

		Get(routeName string) (route *types.Route, err types.Error)

		Create(newRoute types.Route, who Requester) (types.Route, types.Error)

		Update(updatedRoute types.Route, who Requester) (types.Route, types.Error)

		Delete(routeName string, who Requester) (e types.Error)
	}

	// Cluster is the service interface to manipulate Cluster entities
	Cluster interface {
		GetAll() (clusters types.Clusters, err types.Error)

		Get(clusterName string) (cluster *types.Cluster, err types.Error)

		Create(newCluster types.Cluster, who Requester) (types.Cluster, types.Error)

		Update(updatedCluster types.Cluster, who Requester) (types.Cluster, types.Error)

		Delete(clusterName string, who Requester) (e types.Error)
	}

	Organization interface {
		GetAll() (organizations types.Organizations, err types.Error)

		Get(organizationName string) (organization *types.Organization, err types.Error)

		Create(newOrganization types.Organization, who Requester) (types.Organization, types.Error)

		Update(updatedOrganization types.Organization, who Requester) (types.Organization, types.Error)

		Delete(organizationName string, who Requester) (e types.Error)
	}

	// Developer is the service interface to manipulate Developer entities
	Developer interface {
		GetAll(organizationName string) (developers types.Developers, err types.Error)

		Get(organizationName, developerEmail string) (developer *types.Developer, err types.Error)

		Create(organizationName string, newDeveloper types.Developer, who Requester) (types.Developer, types.Error)

		Update(organizationName, developerEmail string, updatedDeveloper types.Developer, who Requester) (types.Developer, types.Error)

		Delete(organizationName, developerEmail string, who Requester) (e types.Error)
	}

	// DeveloperApp is the service interface to manipulate DeveloperApp entities
	DeveloperApp interface {
		GetAll(organizationName string) (developerApps types.DeveloperApps, err types.Error)

		GetAllByEmail(organizationName, developerEmail string) (developerApps types.DeveloperApps, err types.Error)

		GetByName(organizationName, developerEmail, developerAppName string) (developerApp *types.DeveloperApp, err types.Error)

		GetByID(organizationName, developerAppName string) (developerApp *types.DeveloperApp, err types.Error)

		Create(organizationName, developerEmail string, newDeveloperApp types.DeveloperApp, who Requester) (types.DeveloperApp, types.Error)

		Update(organizationName, developerEmail string, updatedDeveloperApp types.DeveloperApp, who Requester) (types.DeveloperApp, types.Error)

		Delete(organizationName, developerEmail, developerAppName string, who Requester) (e types.Error)
	}

	// Key is the service interface to manipulate Key entities
	Key interface {
		Get(organizationName, developerEmail, appName, consumerKey string) (key *types.Key, err types.Error)

		GetByDeveloperAppID(organizationName, developerAppID string) (keys types.Keys, err types.Error)

		Create(organizationName string, newKey types.Key, developerApp *types.DeveloperApp, who Requester) (types.Key, types.Error)

		Update(organizationName, consumerKey string, updateKey *types.Key, who Requester) (types.Key, types.Error)

		Delete(organizationName, consumerKey string, who Requester) (e types.Error)
	}

	// APIProduct is the service interface to manipulate APIProduct entities
	APIProduct interface {
		GetAll(organizationName string) (apiproducts types.APIProducts, err types.Error)

		Get(organizationName, apiproductName string) (apiproduct *types.APIProduct, err types.Error)

		Create(organizationName string, newAPIProduct types.APIProduct, who Requester) (types.APIProduct, types.Error)

		Update(organizationName string, updatedAPIProduct types.APIProduct, who Requester) (types.APIProduct, types.Error)

		Delete(organizationName string, apiproductName string, who Requester) (e types.Error)
	}

	// User is the service interface to manipulate User entities
	User interface {
		GetAll() (users types.Users, err types.Error)

		Get(userName string) (user *types.User, err types.Error)

		Create(newUser *types.User, who Requester) (*types.User, types.Error)

		Update(updatedUser *types.User, who Requester) (*types.User, types.Error)

		Delete(userName string, who Requester) (e types.Error)
	}

	// Role is the service interface to manipulate Role entities
	Role interface {
		GetAll() (roles types.Roles, err types.Error)

		Get(roleName string) (role *types.Role, err types.Error)

		Create(newRole *types.Role, who Requester) (*types.Role, types.Error)

		Update(updatedRole *types.Role, who Requester) (*types.Role, types.Error)

		Delete(roleName string, who Requester) (e types.Error)
	}
)
