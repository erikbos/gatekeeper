package types

// APIProduct type contains everything about an API product
type APIProduct struct {
	Name             string      `json:"name"`             // Name of apiproduct (not changable)
	DisplayName      string      `json:"displayName"`      // Friendly display name of route
	Description      string      `json:"description"`      // Full description of this api product
	RouteGroup       string      `json:"RouteGroup"`       // Routegroup this apiproduct should match to
	Paths            StringSlice `json:"paths"`            // List of paths this apiproduct applies to
	Attributes       Attributes  `json:"attributes"`       // Attributes of this apiproduct
	Policies         string      `json:"policies"`         // Comma separated list of policynames, to apply to requests
	OrganizationName string      `json:"organizationName"` // Organization this api product belongs to (not used)
	CreatedAt        int64       `json:"createdAt"`        // Created at timestamp in epoch milliseconds
	CreatedBy        string      `json:"createdBy"`        // Name of user who created this apiproduct
	LastmodifiedAt   int64       `json:"lastmodifiedAt"`   // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy   string      `json:"lastmodifiedBy"`   // Name of user who last updated this apiproduct
}

// APIProducts holds one or more apiproducts
type APIProducts []APIProduct
