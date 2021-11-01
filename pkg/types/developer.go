package types

// Developer contains everything about a Developer
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type Developer struct {
	// Id of developer (not changable)
	DeveloperID string `json:"developerId"`

	// Status of developer (should be "approved" to allow access)
	Status string `json:"status"`

	// Name of developer applications of this developer
	Apps StringSlice `json:"apps"`

	// Attributes of developer
	Attributes Attributes `json:"attributes"`

	// Email address
	Email string `json:"email" binding:"required"`

	// Username
	UserName string `json:"userName" binding:"required"`

	// First name
	FirstName string `json:"firstName" binding:"required"`

	// Last name
	LastName string `json:"lastName" binding:"required"`

	// Created at timestamp in epoch milliseconds
	CreatedAt int64 `json:"createdAt"`

	// Name of user who created this organiz
	CreatedBy string `json:"createdBy"`

	// Last modified at timestamp in epoch milliseconds
	LastmodifiedAt int64 `json:"lastModifiedAt"`

	// Name of user who last updated this developer
	LastmodifiedBy string `json:"lastModifiedBy"`
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
