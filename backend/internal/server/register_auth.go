// 鉴权域：登录（公开，独立限流）+ 当前用户信息/改密/登出（鉴权）
// + 用户 API Key（创建/列表/吊销凭登录态；exchange 凭 Key 换 JWT，公开）。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerAuth(apiV1, authed *gin.RouterGroup, d *Deps) {
	authHandler := handler.NewAuthHandler(d.UserRepo, d.JWT, d.PermSvc)

	// 登录端点：双层防护——
	//   1) LoginRateLimit(30)：按 IP 请求速率限流，防刷屏。
	//   2) LoginLockout：按 IP+username 失败次数锁定（5 次失败锁 15 分钟），防密码爆破。
	apiV1.POST("/auth/login", middleware.LoginRateLimit(30), middleware.LoginLockout(), authHandler.Login)

	authed.GET("/auth/me", authHandler.Me)
	authed.GET("/auth/me/permissions", authHandler.MyPermissions)
	authed.POST("/auth/change-password", authHandler.ChangePassword)
	authed.POST("/auth/logout", authHandler.Logout) // 清除登录 Cookie(P1-8)

	// 用户 API Key（凭登录态管理自己的 Key；登录即可，无需额外权限码）
	keyRepo := db.NewAPIKeyRepository(d.Pool)
	keyHandler := handler.NewAPIKeyHandler(keyRepo, d.UserRepo, d.JWT)
	authed.POST("/auth/api-keys", keyHandler.Create)
	authed.GET("/auth/api-keys", keyHandler.List)
	authed.DELETE("/auth/api-keys/:id", keyHandler.Revoke)
	// exchange：凭 API Key 换 JWT（公开路由，凭 Key 本身鉴权）。
	// 复用登录限流防 Key 爆破；ExtractAPIKey 从 Authorization: Bearer 取 Key。
	apiV1.POST("/auth/api-key/exchange",
		middleware.LoginRateLimit(30),
		handler.ExtractAPIKey(),
		keyHandler.Exchange)
}
