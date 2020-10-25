package cache

import (
	"github.com/coocood/freecache"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// Config contains our start configuration
type Config struct {
	Size        int `yaml:"size"`
	TTL         int `yaml:"ttl"`
	NegativeTTL int `yaml:"negativettl"`
}

// Cache holds our runtime parameters
type Cache struct {
	config    *Config
	db        *db.Database
	freecache *freecache.Cache
	logger    *zap.Logger
	metrics   *metrics
}

// New initializes read through cache for database access
func New(config *Config, d *db.Database, applicationName string, logger *zap.Logger) (*db.Database, error) {

	c := &Cache{
		config:    config,
		db:        d,
		freecache: freecache.NewCache(config.Size),
		logger:    logger.With(zap.String("system", "cache")),
	}
	c.metrics = newMetrics(c)
	c.metrics.registerMetricsWithPrometheus(applicationName)

	c.logger.Info("new",
		zap.Int("size", config.Size),
		zap.Int("ttl", config.TTL),
		zap.Int("negativettl", config.NegativeTTL))

	return &db.Database{
		Listener:     d.Listener,
		Route:        d.Route,
		Cluster:      d.Cluster,
		Developer:    NewDeveloperCache(c, d.Developer),
		DeveloperApp: NewDeveloperAppCache(c, d.DeveloperApp),
		APIProduct:   NewAPIProductCache(c, d.APIProduct),
		Credential:   NewCredentialCache(c, d.Credential),
		OAuth:        NewOAuthCache(c, d.OAuth),
		User:         NewUserCache(c, d.User),
		Role:         NewRoleCache(c, d.Role),
		Readiness:    d.Readiness,
	}, nil
}
