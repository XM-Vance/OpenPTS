// 审批流域：读/发起/撤回需 system:read；审批/驳回需 system:write。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerApproval(g *gin.RouterGroup, d *Deps) {
	approvalH := handler.NewApprovalHandler(d.ApprovalRepo, d.ApprovalReg, d.SSEHub)

	reqSYRead := middleware.RequirePermission(d.PermSvc, "system:read")

	// N3 审批流（读/发起需 system:read；审批权用 system:write）
	g.GET("/approvals", reqSYRead, approvalH.List)
	g.GET("/approvals/templates", reqSYRead, approvalH.ListTemplates)
	g.GET("/approvals/by-resource", reqSYRead, approvalH.ByResource)
	g.GET("/approvals/:id", reqSYRead, approvalH.Get)
	g.POST("/approvals", reqSYRead, approvalH.Submit)
	g.POST("/approvals/:id/approve", middleware.RequirePermission(d.PermSvc, "system:write"), approvalH.Approve)
	g.POST("/approvals/:id/reject", middleware.RequirePermission(d.PermSvc, "system:write"), approvalH.Reject)
	g.POST("/approvals/:id/withdraw", reqSYRead, approvalH.Withdraw)
}
