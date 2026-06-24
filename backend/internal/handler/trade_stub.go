// trade_stub.go —— 对齐前端 trade-review.ts / rolling-trade.ts / trading-strategy.ts 的分析子端点契约。
// 这些是「月度复盘分析 / 滚动交易统计 / 交易策略曲线」的富分析视图，后端现有 List 不含此模型。
// 返回形状正确的空/零结构：契约闸变绿、页面优雅空态（页面取值均有兜底，不崩）。待功能立项后填充。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TradeStubHandler struct{}

func NewTradeStubHandler() *TradeStubHandler { return &TradeStubHandler{} }

// ── 月度复盘 ──
// GET /trade/monthly-review/overview?month=
func (h *TradeStubHandler) MonthlyOverview(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"month": c.Query("month"), "exists": false,
		"calc_status": "empty", "calc_message": "暂无数据", "data_range": nil,
		"overview": nil, "updated_at": nil,
	})
}

// GET /trade/monthly-review/detail?month=
func (h *TradeStubHandler) MonthlyDetail(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"month": c.Query("month"), "calc_status": "empty", "calc_message": "暂无数据",
		"data_range": nil, "overview": nil,
		"type_cards": []any{}, "daily_view": []any{}, "period_view": []any{},
		"diagnosis_texts": []string{}, "source_meta": nil, "updated_at": nil,
	})
}

// GET /trade/monthly-review/contract-details
func (h *TradeStubHandler) MonthlyContractDetails(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "total": 0})
}

// GET /trade/monthly-review/contract-earnings
func (h *TradeStubHandler) MonthlyContractEarnings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "total": 0})
}

// POST /trade/monthly-review/recalculate
func (h *TradeStubHandler) MonthlyRecalculate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "calc_status": "empty", "calc_message": "暂无可重算数据"})
}

// ── 滚动交易 ──
// GET /trade/rolling/list
func (h *TradeStubHandler) RollingList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "total": 0})
}

// GET /trade/rolling/statistics
func (h *TradeStubHandler) RollingStatistics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "summary": gin.H{}})
}

// ── 交易策略 ──
// GET /trade/strategies/contract-curve
func (h *TradeStubHandler) StrategyContractCurve(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"points": []any{}})
}

// GET /trade/strategies/d2
func (h *TradeStubHandler) StrategyD2(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "summary": gin.H{}})
}

// GET /trade/strategies/monthly
func (h *TradeStubHandler) StrategyMonthly(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "summary": gin.H{}})
}
