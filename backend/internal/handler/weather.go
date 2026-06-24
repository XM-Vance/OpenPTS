// 气象handler。
// 2026-06 自 v1clone.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── D4 气象 ───
type WeatherHandler struct{ repo *db.WeatherRepository }

func NewWeatherHandler(repo *db.WeatherRepository) *WeatherHandler {
	return &WeatherHandler{repo: repo}
}

func (h *WeatherHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	list, err := h.repo.List(c.Request.Context(), c.Query("location"), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *WeatherHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"records": n, "message": "已生成气象演示数据"})
}
