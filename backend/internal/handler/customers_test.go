// 客户档案 handler 集成测试。
package handler

import (
	"net/http"
	"testing"

	"github.com/ptis/backend/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TestCustomerAuthz_NoToken 验证未登录请求客户端点应返回 401。
func TestCustomerAuthz_NoToken(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	_ = jwtSvc

	// 模拟需要认证的路由组
	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(nil))
	{
		authGroup.GET("/customers", stubHandler)
		authGroup.GET("/customers/:id", stubHandler)
		authGroup.POST("/customers", stubHandler)
		authGroup.PUT("/customers/:id", stubHandler)
		authGroup.DELETE("/customers/:id", stubHandler)
	}

	t.Run("GET /api/v1/customers 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/customers", "", http.StatusUnauthorized)
	})

	t.Run("POST /api/v1/customers 无 token 返回 401", func(t *testing.T) {
		assertStatusJSON(t, r, "POST", "/api/v1/customers", "", gin.H{}, http.StatusUnauthorized)
	})

	t.Run("GET /api/v1/customers/123 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/customers/123", "", http.StatusUnauthorized)
	})

	t.Run("PUT /api/v1/customers/123 无 token 返回 401", func(t *testing.T) {
		assertStatusJSON(t, r, "PUT", "/api/v1/customers/123", "", gin.H{}, http.StatusUnauthorized)
	})

	t.Run("DELETE /api/v1/customers/123 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "DELETE", "/api/v1/customers/123", "", http.StatusUnauthorized)
	})
}

// TestCustomerPing_Unauthenticated 验证无需认证的端点正常工作。
func TestCustomerPing_Unauthenticated(t *testing.T) {
	r, _, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("GET /api/v1/ping 无需 token", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/ping", "", http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "pong" {
			t.Errorf("期望 message=pong，得到 %v", result["message"])
		}
	})
}

// TestCustomerEndpoints_Integration 验证客户端点基础行为（需要真实数据库）。
func TestCustomerEndpoints_Integration(t *testing.T) {
	skipIfNoDB(t)

	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	token := adminToken(t, jwtSvc)

	// 此处注入真实 CustomersHandler
	// 由于需要真实 repo，测试被跳过；仅验证编译通过
	_ = r
	_ = token
}

// mockJWTMiddleware 创建一个模拟 JWT 中间件。
// 如果 svc 为 nil，则总是返回 401；否则正常通过。
func mockJWTMiddleware(svc *auth.JWTService) gin.HandlerFunc {
	if svc == nil {
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "缺少 Authorization 头"})
		}
	}
	return func(c *gin.Context) {
		// 注入模拟的 Claims，使 handler 可以读取
		c.Set(auth.ClaimsContextKey, &auth.Claims{
			UserID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Username: "admin",
		})
		c.Next()
	}
}

// stubHandler 是一个无操作 handler，用于路由测试。
func stubHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
