package types

import "github.com/go-playground/validator/v10"

// Organization holds details of an organization
type (
	Organization struct {
		// Name of organization (not changable)
		Name string `validate:"required,min=1,max=100"`

		// Friendly display name of organization
		DisplayName string

		// Attributes of this organization
		Attributes Attributes

		// Created at timestamp in epoch milliseconds
		CreatedAt int64

		// Name of user who created this organization
		CreatedBy string

		// Last modified at timestamp in epoch milliseconds
		LastModifiedAt int64

		// Name of user who last updated this organization
		LastModifiedBy string
	}

	// Organizations holds one or more Organizations
	Organizations []Organization
)

var (
	// NullOrganization is an empty organization type
	NullOrganization = Organization{}

	// NullOrganizations is an empty organization slice
	NullOrganizations = Organizations{}
)

// Validate checks if field values are set correct and are allowed
func (o *Organization) Validate() error {

	validate := validator.New()
	return validate.Struct(o)
}
