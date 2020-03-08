package db

import (
	"bytes"
	"encoding/gob"

	"github.com/coocood/freecache"
	"github.com/erikbos/apiauth/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// FIXME:
// - duplicate code: can we use golang interface?
// - cleaner way to wrap the real db look functions?
// - uses one freecache instance to store developer, app, appcredentials

// Cache holds all details of our cached database records and cache performance counters
type Cache struct {
	cache                  *freecache.Cache
	cacheTTL               int
	negativeTTL            int
	dbCacheHitsCounter     *prometheus.CounterVec
	dbCacheMissesCounter   *prometheus.CounterVec
	dbCacheLookupHistogram prometheus.Summary
}

// CacheInit initializes in memory cache and registers metrics
//
func CacheInit(size, cachettl, negativettl int) *Cache {
	c := Cache{}

	c.cache = freecache.NewCache(size)
	c.cacheTTL = cachettl
	c.negativeTTL = negativettl

	c.dbCacheHitsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "apiauth_database_cache_hits_total",
			Help: "Number of database cache hits.",
		}, []string{"table"})
	prometheus.MustRegister(c.dbCacheHitsCounter)

	c.dbCacheMissesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "apiauth_database_cache_misses_total",
			Help: "Number of database cache misses.",
		}, []string{"table"})
	prometheus.MustRegister(c.dbCacheMissesCounter)

	prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "apiauth_database_cache_entries",
			Help: "Number of entries in database cache.",
		},
		func() float64 { return (float64(c.cache.EntryCount())) },
	))

	prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "apiauth_database_cache_hitratio",
			Help: "Hitratio of database cache.",
		},
		func() float64 { return (c.cache.HitRate()) },
	))

	prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "apiauth_database_cache_hits",
			Help: "Number of database cache hits.",
		},
		func() float64 { return (float64(c.cache.HitCount())) },
	))

	prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "apiauth_database_cache_misses",
			Help: "Number of database cache misses.",
		},
		func() float64 { return (float64(c.cache.MissCount())) },
	))

	prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "apiauth_database_cache_evictions",
			Help: "Number of database cache entries that where evicted.",
		},
		func() float64 { return (float64(c.cache.EvacuateCount())) },
	))

	prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "apiauth_database_cache_expired",
			Help: "Number of database cache entries that expired.",
		},
		func() float64 { return (float64(c.cache.ExpiredCount())) },
	))

	c.dbCacheLookupHistogram = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "apiauth_database_cache_latency",
			Help:       "Database retrieval latency (cached & non-cached) in seconds.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		})
	prometheus.MustRegister(c.dbCacheLookupHistogram)

	return &c
}

//GetAppCredentialCached retrieves entry from database (or cache if entry present)
//
func (c *Cache) GetAppCredentialCached(d *Database, organization, key string) (types.AppCredential, error) {
	var appcredential types.AppCredential
	var err error

	timer := prometheus.NewTimer(c.dbCacheLookupHistogram)
	defer timer.ObserveDuration()

	// if we have a cached entry return that one
	cached, err := c.cache.Get([]byte(key))
	if err == nil && cached != nil {
		buf := bytes.NewBuffer(cached)
		decoder := gob.NewDecoder(buf)
		err = decoder.Decode(&appcredential)
		if err == nil {
			c.dbCacheHitsCounter.WithLabelValues("app_credentials").Inc()
			return appcredential, nil
		}
		log.Infof("GetAppCredentialCached, could not decode cache entry of %s", key)
	}

	// No previous cache entry, let's fetch it from database
	appcredential, err = d.GetAppCredentialByKey(organization, key)
	if err != nil {
		c.dbCacheMissesCounter.WithLabelValues("app_credentials").Inc()
		return appcredential, nil
	}

	// Store entry in cache
	var buf bytes.Buffer
	encode := gob.NewEncoder(&buf)
	err = encode.Encode(appcredential)
	if err == nil {
		c.cache.Set([]byte(key), buf.Bytes(), c.cacheTTL)
	}
	c.dbCacheMissesCounter.WithLabelValues("app_credentials").Inc()
	return appcredential, nil
}

//GetAPIProductCached retrieves entry from database (or cache if entry present)
//
func (c *Cache) GetAPIProductCached(d *Database, apiproductname string) (types.APIProduct, error) {
	var apiproduct types.APIProduct
	var err error

	timer := prometheus.NewTimer(c.dbCacheLookupHistogram)
	defer timer.ObserveDuration()

	// if we have a cached entry return that one
	cached, err := c.cache.Get([]byte(apiproductname))
	if err == nil && cached != nil {
		buf := bytes.NewBuffer(cached)
		decoder := gob.NewDecoder(buf)
		err = decoder.Decode(&apiproduct)
		if err == nil {
			c.dbCacheHitsCounter.WithLabelValues("api_products").Inc()
			return apiproduct, nil
		}
		log.Info("GetAPIProductCached, could not decode cache entry of %s", apiproductname)
	}

	// No cache entry, let's fetch it from database
	apiproduct, err = d.GetAPIProductByName(apiproductname)
	if err != nil {
		c.dbCacheMissesCounter.WithLabelValues("api_products").Inc()
		return apiproduct, nil
	}

	// Store entry in cache
	var buf bytes.Buffer
	encode := gob.NewEncoder(&buf)
	err = encode.Encode(apiproduct)
	if err == nil {
		c.cache.Set([]byte(apiproductname), buf.Bytes(), c.cacheTTL)
	}
	c.dbCacheMissesCounter.WithLabelValues("api_products").Inc()
	return apiproduct, nil
}

//GetDeveloperAppCached retrieves entry from database (or cache if entry present)
//
func (c *Cache) GetDeveloperAppCached(d *Database, developerAppID string) (types.DeveloperApp, error) {
	var developerapp types.DeveloperApp
	var err error

	timer := prometheus.NewTimer(c.dbCacheLookupHistogram)
	defer timer.ObserveDuration()

	// If we have a cached entry return that one
	cached, err := c.cache.Get([]byte(developerAppID))
	if err == nil && cached != nil {
		buf := bytes.NewBuffer(cached)
		decoder := gob.NewDecoder(buf)
		err = decoder.Decode(&developerapp)
		if err == nil {
			c.dbCacheHitsCounter.WithLabelValues("apps").Inc()
			return developerapp, nil
		}
		log.Infof("GetDeveloperAppCached, could not decode cache entry of %s", developerAppID)
	}

	developerapp, err = d.GetDeveloperAppByID("", developerAppID)
	if err != nil {
		c.dbCacheMissesCounter.WithLabelValues("apps").Inc()
		return developerapp, nil
	}

	// Store entry in cache
	var buf bytes.Buffer
	encode := gob.NewEncoder(&buf)
	err = encode.Encode(developerapp)
	if err == nil {
		c.cache.Set([]byte(developerAppID), buf.Bytes(), c.cacheTTL)
	}
	c.dbCacheMissesCounter.WithLabelValues("apps").Inc()
	return developerapp, nil
}
