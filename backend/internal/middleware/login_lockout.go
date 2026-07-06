// 登录失败次数锁定：针对 IP + 用户名组合。
//
// 与 rate_limit_login.go 的区别：
//   - rate_limit_login：按 IP 的请求速率限流（防刷屏），不区分成功/失败。
//   - 本锁定：按 IP+username 的失败次数锁定（防密码爆破）。
//
// 规则（与 LoginRateLimit 串联挂在 /auth/login）：
//   - 连续失败累计 5 次 → 锁定该 ip+username 组合 15 分钟，期间该组合登录直接 423。
//   - 登录成功（非 401 响应）→ 清零该组合计数。
//   - 计数存于进程内 TTLCache（懒过期），无需 Redis；TTL 即锁定时长。
//
// username 提取：中间件在 c.Next 前预读 body（再塞回，供后续 handler 读取）。
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/cache"
)

const (
	loginLockoutThreshold = 5              // 连续失败 5 次锁定
	loginLockoutTTL       = 15 * time.Minute // 锁定 15 分钟
	loginCounterTTL       = 15 * time.Minute // 计数器窗口（与锁定窗口对齐）
)

// loginLockout 维护失败计数与锁定状态。
// 复用进程内 TTLCache，避免引入 Redis 依赖。
type loginLockout struct {
	mu      sync.Mutex                 // 保护 counter map 的整数自增（TTLCache 存 any）
	counts  *cache.TTLCache            // key -> int 失败次数
	locks   *cache.TTLCache            // key -> struct{} 锁定标记
}

var defaultLoginLockout = func() *loginLockout {
	return &loginLockout{
		counts: cache.New("login_lockout_counts"),
		locks:  cache.New("login_lockout_locked"),
	}
}()

// newLoginLockoutForTest 构造独立实例供单元测试用（不污染默认单例）。
func newLoginLockoutForTest() *loginLockout {
	return &loginLockout{
		counts: cache.New("login_lockout_counts_test"),
		locks:  cache.New("login_lockout_locked_test"),
	}
}

// IsLocked 返回该组合是否被锁定。
func (l *loginLockout) IsLocked(key string) bool {
	_, ok := l.locks.Get(key)
	return ok
}

// RecordFailure 记录一次失败；达到阈值则置锁定。
func (l *loginLockout) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var n int
	if v, ok := l.counts.Get(key); ok {
		n, _ = v.(int)
	}
	n++
	if n >= loginLockoutThreshold {
		l.locks.Set(key, struct{}{}, loginLockoutTTL)
		l.counts.Invalidate(key)
	} else {
		l.counts.Set(key, n, loginCounterTTL)
	}
}

// RecordSuccess 成功登录清零计数。
func (l *loginLockout) RecordSuccess(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.counts.Invalidate(key)
}

// loginLockoutKey 组合 IP 与请求体 username（小写归一）。
// username 为空（body 非 JSON / 无 username 字段）时回退为纯 IP，避免绕过。
func loginLockoutKey(c *gin.Context) string {
	ip := c.ClientIP()
	username := c.GetString("_login_username")
	if username == "" {
		return ip
	}
	return ip + ":" + username
}

// LoginLockout 登录失败锁定中间件。
// 前置：预读 body 解析 username（再塞回），并拒绝已锁定的组合；
// 后置：按响应状态记录成功/失败。
func LoginLockout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 预读 body 解析 username，再塞回供 handler 读取
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			_ = c.Request.Body.Close()
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			var probe struct {
				Username string `json:"username"`
			}
			if len(bodyBytes) > 0 {
				_ = json.Unmarshal(bodyBytes, &probe)
			}
			c.Set("_login_username", probe.Username)
		}

		key := loginLockoutKey(c)
		if defaultLoginLockout.IsLocked(key) {
			c.AbortWithStatusJSON(http.StatusLocked, gin.H{
				"error": "登录失败次数过多，请 15 分钟后再试",
			})
			return
		}

		c.Next()

		// 响应后判定：401 = 失败计数，其他（200 登录成功 / 400 参数错误等）= 成功清零
		if c.Writer.Status() == http.StatusUnauthorized {
			defaultLoginLockout.RecordFailure(key)
		} else {
			defaultLoginLockout.RecordSuccess(key)
		}
	}
}
