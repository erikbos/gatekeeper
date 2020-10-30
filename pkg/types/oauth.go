package types

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
