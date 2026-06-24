// RBAC 中间件：要求请求拥有指定权限码。必须在 JWT() 之后使用。
package middleware

import (
	"net/http"

	"github.com/ptis/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

// RequirePermission 校验当前用户拥有 permCode 权限，否则 403。
func RequirePermission(svc *auth.PermissionService, permCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsAny, ok := c.Get(auth.ClaimsContextKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		claims := claimsAny.(*auth.Claims)

		has, err := svc.Has(c.Request.Context(), claims.UserID, permCode)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "权限检查失败"})
			return
		}
		if !has {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权限: " + permCode})
			return
		}
		c.Next()
	}
}
