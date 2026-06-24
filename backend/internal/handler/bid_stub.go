// bid_stub.go —— 对齐前端 lib/api/day-ahead-bid.ts 契约。
// 注意：日前竞价（交易源管理 + 次日申报模拟 + 盈亏分析 + 复盘）是一整套尚未实现的交易功能，
// 后端无对应数据模型。此处按前端形状返回「空/零」结构：契约闸变绿、页面优雅空态不崩，
// 待该功能正式立项后用真实逻辑替换。
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type BidStubHandler struct{}

func NewBidStubHandler() *BidStubHandler { return &BidStubHandler{} }

// 空的 SimulationDetail（字段与前端 SimulationDetail 接口一致）
func emptySimulation(tradeSourceID, targetDate string) gin.H {
	now := time.Now().Format("2006-01-02 15:04:05")
	return gin.H{
		"trade_source_id":         tradeSourceID,
		"target_date":             targetDate,
		"current_server_time":     now,
		"declaration_time":        "",
		"trade_type":              "auto",
		"strategy_name":           "",
		"strategy_id":             "",
		"strategy_code":           "",
		"next_day_declare_status": "未申报",
		"summary": gin.H{
			"total_bid_mwh":          0,
			"active_period_count":    0,
			"max_bid_mwh_per_period": 0,
		},
		"expected_pnl_yuan": nil,
		"price_forecast_30m": []float64{},
		"bid_mwh_30m":        []float64{},
		"is_editable":        false,
		"lock_reason":        nil,
	}
}

// GET /bid/trade-sources
func (h *BidStubHandler) TradeSources(c *gin.Context) { c.JSON(http.StatusOK, []any{}) }

// GET /bid/trade-sources/:id
func (h *BidStubHandler) TradeSourceDetail(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"trade_source_id": c.Param("id"), "trade_source_name": "", "trade_type": "auto",
		"strategy_id": "", "strategy_code": "", "trade_source_status": "停用",
		"next_day_declare_status": "未申报", "source_kind": "simulation", "readonly": true,
		"description": "", "params": []any{}, "created_at": "", "updated_at": "",
	})
}

// POST /bid/trade-sources/auto | /bid/trade-sources/manual  → 创建（占位，返回明确未实现）
func (h *BidStubHandler) CreateTradeSource(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "日前竞价交易源管理尚未实现"})
}

// PUT /bid/trade-sources/:id
func (h *BidStubHandler) UpdateTradeSource(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "日前竞价交易源管理尚未实现"})
}

// POST /bid/trade-sources/:id/status
func (h *BidStubHandler) SetTradeSourceStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /bid/trade-sources/:id
func (h *BidStubHandler) DeleteTradeSource(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /bid/simulations/next-day
func (h *BidStubHandler) NextDaySimulation(c *gin.Context) {
	c.JSON(http.StatusOK, emptySimulation(c.Query("trade_source_id"), time.Now().AddDate(0, 0, 1).Format("2006-01-02")))
}

// POST /bid/simulations/manual-save  /  /bid/simulations/manual-reset
func (h *BidStubHandler) ManualSimulation(c *gin.Context) {
	var body struct {
		TradeSourceID string `json:"trade_source_id"`
		TargetDate    string `json:"target_date"`
	}
	_ = c.ShouldBindJSON(&body)
	c.JSON(http.StatusOK, emptySimulation(body.TradeSourceID, body.TargetDate))
}

// GET /bid/analysis/summary  → ProfitSummary（全字段零值）
func (h *BidStubHandler) ProfitSummary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"trade_source_id": c.Query("trade_source_id"),
		"start_date":      c.Query("start_date"), "end_date": c.Query("end_date"),
		"total_realized_pnl_yuan": 0, "avg_daily_realized_pnl_yuan": 0,
		"daily_win_rate": 0, "period_win_rate": 0,
		"profitable_amount_yuan": 0, "loss_amount_yuan": 0, "profit_loss_ratio": 0,
		"avg_profit_yuan": 0, "avg_loss_yuan": 0, "avg_profit_loss_ratio": 0,
		"max_single_day_profit_yuan": 0, "max_single_day_loss_yuan": 0, "max_profit_loss_ratio": 0,
		"max_drawdown_yuan": 0, "unit_pnl_yuan_per_mwh": 0,
		"avg_bid_mwh_per_active_period": 0, "avg_period_pnl_yuan": 0, "trading_days": 0,
	})
}

// GET /bid/analysis/profit-curve → ProfitCurveResponse
func (h *BidStubHandler) ProfitCurve(c *gin.Context) {
	metric := c.DefaultQuery("metric", "amount")
	c.JSON(http.StatusOK, gin.H{
		"trade_source_id": c.Query("trade_source_id"), "metric": metric, "points": []any{},
	})
}

// GET /bid/analysis/daily → ProfitDailyResponse
func (h *BidStubHandler) ProfitDaily(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"rows": []any{},
		"summary_row": gin.H{
			"bid_total_mwh": 0, "realized_pnl_yuan": 0, "unit_pnl_yuan_per_mwh": 0,
			"win_periods": 0, "loss_periods": 0,
		},
	})
}

// GET /bid/analysis/daily-review/:date → DailyReviewDetail
func (h *BidStubHandler) DailyReview(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"trade_source_id": c.Query("trade_source_id"), "trade_source_name": "",
		"target_date": c.Param("date"),
		"summary": gin.H{
			"expected_pnl_yuan": nil, "realized_pnl_yuan": 0, "total_bid_mwh": 0,
			"win_periods": 0, "loss_periods": 0, "avg_spread_yuan_per_mwh": 0,
		},
		"chart_rows": []any{}, "period_profit_rows": []any{},
	})
}
