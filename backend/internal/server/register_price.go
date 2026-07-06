// 价格域：价格预测 / 现货趋势 / TOU 时段规则 / 电网代理价 / 现货市场 / 市场分析。
// 统一走价格管理(price_management)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerPrice(g *gin.RouterGroup, d *Deps) {
	priceH := handler.NewPriceHandler(d.PriceRepo, d.Config.DemoMode)
	spotTrendH := handler.NewSpotTrendHandler(d.SpotTrendRepo)
	touH := handler.NewTOUHandler(d.TOURepo)
	gridAgencyH := handler.NewGridAgencyHandler(d.GridAgencyRepo)
	spotMarketH := handler.NewSpotMarketHandler(d.SpotMarketRepo)
	marketH := handler.NewMarketAnalysisHandler(d.MarketAnalysisRepo)
	priceTrendH := handler.NewPriceTrendHandler(d.PriceTrendRepo)

	reqPMRead := middleware.RequirePermission(d.PermSvc, "price_management:read")
	reqPMWrite := middleware.RequirePermission(d.PermSvc, "price_management:write")
	reqPMDelete := middleware.RequirePermission(d.PermSvc, "price_management:delete")

	// 价格预测
	g.POST("/price/forecast", reqPMRead, priceH.Forecast)
	g.POST("/price/demo-data", reqPMWrite, priceH.GenerateDemoData)

	// 现货价格趋势（沿用价格模块权限）
	g.GET("/price/trend/daily", reqPMRead, spotTrendH.DailyAvg)
	g.GET("/price/trend/hourly", reqPMRead, spotTrendH.HourlyAvg)
	g.GET("/price/trend/price-trend", reqPMRead, priceTrendH.PriceTrend) // 现货-合同价差，替代原 stub

	// E4 TOU 时段规则（沿用价格模块权限）
	g.GET("/price/tou-rules", reqPMRead, touH.List)
	g.POST("/price/tou-rules", reqPMWrite, touH.Create)
	g.PUT("/price/tou-rules/:id", reqPMWrite, touH.Update)
	g.DELETE("/price/tou-rules/:id", reqPMDelete, touH.Delete)
	g.POST("/price/tou-rules/demo-data", reqPMWrite, touH.GenerateDemoData)

	// E5 电网代理价（沿用价格模块权限）
	g.GET("/price/grid-agency", reqPMRead, gridAgencyH.List)
	g.POST("/price/grid-agency", reqPMWrite, gridAgencyH.Create)
	g.PUT("/price/grid-agency/:id", reqPMWrite, gridAgencyH.Update)
	g.DELETE("/price/grid-agency/:id", reqPMDelete, gridAgencyH.Delete)
	g.POST("/price/grid-agency/demo-data", reqPMWrite, gridAgencyH.GenerateDemoData)

	// G5 现货市场（沿用价格模块权限）
	g.GET("/price/spot-market", reqPMRead, spotMarketH.List)
	g.GET("/price/spot-market/statistics", reqPMRead, spotMarketH.Statistics)
	g.GET("/price/spot-market/price-curve", reqPMRead, spotMarketH.PriceCurve)
	g.POST("/price/spot-market/demo-data", reqPMWrite, spotMarketH.GenerateDemoData)

	// V5 市场分析（沿用价格权限）
	g.GET("/price/market-analysis", reqPMRead, marketH.List)
	g.POST("/price/market-analysis/demo-data", reqPMWrite, marketH.GenerateDemoData)
}
