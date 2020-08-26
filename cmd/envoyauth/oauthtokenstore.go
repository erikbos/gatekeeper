package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

// TokenStore holds our database config
type TokenStore struct {
	db    *db.Database
	cache *Cache
}

// NewOAuthTokenStore creates token store instance
func NewOAuthTokenStore(database *db.Database, cache *Cache) oauth2.TokenStore {

	return &TokenStore{
		db:    database,
		cache: cache,
	}
}

// Create stores token in database
func (tokenstore *TokenStore) Create(info oauth2.TokenInfo) (err error) {

	log.Debugf("OAuthTokenStore: Create: %s", info.GetAccess())

	token := shared.OAuthAccessToken{
		// FIXME do we need all fields
		ClientID:         info.GetClientID(),
		UserID:           info.GetUserID(),
		RedirectURI:      info.GetRedirectURI(),
		Scope:            info.GetScope(),
		Code:             info.GetCode(),
		CodeCreatedAt:    shared.TimeMillisecondsToInt64(info.GetCodeCreateAt()),
		CodeExpiresIn:    int64(info.GetCodeExpiresIn().Milliseconds()),
		Access:           info.GetAccess(),
		AccessCreatedAt:  shared.TimeMillisecondsToInt64(info.GetAccessCreateAt()),
		AccessExpiresIn:  int64(info.GetAccessExpiresIn().Milliseconds()),
		Refresh:          info.GetRefresh(),
		RefreshCreatedAt: shared.TimeMillisecondsToInt64(info.GetRefreshCreateAt()),
		RefreshExpiresIn: int64(info.GetRefreshExpiresIn().Milliseconds()),
	}
	return tokenstore.db.OAuth.OAuthAccessTokenCreate(&token)
}

// GetByAccess gets token by access name
func (tokenstore *TokenStore) GetByAccess(access string) (oauth2.TokenInfo, error) {

	// log.Debugf("OAuthTokenStore: GetByAccess: %s", access)
	if access == "" {
		return nil, nil
	}

	token, err := tokenstore.cache.GetAccessToken(&access)
	// in case we do not have this token in cache let's try to retrieve it from database
	if err != nil {
		token, err = tokenstore.db.OAuth.OAuthAccessTokenGetByAccess(access)
		if err != nil {
			// TODO increase unknown oauth access counter (not an error state)
			return nil, err
		}
		// Store retrieved token in cache, in case of error we proceed as we can
		// statisfy the request as we did retrieve succesful from database
		if err = tokenstore.cache.StoreAccessToken(&access, token); err != nil {
			log.Debugf("Could not store OAuth access token '%s' in cache", access)
		}
	}
	return toOAuthTokenStore(token)
}

// GetByCode gets token by code name
func (tokenstore *TokenStore) GetByCode(code string) (oauth2.TokenInfo, error) {

	log.Debugf("OAuthTokenStore: GetByCode: %s", code)
	if code == "" {
		return nil, nil
	}
	token, err := tokenstore.db.OAuth.OAuthAccessTokenGetByCode(code)
	if err != nil {
		return nil, err
	}
	return toOAuthTokenStore(token)
}

// GetByRefresh gets token by refresh name
func (tokenstore *TokenStore) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {

	log.Debugf("OAuthTokenStore: GetByRefresh: %s", refresh)
	if refresh == "" {
		return nil, nil
	}
	token, err := tokenstore.db.OAuth.OAuthAccessTokenGetByRefresh(refresh)
	if err != nil {
		return nil, err
	}
	return toOAuthTokenStore(token)
}

func toOAuthTokenStore(token *shared.OAuthAccessToken) (oauth2.TokenInfo, error) {

	return &models.Token{
		// FIXME do we need all fields
		ClientID:         token.ClientID,
		UserID:           token.UserID,
		RedirectURI:      token.RedirectURI,
		Scope:            token.Scope,
		Code:             token.Code,
		CodeCreateAt:     time.Unix(0, token.CodeCreatedAt*int64(time.Millisecond)),
		CodeExpiresIn:    time.Duration(token.CodeExpiresIn) * time.Millisecond,
		Access:           token.Access,
		AccessCreateAt:   time.Unix(0, token.AccessCreatedAt*int64(time.Millisecond)),
		AccessExpiresIn:  time.Duration(token.AccessExpiresIn) * time.Millisecond,
		Refresh:          token.Refresh,
		RefreshCreateAt:  time.Unix(0, token.RefreshCreatedAt*int64(time.Millisecond)),
		RefreshExpiresIn: time.Duration(token.RefreshExpiresIn) * time.Millisecond,
	}, nil
}

// RemoveByAccess removes token from database
func (tokenstore *TokenStore) RemoveByAccess(access string) error {

	log.Debugf("OAuthTokenStore: RemoveByAccess: %s", access)
	return tokenstore.db.OAuth.OAuthAccessTokenRemoveByAccess(&access)
}

// RemoveByCode removes token from database
func (tokenstore *TokenStore) RemoveByCode(code string) (err error) {

	log.Debugf("OAuthTokenStore: RemoveByCode: %s", code)
	return tokenstore.db.OAuth.OAuthAccessTokenRemoveByCode(&code)
}

// RemoveByRefresh removes token from database
func (tokenstore *TokenStore) RemoveByRefresh(refresh string) error {

	log.Debugf("OAuthTokenStore: RemoveByRefresh: %s", refresh)
	return tokenstore.db.OAuth.OAuthAccessTokenRemoveByRefresh(&refresh)
}
