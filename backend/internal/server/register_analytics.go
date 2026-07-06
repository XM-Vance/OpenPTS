// 客户分析域：告警(统计·列表·确认) / 客户特性 / 客户负荷 / 客户利润 / 客户分析。
// 统一走分析(analytics)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerAnalytics(g *gin.RouterGroup, d *Deps) {
	analyticsH := handler.NewAnalyticsHandler(d.AnalyticsRepo, d.CustomerRepo, d.LoadCharExtRepo)
	custLoadH := handler.NewCustomerLoadHandler(d.CustLoadRepo)
	custProfitH := handler.NewCustomerProfitHandler(d.CustProfitRepo)
	custAnalysisH := handler.NewCustomerAnalysisHandler(d.CustAnalysisRepo)

	reqANRead := middleware.RequirePermission(d.PermSvc, "analytics:read")
	reqANWrite := middleware.RequirePermission(d.PermSvc, "analytics:write")

	// 客户分析（告警 + 特性）
	g.GET("/analytics/alerts/stats", reqANRead, analyticsH.Stats)
	g.GET("/analytics/alerts", reqANRead, analyticsH.ListAlerts)
	g.POST("/analytics/alerts/:id/ack", reqANWrite, analyticsH.AckAlert)
	g.GET("/analytics/characteristics", reqANRead, analyticsH.ListCharacteristics)
	g.POST("/analytics/demo-data", reqANWrite, analyticsH.GenerateDemoData)

	// E2 客户负荷分析（沿用客户分析模块权限）
	g.GET("/analytics/customer-load/summary", reqANRead, custLoadH.Summary)
	g.GET("/analytics/customer-load/:id/curve", reqANRead, custLoadH.LatestCurve)

	// F1 客户利润（沿用客户分析权限）
	g.GET("/analytics/customer-profit", reqANRead, custProfitH.List)
	g.POST("/analytics/customer-profit/demo-data", reqANWrite, custProfitH.GenerateDemoData)

	// G9 客户分析（沿用客户分析权限）
	g.GET("/analytics/customer-analysis", reqANRead, custAnalysisH.List)
	g.POST("/analytics/customer-analysis/demo-data", reqANWrite, custAnalysisH.GenerateDemoData)
}
