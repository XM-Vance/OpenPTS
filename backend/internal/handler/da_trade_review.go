// 日前交易复盘handler。
// 2026-06 自 v1clone.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── D3 日前交易复盘 ───
type DATradeReviewHandler struct{ repo *db.DATradeReviewRepository }

func NewDATradeReviewHandler(repo *db.DATradeReviewRepository) *DATradeReviewHandler {
	return &DATradeReviewHandler{repo: repo}
}

func (h *DATradeReviewHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	list, err := h.repo.List(c.Request.Context(), limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *DATradeReviewHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"days": n, "message": "已生成日前复盘演示数据"})
}
