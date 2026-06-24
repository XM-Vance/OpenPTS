// 进程内 TTL 缓存：sync.Map + lazy expiration + Prom 指标。
// 适合中低吞吐的读多写少场景（如仪表盘 KPI、字典数据）。
package cache

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type entry struct {
	value     any
	expiresAt time.Time
}

type TTLCache struct {
	store sync.Map
	hits  prometheus.Counter
	miss  prometheus.Counter
}

var (
	cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "ptis_cache_hits_total", Help: "缓存命中次数"},
		[]string{"name"},
	)
	cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "ptis_cache_misses_total", Help: "缓存未命中次数"},
		[]string{"name"},
	)
	_ = func() bool {
		prometheus.MustRegister(cacheHits, cacheMisses)
		return true
	}()
)

// New 创建一个命名缓存实例；name 用作 Prometheus label。
func New(name string) *TTLCache {
	return &TTLCache{
		hits: cacheHits.WithLabelValues(name),
		miss: cacheMisses.WithLabelValues(name),
	}
}

// Get 取值。过期视为 miss 并自动清除。
func (c *TTLCache) Get(key string) (any, bool) {
	v, ok := c.store.Load(key)
	if !ok {
		c.miss.Inc()
		return nil, false
	}
	e := v.(*entry)
	if time.Now().After(e.expiresAt) {
		c.store.Delete(key)
		c.miss.Inc()
		return nil, false
	}
	c.hits.Inc()
	return e.value, true
}

// Set 写入值，并附 TTL。
func (c *TTLCache) Set(key string, value any, ttl time.Duration) {
	c.store.Store(key, &entry{value: value, expiresAt: time.Now().Add(ttl)})
}

// Invalidate 主动失效（写操作后清缓存用）。
func (c *TTLCache) Invalidate(key string) {
	c.store.Delete(key)
}

// GetOrLoad 经典模式：命中返回，未命中调用 loader 并缓存。
// 注意：并发未命中会让 loader 多次执行；如需严格 single-flight，需另封装。
func GetOrLoad[T any](c *TTLCache, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	if v, ok := c.Get(key); ok {
		return v.(T), nil
	}
	val, err := loader()
	if err != nil {
		var zero T
		return zero, err
	}
	c.Set(key, val, ttl)
	return val, nil
}
