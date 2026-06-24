// 大屏数据聚合。
// 2026-06 自 new_modules2.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── 大屏数据聚合 ───

type DisplayScreenHandler struct{ repo *db.DisplayScreenRepository }

func NewDisplayScreenHandler(repo *db.DisplayScreenRepository) *DisplayScreenHandler {
	return &DisplayScreenHandler{repo: repo}
}

func (h *DisplayScreenHandler) Overview(c *gin.Context) {
	overview, err := h.repo.Overview(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *DisplayScreenHandler) Trend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	list, err := h.repo.Trend(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *DisplayScreenHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成大屏演示数据"})
}
