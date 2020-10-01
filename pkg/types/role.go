package types

import (
	"encoding/json"
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

// PathAllowed checks whether a role allows to access a path
func (role *Role) PathAllowed(requestMethod, requestPath string) bool {

	// log.Printf("PathAllowed: %s, %s, %s", role.Name, requestMethod, requestPath)
	for _, allow := range role.Allows {
		if isMethodAllowed(allow.Methods, requestMethod) &&
			isPathAllowed(allow.Paths, requestPath) {
			return true
		}
	}
	// by default we do not allow access
	return false
}

// isMethodAllowed checks if methods exists in a slice of methods
func isMethodAllowed(methods []string, requestMethod string) bool {

	// log.Printf("checkIsMethodAllowed: %+v %s", methods, requestMethod)
	for _, method := range methods {
		if requestMethod == strings.ToUpper(method) {
			return true
		}
	}
	return false
}

// isMethodAllowed checks if path matches one of the paths
func isPathAllowed(paths []string, requestPath string) bool {

	// log.Printf("checkIsPathAllowed: %+v, %s", paths, requestPath)
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
