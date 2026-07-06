// 零售域：套餐 / 合同 / 合同日价 / 月度结算 / 签约进度 / 合同 PDF / 绿电交易。
// 统一走零售管理(retail_management)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerRetail(g *gin.RouterGroup, d *Deps) {
	retailH := handler.NewRetailHandler(d.RetailRepo)
	contractPriceH := handler.NewContractPriceHandler(d.ContractPriceRepo)
	retailMonthlyH := handler.NewRetailMonthlyHandler(d.RetailMonthlyRepo)
	contractProgressH := handler.NewContractProgressHandler(d.ContractProgressRepo)
	priceTrendH := handler.NewPriceTrendHandler(d.PriceTrendRepo)
	contractPdfH := handler.NewContractPDFHandler(d.RetailRepo, d.AttachmentRepo, d.ObjectStore)
	greenPowerH := handler.NewGreenPowerHandler(d.GreenPowerRepo)

	reqRMRead := middleware.RequirePermission(d.PermSvc, "retail_management:read")
	reqRMWrite := middleware.RequirePermission(d.PermSvc, "retail_management:write")
	reqRMDelete := middleware.RequirePermission(d.PermSvc, "retail_management:delete")

	// 零售套餐 / 合同
	g.GET("/retail/pricing-models", reqRMRead, retailH.ListPricingModels)
	g.GET("/retail/packages", reqRMRead, retailH.ListPackages)
	g.POST("/retail/packages", reqRMWrite, retailH.CreatePackage)
	g.PUT("/retail/packages/:id", reqRMWrite, retailH.UpdatePackage)
	g.DELETE("/retail/packages/:id", reqRMDelete, retailH.DeletePackage)
	g.GET("/retail/contracts", reqRMRead, retailH.ListContracts)
	g.POST("/retail/contracts", reqRMWrite, retailH.CreateContract)
	g.GET("/retail/contracts/:id", reqRMRead, retailH.GetContract)
	g.PUT("/retail/contracts/:id", reqRMWrite, retailH.UpdateContract)
	g.DELETE("/retail/contracts/:id", reqRMDelete, retailH.DeleteContract)

	// 合同电价日维度（沿用零售模块权限）
	g.GET("/retail/price-daily", reqRMRead, contractPriceH.List)
	g.GET("/retail/price-daily/daily-summary", reqRMRead, priceTrendH.DailySummary) // 合同电价汇总，替代原 stub
	g.POST("/retail/price-daily/demo-data", reqRMWrite, contractPriceH.GenerateDemoData)

	// U1 零售月度结算
	g.GET("/retail/monthly-settlement", reqRMRead, retailMonthlyH.List)
	g.POST("/retail/monthly-settlement/demo-data", reqRMWrite, retailMonthlyH.GenerateDemoData)

	// U4 合同 PDF 生成
	g.POST("/retail/contracts/:id/pdf", reqRMWrite, contractPdfH.Generate)

	// G1 签约进度跟踪（沿用零售模块权限）
	g.GET("/retail/contract-progress", reqRMRead, contractProgressH.List)
	g.POST("/retail/contract-progress", reqRMWrite, contractProgressH.Create)
	g.POST("/retail/contract-progress/demo-data", reqRMWrite, contractProgressH.GenerateDemoData)

	// G3 绿电交易（沿用零售模块权限）
	g.GET("/trade/green-power", reqRMRead, greenPowerH.List)
	g.POST("/trade/green-power", reqRMWrite, greenPowerH.Create)
	g.DELETE("/trade/green-power/:id", reqRMDelete, greenPowerH.Delete)
	g.POST("/trade/green-power/demo-data", reqRMWrite, greenPowerH.GenerateDemoData)
}
