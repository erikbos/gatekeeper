package shared

import (
	"encoding/json"
)

// Organization contains everything about an organization
type Organization struct {
	Name           string     `json:"name" binding:"required"` // Name of organization (not changable)
	DisplayName    string     `json:"displayName"`             // Friendly name of organization, can be updated afterwards
	Attributes     Attributes `json:"attributes"`              // Attributes of organization
	CreatedAt      int64      `json:"createdAt"`               // Created at timestamp in epoch milliseconds
	CreatedBy      string     `json:"createdBy"`               // Name of user who created this organization
	LastmodifiedAt int64      `json:"lastmodifiedAt"`          // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy string     `json:"lastmodifiedBy"`          // Name of user who last updated this organization
}

// Developer contains everything about a Developer
type Developer struct {
	DeveloperID      string      `json:"developerId"`                  // Id of developer (not changable)
	Status           string      `json:"status"`                       // Status of developer (should be "approved" to allow access)
	OrganizationName string      `json:"organizationName"`             // Organization this developer belongs to
	Apps             StringSlice `json:"apps"`                         // Name of developer applications of this developer
	Attributes       Attributes  `json:"attributes"`                   // Attributes of developer
	Email            string      `json:"email" binding:"required"`     // Email address
	UserName         string      `json:"userName" binding:"required"`  // Username
	FirstName        string      `json:"firstName" binding:"required"` // First name
	LastName         string      `json:"lastName" binding:"required"`  // Last name
	CreatedAt        int64       `json:"createdAt"`                    // Created at timestamp in epoch milliseconds
	CreatedBy        string      `json:"createdBy"`                    // Name of user who created this organiz
	SuspendedTill    int64       `json:"suspendedTill"`                // if set developer is suspend till this time, epoch milliseconds
	LastmodifiedAt   int64       `json:"lastmodifiedAt"`               // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy   string      `json:"lastmodifiedBy"`               // Name of user who last updated this organization
}

// DeveloperApp contains everything about a Developer Application
type DeveloperApp struct {
	AppID            string     `json:"appId"`                   // Id of developer app (not changable)
	DeveloperID      string     `json:"developerId"`             // Id of developer (not changable)
	OrganizationName string     `json:"organizationName"`        // Organization this virtual hosts belongs to (not used)
	Status           string     `json:"status"`                  // Activation status of developer application
	Attributes       Attributes `json:"attributes"`              // Attributes of developer application
	Name             string     `json:"name" binding:"required"` // Name of developer application
	DisplayName      string     `json:"displayName"`             // Friendly name of developer app , can be updated afterwards
	CreatedAt        int64      `json:"createdAt"`               // Created at timestamp in epoch milliseconds
	CreatedBy        string     `json:"createdBy"`               // Name of user who created this organization
	LastmodifiedAt   int64      `json:"lastmodifiedAt"`          // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy   string     `json:"lastmodifiedBy"`          // Name of user who last updated this organization
}

// DeveloperAppKey contains an apikey entitlement
type DeveloperAppKey struct {
	ConsumerKey      string             `json:"consumerKey"`      // apikey of this credential
	ConsumerSecret   string             `json:"consumerSecret"`   // secretid of crendetial, to be used in OAuth2 token rq
	APIProducts      APIProductStatuses `json:"apiProducts"`      // List of apiproduct which can be accessed
	Attributes       Attributes         `json:"attributes"`       // Attributes of credential
	ExpiresAt        int64              `json:"expiresAt"`        // Expiry date in epoch milliseconds
	IssuedAt         int64              `json:"issuesAt"`         // Issue date in epoch milliseconds
	AppID            string             `json:"AppId"`            // Developer app id
	OrganizationName string             `json:"organizationName"` // Organization this virtual hosts belongs to
	Status           string             `json:"status"`           // Status (should be "approved" to allow access)
}

// APIProductStatus contains whether an apikey's assigned apiproduct has been approved
type APIProductStatus struct {
	Status     string `json:"status"`     // Status (should be "approved" to allow access)
	Apiproduct string `json:"apiProduct"` // APIProduct name
}

// APIProductStatuses contains list of apiproducts
type APIProductStatuses []APIProductStatus

// APIProduct type contains everything about an API product
type APIProduct struct {
	Name             string      `json:"name"`             // Name of apiproduct (not changable)
	DisplayName      string      `json:"displayName"`      // Friendly display name of route
	Description      string      `json:"description"`      // Full description of this api product
	RouteGroup       string      `json:"RouteGroup"`       // Routegroup this apiproduct should match to
	Paths            StringSlice `json:"paths"`            // List of paths this apiproduct applies to
	Attributes       Attributes  `json:"attributes"`       // Attributes of this apiproduct
	Policies         string      `json:"policies"`         // Comma separated list of policynames, to apply to requests
	OrganizationName string      `json:"organizationName"` // Organization this virtual hosts belongs to (not used)
	CreatedAt        int64       `json:"createdAt"`        // Created at timestamp in epoch milliseconds
	CreatedBy        string      `json:"createdBy"`        // Name of user who created this apiproduct
	LastmodifiedAt   int64       `json:"lastmodifiedAt"`   // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy   string      `json:"lastmodifiedBy"`   // Name of user who last updated this apiproduct
}

// VirtualHost contains everything about downstream configuration of virtual hosts
type VirtualHost struct {
	Name             string      `json:"name"`             // Name of virtual host (not changable)
	DisplayName      string      `json:"displayName"`      // Friendly display name of route
	VirtualHosts     StringSlice `json:"virtualHosts"`     // List of virtualhosts
	Port             int         `json:"port"`             // tcp port to listen on
	RouteGroup       string      `json:"routeGroup"`       // Routegroup to forward traffic to
	Policies         string      `json:"policies"`         // Comma separated list of policynames, to apply to requests
	Attributes       Attributes  `json:"attributes"`       // Attributes of this virtual host
	OrganizationName string      `json:"organizationName"` // Organization this virtual hosts belongs to (not used)
	CreatedAt        int64       `json:"createdAt"`        // Created at timestamp in epoch milliseconds
	CreatedBy        string      `json:"createdBy"`        // Name of user who created this virtualhost
	LastmodifiedAt   int64       `json:"lastmodifiedAt"`   // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy   string      `json:"lastmodifiedBy"`   // Name of user who last updated this virtualhost
}

// Route holds configuration of one or more routes
type Route struct {
	Name           string     `json:"name"`           // Name of route (not changable)
	DisplayName    string     `json:"displayName"`    // Friendly display name of route
	RouteGroup     string     `json:"RouteGroup"`     // Routegroup this route is part of
	Path           string     `json:"path"`           // Path of route
	PathType       string     `json:"pathType"`       // Type of pathmatching: path, prefix, regexp
	Cluster        string     `json:"cluster"`        // Name of cluster to forward traffic to
	Attributes     Attributes `json:"attributes"`     // Attributes of this route
	CreatedAt      int64      `json:"createdAt"`      // Created at timestamp in epoch milliseconds
	CreatedBy      string     `json:"createdBy"`      // Name of user who created this route
	LastmodifiedAt int64      `json:"lastmodifiedAt"` // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy string     `json:"lastmodifiedBy"` // Name of user who last updated this route
}

// Cluster holds configuration of an upstream cluster
type Cluster struct {
	Name           string     `json:"name"`           // Name of cluster (not changable)
	DisplayName    string     `json:"displayName"`    // Friendly display name of cluster
	HostName       string     `json:"hostName"`       // Hostname of cluster
	Port           int        `json:"port"`           // tcp port of cluster
	Attributes     Attributes `json:"attributes"`     // Attributes of this cluster
	CreatedAt      int64      `json:"createdAt"`      // Created at timestamp in epoch milliseconds
	CreatedBy      string     `json:"createdBy"`      // Name of user who created this cluster
	LastmodifiedAt int64      `json:"lastmodifiedAt"` // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy string     `json:"lastmodifiedBy"` // Name of user who last updated this cluster
}

// OAuthAccessToken holds details of an issued OAuth token
type OAuthAccessToken struct {
	ClientID         string `json:"client_id"`
	UserID           string `json:"user_id"`
	RedirectURI      string `json:"redirect_uri"`
	Scope            string `json:"scope"`
	Code             string `json:"code"`
	CodeCreatedAt    int64  `json:"code_created_at"`
	CodeExpiresIn    int64  `json:"code_expires_in"`
	Access           string `json:"access"`
	AccessCreatedAt  int64  `json:"access_created_at"`
	AccessExpiresIn  int64  `json:"access_expires_in"`
	Refresh          string `json:"refresh"`
	RefreshCreatedAt int64  `json:"refresh_created_at"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

// Unmarshal unpacks JSON array of attribute bags
// Example input: [{"name":"S","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (ps APIProductStatuses) Unmarshal(jsonProductStatuses string) []APIProductStatus {

	if jsonProductStatuses != "" {
		var productStatus = make([]APIProductStatus, 0)
		if err := json.Unmarshal([]byte(jsonProductStatuses), &productStatus); err == nil {
			return productStatus
		}
	}
	return []APIProductStatus{}
}

// Marshal packs array of attributes into JSON
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (ps APIProductStatuses) Marshal() string {

	if len(ps) > 0 {
		ArrayOfAttributesInJSON, err := json.Marshal(ps)
		if err == nil {
			return string(ArrayOfAttributesInJSON)
		}
	}
	return "[]"
}

// StringSlice holds a number of strings
type StringSlice []string

// Unmarshal unpacks JSON to slice of strings
// e.g. [\"PetStore5\",\"PizzaShop1\"] to []string
//
func (s StringSlice) Unmarshal(jsonArrayOfStrings string) StringSlice {

	if jsonArrayOfStrings != "" {
		var StringValues []string
		err := json.Unmarshal([]byte(jsonArrayOfStrings), &StringValues)
		if err == nil {
			return StringValues
		}
	}
	return StringSlice{}
}

// Marshal packs slice of strings into JSON
// e.g. []string to [\"PetStore5\",\"PizzaShop1\"]
//
func (s StringSlice) Marshal() string {

	if len(s) > 0 {
		ArrayOfStringsInJSON, err := json.Marshal(s)
		if err == nil {
			return string(ArrayOfStringsInJSON)
		}
	}
	return "[]"
}
