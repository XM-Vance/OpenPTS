// 附件域：读需 system:read（防签名 URL 泄漏）；写/删需客户管理(customer_management)权限，
// 上传额外加严格限流（防滥用上传大文件）。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerAttachment(g *gin.RouterGroup, d *Deps) {
	attachmentH := handler.NewAttachmentHandler(d.AttachmentRepo, d.ObjectStore)

	reqCMWrite := middleware.RequirePermission(d.PermSvc, "customer_management:write")
	reqCMDelete := middleware.RequirePermission(d.PermSvc, "customer_management:delete")
	reqSYRead := middleware.RequirePermission(d.PermSvc, "system:read")

	// N2 附件（读需 system:read 防签名 URL 泄漏；写/删需 customer_management 权限）
	// 上传严格限流（防滥用上传大文件）：每用户 5 秒 1 次
	g.GET("/attachments", reqSYRead, attachmentH.List)
	g.GET("/attachments/:id/url", reqSYRead, attachmentH.DownloadURL)
	g.POST("/attachments", reqCMWrite, middleware.StrictRateLimit(0.2, 3), attachmentH.Upload)
	g.DELETE("/attachments/:id", reqCMDelete, attachmentH.Delete)
}
