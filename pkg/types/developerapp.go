package types

import "github.com/go-playground/validator/v10"

// DeveloperApp contains everything about a Developer Application
type (
	DeveloperApp struct {
		// Id of developer app (not changable)
		AppID string

		// Id of developer (not changable)
		DeveloperID string

		// Activation status of developer application
		Status string

		// Attributes of developer application
		Attributes Attributes

		// Name of developer application
		Name string `validate:"required,min=1"`

		// Friendly name of developer app
		DisplayName string

		// OAuth scopes
		Scopes []string

		// OAuth call back URL
		CallbackUrl string

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
	DeveloperApps []DeveloperApp
)

var (
	// NullDeveloperApp is an empty developer app type
	NullDeveloperApp = DeveloperApp{}

	// NullDeveloperApps is an empty developer app slice
	NullDeveloperApps = DeveloperApps{}
)

// Validate checks if field values are set correct and are allowed
func (d *DeveloperApp) Validate() error {

	validate := validator.New()
	return validate.Struct(d)
}

const (
	applicationStatusApproved = "approved"
	applicationStatusRevoked  = "revoked"
)

// Activate marks a developer as approved
func (d *DeveloperApp) Approve() {

	d.Status = applicationStatusApproved
}

// Deactivate marks a developer as inactive
func (d *DeveloperApp) Revoke() {

	d.Status = applicationStatusRevoked
}

// IsActive returns true in case developer's status is active
func (d *DeveloperApp) IsActive() bool {

	return d.Status == applicationStatusApproved
}
