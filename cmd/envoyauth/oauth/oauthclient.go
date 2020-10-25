package oauth

import (
	"errors"

	"go.uber.org/zap"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// ClientTokenStore is our interface to our credential database

// ClientTokenStore holds our database config
type ClientTokenStore struct {
	db      *db.Database
	metrics *metrics
	logger  *zap.Logger
}

// NewOAuthClientTokenStore creates client token store instance
func NewOAuthClientTokenStore(database *db.Database, metrics *metrics,
	logger *zap.Logger) oauth2.ClientStore {

	return &ClientTokenStore{
		db:      database,
		metrics: metrics,
		logger:  logger.With(zap.String("system", "oauthclientstore")),
	}
}

// GetByID retrieves access token based upon tokenid (which is OAuth consumerkey)
func (clientstore *ClientTokenStore) GetByID(id string) (oauth2.ClientInfo, error) {

	if clientstore == nil || id == "" {
		return nil, errors.New("Cannot handle request")
	}
	clientstore.logger.Debug("GetByID", zap.String("id", id))

	credential, err := clientstore.db.Credential.GetByKey(&id)
	if err != nil {
		clientstore.metrics.IncClientStoreMisses()
		return nil, err
	}
	clientstore.metrics.IncClientStoreHits()
	return &models.Client{
		ID:     credential.ConsumerKey,
		Secret: credential.ConsumerSecret,
	}, nil
}
