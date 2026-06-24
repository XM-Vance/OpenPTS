// 中期负荷预测。
// 2026-06 自 p1.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── V2 中期预测 ───
type MediumForecastHandler struct{ repo *db.MediumForecastRepository }

func NewMediumForecastHandler(repo *db.MediumForecastRepository) *MediumForecastHandler {
	return &MediumForecastHandler{repo: repo}
}

func (h *MediumForecastHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "18"))
	list, err := h.repo.List(c.Request.Context(), limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *MediumForecastHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"months": n, "message": "已生成中期负荷预测演示数据"})
}
