// 市场分析。
// 2026-06 自 p1.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── V5 市场分析 ───
type MarketAnalysisHandler struct{ repo *db.MarketAnalysisRepository }

func NewMarketAnalysisHandler(repo *db.MarketAnalysisRepository) *MarketAnalysisHandler {
	return &MarketAnalysisHandler{repo: repo}
}

func (h *MarketAnalysisHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	list, err := h.repo.List(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *MarketAnalysisHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"days": n, "message": "已生成市场分析演示数据"})
}
