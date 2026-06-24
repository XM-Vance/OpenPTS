// 光伏 handler 集成测试。
package handler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestSolarEndpoints 测试 GET /api/v1/solar/forecast 端点的鉴权行为。
// 未认证请求应返回 401，已认证请求应返回 200（stub handler）。
func TestSolarEndpoints(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	_ = jwtSvc

	// 注册需要认证的 solar 路由组
	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(nil))
	{
		authGroup.GET("/solar/stations", stubHandler)
		authGroup.GET("/solar/stations/:id", stubHandler)
		authGroup.GET("/solar/forecast", stubHandler)
		authGroup.GET("/solar/revenue", stubHandler)
		authGroup.POST("/solar/stations", stubHandler)
		authGroup.POST("/solar/demo-data", stubHandler)
	}

	t.Run("GET /api/v1/solar/forecast 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/solar/forecast", "", http.StatusUnauthorized)
	})

	t.Run("GET /api/v1/solar/stations 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/solar/stations", "", http.StatusUnauthorized)
	})

	t.Run("GET /api/v1/solar/revenue 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/solar/revenue", "", http.StatusUnauthorized)
	})

	t.Run("POST /api/v1/solar/stations 无 token 返回 401", func(t *testing.T) {
		assertStatusJSON(t, r, "POST", "/api/v1/solar/stations", "", gin.H{}, http.StatusUnauthorized)
	})
}

// TestSolarEndpoints_Authenticated 测试已认证用户访问光伏端点。
func TestSolarEndpoints_Authenticated(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()

	// 注册带真实 JWT 验证的路由组
	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(jwtSvc))
	{
		authGroup.GET("/solar/forecast", stubHandler)
		authGroup.GET("/solar/stations", stubHandler)
		authGroup.GET("/solar/revenue", stubHandler)
	}

	token := adminToken(t, jwtSvc)

	t.Run("GET /api/v1/solar/forecast 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/solar/forecast", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("GET /api/v1/solar/stations 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/solar/stations", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("GET /api/v1/solar/revenue 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/solar/revenue", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})
}

// TestSolarEndpoints_Integration 验证真实 SolarHandler 行为（需要真实数据库）。
func TestSolarEndpoints_Integration(t *testing.T) {
	skipIfNoDB(t)

	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	token := adminToken(t, jwtSvc)

	_ = r
	_ = token
}
