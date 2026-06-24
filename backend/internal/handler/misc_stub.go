// misc_stub.go —— 收尾批：对齐前端各分析视图契约（da-review/合同电价趋势/批发月结算/
// 调频补偿/客户利润看板/负荷特性散点 等）。后端现有 List/明细不含这些聚合视图，
// 返回形状正确的空/零结构（页面取值有兜底，空态不崩）。待各功能立项后填充。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MiscStubHandler struct{}

func NewMiscStubHandler() *MiscStubHandler { return &MiscStubHandler{} }

// 通用空容器（绝大多数分析视图页面以 `?? []/{}` 兜底读取）
func (h *MiscStubHandler) EmptyContainer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "summary": gin.H{}, "total": 0})
}

// 导入类（POST）：返回已导入 0 条
func (h *MiscStubHandler) EmptyImport(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "imported": 0, "message": "暂未实现导入"})
}

// 日前复盘交易日列表
func (h *MiscStubHandler) TradeDates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"trade_dates": []any{}, "dates": []any{}})
}

// 批发月结算 年份列表
func (h *MiscStubHandler) SettlementYears(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"years": []any{}, "items": []any{}})
}

// 客户利润看板（CustomerProfitDashboard 全字段零值）
func (h *MiscStubHandler) ProfitDashboard(c *gin.Context) {
	emptyContribution := func(t string) gin.H {
		return gin.H{
			"top5": []any{},
			"others": gin.H{"profit": 0, "percentage": 0, "contribution_value": 0},
			"customer_count": 0, "total_profit": 0, "contribution_type": t,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"kpi": gin.H{
			"customer_count": 0, "total_energy_mwh": 0, "retail_revenue": 0, "retail_avg_price": 0,
			"wholesale_cost": 0, "wholesale_avg_price": 0, "gross_profit": 0, "avg_spread": 0,
			"source_summary": gin.H{"monthly_months": []any{}, "platform_daily_months": []any{}},
		},
		"positive_contribution": emptyContribution("positive"),
		"negative_contribution": emptyContribution("negative"),
		"rankings": gin.H{
			"profit": gin.H{"top5": []any{}, "bottom5": []any{}},
			"spread": gin.H{"top5": []any{}, "bottom5": []any{}},
		},
		"customer_list": gin.H{"total": 0, "page": 1, "page_size": 20, "items": []any{}},
	})
}
