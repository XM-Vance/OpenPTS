// Handler 集成测试工具函数。
package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/auth"
)

// setupTestRouter 创建测试用的 Gin Engine，并注册基础路由。
// 测试用例可以继续注册额外的 handler 路由。
// repo 和 svc 依赖需由测试者自行注入；如需跳过数据库，请使用 t.Skip。
//
// 返回值：
//   - router: 已注册基础 middleware 的 Gin Engine
//   - jwtSvc: 用于生成测试 token
//   - cleanup: 清理函数
func setupTestRouter(t *testing.T) (*gin.Engine, *auth.JWTService, func()) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	jwtSvc := auth.NewJWTService("test-secret-at-least-32-characters-xxxx", 1*time.Hour)

	r := gin.New()
	r.Use(gin.Recovery())

	// 基础端点（不需要认证）
	r.GET("/api/v1/ping", Ping)

	return r, jwtSvc, func() {}
}

// loginToken 用指定身份生成一个有效的 JWT token。
func loginToken(t *testing.T, jwtSvc *auth.JWTService, userID uuid.UUID, username string) string {
	t.Helper()
	token, err := jwtSvc.Sign(userID, username, "", false)
	if err != nil {
		t.Fatalf("生成测试 token 失败: %v", err)
	}
	return token
}

// adminToken 生成一个管理员 JWT token（固定 UUID 方便断言）。
func adminToken(t *testing.T, jwtSvc *auth.JWTService) string {
	t.Helper()
	return loginToken(t, jwtSvc,
		uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		"admin")
}

// doJSON 执行一个 HTTP 请求（JSON body），返回响应体字符串。
func doJSON(t *testing.T, r *gin.Engine, method, url, token string, body any) string {
	t.Helper()

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("序列化请求体失败: %v", err)
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w.Body.String()
}

// doRequest 执行一个无 body 的 HTTP 请求，返回响应体字符串。
func doRequest(t *testing.T, r *gin.Engine, method, url, token string) string {
	t.Helper()

	req := httptest.NewRequest(method, url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w.Body.String()
}

// assertStatus 请求（GET 无 body）并断言 HTTP 状态码，返回响应体。
func assertStatus(t *testing.T, r *gin.Engine, method, url, token string, expectedCode int) string {
	t.Helper()
	req := httptest.NewRequest(method, url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != expectedCode {
		t.Errorf("请求 %s %s 期望状态码 %d，得到 %d，响应: %s",
			method, url, expectedCode, w.Code, w.Body.String())
	}
	return w.Body.String()
}

// assertStatusJSON 请求（带 JSON body）并断言 HTTP 状态码，返回响应体。
func assertStatusJSON(t *testing.T, r *gin.Engine, method, url, token string, body any, expectedCode int) string {
	t.Helper()
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("序列化请求体失败: %v", err)
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != expectedCode {
		t.Errorf("请求 %s %s 期望状态码 %d，得到 %d，响应: %s",
			method, url, expectedCode, w.Code, w.Body.String())
	}
	return w.Body.String()
}

// parseJSON 将 JSON 字符串解析到目标结构体。
func parseJSON(t *testing.T, data string, dst any) {
	t.Helper()
	if err := json.Unmarshal([]byte(data), dst); err != nil {
		t.Fatalf("解析 JSON 失败: %v，原始内容: %s", err, data)
	}
}

// skipIfNoDB 跳过测试如果数据库不可用。
func skipIfNoDB(t *testing.T) {
	t.Helper()
	t.Skip("需要真实数据库")
}
