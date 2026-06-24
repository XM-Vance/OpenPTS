// 细粒度限流：按「key（IP 或 username）+ 路径」组合限流。
// 用于敏感端点（login、upload）单独低配额，防止暴力或滥用。
package middleware

import (
	"net/http"
	"sync"

	"github.com/ptis/backend/internal/auth"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// StrictKeyFunc 决定限流键：默认按 IP，登录用户优先用 username。
func defaultKey(c *gin.Context) string {
	if v, ok := c.Get(auth.ClaimsContextKey); ok {
		if cl, ok := v.(*auth.Claims); ok && cl.Username != "" {
			return "u:" + cl.Username
		}
	}
	return "ip:" + c.ClientIP()
}

type keyedLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

func newKeyedLimiter(rps float64, burst int) *keyedLimiter {
	return &keyedLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        rate.Limit(rps),
		burst:    burst,
	}
}

func (l *keyedLimiter) get(key string) *rate.Limiter {
	l.mu.RLock()
	lim, ok := l.limiters[key]
	l.mu.RUnlock()
	if ok {
		return lim
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.limiters[key]; ok {
		return lim
	}
	lim = rate.NewLimiter(l.r, l.burst)
	l.limiters[key] = lim
	return lim
}

// StrictRateLimit 严格限流：默认按 user/IP 限流，rps + burst 决定速率。
// 适用于：登录端点（防爆破）、上传端点（防滥用）、生成演示数据（避免误触）等。
func StrictRateLimit(rps float64, burst int) gin.HandlerFunc {
	l := newKeyedLimiter(rps, burst)
	return func(c *gin.Context) {
		key := defaultKey(c) + ":" + c.FullPath()
		if !l.get(key).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "操作过于频繁，请稍后再试",
				"path":  c.FullPath(),
			})
			return
		}
		c.Next()
	}
}
