package types

import (
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/go-playground/validator/v10"
)

// Role holds an role
type (
	Role struct {
		// Name of role (not changable)
		Name string `validate:"required,min=1,max=100"`

		// Display name
		DisplayName string

		// Allowed methods & paths
		Permissions `validate:"required"`

		// Created at timestamp in epoch milliseconds
		CreatedAt int64

		// Name of user who created this role
		CreatedBy string

		// Last modified at timestamp in epoch milliseconds
		LastModifiedAt int64

		// Name of user who last updated this role
		LastModifiedBy string
	}

	// Roles holds one or more roles
	Roles []Role
)

// Permission holds the criteria a role will allow request
type Permission struct {
	// Request methods which are allowed
	// FIXME these bindings settings do not get used...
	Methods []string `validate:"dive,oneof=GET POST PUT PATCH DELETE,required"`

	// Request paths (regexp) which are allowed
	Paths []string `validate:"dive,startswith=/,required"`
}

// Permissions holds one or more allow
type Permissions []Permission

var (
	// NullRole is an empty role type
	NullRole = Role{}

	// NullRoles is an empty role slice
	NullRoles = Roles{}

	// NullPermission is an allow type
	NullPermission = Permission{}

	// NullPermissions is an allows type
	NullPermissions = Permissions{}
)

// Validate checks if field values are set correct and are allowed
func (r *Role) Validate() error {

	validate := validator.New()
	return validate.Struct(r)
}

// IsPathAllowed checks whether role is allowed to access a path
func (role *Role) IsPathAllowed(requestMethod, requestPath string) bool {

	for _, allow := range role.Permissions {
		if methodMatch(allow.Methods, requestMethod) &&
			pathMatch(allow.Paths, requestPath) {
			return true
		}
	}
	// by default we do not allow access
	return false
}

// methodMatch checks if methods exists in a slice of methods
func methodMatch(methods []string, requestMethod string) bool {

	for _, method := range methods {
		if requestMethod == strings.ToUpper(method) {
			return true
		}
	}
	return false
}

// pathMatch checks if path matches one of the paths
func pathMatch(paths []string, requestPath string) bool {

	for _, path := range paths {
		matched, err := doublestar.Match(path, requestPath)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// Sort a slice of users
func (roles Roles) Sort() {
	// Sort roles by name
	sort.SliceStable(roles, func(i, j int) bool {
		return roles[i].Name < roles[j].Name
	})
}
