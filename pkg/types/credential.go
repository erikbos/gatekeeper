package types

import "encoding/json"

// DeveloperAppKey contains an apikey entitlement
type DeveloperAppKey struct {
	ConsumerKey      string             `json:"consumerKey"`      // apikey of this credential
	ConsumerSecret   string             `json:"consumerSecret"`   // secretid of crendetial, to be used in OAuth2 token rq
	APIProducts      APIProductStatuses `json:"apiProducts"`      // List of apiproduct which can be accessed
	Attributes       Attributes         `json:"attributes"`       // Attributes of credential
	ExpiresAt        int64              `json:"expiresAt"`        // Expiry date in epoch milliseconds
	IssuedAt         int64              `json:"issuesAt"`         // Issue date in epoch milliseconds
	AppID            string             `json:"AppId"`            // Developer app id
	OrganizationName string             `json:"organizationName"` // Organization this credential belongs to
	Status           string             `json:"status"`           // Status (should be "approved" to allow access)
}

// DeveloperAppKeys holds one or more apikeys
type DeveloperAppKeys []DeveloperAppKey

// APIProductStatus contains whether an apikey's assigned apiproduct has been approved
type APIProductStatus struct {
	Status     string `json:"status"`     // Status (should be "approved" to allow access)
	Apiproduct string `json:"apiProduct"` // Name of apiproduct
}

// APIProductStatuses contains list of apiproducts
type APIProductStatuses []APIProductStatus

// OAuthAccessToken holds details of an issued OAuth token
type OAuthAccessToken struct {
	ClientID         string `json:"client_id"`
	UserID           string `json:"user_id"`
	RedirectURI      string `json:"redirect_uri"`
	Scope            string `json:"scope"`
	Code             string `json:"code"`
	CodeCreatedAt    int64  `json:"code_created_at"`
	CodeExpiresIn    int64  `json:"code_expires_in"`
	Access           string `json:"access"`
	AccessCreatedAt  int64  `json:"access_created_at"`
	AccessExpiresIn  int64  `json:"access_expires_in"`
	Refresh          string `json:"refresh"`
	RefreshCreatedAt int64  `json:"refresh_created_at"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

// Unmarshal unpacks JSON array of attribute bags
// Example input: [{"name":"S","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (ps APIProductStatuses) Unmarshal(jsonProductStatuses string) []APIProductStatus {

	if jsonProductStatuses != "" {
		var productStatus = make([]APIProductStatus, 0)
		if err := json.Unmarshal([]byte(jsonProductStatuses), &productStatus); err == nil {
			return productStatus
		}
	}
	return []APIProductStatus{}
}

// Marshal packs array of attributes into JSON
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (ps APIProductStatuses) Marshal() string {

	if len(ps) > 0 {
		ArrayOfAttributesInJSON, err := json.Marshal(ps)
		if err == nil {
			return string(ArrayOfAttributesInJSON)
		}
	}
	return "[]"
}
