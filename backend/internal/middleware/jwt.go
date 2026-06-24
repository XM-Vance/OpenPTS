// JWT 鉴权中间件：凭证按 Authorization 头(Bearer)→ ptis_token Cookie 顺序提取，
// 校验通过后写入 context。
// Cookie 路径是 P1-8 引入的默认通道(httpOnly,防 XSS 窃取)；
// Authorization 头保留给脚本/工具与过渡期旧客户端。
package middleware

import (
	"net/http"
	"strings"

	"github.com/ptis/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

// extractToken 依次尝试 Authorization 头、Cookie；都没有时返回空串。
func extractToken(c *gin.Context) (string, string) {
	if header := c.GetHeader("Authorization"); header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return "", "Authorization 格式应为 'Bearer <token>'"
		}
		return parts[1], ""
	}
	if tok, err := c.Cookie(auth.TokenCookieName); err == nil && tok != "" {
		return tok, ""
	}
	return "", "缺少登录凭证(Authorization 头或登录 Cookie)"
}

func JWT(svc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok, errMsg := extractToken(c)
		if tok == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errMsg})
			return
		}
		claims, err := svc.Parse(tok)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 无效"})
			return
		}
		c.Set(auth.ClaimsContextKey, claims)
		c.Next()
	}
}
