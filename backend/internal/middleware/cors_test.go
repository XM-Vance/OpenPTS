package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// 复现并验证修复：生产模式（devMode=false）且 ALLOWED_ORIGINS 为空时，
// CORS(false) 不应 panic（PR #137 引入的 bug：旧代码 AllowOrigins=[] 触发
// gin-contrib/cors 的 panic("conflict settings: all origins disabled")，
// 导致生产 backend 启动即崩，见部署报告 3.1）。
func TestCORSProdEmptyOriginsDoesNotPanic(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")

	// 旧代码在此行 panic；修复后正常返回 handler。
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CORS(false) 在 ALLOWED_ORIGINS 为空时 panic: %v", r)
		}
	}()
	h := CORS(false)
	if h == nil {
		t.Fatal("CORS(false) 返回 nil")
	}
}

// 生产模式空白名单：跨域请求被库直接拒绝（403 Forbidden，比"放行但不写头"更安全）。
func TestCORSProdEmptyOriginsRejectsCrossOrigin(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")

	r := gin.New()
	r.Use(CORS(false))
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	// 模拟跨域请求（Origin 与 Host 不同源）
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	req.Host = "ptis.local"
	r.ServeHTTP(w, req)

	// gin-contrib/cors 对不在白名单/非同源的 Origin 直接 AbortWithStatus(403)
	if w.Code != http.StatusForbidden {
		t.Fatalf("跨域请求应被拒绝（403），实际 %d", w.Code)
	}
}

// 生产模式空白名单：同源请求（origin == scheme://host）正常放行。
// 注意：库对同源请求（fetch 自带 Origin）不写 Access-Control-Allow-Origin 头——
// 同源本就不需要 CORS，浏览器不读该头。仅验证请求未被 403 拦截。
func TestCORSProdEmptyOriginsAllowsSameOrigin(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")

	r := gin.New()
	r.Use(CORS(false))
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	// 同源：Origin == "http://" + Host（库的精确匹配规则）
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "http://ptis.local")
	req.Host = "ptis.local"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("同源请求应正常处理（200），实际 %d", w.Code)
	}
}

// 生产模式配白名单：白名单内 Origin 放行并带 credentials。
func TestCORSProdWhitelistAllowsListedOrigin(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://app.ptis.local,https://admin.ptis.local")

	r := gin.New()
	r.Use(CORS(false))
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "https://app.ptis.local")
	req.Host = "ptis.local"
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.ptis.local" {
		t.Errorf("白名单内 Origin 应放行，期望 https://app.ptis.local，实际 %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("白名单非空时应带 credentials，实际 %q", got)
	}
}

// 开发模式：放行所有来源。
func TestCORSDevAllowsAll(t *testing.T) {
	r := gin.New()
	r.Use(CORS(true))
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "https://anything.example.com")
	r.ServeHTTP(w, req)

	// dev 模式 AllowAllOrigins=true，响应带 * 或回显 origin
	if got := w.Header().Get("Access-Control-Allow-Origin"); got == "" {
		t.Error("dev 模式应放行所有来源")
	}
}
