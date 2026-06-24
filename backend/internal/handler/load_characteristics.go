// 负荷特性分析。
// 2026-06 自 new_modules2.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── 负荷特性分析 ───

type LoadCharacteristicsHandler struct {
	repo *db.LoadCharacteristicsRepository
}

func NewLoadCharacteristicsHandler(repo *db.LoadCharacteristicsRepository) *LoadCharacteristicsHandler {
	return &LoadCharacteristicsHandler{repo: repo}
}

func (h *LoadCharacteristicsHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := h.repo.List(c.Request.Context(), c.Query("month"), limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *LoadCharacteristicsHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成负荷特性演示数据"})
}
