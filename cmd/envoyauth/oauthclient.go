package main

import (
	"go.uber.org/zap"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// ClientTokenStore holds our database config
type ClientTokenStore struct {
	db     *db.Database
	cache  *Cache
	logger *zap.Logger
}

// NewOAuthClientTokenStore creates client token store instance
func NewOAuthClientTokenStore(database *db.Database, cache *Cache,
	logger *zap.Logger) oauth2.ClientStore {

	return &ClientTokenStore{
		db:     database,
		cache:  cache,
		logger: logger.With(zap.String("system", "oauthclientstore")),
	}
}

// GetByID retrieves access token based upon tokenid
func (clientstore *ClientTokenStore) GetByID(id string) (oauth2.ClientInfo, error) {

	if id == "" {
		return nil, nil
	}
	clientstore.logger.Debug("GetByID", zap.String("id", id))

	credential, err := clientstore.cache.GetDeveloperAppKey(&id)
	// in case we do not have this apikey in cache let's try to retrieve it from database
	if err != nil {
		credential, err = clientstore.db.Credential.GetByKey(nil, &id)
		if err != nil {
			// FIX ME increase unknown apikey counter (not an error state)
			return nil, err
		}
		// Store retrieved credential in cache, in case of error we proceed as we can
		// statisfy the request as we did retrieve succesful from database
		if err = clientstore.cache.StoreDeveloperAppKey(&id, credential); err != nil {
			clientstore.logger.Debug("GetByID cannot store in cache", zap.String("id", id))
		}
	}

	// TODO increase fetch client id metric, label what=reject (good!)
	return &models.Client{
		ID:     credential.ConsumerKey,
		Secret: credential.ConsumerSecret,
		// Domain: "www.example.com",
		// UserID: "joe",
	}, nil
}
