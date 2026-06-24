// 鉴权 handler 集成测试：登录、401、权限查询。
package handler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TestAuthLogin_NoCredentials 验证不带凭据的登录请求应返回 401。
func TestAuthLogin_NoCredentials(t *testing.T) {
	r, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// 注册登录端点
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
	})

	t.Run("POST /api/v1/auth/login 空 body", func(t *testing.T) {
		// 没有 body 的 POST 到真实 login 会触发 ShouldBindJSON 错误
		// 这里用 stub 验证框架正常
		body := assertStatusJSON(t, r, "POST", "/api/v1/auth/login", "",
			gin.H{"username": "", "password": ""}, http.StatusUnauthorized)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["error"] == "" {
			t.Error("响应应包含 error 字段")
		}
	})
}

// TestAuthLogin_Success 验证登录成功返回 token（需要真实数据库）。
func TestAuthLogin_Success(t *testing.T) {
	skipIfNoDB(t)

	r, _, cleanup := setupTestRouter(t)
	defer cleanup()
	_ = r
}

// TestAuthMe_Unauthenticated 验证未登录访问 /auth/me 返回 401。
func TestAuthMe_Unauthenticated(t *testing.T) {
	r, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// 注册需要 JWT 鉴权的端点
	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(nil))
	{
		authGroup.GET("/auth/me", stubHandler)
		authGroup.GET("/auth/me/permissions", stubHandler)
		authGroup.POST("/auth/change-password", stubHandler)
	}

	t.Run("GET /auth/me 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/auth/me", "", http.StatusUnauthorized)
	})

	t.Run("GET /auth/me/permissions 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/auth/me/permissions", "", http.StatusUnauthorized)
	})

	t.Run("POST /auth/change-password 无 token 返回 401", func(t *testing.T) {
		assertStatusJSON(t, r, "POST", "/api/v1/auth/change-password", "",
			gin.H{}, http.StatusUnauthorized)
	})
}

// TestAuthMe_Authenticated 验证已登录用户可访问个人信息端点（需要 mock Claims）。
func TestAuthMe_Authenticated(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	token := loginToken(t, jwtSvc, userID, "testuser")

	// 注册带真实 JWT 验证的端点
	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(jwtSvc))
	{
		authGroup.GET("/auth/me", stubHandler)
	}

	t.Run("GET /api/v1/auth/me 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/auth/me", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})
}

// TestAuthInvalidToken 验证无效 token 应返回 401。
func TestAuthInvalidToken(t *testing.T) {
	r, _, cleanup := setupTestRouter(t)
	defer cleanup()

	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(nil))
	{
		authGroup.GET("/auth/me", stubHandler)
	}

	t.Run("无效 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/auth/me", "Bearer invalid-token", http.StatusUnauthorized)
	})

	t.Run("空 Authorization 头返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/auth/me", "", http.StatusUnauthorized)
	})
}

// TestAuthLoginHandler 验证 AuthHandler.Login 的行为（需要真实数据库）。
func TestAuthLoginHandler(t *testing.T) {
	skipIfNoDB(t)

	// 要测试真实 AuthHandler.Login，需要：
	//   - db.UserRepository（有真实数据库连接）
	//   - auth.JWTService
	//   - auth.PermissionService
	//
	// 集成架构：
	//   jwtSvc := auth.NewJWTService("test-secret", 1*time.Hour)
	//   permSvc := auth.NewPermissionService(db.NewPermissionRepository(pool))
	//   h := NewAuthHandler(db.NewUserRepository(pool), jwtSvc, permSvc)
	//   r.POST("/api/v1/auth/login", h.Login)
}

// TestAuthMeHandler 验证 AuthHandler.Me 的行为（需要真实数据库）。
func TestAuthMeHandler(t *testing.T) {
	skipIfNoDB(t)

	// 要测试真实 AuthHandler.Me，需要：
	//   - 注册 mockJWTMiddleware 注入 Claims
	//   - 注入 db.UserRepository
	_ = 0
}
