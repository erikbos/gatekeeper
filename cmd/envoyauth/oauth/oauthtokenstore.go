package oauth

import (
	"errors"
	"time"

	"go.uber.org/zap"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// TokenStore holds our database config
type TokenStore struct {
	db      *db.Database
	metrics *metrics
	logger  *zap.Logger
}

// NewTokenStore creates token store instance
func NewTokenStore(database *db.Database, metrics *metrics,
	logger *zap.Logger) oauth2.TokenStore {

	return &TokenStore{
		db:      database,
		metrics: metrics,
		logger:  logger.With(zap.String("system", "oauthtokenstore")),
	}
}

// Create stores token in database
func (tokenstore *TokenStore) Create(info oauth2.TokenInfo) (err error) {

	tokenstore.logger.Debug("Create", zap.String("token", info.GetAccess()))
	token := types.OAuthAccessToken{
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
	if err := tokenstore.db.OAuth.OAuthAccessTokenCreate(&token); err != nil {
		tokenstore.metrics.IncTokenStoreIssueFailures()
		return err
	}
	tokenstore.metrics.IncTokenStoreIssueSuccesses()
	return nil
}

// GetByAccess gets token by access name
func (tokenstore *TokenStore) GetByAccess(access string) (oauth2.TokenInfo, error) {

	const method string = "access"

	tokenstore.logger.Debug("GetByAccess", zap.String(method, access))
	if access == "" {
		tokenstore.metrics.IncTokenStoreLookupMisses(method)
		return nil, errors.New("Empty token provided")
	}
	token, err := tokenstore.db.OAuth.OAuthAccessTokenGetByAccess(access)
	if err != nil {
		tokenstore.metrics.IncTokenStoreLookupMisses(method)
		return nil, err
	}
	tokenstore.metrics.IncTokenStoreLookupHits(method)
	return toOAuthTokenStore(token)
}

// GetByCode gets token by code name
func (tokenstore *TokenStore) GetByCode(code string) (oauth2.TokenInfo, error) {

	const method string = "code"

	tokenstore.logger.Debug("GetByCode", zap.String(method, code))
	if code == "" {
		tokenstore.metrics.IncTokenStoreLookupMisses(method)
		return nil, nil
	}
	token, err := tokenstore.db.OAuth.OAuthAccessTokenGetByCode(code)
	if err != nil {
		tokenstore.metrics.IncTokenStoreLookupMisses(method)
		return nil, err
	}
	tokenstore.metrics.IncTokenStoreLookupHits(method)
	return toOAuthTokenStore(token)
}

// GetByRefresh gets token by refresh name
func (tokenstore *TokenStore) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {

	const method string = "refresh"

	tokenstore.logger.Debug("GetByRefresh", zap.String(method, refresh))
	if refresh == "" {
		tokenstore.metrics.IncTokenStoreLookupMisses(method)
		return nil, nil
	}
	token, err := tokenstore.db.OAuth.OAuthAccessTokenGetByRefresh(refresh)
	if err != nil {
		tokenstore.metrics.IncTokenStoreLookupMisses(method)
		return nil, err
	}
	tokenstore.metrics.IncTokenStoreLookupHits(method)
	return toOAuthTokenStore(token)
}

func toOAuthTokenStore(token *types.OAuthAccessToken) (oauth2.TokenInfo, error) {

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

	tokenstore.logger.Debug("RemoveByAccess", zap.String("access", access))
	return tokenstore.db.OAuth.OAuthAccessTokenRemoveByAccess(access)
}

// RemoveByCode removes token from database
func (tokenstore *TokenStore) RemoveByCode(code string) (err error) {

	tokenstore.logger.Debug("RemoveByCode", zap.String("code", code))
	return tokenstore.db.OAuth.OAuthAccessTokenRemoveByCode(code)
}

// RemoveByRefresh removes token from database
func (tokenstore *TokenStore) RemoveByRefresh(refresh string) error {

	tokenstore.logger.Debug("RemoveByRefresh", zap.String("refresh", refresh))
	return tokenstore.db.OAuth.OAuthAccessTokenRemoveByRefresh(refresh)
}
