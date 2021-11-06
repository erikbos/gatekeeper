package types

// Developer contains everything about a Developer
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type Developer struct {
	// Id of developer (not changable)
	DeveloperID string

	// Status of developer (should be "approved" to allow access)
	Status string

	// Name of developer applications of this developer
	Apps StringSlice

	// Attributes of developer
	Attributes Attributes

	// Email address
	Email string

	// Username
	UserName string

	// First name
	FirstName string

	// Last name
	LastName string

	// Created at timestamp in epoch milliseconds
	CreatedAt int64

	// Name of user who created this organiz
	CreatedBy string

	// Last modified at timestamp in epoch milliseconds
	LastModifiedAt int64

	// Name of user who last updated this developer
	LastModifiedBy string
}

// Developers holds one or more developers
type Developers []Developer

var (
	// NullDeveloper is an empty developer type
	NullDeveloper = Developer{}

	// NullDevelopers is an empty developer slice
	NullDevelopers = Developers{}
)

// Activate marks a developer as approved
func (d *Developer) Activate() {

	d.Status = "active"
}

// Deactivate marks a developer as inactive
func (d *Developer) Deactivate() {

	d.Status = "inactive"
}

// IsActive returns true in case developer's status is active
func (d *Developer) IsActive() bool {

	return d.Status == "active"
}
