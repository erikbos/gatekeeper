package types

// DeveloperApp contains everything about a Developer Application
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type DeveloperApp struct {
	// Id of developer app (not changable)
	AppID string

	// Id of developer (not changable)
	DeveloperID string

	// Activation status of developer application
	Status string

	// Attributes of developer application
	Attributes Attributes

	// Name of developer application
	Name string

	// Friendly name of developer app
	DisplayName string

	// Created at timestamp in epoch milliseconds
	CreatedAt int64

	// Name of user who created this app
	CreatedBy string

	// Last modified at timestamp in epoch milliseconds
	LastModifiedAt int64

	// Name of user who last updated this app
	LastModifiedBy string
}

// DeveloperApps holds one or more developer apps
type DeveloperApps []DeveloperApp

var (
	// NullDeveloperApp is an empty developer app type
	NullDeveloperApp = DeveloperApp{}

	// NullDeveloperApps is an empty developer app slice
	NullDeveloperApps = DeveloperApps{}
)
