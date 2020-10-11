package cache

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	cache *Cache

	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec
}

func newMetrics(cache *Cache) *metrics {

	return &metrics{cache: cache}
}

func (m *metrics) registerMetricsWithPrometheus(applicationName string) {

	m.cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "cache_hits_total",
			Help:      "Number of cache hits.",
		}, []string{"entity"})
	prometheus.MustRegister(m.cacheHits)

	m.cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "cache_misses_total",
			Help:      "Number of cache misses.",
		}, []string{"entity"})
	prometheus.MustRegister(m.cacheMisses)

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "cache_entries",
			Help:      "Number of entries in cache.",
		},
		func() float64 {
			return float64(m.cache.freecache.EntryCount())
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "cache_hitratio",
			Help:      "Hitratio of cache.",
		},
		func() float64 {
			return m.cache.freecache.HitRate()
		},
	))

	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "cache_hits",
			Help:      "Number of cache hits.",
		},
		func() float64 {
			return float64(m.cache.freecache.HitCount())
		},
	))

	prometheus.MustRegister(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "cache_misses",
			Help:      "Number of cache misses.",
		},
		func() float64 {
			return float64(m.cache.freecache.MissCount())
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "cache_evictions",
			Help:      "Number of cache entries that where evicted.",
		},
		func() float64 {
			return float64(m.cache.freecache.EvacuateCount())
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: applicationName,
			Name:      "cache_expired",
			Help:      "Number of cache entries that expired.",
		},
		func() float64 {
			return float64(m.cache.freecache.ExpiredCount())
		},
	))
}

// EntityCacheHit
func (m *metrics) EntityCacheHit(entityType string) {

	m.cacheHits.WithLabelValues(entityType).Inc()
}

// EntityCacheMiss
func (m *metrics) EntityCacheMiss(entityType string) {

	m.cacheMisses.WithLabelValues(entityType).Inc()
}
