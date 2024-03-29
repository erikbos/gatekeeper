package db

import (
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
		Key
		Company
		APIProduct
		OAuth
		User
		Role
		Audit
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
		GetAll(organizationName string) (types.Developers, types.Error)

		// GetByEmail retrieves a developer
		GetByEmail(organizationName, developerEmail string) (*types.Developer, types.Error)

		// GetByID retrieves a developer
		GetByID(organizationName, developerID string) (*types.Developer, types.Error)

		// Update UPSERTs a developer
		Update(organizationName string, dev *types.Developer) types.Error

		// DeleteByID deletes a developer
		DeleteByID(organizationName, developerID string) types.Error
	}

	// DeveloperApp the developer app information storage interface
	DeveloperApp interface {
		// GetAll retrieves all developer apps
		GetAll(organizationName string) (types.DeveloperApps, types.Error)

		// GetAllByDeveloperID retrieves all developer apps from one developer
		GetAllByDeveloperID(organizationName, developerID string) (types.DeveloperApps, types.Error)

		// GetByName returns a developer app
		GetByName(organizationName, developerEmail, developerAppName string) (*types.DeveloperApp, types.Error)

		// GetByID returns a developer app
		GetByID(organizationName, developerAppID string) (*types.DeveloperApp, types.Error)

		// GetCountByDeveloperID retrieves number of apps belonging to a developer
		GetCountByDeveloperID(organizationName, developerID string) (int, types.Error)

		// UpdateByName UPSERTs a developer app
		Update(organizationName string, app *types.DeveloperApp) types.Error

		// DeleteByID deletes a developer app
		DeleteByID(organizationName, developerAppID string) types.Error
	}

	// Key the key information storage interface
	Key interface {
		// GetByKey returns details of a single apikey
		GetByKey(organizationName, key *string) (*types.Key, types.Error)

		// GetCountByAPIProductName returns number of keys that has apiproduct assigned
		GetCountByAPIProductName(organizationName, apiProductName string) (int, types.Error)

		// GetByDeveloperAppID returns an array with apikey details of a developer app
		GetByDeveloperAppID(organizationName, developerAppID string) (types.Keys, types.Error)

		// UpdateByKey UPSERTs key
		UpdateByKey(organizationName string, c *types.Key) types.Error

		// DeleteByKey deletes key
		DeleteByKey(organizationName, consumerKey string) types.Error
	}

	// Company the company information storage interface
	Company interface {
		GetAll(organizationName string) (types.Companies, types.Error)

		Get(organizationName, companyName string) (*types.Company, types.Error)

		Update(organizationName string, c *types.Company) types.Error

		Delete(organizationName, company string) types.Error
	}

	// APIProduct the apiproduct information storage interface
	APIProduct interface {
		// GetAll retrieves all api products
		GetAll(organizationName string) (types.APIProducts, types.Error)

		// Get returns an apiproduct
		Get(organizationName, apiproductName string) (*types.APIProduct, types.Error)

		// Update UPSERTs an apiproduct in database
		Update(organizationName string, p *types.APIProduct) types.Error

		// Delete deletes an apiproduct
		Delete(organizationName, apiProduct string) types.Error
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

		// Get retrieves a role
		Get(roleName string) (*types.Role, types.Error)

		// Update UPSERTs a role
		Update(c *types.Role) types.Error

		// Delete removes a role
		Delete(roleToDelete string) types.Error
	}

	Audit interface {
		// GetOrganization retrieves audit records of an organization
		GetOrganization(organizationName string, params AuditFilterParams) (types.Audits, types.Error)

		// GetAPIProduct retrieves audit records of an apiproduct
		GetAPIProduct(organizationName, apiproductName string, params AuditFilterParams) (types.Audits, types.Error)

		// GetDeveloper retrieves audit records of a developer
		GetDeveloper(organizationName, developerID string, params AuditFilterParams) (types.Audits, types.Error)

		// GetApplication retrieves audit records of an application
		GetApplication(organizationName, developerID, appID string, params AuditFilterParams) (types.Audits, types.Error)

		// GetUser retrieves audit records of a user
		GetUser(auditName string, params AuditFilterParams) (types.Audits, types.Error)

		// Write inserts an audit record
		Write(l *types.Audit) types.Error
	}

	AuditFilterParams struct {
		// Start timestamp in epoch milliseconds
		StartTime int64
		// End timestamp in epoch milliseconds
		EndTime int64
		// Maximum number of entities to return
		Count int64
	}
)
