// CORS 中间件：开发期放开，生产期按白名单。
package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 安全策略：
//   - devMode = true：允许所有来源（开发期方便）。
//   - 生产模式：从 ALLOWED_ORIGINS 环境变量读取白名单（逗号分隔）。
//     未配置白名单时**仅放行同源请求、拒绝一切跨域**，防 CSRF / CSWSH。
//     与 ws.go 的 CheckOrigin 共用 OriginAllowed（见 origin.go），规则一致。
//
// ⚠️ 实现注意（gin-contrib/cors v1.7.2 的坑，曾导致生产 panic，见部署报告 3.1）：
//   该库在 newCors→Validate 阶段对「!AllowAllOrigins && 无 origin 函数 && AllowOrigins 为空」
//   的配置直接 panic("conflict settings: all origins disabled")。
//   因此「拒绝所有跨域」绝不能用 AllowOrigins=[] 实现；必须提供 AllowOriginWithContextFunc
//   （让 Validate 跳过该校验），在函数内按同源/白名单判定。本实现统一用此函数。
func CORS(devMode bool) gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders: []string{"Content-Length", "X-Request-ID"},
		MaxAge:        12 * time.Hour,
	}
	if devMode {
		cfg.AllowAllOrigins = true
		cfg.AllowCredentials = false
	} else {
		// 生产模式：用 AllowOriginWithContextFunc 判定（避免库对空 AllowOrigins 的
		// panic），复用 OriginAllowed 的同源+白名单规则。
		whitelistNonEmpty := len(loadAllowedOrigins()) > 0
		cfg.AllowOrigins = nil
		cfg.AllowOriginWithContextFunc = func(c *gin.Context, origin string) bool {
			return OriginAllowed(false, origin, c.Request.Host)
		}
		cfg.AllowCredentials = whitelistNonEmpty // 白名单非空才带 credentials
	}
	return cors.New(cfg)
}
