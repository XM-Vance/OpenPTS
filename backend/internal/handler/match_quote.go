// 滚动撮合报价handler。
// 2026-06 自 v1clone_f.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── F3 撮合报价 ───
type MatchQuoteHandler struct{ repo *db.MatchQuoteRepository }

func NewMatchQuoteHandler(repo *db.MatchQuoteRepository) *MatchQuoteHandler {
	return &MatchQuoteHandler{repo: repo}
}

func (h *MatchQuoteHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	list, err := h.repo.List(c.Request.Context(), days, limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *MatchQuoteHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"quotes": n, "message": "已生成撮合演示数据"})
}
