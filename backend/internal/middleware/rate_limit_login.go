// 登录专用速率限制：基于滑动窗口的每 IP 限流。
// 默认每 IP 每分钟最多 30 次请求，防止暴力破解密码。
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// loginAttempt 记录某 IP 在窗口内的请求时间戳
type loginAttempt struct {
	mu       sync.Mutex
	times    []time.Time
	maxPerMin int
}

type loginRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*loginAttempt
	maxPerMin int
	window    time.Duration
}

func newLoginRateLimiter(maxPerMin int) *loginRateLimiter {
	return &loginRateLimiter{
		limiters:  make(map[string]*loginAttempt),
		maxPerMin: maxPerMin,
		window:    time.Minute,
	}
}

func (l *loginRateLimiter) get(ip string) *loginAttempt {
	l.mu.RLock()
	a, ok := l.limiters[ip]
	l.mu.RUnlock()
	if ok {
		return a
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if a, ok := l.limiters[ip]; ok {
		return a
	}
	a = &loginAttempt{
		times:     make([]time.Time, 0, l.maxPerMin),
		maxPerMin: l.maxPerMin,
	}
	l.limiters[ip] = a
	return a
}

// allow 检查并清理过期时间戳，返回是否允许请求
func (a *loginAttempt) allow() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Minute)

	// 清理窗口外的记录
	valid := a.times[:0]
	for _, t := range a.times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	a.times = valid

	if len(a.times) >= a.maxPerMin {
		return false
	}
	a.times = append(a.times, now)
	return true
}

// LoginRateLimit 创建登录专用的速率限制中间件。
// maxPerMin: 每个 IP 每分钟允许的最大请求数。
func LoginRateLimit(maxPerMin int) gin.HandlerFunc {
	l := newLoginRateLimiter(maxPerMin)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !l.get(ip).allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "登录请求过于频繁，请一分钟后再试",
			})
			return
		}
		c.Next()
	}
}
