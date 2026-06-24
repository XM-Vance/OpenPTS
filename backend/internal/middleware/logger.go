// 结构化日志中间件 + 全局 logger 初始化。
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger 初始化全局 logger。main 启动时调用一次。
func InitLogger(level string) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		l = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(l)
	zerolog.TimeFieldFormat = time.RFC3339
}

// Logger 记录每个 HTTP 请求的方法、路径、状态、延迟。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}
		latency := time.Since(start)
		status := c.Writer.Status()

		evt := log.Info()
		switch {
		case status >= 500:
			evt = log.Error()
		case status >= 400:
			evt = log.Warn()
		}

		rid, _ := c.Get(RequestIDKey)
		evt.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Interface("request_id", rid).
			Msg("HTTP")
	}
}
