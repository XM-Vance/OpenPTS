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

func (h *MonthlyTradeReviewHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"months": n, "message": "已生成月度复盘演示数据"})
}
