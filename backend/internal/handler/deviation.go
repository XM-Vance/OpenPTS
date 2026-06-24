// 偏差结算。
// 2026-06 自 new_modules.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── 偏差结算 ───

type DeviationHandler struct{ repo *db.DeviationRepository }

func NewDeviationHandler(repo *db.DeviationRepository) *DeviationHandler {
	return &DeviationHandler{repo: repo}
}

func (h *DeviationHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	list, err := h.repo.List(c.Request.Context(), c.Query("category"), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *DeviationHandler) Summary(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	list, err := h.repo.Summary(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *DeviationHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成偏差结算演示数据"})
}
