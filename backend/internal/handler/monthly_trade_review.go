// 月度交易复盘handler。
// 2026-06 自 v1clone_f.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── F2 月度复盘 ───
type MonthlyTradeReviewHandler struct {
	repo *db.MonthlyTradeReviewRepository
}

func NewMonthlyTradeReviewHandler(repo *db.MonthlyTradeReviewRepository) *MonthlyTradeReviewHandler {
	return &MonthlyTradeReviewHandler{repo: repo}
}

func (h *MonthlyTradeReviewHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	list, err := h.repo.List(c.Request.Context(), limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Overview 月度复盘概览（读 monthly_trade_review 表，替代原 stub 空态）。
func (h *MonthlyTradeReviewHandler) Overview(c *gin.Context) {
	month := c.Query("month")
	m, err := h.repo.GetByMonth(c.Request.Context(), month)
	if err != nil {
		// 无数据返回空态（exists:false），不报错——月份未结算属正常
		c.JSON(http.StatusOK, gin.H{
			"month": month, "exists": false,
			"calc_status": "empty", "calc_message": "该月份暂无复盘数据",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"month":   m.OperatingMonth,
		"exists":  true,
		"overview": m,
	})
}

func (h *MonthlyTradeReviewHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"months": n, "message": "已生成月度复盘演示数据"})
}
