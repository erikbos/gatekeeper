package types

// Developer contains everything about a Developer
type Developer struct {
	DeveloperID      string      `json:"developerId"`                  // Id of developer (not changable)
	Status           string      `json:"status"`                       // Status of developer (should be "approved" to allow access)
	OrganizationName string      `json:"organizationName"`             // Organization this developer belongs to
	Apps             StringSlice `json:"apps"`                         // Name of developer applications of this developer
	Attributes       Attributes  `json:"attributes"`                   // Attributes of developer
	Email            string      `json:"email" binding:"required"`     // Email address
	UserName         string      `json:"userName" binding:"required"`  // Username
	FirstName        string      `json:"firstName" binding:"required"` // First name
	LastName         string      `json:"lastName" binding:"required"`  // Last name
	CreatedAt        int64       `json:"createdAt"`                    // Created at timestamp in epoch milliseconds
	CreatedBy        string      `json:"createdBy"`                    // Name of user who created this organiz
	SuspendedTill    int64       `json:"suspendedTill"`                // if set developer is suspend till this time, epoch milliseconds
	LastmodifiedAt   int64       `json:"lastmodifiedAt"`               // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy   string      `json:"lastmodifiedBy"`               // Name of user who last updated this organization
}

// Developers holds one or more developers
type Developers []Developer
