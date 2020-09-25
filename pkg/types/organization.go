package types

// Organization contains everything about an organization
type Organization struct {
	// Name of organization (not changable)
	Name string `json:"name" binding:"required"`

	// Friendly name of organization, can be updated afterwards
	DisplayName string `json:"displayName"`

	// Attributes of organization
	Attributes Attributes `json:"attributes"`

	// Created at timestamp in epoch milliseconds
	CreatedAt int64 `json:"createdAt"`

	// Name of user who created this organization
	CreatedBy string `json:"createdBy"`

	// Last modified at timestamp in epoch milliseconds
	LastmodifiedAt int64 `json:"lastmodifiedAt"`

	// Name of user who last updated this organization
	LastmodifiedBy string `json:"lastmodifiedBy"`
}

// Organizations holds one or more organizations
type Organizations []Organization

var (
	// NullOrganization is an empty organization type
	NullOrganization = Organization{}

	// NullOrganizations is an empty organization slice
	NullOrganizations = Organizations{}
)
