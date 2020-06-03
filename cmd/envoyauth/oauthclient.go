package main

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// ClientTokenStore holds our database config
type ClientTokenStore struct {
	db *db.Database
}

// NewOAuthClientTokenStore creates client token store instance
func NewOAuthClientTokenStore(database *db.Database) oauth2.ClientStore {

	return &ClientTokenStore{
		db: database,
	}
}

// GetByID gets client id token
func (clientstore *ClientTokenStore) GetByID(id string) (oauth2.ClientInfo, error) {

	if id == "" {
		return nil, nil
	}
	log.Infof("OAuthClientTokenStore: GetByID: %s", id)

	// FIXME
	credential, err := clientstore.db.GetAppCredentialByKey(nil, &id)
	if err != nil {
		// FIXME increase fetch client id metric, label what=reject (not an error state)
		return nil, nil
	}

	// FIXME increase fetch client id metric, label what=reject (good!)
	return &models.Client{
		ID:     credential.ConsumerKey,
		Secret: credential.ConsumerSecret,
		// Domain: "www.example.com",
		// UserID: "joe",
	}, nil
}
