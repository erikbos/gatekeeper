package types

// DeveloperApp contains everything about a Developer Application
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type DeveloperApp struct {
	// Id of developer app (not changable)
	AppID string `json:"appId"`

	// Id of developer (not changable)
	DeveloperID string `json:"developerId"`

	// Organization this developer app belongs to (not used)
	OrganizationName string `json:"organizationName"`

	// Activation status of developer application
	Status string `json:"status"`

	// Attributes of developer application
	Attributes Attributes `json:"attributes"`

	// Name of developer application
	Name string `json:"name" binding:"required"`

	// Friendly name of developer app , can be updated afterwards
	DisplayName string `json:"displayName"`

	// Created at timestamp in epoch milliseconds
	CreatedAt int64 `json:"createdAt"`

	// Name of user who created this organization
	CreatedBy string `json:"createdBy"`

	// Last modified at timestamp in epoch milliseconds
	LastmodifiedAt int64 `json:"lastmodifiedAt"`

	// Name of user who last updated this organization
	LastmodifiedBy string `json:"lastmodifiedBy"`
}

// DeveloperApps holds one or more developer apps
type DeveloperApps []DeveloperApp

var (
	// NullDeveloperApp is an empty developer app type
	NullDeveloperApp = DeveloperApp{}

	// NullDeveloperApps is an empty developer app slice
	NullDeveloperApps = DeveloperApps{}
)
