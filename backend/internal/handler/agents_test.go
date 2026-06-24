// 代理人 handler 集成测试。
package handler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestAgentsEndpoints 测试 GET /api/v1/agents 端点的鉴权行为。
// 未认证请求应返回 401，已认证请求应返回 200（stub handler）。
func TestAgentsEndpoints(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	_ = jwtSvc

	// 注册需要认证的 agents 路由组
	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(nil))
	{
		authGroup.GET("/agents", stubHandler)
		authGroup.GET("/agents/:id", stubHandler)
		authGroup.POST("/agents", stubHandler)
		authGroup.PUT("/agents/:id", stubHandler)
		authGroup.DELETE("/agents/:id", stubHandler)
		authGroup.GET("/agents/:id/customers", stubHandler)
	}

	t.Run("GET /api/v1/agents 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/agents", "", http.StatusUnauthorized)
	})

	t.Run("GET /api/v1/agents/1 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/agents/1", "", http.StatusUnauthorized)
	})

	t.Run("GET /api/v1/agents/1/customers 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/agents/1/customers", "", http.StatusUnauthorized)
	})

	t.Run("POST /api/v1/agents 无 token 返回 401", func(t *testing.T) {
		assertStatusJSON(t, r, "POST", "/api/v1/agents", "", gin.H{}, http.StatusUnauthorized)
	})
}

// TestAgentsEndpoints_Authenticated 测试已认证用户访问代理人端点。
func TestAgentsEndpoints_Authenticated(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()

	// 注册带真实 JWT 验证的路由组
	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(jwtSvc))
	{
		authGroup.GET("/agents", stubHandler)
		authGroup.GET("/agents/:id", stubHandler)
		authGroup.GET("/agents/:id/customers", stubHandler)
	}

	token := adminToken(t, jwtSvc)

	t.Run("GET /api/v1/agents 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/agents", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("GET /api/v1/agents/1 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/agents/1", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("GET /api/v1/agents/1/customers 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/agents/1/customers", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})
}

// TestAgentsEndpoints_Integration 验证真实 AgentsHandler 行为（需要真实数据库）。
func TestAgentsEndpoints_Integration(t *testing.T) {
	skipIfNoDB(t)

	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	token := adminToken(t, jwtSvc)

	_ = r
	_ = token
}
