package main

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// ClientTokenStore holds our database config
type ClientTokenStore struct {
	db    *db.Database
	cache *Cache
}

// NewOAuthClientTokenStore creates client token store instance
func NewOAuthClientTokenStore(database *db.Database, cache *Cache) oauth2.ClientStore {

	return &ClientTokenStore{
		db:    database,
		cache: cache,
	}
}

// GetByID retrieves access token based upon tokenid
func (clientstore *ClientTokenStore) GetByID(id string) (oauth2.ClientInfo, error) {

	if id == "" {
		return nil, nil
	}
	log.Infof("OAuthClientTokenStore: GetByID: %s", id)

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
			log.Debugf("Could not store OAuth2 credential '%s' in cache", id)
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
