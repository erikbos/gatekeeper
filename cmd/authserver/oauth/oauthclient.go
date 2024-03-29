package oauth

import (
	"context"
	"errors"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/authserver/metrics"
	"github.com/erikbos/gatekeeper/pkg/db"
)

// ClientTokenStore is our interface to our OAuth token database.
type ClientTokenStore struct {
	db      *db.Database
	metrics *metrics.Metrics
	logger  *zap.Logger
}

// NewClientTokenStore creates client token store instance
func NewClientTokenStore(database *db.Database, metrics *metrics.Metrics,
	logger *zap.Logger) oauth2.ClientStore {

	return &ClientTokenStore{
		db:      database,
		metrics: metrics,
		logger:  logger.With(zap.String("system", "oauthclientstore")),
	}
}

// GetByID retrieves access token based upon tokenid (which is OAuth consumerkey)
func (clientstore *ClientTokenStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {

	if clientstore == nil || id == "" {
		return nil, errors.New("cannot handle nil GetByID request")
	}
	clientstore.logger.Debug("GetByID", zap.String("id", id))

	key, err := clientstore.db.Key.GetByKey(nil, &id)
	if err != nil {
		clientstore.metrics.IncOAuthClientStoreMisses()
		return nil, err
	}
	clientstore.metrics.IncOAuthClientStoreHits()
	return &models.Client{
		ID:     key.ConsumerKey,
		Secret: key.ConsumerSecret,
	}, nil
}
