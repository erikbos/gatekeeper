package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// OAuthStore the oauth information storage interface
	OAuthStore interface {
		OAuthAccessTokenGetByAccess(accessToken string) (*shared.OAuthAccessToken, error)

		OAuthAccessTokenGetByCode(code string) (*shared.OAuthAccessToken, error)

		OAuthAccessTokenGetByRefresh(refresh string) (*shared.OAuthAccessToken, error)

		OAuthAccessTokenCreate(t *shared.OAuthAccessToken) error

		OAuthAccessTokenRemoveByAccess(accessTokenToDelete *string) error

		OAuthAccessTokenRemoveByCode(codeToDelete *string) error

		OAuthAccessTokenRemoveByRefresh(refreshToDelete *string) error
	}
)
