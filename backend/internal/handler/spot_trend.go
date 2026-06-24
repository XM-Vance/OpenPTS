// 现货价格趋势handler。
// 2026-06 自 v1clone.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── D2 现货价格趋势 ───
type SpotTrendHandler struct{ repo *db.SpotTrendRepository }

func NewSpotTrendHandler(repo *db.SpotTrendRepository) *SpotTrendHandler {
	return &SpotTrendHandler{repo: repo}
}

func (h *SpotTrendHandler) DailyAvg(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	list, err := h.repo.DailyAvg(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *SpotTrendHandler) HourlyAvg(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	list, err := h.repo.HourlyAvg(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}
