package types

// DeveloperApp contains everything about a Developer Application
type DeveloperApp struct {
	AppID            string     `json:"appId"`                   // Id of developer app (not changable)
	DeveloperID      string     `json:"developerId"`             // Id of developer (not changable)
	OrganizationName string     `json:"organizationName"`        // Organization this developer app belongs to (not used)
	Status           string     `json:"status"`                  // Activation status of developer application
	Attributes       Attributes `json:"attributes"`              // Attributes of developer application
	Name             string     `json:"name" binding:"required"` // Name of developer application
	DisplayName      string     `json:"displayName"`             // Friendly name of developer app , can be updated afterwards
	CreatedAt        int64      `json:"createdAt"`               // Created at timestamp in epoch milliseconds
	CreatedBy        string     `json:"createdBy"`               // Name of user who created this organization
	LastmodifiedAt   int64      `json:"lastmodifiedAt"`          // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy   string     `json:"lastmodifiedBy"`          // Name of user who last updated this organization
}

// DeveloperApps holds one or more developer apps
type DeveloperApps []DeveloperApp
