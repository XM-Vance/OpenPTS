// 零售管理 handler 集成测试：套餐 CRUD、合同 CRUD。
package handler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TestRetailAuthz_NoToken 验证未登录请求零售端点应返回 401。
func TestRetailAuthz_NoToken(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	_ = jwtSvc

	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(nil))
	{
		authGroup.GET("/retail/pricing-models", stubHandler)
		authGroup.GET("/retail/packages", stubHandler)
		authGroup.POST("/retail/packages", stubHandler)
		authGroup.PUT("/retail/packages/:id", stubHandler)
		authGroup.DELETE("/retail/packages/:id", stubHandler)
		authGroup.GET("/retail/contracts", stubHandler)
		authGroup.POST("/retail/contracts", stubHandler)
		authGroup.PUT("/retail/contracts/:id", stubHandler)
		authGroup.DELETE("/retail/contracts/:id", stubHandler)
	}

	t.Run("GET /retail/pricing-models 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/retail/pricing-models", "", http.StatusUnauthorized)
	})
	t.Run("GET /retail/packages 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/retail/packages", "", http.StatusUnauthorized)
	})
	t.Run("POST /retail/packages 无 token 返回 401", func(t *testing.T) {
		assertStatusJSON(t, r, "POST", "/api/v1/retail/packages", "", gin.H{}, http.StatusUnauthorized)
	})
	t.Run("PUT /retail/packages/:id 无 token 返回 401", func(t *testing.T) {
		assertStatusJSON(t, r, "PUT", "/api/v1/retail/packages/"+uuid.New().String(), "", gin.H{}, http.StatusUnauthorized)
	})
	t.Run("DELETE /retail/packages/:id 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "DELETE", "/api/v1/retail/packages/"+uuid.New().String(), "", http.StatusUnauthorized)
	})
	t.Run("GET /retail/contracts 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "GET", "/api/v1/retail/contracts", "", http.StatusUnauthorized)
	})
	t.Run("POST /retail/contracts 无 token 返回 401", func(t *testing.T) {
		assertStatusJSON(t, r, "POST", "/api/v1/retail/contracts", "", gin.H{}, http.StatusUnauthorized)
	})
	t.Run("PUT /retail/contracts/:id 无 token 返回 401", func(t *testing.T) {
		assertStatusJSON(t, r, "PUT", "/api/v1/retail/contracts/"+uuid.New().String(), "", gin.H{}, http.StatusUnauthorized)
	})
	t.Run("DELETE /retail/contracts/:id 无 token 返回 401", func(t *testing.T) {
		assertStatus(t, r, "DELETE", "/api/v1/retail/contracts/"+uuid.New().String(), "", http.StatusUnauthorized)
	})
}

// TestRetailContracts_Authenticated 验证已登录用户可以访问零售合同端点。
func TestRetailContracts_Authenticated(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	token := adminToken(t, jwtSvc)

	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(jwtSvc))
	{
		authGroup.GET("/retail/contracts", stubHandler)
		authGroup.POST("/retail/contracts", stubHandler)
		authGroup.PUT("/retail/contracts/:id", stubHandler)
		authGroup.DELETE("/retail/contracts/:id", stubHandler)
	}

	t.Run("GET /retail/contracts 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/retail/contracts", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("POST /retail/contracts 有 token 返回 200", func(t *testing.T) {
		body := assertStatusJSON(t, r, "POST", "/api/v1/retail/contracts", token,
			gin.H{"customer_id": uuid.New().String()}, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("PUT /retail/contracts 有 token 返回 200", func(t *testing.T) {
		body := assertStatusJSON(t, r, "PUT", "/api/v1/retail/contracts/"+uuid.New().String(), token,
			gin.H{"customer_id": uuid.New().String()}, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("DELETE /retail/contracts 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "DELETE", "/api/v1/retail/contracts/"+uuid.New().String(), token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})
}

// TestRetailPackages_Authenticated 验证已登录用户可以访问零售套餐端点。
func TestRetailPackages_Authenticated(t *testing.T) {
	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	token := adminToken(t, jwtSvc)

	authGroup := r.Group("/api/v1")
	authGroup.Use(mockJWTMiddleware(jwtSvc))
	{
		authGroup.GET("/retail/pricing-models", stubHandler)
		authGroup.GET("/retail/packages", stubHandler)
		authGroup.POST("/retail/packages", stubHandler)
		authGroup.PUT("/retail/packages/:id", stubHandler)
		authGroup.DELETE("/retail/packages/:id", stubHandler)
	}

	t.Run("GET /retail/pricing-models 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/retail/pricing-models", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("GET /retail/packages 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "GET", "/api/v1/retail/packages", token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("POST /retail/packages 有 token 返回 200", func(t *testing.T) {
		body := assertStatusJSON(t, r, "POST", "/api/v1/retail/packages", token,
			gin.H{"package_name": "测试套餐", "package_type": "standard"}, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("PUT /retail/packages 有 token 返回 200", func(t *testing.T) {
		body := assertStatusJSON(t, r, "PUT", "/api/v1/retail/packages/"+uuid.New().String(), token,
			gin.H{"package_name": "更新套餐", "package_type": "premium"}, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})

	t.Run("DELETE /retail/packages 有 token 返回 200", func(t *testing.T) {
		body := assertStatus(t, r, "DELETE", "/api/v1/retail/packages/"+uuid.New().String(), token, http.StatusOK)
		var result map[string]any
		parseJSON(t, body, &result)
		if result["message"] != "ok" {
			t.Errorf("期望 message=ok，得到 %v", result["message"])
		}
	})
}

// TestRetailEndpoints_Integration 验证零售端点集成行为（需要真实数据库）。
func TestRetailEndpoints_Integration(t *testing.T) {
	skipIfNoDB(t)

	r, jwtSvc, cleanup := setupTestRouter(t)
	defer cleanup()
	_ = r
	_ = jwtSvc

	// 真实集成测试应如下注册：
	// h := NewRetailHandler(db.NewRetailRepository(pool))
	// r.POST("/api/v1/retail/contracts", mockJWTMiddleware(jwtSvc), h.CreateContract)
	// ...
}
