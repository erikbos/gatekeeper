package types

import "encoding/json"

// DeveloperAppKey contains an apikey entitlement
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type DeveloperAppKey struct {
	// apikey of this credential
	ConsumerKey string `json:"consumerKey"`

	// secretid of credential, used in OAuth2 flow
	ConsumerSecret string `json:"consumerSecret"`

	// List of apiproduct which can be accessed using this key
	APIProducts APIProductStatuses `json:"apiProducts"`

	// Attributes of credential
	Attributes Attributes `json:"attributes"`

	// Expiry date in epoch milliseconds
	ExpiresAt int64 `json:"expiresAt"`

	// Issue date in epoch milliseconds
	IssuedAt int64 `json:"issuesAt"`

	// Developer app id
	AppID string `json:"AppId"`

	// Status (should be "approved" to allow access)
	Status string `json:"status"`
}

// DeveloperAppKeys holds one or more apikeys
type DeveloperAppKeys []DeveloperAppKey

var (
	// NullDeveloperAppKey is an empty key type
	NullDeveloperAppKey = DeveloperAppKey{}

	// NullDeveloperAppKeys is an empty key slice
	NullDeveloperAppKeys = DeveloperAppKeys{}
)

// APIProductStatus contains whether an apikey's assigned apiproduct has been approved
type APIProductStatus struct {
	// Name of apiproduct
	Apiproduct string `json:"apiProduct"`

	// Status (should be "approved" to allow access)
	Status string `json:"status"`
}

// APIProductStatuses contains list of apiproducts
type APIProductStatuses []APIProductStatus

// SetApproved marks a credential as approved
func (k *DeveloperAppKey) SetApproved() {

	k.Status = "approved"
}

// IsApproved returns true in case credential's status is approved
func (k *DeveloperAppKey) IsApproved() bool {

	return k.Status == "approved"
}

// IsExpired returns true in case credential is expired
func (k *DeveloperAppKey) IsExpired(now int64) bool {

	if k.ExpiresAt == 0 || k.ExpiresAt == -1 {
		return false
	}
	if now < k.ExpiresAt {
		return false
	}
	return true
}

// SetApproved marks a credential's apiproduct as approved
func (p *APIProductStatus) SetApproved() {

	p.Status = "approved"
}

// IsApproved returns true in case credential's apiproduct status is approved
func (p *APIProductStatus) IsApproved() bool {

	return p.Status == "approved"
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
