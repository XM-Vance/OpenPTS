// 审计中间件：拦截写操作（POST/PUT/DELETE），经批量写入器异步落库 audit_logs。
package middleware

import (
	"regexp"
	"strings"
	"time"

	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
)

// uuid 形式的 path 参数（用于 resource_id 提取）。
var uuidRe = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

// 路径前缀 → resource 名映射，用于审计列表筛选。
var resourceByPrefix = map[string]string{
	"/api/v1/users":              "users",
	"/api/v1/roles":              "roles",
	"/api/v1/customers":          "customers",
	"/api/v1/retail/contracts":   "retail_contracts",
	"/api/v1/retail/packages":    "retail_packages",
	"/api/v1/load":               "load",
	"/api/v1/price":              "price",
	"/api/v1/settlement":         "settlement",
	"/api/v1/freq":               "freq",
	"/api/v1/storage":            "storage",
	"/api/v1/analytics":          "analytics",
	"/api/v1/scheduler":          "scheduler",
	"/api/v1/auth":               "auth",
}

// Audit 拦截写操作,在请求 goroutine 内同步构造审计记录后非阻塞投递给批量写入器。
// 不再 per-request 起 goroutine(P2-15):高峰下 goroutine 不再无界,且落库改为批量。
func Audit(w *AuditWriter) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		// 只审计写操作。
		if method != "POST" && method != "PUT" && method != "DELETE" {
			c.Next()
			return
		}
		start := time.Now()
		c.Next()
		w.enqueue(buildAuditInput(c, method, time.Since(start)))
	}
}

// buildAuditInput 从已完成的请求上下文同步提取审计字段。
// 同步执行(而非旧版的 go writeAudit)使其在请求 goroutine 生命周期内完成,
// 既避免读已回收的 gin.Context,也让优雅关闭能等齐所有投递。
func buildAuditInput(c *gin.Context, method string, dur time.Duration) db.AuditCreateInput {
	in := db.AuditCreateInput{
		Method:     method,
		Path:       c.Request.URL.Path,
		StatusCode: c.Writer.Status(),
		DurationMs: int(dur / time.Millisecond),
	}

	// 用户信息（JWT claims）。
	if v, ok := c.Get(auth.ClaimsContextKey); ok {
		if claims, ok := v.(*auth.Claims); ok {
			uid := claims.UserID.String()
			in.UserID = &uid
			uname := claims.Username
			in.Username = &uname
		}
	}

	// resource + resource_id 提取。
	in.Resource = resolveResource(c.Request.URL.Path)
	if id := uuidRe.FindString(c.Request.URL.Path); id != "" {
		in.ResourceID = &id
	}

	// IP / UA。
	ip := c.ClientIP()
	if ip != "" {
		in.IP = &ip
	}
	if ua := c.Request.UserAgent(); ua != "" {
		in.UserAgent = &ua
	}

	// 错误信息（如果 handler 调用了 c.Errors）。
	if len(c.Errors) > 0 {
		msg := c.Errors.String()
		if len(msg) > 500 {
			msg = msg[:500]
		}
		in.ErrorMessage = &msg
	}
	return in
}

func resolveResource(path string) *string {
	// 先按更长前缀匹配。
	best := ""
	for prefix := range resourceByPrefix {
		if strings.HasPrefix(path, prefix) && len(prefix) > len(best) {
			best = prefix
		}
	}
	if best == "" {
		return nil
	}
	r := resourceByPrefix[best]
	return &r
}
