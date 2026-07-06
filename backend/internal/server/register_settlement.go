// 结算域：日结算 / 月结算 / 月度手工数据 / 预结算 / 偏差结算 / 机制电量 / 交易规则。
// 统一走结算管理(settlement_management)权限码。
//
// 注意：本开源骨架不含结算引擎（preview/settle/calculate 端点已移除），
// 也不含具体省份的偏差考核算法。相关端点可由二次开发者在此接入。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerSettlement(g *gin.RouterGroup, d *Deps) {
	settlementH := handler.NewSettlementHandler(d.SettlementRepo)
	monthlyH := handler.NewMonthlySettlementHandler(d.MonthlyRepo)
	manualH := handler.NewMonthlyManualHandler(d.ManualRepo)
	preSettleH := handler.NewPreSettleHandler(d.PreSettleRepo)
	deviationH := handler.NewDeviationHandler(d.DeviationRepo)
	mechH := handler.NewMechanismEnergyHandler(d.MechEnergyRepo)
	tradeRuleH := handler.NewTradeRuleHandler(db.NewTradeRuleRepository(d.Pool))

	reqSMRead := middleware.RequirePermission(d.PermSvc, "settlement_management:read")
	reqSMWrite := middleware.RequirePermission(d.PermSvc, "settlement_management:write")

	// 日结算
	g.GET("/settlement/daily", reqSMRead, settlementH.List)
	g.GET("/settlement/daily/:date", reqSMRead, settlementH.Get)
	g.POST("/settlement/demo-data", reqSMWrite, settlementH.GenerateDemoData)

	// 月度结算（沿用结算模块权限）
	g.GET("/settlement/monthly", reqSMRead, monthlyH.List)
	g.POST("/settlement/monthly/demo-data", reqSMWrite, monthlyH.GenerateDemoData)

	// F4 月度手工数据（沿用结算模块权限）
	g.GET("/settlement/manual-data", reqSMRead, manualH.List)
	g.POST("/settlement/manual-data", reqSMWrite, manualH.Create)
	g.POST("/settlement/manual-data/demo-data", reqSMWrite, manualH.GenerateDemoData)

	// U2 预结算
	g.GET("/settlement/pre", reqSMRead, preSettleH.List)
	g.GET("/settlement/pre/:date", reqSMRead, preSettleH.Get)
	g.POST("/settlement/pre/demo-data", reqSMWrite, preSettleH.GenerateDemoData)

	// G2 偏差结算（沿用结算模块权限；calculate 端点为算法接入点，未内置）
	g.GET("/settlement/deviation", reqSMRead, deviationH.List)
	g.GET("/settlement/deviation/summary", reqSMRead, deviationH.Summary)
	g.POST("/settlement/deviation/demo-data", reqSMWrite, deviationH.GenerateDemoData)

	// V4 机制电量（沿用结算权限）
	g.GET("/settlement/mechanism-energy", reqSMRead, mechH.List)
	g.POST("/settlement/mechanism-energy/demo-data", reqSMWrite, mechH.GenerateDemoData)

	// 交易规则（通用 key/value 规则表，规则数据由部署方填充）
	g.GET("/trade-rules", reqSMRead, tradeRuleH.List)
	g.GET("/trade-rules/export", reqSMRead, tradeRuleH.Export)
	g.POST("/trade-rules", reqSMWrite, tradeRuleH.Create)
	g.PUT("/trade-rules/:id", reqSMWrite, tradeRuleH.Update)
	g.DELETE("/trade-rules/:id", reqSMWrite, tradeRuleH.Delete)
}
