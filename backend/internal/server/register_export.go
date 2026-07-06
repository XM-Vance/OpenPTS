// 数据导入导出域：基础数据导出（需 system:read，防批量导出敏感数据）+ 客户批量导入（写权限 + 限流）。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerExport(g *gin.RouterGroup, d *Deps) {
	exportH := handler.NewExportHandler(d.CustomerRepo, d.MonthlyRepo, d.ContractPriceRepo)
	importH := handler.NewImportHandler(d.CustomerRepo)

	reqCMWrite := middleware.RequirePermission(d.PermSvc, "customer_management:write")
	reqSYRead := middleware.RequirePermission(d.PermSvc, "system:read")

	// M1 数据导出（需 system:read，防任意用户批量导出敏感数据）
	g.GET("/export/:resource", reqSYRead, exportH.Resource)

	// N1 数据导入（写权限）+ 限流：每用户 30 秒 1 次（rps=1/30=0.033，burst=2）
	g.POST("/import/customers", reqCMWrite, middleware.StrictRateLimit(0.033, 2), importH.Customers)
}
