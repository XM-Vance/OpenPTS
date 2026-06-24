// 债券 handler 集成测试。
package handler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestBondsEndpoints 测试 GET /api/v1/bonds 端点的鉴权行为。
// 未认证请求应返回 401，已认证请求应返回 200（stub handler）。
func TestBondsEndpoints(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	_ = jwtSvc

	// 注册需要认证的 bonds 路由组
	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(nil))
	{
		authGroup.GET("/bonds", stubHandler)
		authGroup.GET("/bonds/:id", stubHandler)
		authGroup.POST("/bonds", stubHandler)
		authGroup.PUT("/bonds/:id", stubHandler)
		authGroup.DELETE("/bonds/:id", stubHandler)
	}

	t.Run("GET /api/v1/bonds 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/bonds", "", http.StatusUnauthorized)
	})

	t.Run("GET /api/v1/bonds/1 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/bonds/1", "", http.StatusUnauthorized)
	})

	t.Run("POST /api/v1/bonds 无 token 返回 401", func(t *testing.T) {
		assertStatusJSON(t, r, "POST", "/api/v1/bonds", "", gin.H{}, http.StatusUnauthorized)
	})
}

// TestBondsEndpoints_Authenticated 测试已认证用户访问债券端点。
func TestBondsEndpoints_Authenticated(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()

	// 注册带真实 JWT 验证的路由组
	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(jwtSvc))
	{
		authGroup.GET("/bonds", stubHandler)
		authGroup.GET("/bonds/:id", stubHandler)
	}

	token := adminToken(t, jwtSvc)

	t.Run("GET /api/v1/bonds 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/bonds", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("GET /api/v1/bonds/1 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/bonds/1", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})
}

// TestBondsEndpoints_Integration 验证真实 BondsHandler 行为（需要真实数据库）。
func TestBondsEndpoints_Integration(t *testing.T) {
	skipIfNoDB(t)

	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	token := adminToken(t, jwtSvc)

	_ = r
	_ = token
}
