// 实时推送域：SSE 告警流 / WebSocket（公开，JWT 在 handler 内按 query 解析）
// + SSE 测试推送 / 在线用户（鉴权）。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerRealtime(apiV1, authed *gin.RouterGroup, d *Deps) {
	sseH := handler.NewSSEHandler(d.SSEHub, d.JWT)
	wsH := handler.NewWebSocketHandler(d.JWT)

	reqSYRead := middleware.RequirePermission(d.PermSvc, "system:read")

	// SSE 流：JWT 在 handler 内通过 query 参数解析
	apiV1.GET("/stream/alerts", sseH.Stream)
	// WebSocket 双向（演示用：echo）
	apiV1.GET("/ws/echo", wsH.Echo)

	// M2 SSE 测试推送（仅 system:read 即可触发，便于运维联调）
	authed.POST("/stream/test", reqSYRead, sseH.PublishTest)
	// T4 在线用户（任何已登录用户可查）
	authed.GET("/online", sseH.Online)
}
