package types

// Organization
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type (
	Organization struct {
		// Name of listener (not changable)
		Name string `binding:"required,min=4"`

		// Friendly display name of listener
		DisplayName string

		// Attributes of this listener
		Attributes Attributes

		// Created at timestamp in epoch milliseconds
		CreatedAt int64

		// Name of user who created this listener
		CreatedBy string

		// Last modified at timestamp in epoch milliseconds
		LastModifiedAt int64

		// Name of user who last updated this listener
		LastModifiedBy string
	}

	// Organizations holds one or more Organizations
	Organizations []Organization
)

var (
	// NullOrganization is an empty listener type
	NullOrganization = Organization{}

	// NullOrganizations is an empty listener slice
	NullOrganizations = Organizations{}
)
