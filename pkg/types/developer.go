package types

// Developer contains everything about a Developer
type Developer struct {
	// Id of developer (not changable)
	DeveloperID string `json:"developerId"`

	// Status of developer (should be "approved" to allow access)
	Status string `json:"status"`

	// Organization this developer belongs to
	OrganizationName string `json:"organizationName"`

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

	// if set developer is suspend till this time, epoch milliseconds
	SuspendedTill int64 `json:"suspendedTill"`

	// Last modified at timestamp in epoch milliseconds
	LastmodifiedAt int64 `json:"lastmodifiedAt"`

	// Name of user who last updated this organization
	LastmodifiedBy string `json:"lastmodifiedBy"`
}

// Developers holds one or more developers
type Developers []Developer

var (
	// NullDeveloper is an empty developer type
	NullDeveloper = Developer{}

	// NullDevelopers is an empty developer slice
	NullDevelopers = Developers{}
)
