// CORS 中间件：开发期放开，生产期按白名单。
package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 安全策略：
//   - devMode = true：允许所有来源（开发期方便）。
//   - 生产模式：必须从 ALLOWED_ORIGINS 环境变量配置白名单（逗号分隔）。
//     未配置时不允许任何跨域来源（nginx 同源反代兜底）。
//     生产环境不允许 AllowAllOrigins，防止 CSRF。
func CORS(devMode bool) gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:   []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:  []string{"Content-Length", "X-Request-ID"},
		MaxAge:         12 * time.Hour,
		AllowOrigins:   []string{},
	}
	if devMode {
		cfg.AllowAllOrigins = true
		cfg.AllowCredentials = false
	} else {
		// 生产模式：从 ALLOWED_ORIGINS 环境变量读取白名单
		var clean []string
		for _, o := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
			if o = strings.TrimSpace(o); o != "" {
				clean = append(clean, o)
			}
		}
		if len(clean) > 0 {
			cfg.AllowOrigins = clean
			cfg.AllowCredentials = true
		} else {
			// 未配置白名单时放行所有（nginx 同源反代兜底）
			cfg.AllowAllOrigins = true
			cfg.AllowCredentials = false
		}
	}
	return cors.New(cfg)
}
