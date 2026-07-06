// 滚动撮合交易。
// 2026-06 自 new_modules.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── 滚动撮合交易 ───

type RollingTradeHandler struct{ repo *db.RollingTradeRepository }

func NewRollingTradeHandler(repo *db.RollingTradeRepository) *RollingTradeHandler {
	return &RollingTradeHandler{repo: repo}
}

func (h *RollingTradeHandler) List(c *gin.Context) {
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

// Statistics 滚动撮合交易统计（买入/卖出/净/月度总量）。
func (h *RollingTradeHandler) Statistics(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	s, err := h.repo.Statistics(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *RollingTradeHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"trades": n, "message": "已生成滚动撮合交易演示数据"})
}
