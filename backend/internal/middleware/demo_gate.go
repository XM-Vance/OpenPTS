package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DemoGate 在生产环境下拦截所有 /demo-data 端点。
// 仅当 isProd == true 时生效；非生产环境直接放行。
func DemoGate(isProd bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isProd && len(c.Request.URL.Path) >= 10 && c.Request.URL.Path[len(c.Request.URL.Path)-10:] == "/demo-data" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "生产环境已禁用 demo 数据端点",
			})
			return
		}
		c.Next()
	}
}
