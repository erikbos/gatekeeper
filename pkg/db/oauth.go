package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// OAuthStore the oauth information storage interface
	OAuthStore interface {
		// OAuthAccessTokenGetByAccess retrieves an access token
		OAuthAccessTokenGetByAccess(accessToken string) (*shared.OAuthAccessToken, error)

		// OAuthAccessTokenGetByCode retrieves token by code
		OAuthAccessTokenGetByCode(code string) (*shared.OAuthAccessToken, error)

		// OAuthAccessTokenGetByRefresh retrieves token by refreshcode
		OAuthAccessTokenGetByRefresh(refresh string) (*shared.OAuthAccessToken, error)

		// OAuthAccessTokenCreate creates an access token
		OAuthAccessTokenCreate(t *shared.OAuthAccessToken) error

		// OAuthAccessTokenRemoveByAccess deletes an access token
		OAuthAccessTokenRemoveByAccess(accessTokenToDelete *string) error

		// OAuthAccessTokenRemoveByCode deletes an access token
		OAuthAccessTokenRemoveByCode(codeToDelete *string) error

		// OAuthAccessTokenRemoveByRefresh deletes an access token
		OAuthAccessTokenRemoveByRefresh(refreshToDelete *string) error
	}
)
