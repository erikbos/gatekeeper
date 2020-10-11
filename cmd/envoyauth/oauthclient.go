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
	logger *zap.Logger
}

// NewOAuthClientTokenStore creates client token store instance
func NewOAuthClientTokenStore(database *db.Database, logger *zap.Logger) oauth2.ClientStore {

	return &ClientTokenStore{
		db:     database,
		logger: logger.With(zap.String("system", "oauthclientstore")),
	}
}

// GetByID retrieves access token based upon tokenid
func (clientstore *ClientTokenStore) GetByID(id string) (oauth2.ClientInfo, error) {

	if id == "" {
		return nil, nil
	}
	clientstore.logger.Debug("GetByID", zap.String("id", id))

	credential, err := clientstore.db.Credential.GetByKey(nil, &id)
	if err != nil {
		// FIX ME increase unknown apikey counter (not an error state)
		return nil, err
	}
	// TODO increase fetch client id metric, label accepted (good!)

	return &models.Client{
		ID:     credential.ConsumerKey,
		Secret: credential.ConsumerSecret,
		// Domain: "www.example.com",
		// UserID: "joe",
	}, nil
}
