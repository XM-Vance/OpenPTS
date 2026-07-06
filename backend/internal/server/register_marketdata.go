// 市场行情域：通用市场行情数据 + 碳交易行情（CEA/CCER/EUA，全国统一，不分省）。
// 统一走价格管理(price_management)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerMarketData(g *gin.RouterGroup, d *Deps) {
	marketDataH := handler.NewMarketDataHandler(d.MarketDataRepo)
	carbonH := handler.NewCarbonHandler(d.CarbonRepo)

	reqPMRead := middleware.RequirePermission(d.PermSvc, "price_management:read")
	reqPMWrite := middleware.RequirePermission(d.PermSvc, "price_management:write")

	// Q1 市场行情数据（沿用价格模块权限）
	g.GET("/market-data", reqPMRead, marketDataH.Overview)
	g.GET("/market-data/overview", reqPMRead, marketDataH.Overview)
	g.GET("/market-data/tables", reqPMRead, marketDataH.ListTables)
	g.GET("/market-data/:table", reqPMRead, marketDataH.Query)

	// 碳交易行情（CEA/CCER/EUA，全国统一行情，不分省）
	g.GET("/carbon/summary", reqPMRead, carbonH.Summary)
	g.GET("/carbon/quotes", reqPMRead, carbonH.List)
	g.POST("/carbon/demo-data", reqPMWrite, carbonH.GenerateDemoData)
}
