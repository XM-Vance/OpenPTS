// Panic 恢复中间件：避免单个请求崩溃整个进程。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		rid, _ := c.Get(RequestIDKey)
		log.Error().
			Interface("error", recovered).
			Str("path", c.Request.URL.Path).
			Interface("request_id", rid).
			Msg("捕获到 panic")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "服务器内部错误",
		})
	})
}
