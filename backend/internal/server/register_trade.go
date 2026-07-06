// 交易域：日前复盘 / 月度复盘 / 撮合报价 / 滚动撮合 / 竞价管理 / 交易策略 / 日前模拟。
// 统一走价格管理(price_management)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerTrade(g *gin.RouterGroup, d *Deps) {
	daReviewH := handler.NewDATradeReviewHandler(d.DAReviewRepo)
	mtrReviewH := handler.NewMonthlyTradeReviewHandler(d.MTRReviewRepo)
	matchQuoteH := handler.NewMatchQuoteHandler(d.MatchQuoteRepo)
	rollingTradeH := handler.NewRollingTradeHandler(d.RollingTradeRepo)
	biddingH := handler.NewBiddingHandler(d.BiddingRepo)
	tradeStrategyH := handler.NewTradeStrategyHandler(d.TradeStrategyRepo)
	daSimH := handler.NewDASimulationHandler(d.DASimRepo)

	reqPMRead := middleware.RequirePermission(d.PermSvc, "price_management:read")
	reqPMWrite := middleware.RequirePermission(d.PermSvc, "price_management:write")

	// 日前交易复盘（沿用价格模块权限）
	g.GET("/trade/da-review", reqPMRead, daReviewH.List)
	g.POST("/trade/da-review/demo-data", reqPMWrite, daReviewH.GenerateDemoData)

	// F2 月度交易复盘（沿用价格模块权限）
	g.GET("/trade/monthly-review", reqPMRead, mtrReviewH.List)
	g.GET("/trade/monthly-review/overview", reqPMRead, mtrReviewH.Overview) // 真数据，替代原 stub
	g.POST("/trade/monthly-review/demo-data", reqPMWrite, mtrReviewH.GenerateDemoData)

	// F3 撮合报价（沿用价格模块权限）
	g.GET("/trade/match-quotes", reqPMRead, matchQuoteH.List)
	g.POST("/trade/match-quotes/demo-data", reqPMWrite, matchQuoteH.GenerateDemoData)
	// G4a 撮合报价补充路由
	g.GET("/trade/match-quotes/days", reqPMRead, matchQuoteH.List)
	g.GET("/trade/match-quotes/quotes", reqPMRead, matchQuoteH.List)

	// G4 滚动撮合交易（沿用价格模块权限）
	g.GET("/trade/rolling", reqPMRead, rollingTradeH.List)
	g.GET("/trade/rolling/rounds", reqPMRead, rollingTradeH.List)
	g.GET("/trade/rolling/period-history", reqPMRead, rollingTradeH.List)
	g.GET("/trade/rolling/statistics", reqPMRead, rollingTradeH.Statistics)
	g.POST("/trade/rolling/demo-data", reqPMWrite, rollingTradeH.GenerateDemoData)

	// G7 竞价管理（沿用价格模块权限）
	g.GET("/trade/bidding", reqPMRead, biddingH.List)
	g.GET("/trade/bidding/statistics", reqPMRead, biddingH.Statistics)
	g.POST("/trade/bidding", reqPMWrite, biddingH.Create)
	g.POST("/trade/bidding/demo-data", reqPMWrite, biddingH.GenerateDemoData)

	// G10 交易策略（沿用价格模块权限）
	g.GET("/trade/strategies", reqPMRead, tradeStrategyH.List)
	g.GET("/trade/strategies/monthly", reqPMRead, tradeStrategyH.Monthly) // 真数据，替代原 stub
	g.POST("/trade/strategies", reqPMWrite, tradeStrategyH.Create)
	g.POST("/trade/strategies/demo-data", reqPMWrite, tradeStrategyH.GenerateDemoData)

	// P2-4 日前模拟（沿用价格模块权限）
	g.GET("/trade/da-simulation", reqPMRead, daSimH.List)
	g.POST("/trade/da-simulation", reqPMWrite, daSimH.Create)
	g.GET("/trade/da-simulation/:id", reqPMRead, daSimH.Get)
	g.POST("/trade/da-simulation/:id/run", reqPMWrite, daSimH.RunSimulation)
	g.DELETE("/trade/da-simulation/:id", reqPMWrite, daSimH.Delete)
	g.POST("/trade/da-simulation/demo-data", reqPMWrite, daSimH.GenerateDemoData)
}
