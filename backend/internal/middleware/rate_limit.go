// 简单的 IP 维度令牌桶限流。开发期足够，阶段 3 可换为 Redis 分布式版本。
package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

func newIPLimiter(rps float64, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        rate.Limit(rps),
		burst:    burst,
	}
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.RLock()
	lim, ok := l.limiters[ip]
	l.mu.RUnlock()
	if ok {
		return lim
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.limiters[ip]; ok {
		return lim
	}
	lim = rate.NewLimiter(l.r, l.burst)
	l.limiters[ip] = lim
	return lim
}

// RateLimit 每秒 rps 个令牌，burst 突发上限。
// 当前不做 IP 过期清理；阶段 3 升级为 LRU/TTL 或 Redis 实现。
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	l := newIPLimiter(rps, burst)
	return func(c *gin.Context) {
		if !l.get(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}
