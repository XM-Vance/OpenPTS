// 契约对齐补端点域：对齐前端实际调用路径，空数据时返回形状正确的空态（待功能立项）。
// 这些是 stub handler，集中在此，避免散落污染各业务域。权限码沿用对应业务模块。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerContractStubs(g *gin.RouterGroup, d *Deps) {
	bidStubH := handler.NewBidStubHandler()
	tradeStubH := handler.NewTradeStubHandler()
	miscStubH := handler.NewMiscStubHandler()

	reqPMRead := middleware.RequirePermission(d.PermSvc, "price_management:read")
	reqPMWrite := middleware.RequirePermission(d.PermSvc, "price_management:write")
	reqPMDelete := middleware.RequirePermission(d.PermSvc, "price_management:delete")
	reqSMRead := middleware.RequirePermission(d.PermSvc, "settlement_management:read")
	reqSMWrite := middleware.RequirePermission(d.PermSvc, "settlement_management:write")
	reqANRead := middleware.RequirePermission(d.PermSvc, "analytics:read")

	// 日前竞价（空态占位，待功能立项）
	g.GET("/bid/trade-sources", reqPMRead, bidStubH.TradeSources)
	g.GET("/bid/trade-sources/:id", reqPMRead, bidStubH.TradeSourceDetail)
	g.POST("/bid/trade-sources/auto", reqPMWrite, bidStubH.CreateTradeSource)
	g.POST("/bid/trade-sources/manual", reqPMWrite, bidStubH.CreateTradeSource)
	g.PUT("/bid/trade-sources/:id", reqPMWrite, bidStubH.UpdateTradeSource)
	g.POST("/bid/trade-sources/:id/status", reqPMWrite, bidStubH.SetTradeSourceStatus)
	g.DELETE("/bid/trade-sources/:id", reqPMDelete, bidStubH.DeleteTradeSource)
	g.GET("/bid/simulations/next-day", reqPMRead, bidStubH.NextDaySimulation)
	g.POST("/bid/simulations/manual-save", reqPMWrite, bidStubH.ManualSimulation)
	g.POST("/bid/simulations/manual-reset", reqPMWrite, bidStubH.ManualSimulation)
	g.GET("/bid/analysis/summary", reqPMRead, bidStubH.ProfitSummary)
	g.GET("/bid/analysis/profit-curve", reqPMRead, bidStubH.ProfitCurve)
	g.GET("/bid/analysis/daily", reqPMRead, bidStubH.ProfitDaily)
	g.GET("/bid/analysis/daily-review", reqPMRead, bidStubH.DailyReview)
	g.GET("/bid/analysis/daily-review/:date", reqPMRead, bidStubH.DailyReview)

	// 月度复盘 / 交易策略 分析子视图（空态占位）
	// 注：rolling/list、rolling/statistics、strategies/monthly、monthly-review/overview
	// 已由真 handler 实现（见 register_trade.go），此处不再重复注册。
	g.GET("/trade/monthly-review/detail", reqPMRead, tradeStubH.MonthlyDetail)
	g.GET("/trade/monthly-review/contract-details", reqPMRead, tradeStubH.MonthlyContractDetails)
	g.GET("/trade/monthly-review/contract-earnings", reqPMRead, tradeStubH.MonthlyContractEarnings)
	g.POST("/trade/monthly-review/recalculate", reqPMWrite, tradeStubH.MonthlyRecalculate)
	g.GET("/trade/strategies/contract-curve", reqPMRead, tradeStubH.StrategyContractCurve)
	g.GET("/trade/strategies/d2", reqPMRead, tradeStubH.StrategyD2)

	// 日前复盘子视图 / 批发月结算 / 合同电价趋势 / 调频补偿 / 客户分析视图（空态占位）
	g.GET("/trade/da-review/overview", reqPMRead, miscStubH.EmptyContainer)
	g.GET("/trade/da-review/detail", reqPMRead, miscStubH.EmptyContainer)
	g.GET("/trade/da-review/day-ahead", reqPMRead, miscStubH.EmptyContainer)
	g.GET("/trade/da-review/operation-detail", reqPMRead, miscStubH.EmptyContainer)
	g.GET("/trade/da-review/trade-dates", reqPMRead, miscStubH.TradeDates)
	g.GET("/wholesale-monthly-settlement", reqSMRead, miscStubH.EmptyContainer)
	g.GET("/wholesale-monthly-settlement/year", reqSMRead, miscStubH.EmptyContainer)
	g.GET("/wholesale-monthly-settlement/years", reqSMRead, miscStubH.SettlementYears)
	g.POST("/wholesale-monthly-settlement/import", reqSMWrite, miscStubH.EmptyImport)
	g.GET("/contract-price-trend/curve-analysis", reqPMRead, miscStubH.EmptyContainer)
	g.GET("/contract-price-trend/price-trend", reqPMRead, miscStubH.EmptyContainer)
	g.GET("/contract-price-trend/quantity-structure", reqPMRead, miscStubH.EmptyContainer)
	g.GET("/contract-price/daily-summary", reqPMRead, miscStubH.EmptyContainer)
	// freq-comp-fee 已移除：前端孤儿模块已删，后端 stub 同步清理（调频补偿费走 /freq/clearing 真路径）
	// load-characteristics/overview/scatter-data（连字符）已移除：前端改用真路径 /load/characteristics/（下划线）
	g.GET("/analytics/customer-load", reqANRead, miscStubH.EmptyContainer)
	g.GET("/customer-profit-analysis/dashboard", reqANRead, miscStubH.ProfitDashboard)
}
