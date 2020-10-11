package cache

import (
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// OAuthCache holds our database config
type OAuthCache struct {
	oauth db.OAuth
	cache *Cache
}

// NewOAuthCache creates user instance
func NewOAuthCache(cache *Cache, oauth db.OAuth) *OAuthCache {
	return &OAuthCache{
		oauth: oauth,
		cache: cache,
	}
}

// OAuthAccessTokenGetByAccess retrieves an access token
func (s *OAuthCache) OAuthAccessTokenGetByAccess(accessToken string) (*types.OAuthAccessToken, error) {

	getTokenByAccess := func() (interface{}, types.Error) {
		token, err := s.oauth.OAuthAccessTokenGetByAccess(accessToken)
		if err != nil {
			return token, types.NewDatabaseError(err)
		}
		return token, nil
	}
	var oauthToken types.OAuthAccessToken
	if err := s.cache.fetchEntry(accessToken, &oauthToken, getTokenByAccess); err != nil {
		return nil, err
	}
	return &oauthToken, nil
}

// OAuthAccessTokenGetByCode retrieves token by code
func (s *OAuthCache) OAuthAccessTokenGetByCode(code string) (*types.OAuthAccessToken, error) {

	getTokenByCode := func() (interface{}, types.Error) {
		token, err := s.oauth.OAuthAccessTokenGetByCode(code)
		if err != nil {
			return token, types.NewDatabaseError(err)
		}
		return token, nil
	}
	var oauthToken types.OAuthAccessToken
	if err := s.cache.fetchEntry(code, &oauthToken, getTokenByCode); err != nil {
		return nil, err
	}
	return &oauthToken, nil
}

// OAuthAccessTokenGetByRefresh retrieves token by refreshcode
func (s *OAuthCache) OAuthAccessTokenGetByRefresh(refresh string) (*types.OAuthAccessToken, error) {

	getTokenByRefresh := func() (interface{}, types.Error) {
		token, err := s.oauth.OAuthAccessTokenGetByRefresh(refresh)
		if err != nil {
			return token, types.NewDatabaseError(err)
		}
		return token, nil
	}
	var oauthToken types.OAuthAccessToken
	if err := s.cache.fetchEntry(refresh, &oauthToken, getTokenByRefresh); err != nil {
		return nil, err
	}
	return &oauthToken, nil
}

// OAuthAccessTokenCreate UPSERTs an organization in database
func (s *OAuthCache) OAuthAccessTokenCreate(t *types.OAuthAccessToken) error {

	return s.oauth.OAuthAccessTokenCreate(t)
}

// OAuthAccessTokenRemoveByAccess deletes an access token
func (s *OAuthCache) OAuthAccessTokenRemoveByAccess(accessTokenToDelete *string) error {

	s.cache.deleteEntry(*accessTokenToDelete, types.OAuthAccessToken{})
	return s.oauth.OAuthAccessTokenRemoveByAccess(accessTokenToDelete)
}

// OAuthAccessTokenRemoveByCode deletes an access token
func (s *OAuthCache) OAuthAccessTokenRemoveByCode(codeToDelete *string) error {

	s.cache.deleteEntry(*codeToDelete, types.OAuthAccessToken{})
	return s.oauth.OAuthAccessTokenRemoveByAccess(codeToDelete)
}

// OAuthAccessTokenRemoveByRefresh deletes an access token
func (s *OAuthCache) OAuthAccessTokenRemoveByRefresh(refreshToDelete *string) error {

	s.cache.deleteEntry(*refreshToDelete, types.OAuthAccessToken{})
	return s.oauth.OAuthAccessTokenRemoveByAccess(refreshToDelete)
}
