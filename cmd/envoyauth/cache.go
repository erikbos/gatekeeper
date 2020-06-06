package main

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/coocood/freecache"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// cacheConfig contains our start configuration
type cacheConfig struct {
	Size        int `yaml:"size"`
	TTL         int `yaml:"ttl"`
	NegativeTTL int `yaml:"negativettl"`
}

// Cache holds our runtime parameters
type Cache struct {
	freecache   *freecache.Cache
	cacheTTL    int
	negativeTTL int
	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec
}

// newCache initializes in memory cache and registers metrics
func newCache(config *cacheConfig) *Cache {

	c := &Cache{
		freecache:   freecache.NewCache(config.Size),
		cacheTTL:    config.TTL,
		negativeTTL: config.NegativeTTL,
	}
	registerCacheMetrics(c)

	return c
}

func getDeveloperCacheKey(id *string) []byte {

	return []byte("dev_" + *id)
}

// StoreDeveloper stores a Developer in cache
func (c *Cache) StoreDeveloper(developerID *string, developer *shared.Developer) error {

	if c.freecache == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(developer); err != nil {
		return err
	}
	return c.freecache.Set(getDeveloperCacheKey(developerID), buf.Bytes(), c.cacheTTL)
}

// GetDeveloper gets a Developer from cache
func (c *Cache) GetDeveloper(developerID *string) (*shared.Developer, error) {

	if c.freecache == nil {
		return nil, nil
	}

	cached, err := c.freecache.Get(getDeveloperCacheKey(developerID))
	if err == nil && cached != nil {
		var developer shared.Developer

		err = gob.NewDecoder(bytes.NewBuffer(cached)).Decode(&developer)
		if err == nil {
			c.cacheHits.WithLabelValues("developer").Inc()
			return &developer, nil
		}
	}
	c.cacheMisses.WithLabelValues("developer").Inc()
	return nil, errors.New("Not found")
}

///

func getDeveloperAppCacheKey(id *string) []byte {

	return []byte("app_" + *id)
}

// StoreDeveloperApp stores a Developer in cache
func (c *Cache) StoreDeveloperApp(AppID *string, app *shared.DeveloperApp) error {

	if c.freecache == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(app); err != nil {
		return err
	}
	return c.freecache.Set(getDeveloperAppCacheKey(AppID), buf.Bytes(), c.cacheTTL)
}

// GetDeveloperApp gets an App from cache
func (c *Cache) GetDeveloperApp(AppID *string) (*shared.DeveloperApp, error) {

	if c.freecache == nil {
		return nil, nil
	}

	cached, err := c.freecache.Get(getDeveloperAppCacheKey(AppID))
	if err == nil && cached != nil {
		var app shared.DeveloperApp

		err = gob.NewDecoder(bytes.NewBuffer(cached)).Decode(&app)
		if err == nil {
			c.cacheHits.WithLabelValues("app").Inc()
			return &app, nil
		}
	}
	c.cacheMisses.WithLabelValues("app").Inc()
	return nil, errors.New("Not found")
}

///

func getDeveloperAppKeyCacheKey(id *string) []byte {

	return []byte("key_" + *id)
}

// StoreDeveloperAppKey stores a DeveloperAppKey in cache
func (c *Cache) StoreDeveloperAppKey(apikey *string, appkey *shared.DeveloperAppKey) error {

	if c.freecache == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(appkey); err != nil {
		return err
	}
	return c.freecache.Set(getDeveloperAppKeyCacheKey(apikey), buf.Bytes(), c.cacheTTL)
}

// GetDeveloperAppKey gets an apikey from cache
func (c *Cache) GetDeveloperAppKey(apikey *string) (*shared.DeveloperAppKey, error) {

	if c.freecache == nil {
		return nil, nil
	}

	cached, err := c.freecache.Get(getDeveloperAppKeyCacheKey(apikey))
	if err == nil && cached != nil {
		var apikey shared.DeveloperAppKey

		err = gob.NewDecoder(bytes.NewBuffer(cached)).Decode(&apikey)
		if err == nil {
			c.cacheHits.WithLabelValues("key").Inc()
			return &apikey, nil
		}
	}
	c.cacheMisses.WithLabelValues("key").Inc()
	return nil, err
}

///

func getAPIProductCacheKey(org, id *string) []byte {

	return []byte("product_" + *org + "_" + *id)
}

// StoreAPIProduct stores an APIProduct in cache
func (c *Cache) StoreAPIProduct(org, productname *string, apiproduct *shared.APIProduct) error {

	if c.freecache == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(apiproduct); err != nil {
		return err
	}
	return c.freecache.Set(getAPIProductCacheKey(org, productname), buf.Bytes(), c.cacheTTL)
}

// GetAPIProduct gets an APIProduct from cache
func (c *Cache) GetAPIProduct(org, productname *string) (*shared.APIProduct, error) {

	if c.freecache == nil {
		return nil, nil
	}

	cached, err := c.freecache.Get(getAPIProductCacheKey(org, productname))
	if err == nil && cached != nil {
		var apiproduct shared.APIProduct

		err = gob.NewDecoder(bytes.NewBuffer(cached)).Decode(&apiproduct)
		if err == nil {
			c.cacheHits.WithLabelValues("apiproduct").Inc()
			return &apiproduct, nil
		}
	}
	c.cacheMisses.WithLabelValues("apiproduct").Inc()
	return nil, err
}

func getAccessTokenCacheKey(name *string) []byte {

	return []byte("at_" + *name)
}

// StoreAccessToken stores an OAuthAccessToken in cache
func (c *Cache) StoreAccessToken(name *string, accessToken *shared.OAuthAccessToken) error {

	if c.freecache == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(accessToken); err != nil {
		return err
	}
	return c.freecache.Set(getAccessTokenCacheKey(name), buf.Bytes(), c.cacheTTL)
}

// GetAccessToken gets an OAuthAccessToken from cache
func (c *Cache) GetAccessToken(name *string) (*shared.OAuthAccessToken, error) {

	if c.freecache == nil {
		return nil, nil
	}

	cached, err := c.freecache.Get(getAccessTokenCacheKey(name))
	if err == nil && cached != nil {
		var accessToken shared.OAuthAccessToken

		err = gob.NewDecoder(bytes.NewBuffer(cached)).Decode(&accessToken)
		if err == nil {
			c.cacheHits.WithLabelValues("accesstoken").Inc()
			return &accessToken, nil
		}
	}
	c.cacheMisses.WithLabelValues("accesstoken").Inc()
	return nil, err
}

func registerCacheMetrics(c *Cache) {

	c.cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "cache_hits_total",
			Help:      "Number of cache hits.",
		}, []string{"table"})
	prometheus.MustRegister(c.cacheHits)

	c.cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "cache_misses_total",
			Help:      "Number of cache misses.",
		}, []string{"table"})
	prometheus.MustRegister(c.cacheMisses)

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: myName,
			Name:      "cache_entries",
			Help:      "Number of entries in cache.",
		},
		func() float64 { return (float64(c.freecache.EntryCount())) },
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: myName,
			Name:      "cache_hitratio",
			Help:      "Hitratio of cache.",
		},
		func() float64 { return (c.freecache.HitRate()) },
	))

	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "cache_hits",
			Help:      "Number of cache hits.",
		},
		func() float64 { return (float64(c.freecache.HitCount())) },
	))

	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "cache_misses",
			Help:      "Number of cache misses.",
		},
		func() float64 { return (float64(c.freecache.MissCount())) },
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: myName,
			Name:      "cache_evictions",
			Help:      "Number of cache entries that where evicted.",
		},
		func() float64 { return (float64(c.freecache.EvacuateCount())) },
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: myName,
			Name:      "cache_expired",
			Help:      "Number of cache entries that expired.",
		},
		func() float64 { return (float64(c.freecache.ExpiredCount())) },
	))
}
