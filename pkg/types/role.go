package types

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar"
)

// Role holds an role
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type Role struct {
	// Name of role (not changable)
	Name string `json:"name" binding:"required,min=4"`

	// Display name
	DisplayName string `json:"displayName"`

	// Allowed methods & paths
	Allows `json:"allows" binding:"required"`

	// Created at timestamp in epoch milliseconds
	CreatedAt int64 `json:"createdAt"`

	// Name of user who created this role
	CreatedBy string `json:"createdBy"`

	// Last modified at timestamp in epoch milliseconds
	LastmodifiedAt int64 `json:"lastmodifiedAt"`

	// Name of user who last updated this role
	LastmodifiedBy string `json:"lastmodifiedBy"`
}

// Roles holds one or more roles
type Roles []Role

// Allow holds the criteria a role will allow request
type Allow struct {
	// Request methods which are allowed
	// FIXME these bindings settings do not get used...
	Methods StringSlice `json:"methods" binding:"required,dive,oneof=GET POST PUT PATCH DELETE"`

	// Request paths (regexp) which are allowed
	Paths StringSlice `json:"paths" binding:"required,dive,startswith=/"`
}

// Allows holds one or more allow
type Allows []Allow

var (
	// NullRole is an empty role type
	NullRole = Role{}

	// NullRoles is an empty role slice
	NullRoles = Roles{}

	// NullAllow is an allow type
	NullAllow = Allow{}

	// NullAllows is an allows type
	NullAllows = Allows{}
)

// IsPathAllowed checks whether role is allowed to access a path
func (role *Role) IsPathAllowed(requestMethod, requestPath string) bool {

	for _, allow := range role.Allows {
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

// Unmarshal unpacks JSON-encoded role allow into Allows
func (a *Allows) Unmarshal(roleAllowsAsJSON string) Allows {

	if roleAllowsAsJSON != "" {
		var allows Allows
		if err := json.Unmarshal([]byte(roleAllowsAsJSON), &allows); err == nil {
			return allows
		}
	}
	return NullAllows
}

// Marshal packs role Allow into JSON
func (a *Allows) Marshal() string {

	if json, err := json.Marshal(a); err == nil {
		return string(json)
	}
	return "[]"
}

// Sort a slice of users
func (roles Roles) Sort() {
	// Sort roles by name
	sort.SliceStable(roles, func(i, j int) bool {
		return roles[i].Name < roles[j].Name
	})
}
