package types

// Organization contains everything about an organization
type Organization struct {
	Name           string     `json:"name" binding:"required"` // Name of organization (not changable)
	DisplayName    string     `json:"displayName"`             // Friendly name of organization, can be updated afterwards
	Attributes     Attributes `json:"attributes"`              // Attributes of organization
	CreatedAt      int64      `json:"createdAt"`               // Created at timestamp in epoch milliseconds
	CreatedBy      string     `json:"createdBy"`               // Name of user who created this organization
	LastmodifiedAt int64      `json:"lastmodifiedAt"`          // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy string     `json:"lastmodifiedBy"`          // Name of user who last updated this organization
}

// Organizations holds one or more organizations
type Organizations []Organization
