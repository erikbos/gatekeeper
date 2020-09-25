package types

// APIProduct type contains everything about an API product
type APIProduct struct {
	// Name of apiproduct (not changable)
	Name string `json:"name"`

	// Friendly display name of route
	DisplayName string `json:"displayName"`

	// Full description of this api product
	Description string `json:"description"`

	// Routegroup this apiproduct should match to
	RouteGroup string `json:"RouteGroup"`

	// List of paths this apiproduct applies to
	Paths StringSlice `json:"paths"`

	// Attributes of this apiproduct
	Attributes Attributes `json:"attributes"`

	// Comma separated list of policynames, to apply to requests
	Policies string `json:"policies"`

	// Organization this api product belongs to (not used)
	OrganizationName string `json:"organizationName"`

	// Created at timestamp in epoch milliseconds
	CreatedAt int64 `json:"createdAt"`

	// Name of user who created this apiproduct
	CreatedBy string `json:"createdBy"`

	// Last modified at timestamp in epoch milliseconds
	LastmodifiedAt int64 `json:"lastmodifiedAt"`

	// Name of user who last updated this apiproduct
	LastmodifiedBy string `json:"lastmodifiedBy"`
}

// APIProducts holds one or more apiproducts
type APIProducts []APIProduct

var (
	// NullAPIProduct is an empty apiproduct type
	NullAPIProduct = APIProduct{}

	// NullAPIProducts is an empty apiproduct slice
	NullAPIProducts = APIProducts{}
)
