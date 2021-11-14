package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

type (
	// Database is our overall database interface
	Database struct {
		Listener
		Route
		Cluster
		Organization
		Developer
		DeveloperApp
		APIProduct
		Key
		OAuth
		User
		Role
		Readiness
	}

	// Listener is the listener information storage interface
	Listener interface {
		// GetAll retrieves all listeners
		GetAll() (types.Listeners, types.Error)

		// Get retrieves a listener
		Get(listener string) (*types.Listener, types.Error)

		// Update updates a listener
		Update(listener *types.Listener) types.Error

		// Delete deletes a listener
		Delete(listenerToDelete string) types.Error
	}

	// Route the route information storage interface
	Route interface {
		// GetAll retrieves all routes
		GetAll() (types.Routes, types.Error)

		// Get retrieves a route from database
		Get(routeName string) (*types.Route, types.Error)

		// Update UPSERTs an route
		Update(route *types.Route) types.Error

		// Delete deletes a route
		Delete(routeToDelete string) types.Error
	}

	// Cluster the cluster information storage interface
	Cluster interface {
		// GetAll retrieves all clusters
		GetAll() (types.Clusters, types.Error)

		// Get retrieves a cluster from database
		Get(clusterName string) (*types.Cluster, types.Error)

		// Update UPSERTs an cluster in database
		Update(c *types.Cluster) types.Error

		// Update UPSERTs an cluster in database
		Delete(clusterToDelete string) types.Error
	}

	Organization interface {
		// GetAll retrieves all users
		GetAll() (types.Organizations, types.Error)

		// Get retrieves a user from database
		Get(userName string) (*types.Organization, types.Error)

		// Update UPSERTs an user in database
		Update(c *types.Organization) types.Error

		// Update UPSERTs an user in database
		Delete(organizationToDelete string) types.Error
	}
	// Developer the developer information storage interface
	Developer interface {
		// GetAll retrieves all developer
		GetAll() (types.Developers, types.Error)

		// GetByEmail retrieves a developer
		GetByEmail(developerEmail string) (*types.Developer, types.Error)

		// GetByID retrieves a developer
		GetByID(developerID string) (*types.Developer, types.Error)

		// Update UPSERTs a developer
		Update(dev *types.Developer) types.Error

		// DeleteByID deletes a developer
		DeleteByID(developerID string) types.Error
	}

	// DeveloperApp the developer app information storage interface
	DeveloperApp interface {
		// GetAll retrieves all developer apps
		GetAll() (types.DeveloperApps, types.Error)

		// GetAllByDeveloperID retrieves all developer apps from one developer
		GetAllByDeveloperID(developerID string) (types.DeveloperApps, types.Error)

		// GetByName returns a developer app
		GetByName(developerAppName string) (*types.DeveloperApp, types.Error)

		// GetByID returns a developer app
		GetByID(developerAppID string) (*types.DeveloperApp, types.Error)

		// GetCountByDeveloperID retrieves number of apps belonging to a developer
		GetCountByDeveloperID(developerID string) (int, types.Error)

		// UpdateByName UPSERTs a developer app
		Update(app *types.DeveloperApp) types.Error

		// DeleteByID deletes a developer app
		DeleteByID(developerAppID string) types.Error
	}

	// APIProduct the apiproduct information storage interface
	APIProduct interface {
		// GetAll retrieves all api products
		GetAll() (types.APIProducts, types.Error)

		// Get returns an apiproduct
		Get(apiproductName string) (*types.APIProduct, types.Error)

		// Update UPSERTs an apiproduct in database
		Update(p *types.APIProduct) types.Error

		// Delete deletes an apiproduct
		Delete(apiProduct string) types.Error
	}

	// Key the cluster information storage interface
	Key interface {
		// GetAll retrieves all listeners
		GetAll() (types.Keys, types.Error)

		// GetByKey returns details of a single apikey
		GetByKey(key *string) (*types.Key, types.Error)

		// GetCountByAPIProductName returns number of keys that has apiproduct assigned
		GetCountByAPIProductName(apiProductName string) (int, types.Error)

		// GetByDeveloperAppID returns an array with apikey details of a developer app
		GetByDeveloperAppID(developerAppID string) (types.Keys, types.Error)

		// UpdateByKey UPSERTs key
		UpdateByKey(c *types.Key) types.Error

		// DeleteByKey deletes key
		DeleteByKey(consumerKey string) types.Error
	}

	// OAuth the oauth information storage interface
	OAuth interface {
		// OAuthAccessTokenGetByAccess retrieves an access token
		OAuthAccessTokenGetByAccess(accessToken string) (*types.OAuthAccessToken, error)

		// OAuthAccessTokenGetByCode retrieves token by code
		OAuthAccessTokenGetByCode(code string) (*types.OAuthAccessToken, error)

		// OAuthAccessTokenGetByRefresh retrieves token by refreshcode
		OAuthAccessTokenGetByRefresh(refresh string) (*types.OAuthAccessToken, error)

		// OAuthAccessTokenCreate creates an access token
		OAuthAccessTokenCreate(t *types.OAuthAccessToken) error

		// OAuthAccessTokenRemoveByAccess deletes an access token
		OAuthAccessTokenRemoveByAccess(accessTokenToDelete string) error

		// OAuthAccessTokenRemoveByCode deletes an access token
		OAuthAccessTokenRemoveByCode(codeToDelete string) error

		// OAuthAccessTokenRemoveByRefresh deletes an access token
		OAuthAccessTokenRemoveByRefresh(refreshToDelete string) error
	}

	// User the user information storage interface
	User interface {
		// GetAll retrieves all users
		GetAll() (types.Users, types.Error)

		// Get retrieves a user from database
		Get(userName string) (*types.User, types.Error)

		// Update UPSERTs an user in database
		Update(c *types.User) types.Error

		// Update UPSERTs an user in database
		Delete(userToDelete string) types.Error
	}

	// Role the role information storage interface
	Role interface {
		// GetAll retrieves all roles
		GetAll() (types.Roles, types.Error)

		// Get retrieves a role from database
		Get(roleName string) (*types.Role, types.Error)

		// Update UPSERTs a role in database
		Update(c *types.Role) types.Error

		// Update UPSERTs a role in database
		Delete(roleToDelete string) types.Error
	}

	// Readiness the readiness storage interface
	Readiness interface {
		// RunReadinessCheck runs a database readiness check continously
		RunReadinessCheck(chan shared.ReadinessMessage)
	}
)
