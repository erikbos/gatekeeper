package main

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/coocood/freecache"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

// CacheConfig contains our start configuration
type CacheConfig struct {
	Size        int `yaml:"size"`
	TTL         int `yaml:"ttl"`
	NegativeTTL int `yaml:"negativettl"`
}

// Cache holds our runtime parameters
type Cache struct {
	db                    *db.Database
	freecache             *freecache.Cache
	cacheTTL              int
	negativeTTL           int
	cacheHits             *prometheus.CounterVec
	cacheMisses           *prometheus.CounterVec
	cacheLatencyHistogram prometheus.Summary
}

// newCache initializes in memory cache and registers metrics
func newCache(config *CacheConfig) *Cache {

	c := &Cache{
		freecache:   freecache.NewCache(config.Size),
		cacheTTL:    config.TTL,
		negativeTTL: config.NegativeTTL,
	}
	registerCacheMetrics(c)

	return c
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

	c.cacheLatencyHistogram = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: myName,
			Name:      "cache_latency",
			Help:      "Latency of cache (cached & non-cached) in seconds.",
			Objectives: map[float64]float64{
				0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001,
			},
		})
	prometheus.MustRegister(c.cacheLatencyHistogram)
}

func getDeveloperCacheKey(id *string) []byte {

	return []byte("dev_" + *id)
}

// StoreDeveloper stores a Developer in cache
func (c *Cache) StoreDeveloper(developer *shared.Developer) error {

	if c.freecache == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(developer); err != nil {
		return err
	}
	return c.freecache.Set(getDeveloperCacheKey(&developer.DeveloperID), buf.Bytes(), c.cacheTTL)
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
func (c *Cache) StoreDeveloperApp(app *shared.DeveloperApp) error {

	if c.freecache == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(app); err != nil {
		return err
	}
	return c.freecache.Set(getDeveloperAppCacheKey(&app.AppID), buf.Bytes(), c.cacheTTL)
}

// GetDeveloperApp gets an App from cache
func (c *Cache) GetDeveloperApp(AppID *string) (*shared.DeveloperApp, error) {

	if c.freecache == nil {
		return nil, nil
	}

	cached, err := c.freecache.Get(getDeveloperCacheKey(AppID))
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
func (c *Cache) StoreDeveloperAppKey(appkey *shared.DeveloperAppKey) error {

	if c.freecache == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(appkey); err != nil {
		return err
	}
	return c.freecache.Set(getDeveloperAppKeyCacheKey(&appkey.ConsumerKey), buf.Bytes(), c.cacheTTL)
}

// GetDeveloperAppKey gets an Apikey from cache
func (c *Cache) GetDeveloperAppKey(key *string) (*shared.DeveloperAppKey, error) {

	if c.freecache == nil {
		return nil, nil
	}

	cached, err := c.freecache.Get(getDeveloperAppKeyCacheKey(key))
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
