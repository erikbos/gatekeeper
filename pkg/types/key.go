package types

// Key contains an apikey entitlement
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type (
	Key struct {
		// ConsumerKey is the key required for authentication
		ConsumerKey string

		// ConsumerSecret is secretid of this key, needed to request OAuth2 access token
		ConsumerSecret string

		// List of apiproducts which can be accessed using this key
		APIProducts KeyAPIProductStatuses

		// Expiry date in epoch milliseconds
		ExpiresAt int64

		// Issue date in epoch milliseconds
		IssuedAt int64

		// Attributes of key
		Attributes Attributes

		// Developer app id this key belongs to
		AppID string

		// Status (should be "approved" to allow access)
		Status string
	}

	// Keys holds one or more apikeys
	Keys []Key
)

var (
	// NullDeveloperAppKey is an empty key type
	NullDeveloperAppKey = Key{}

	// NullDeveloperAppKeys is an empty key slice
	NullDeveloperAppKeys = Keys{}
)

// KeyAPIProductStatus contains whether an apikey's assigned apiproduct has been approved
type KeyAPIProductStatus struct {
	// Name of apiproduct
	Apiproduct string `json:"apiProduct"`

	// Status (should be "approved" to allow access)
	Status string `json:"status"`
}

// KeyAPIProductStatuses contains list of apiproducts
type KeyAPIProductStatuses []KeyAPIProductStatus

// Approve changes this key's status to approved
func (k *Key) Approved() {

	k.Status = "approved"
}

// Revoke change this key's status to revoked
func (k *Key) Revoke() {

	k.Status = "revoked"
}

// IsApproved returns true in case key's status is approved
func (k *Key) IsApproved() bool {

	return k.Status == "approved"
}

// IsExpired returns true in case key is expired
func (k *Key) IsExpired(now int64) bool {

	if k.ExpiresAt == 0 || k.ExpiresAt == -1 {
		return false
	}
	if now < k.ExpiresAt {
		return false
	}
	return true
}

// SetApproved marks a key's apiproduct as approved
func (p *KeyAPIProductStatus) SetApproved() {

	p.Status = "approved"
}

// IsApproved returns true in case key's apiproduct status is approved
func (p *KeyAPIProductStatus) IsApproved() bool {

	return p.Status == "approved"
}
