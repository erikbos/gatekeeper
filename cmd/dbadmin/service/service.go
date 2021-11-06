package service

import (
	"github.com/erikbos/gatekeeper/pkg/types"
)

// Service can manipulate all our entities
type Service struct {
	Listener
	Route
	Cluster
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

		GetAttributes(listenerName string) (attributes *types.Attributes, err types.Error)

		GetAttribute(listenerName, attributeName string) (value string, err types.Error)

		Create(newListener types.Listener, who Requester) (types.Listener, types.Error)

		Update(updatedListener types.Listener, who Requester) (types.Listener, types.Error)

		UpdateAttributes(listenerName string, receivedAttributes types.Attributes, who Requester) types.Error

		UpdateAttribute(listenerName string, attributeValue types.Attribute, who Requester) types.Error

		DeleteAttribute(listenerName, attributeToDelete string, who Requester) (string, types.Error)

		Delete(listenerName string, who Requester) (deletedListener types.Listener, e types.Error)
	}

	// Route is the service interface to manipulate Route entities
	Route interface {
		GetAll() (routes types.Routes, err types.Error)

		Get(routeName string) (route *types.Route, err types.Error)

		GetAttributes(routeName string) (attributes *types.Attributes, err types.Error)

		GetAttribute(routeName, attributeName string) (value string, err types.Error)

		Create(newRoute types.Route, who Requester) (types.Route, types.Error)

		Update(updatedRoute types.Route, who Requester) (types.Route, types.Error)

		UpdateAttributes(routeName string, receivedAttributes types.Attributes, who Requester) types.Error

		UpdateAttribute(routeName string, attributeValue types.Attribute, who Requester) types.Error

		DeleteAttribute(routeName, attributeToDelete string, who Requester) (string, types.Error)

		Delete(routeName string, who Requester) (deletedRoute types.Route, e types.Error)
	}

	// Cluster is the service interface to manipulate Cluster entities
	Cluster interface {
		GetAll() (clusters types.Clusters, err types.Error)

		Get(clusterName string) (cluster *types.Cluster, err types.Error)

		GetAttributes(clusterName string) (attributes *types.Attributes, err types.Error)

		GetAttribute(clusterName, attributeName string) (value string, err types.Error)

		Create(newCluster types.Cluster, who Requester) (types.Cluster, types.Error)

		Update(updatedCluster types.Cluster, who Requester) (types.Cluster, types.Error)

		UpdateAttributes(clusterName string, receivedAttributes types.Attributes, who Requester) types.Error

		UpdateAttribute(clusterName string, attributeValue types.Attribute, who Requester) types.Error

		DeleteAttribute(clusterName, attributeToDelete string, who Requester) (string, types.Error)

		Delete(clusterName string, who Requester) (deletedCluster types.Cluster, e types.Error)
	}

	// Developer is the service interface to manipulate Developer entities
	Developer interface {
		GetAll() (developers types.Developers, err types.Error)

		Get(developerName string) (developer *types.Developer, err types.Error)

		GetAttributes(developerName string) (attributes *types.Attributes, err types.Error)

		GetAttribute(developerName, attributeName string) (value string, err types.Error)

		Create(newDeveloper types.Developer, who Requester) (types.Developer, types.Error)

		Update(developerName string, updatedDeveloper types.Developer, who Requester) (types.Developer, types.Error)

		UpdateAttributes(developerName string, receivedAttributes types.Attributes, who Requester) types.Error

		UpdateAttribute(developerName string, attributeValue types.Attribute, who Requester) types.Error

		DeleteAttribute(developerName, attributeToDelete string, who Requester) (string, types.Error)

		Delete(developerName string, who Requester) (deletedDeveloper types.Developer, e types.Error)
	}

	// DeveloperApp is the service interface to manipulate DeveloperApp entities
	DeveloperApp interface {
		GetAll() (developerApps types.DeveloperApps, err types.Error)

		GetAllByEmail(developerName string) (developerApps types.DeveloperApps, err types.Error)

		GetByName(developerAppName string) (developerApp *types.DeveloperApp, err types.Error)

		GetByID(developerAppName string) (developerApp *types.DeveloperApp, err types.Error)

		GetAttributes(developerAppName string) (attributes *types.Attributes, err types.Error)

		GetAttribute(developerAppName, attributeName string) (value string, err types.Error)

		Create(developerName string, newDeveloperApp types.DeveloperApp, who Requester) (types.DeveloperApp, types.Error)

		Update(updatedDeveloperApp types.DeveloperApp, who Requester) (types.DeveloperApp, types.Error)

		UpdateAttributes(developerAppName string, receivedAttributes types.Attributes, who Requester) types.Error

		UpdateAttribute(developerAppName string, attributeValue types.Attribute, who Requester) types.Error

		DeleteAttribute(developerAppName, attributeToDelete string, who Requester) (string, types.Error)

		Delete(developerID, developerAppName string, who Requester) (deletedDeveloperApp types.DeveloperApp, e types.Error)
	}

	// Key is the service interface to manipulate Key entities
	Key interface {
		GetAll() (keys types.Keys, err types.Error)

		Get(consumerKey string) (key *types.Key, err types.Error)

		GetByDeveloperAppID(developerAppID string) (keys types.Keys, err types.Error)

		Create(newKey types.Key, developerApp *types.DeveloperApp, who Requester) (types.Key, types.Error)

		Update(updateKey types.Key, who Requester) (types.Key, types.Error)

		Delete(consumerKey string, who Requester) (deletedKey types.Key, e types.Error)
	}

	// APIProduct is the service interface to manipulate APIProduct entities
	APIProduct interface {
		GetAll() (apiproducts types.APIProducts, err types.Error)

		Get(apiproductName string) (apiproduct *types.APIProduct, err types.Error)

		GetAttributes(apiproductName string) (attributes *types.Attributes, err types.Error)

		GetAttribute(apiproductName, attributeName string) (value string, err types.Error)

		Create(newAPIProduct types.APIProduct, who Requester) (types.APIProduct, types.Error)

		Update(updatedAPIProduct types.APIProduct, who Requester) (types.APIProduct, types.Error)

		UpdateAttributes(apiproductName string, receivedAttributes types.Attributes, who Requester) types.Error

		UpdateAttribute(apiproductName string, attributeValue types.Attribute, who Requester) types.Error

		DeleteAttribute(apiproductName, attributeToDelete string, who Requester) (string, types.Error)

		Delete(apiproductName string, who Requester) (deletedAPIProduct types.APIProduct, e types.Error)
	}

	// User is the service interface to manipulate User entities
	User interface {
		GetAll() (users types.Users, err types.Error)

		Get(userName string) (user *types.User, err types.Error)

		Create(newUser types.User, who Requester) (*types.User, types.Error)

		Update(updatedUser types.User, who Requester) (*types.User, types.Error)

		Delete(userName string, who Requester) (deletedUser *types.User, e types.Error)
	}

	// Role is the service interface to manipulate Role entities
	Role interface {
		GetAll() (roles types.Roles, err types.Error)

		Get(roleName string) (role *types.Role, err types.Error)

		Create(newRole types.Role, who Requester) (*types.Role, types.Error)

		Update(updatedRole types.Role, who Requester) (*types.Role, types.Error)

		Delete(roleName string, who Requester) (deletedRole *types.Role, e types.Error)
	}
)
