// 预测基础数据。
// 2026-06 自 p0.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── U3 预测基础数据 ───
type ForecastBaseHandler struct{ repo *db.ForecastBaseRepository }

func NewForecastBaseHandler(repo *db.ForecastBaseRepository) *ForecastBaseHandler {
	return &ForecastBaseHandler{repo: repo}
}

func (h *ForecastBaseHandler) ListHolidays(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))
	list, err := h.repo.ListHolidays(c.Request.Context(), year)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *ForecastBaseHandler) ListCurves(c *gin.Context) {
	list, err := h.repo.ListTypicalCurves(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *ForecastBaseHandler) GenerateDemoData(c *gin.Context) {
	hN, cN, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"holidays": hN, "curves": cN, "message": "已生成节假日 + 典型曲线"})
}
