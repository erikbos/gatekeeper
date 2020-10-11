package main

import (
	"github.com/coocood/freecache"
	"github.com/prometheus/client_golang/prometheus"
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
func newCache2(config *cacheConfig) *Cache {

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
			Namespace: applicationName,
			Name:      "cache_hits_total",
			Help:      "Number of cache hits.",
		}, []string{"table"})
	prometheus.MustRegister(c.cacheHits)

	c.cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "cache_misses_total",
			Help:      "Number of cache misses.",
		}, []string{"table"})
	prometheus.MustRegister(c.cacheMisses)

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "cache_entries",
			Help:      "Number of entries in cache.",
		},
		func() float64 { return (float64(c.freecache.EntryCount())) },
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "cache_hitratio",
			Help:      "Hitratio of cache.",
		},
		func() float64 { return (c.freecache.HitRate()) },
	))

	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "cache_hits",
			Help:      "Number of cache hits.",
		},
		func() float64 { return (float64(c.freecache.HitCount())) },
	))

	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "cache_misses",
			Help:      "Number of cache misses.",
		},
		func() float64 { return (float64(c.freecache.MissCount())) },
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "cache_evictions",
			Help:      "Number of cache entries that where evicted.",
		},
		func() float64 { return (float64(c.freecache.EvacuateCount())) },
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "cache_expired",
			Help:      "Number of cache entries that expired.",
		},
		func() float64 { return (float64(c.freecache.ExpiredCount())) },
	))
}
