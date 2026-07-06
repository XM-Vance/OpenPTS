package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DemoGate 在生产环境下拦截所有 demo 数据端点。
// 用 strings.Contains 覆盖各类变体后缀：/demo-data、/obs-demo-data 等。
// 仅当 isProd == true 时生效；非生产环境直接放行。
func DemoGate(isProd bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isProd && strings.Contains(c.Request.URL.Path, "demo-data") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "生产环境已禁用 demo 数据端点",
			})
			return
		}
		c.Next()
	}
}
