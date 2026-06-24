// 预结算。
// 2026-06 自 p0.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── U2 预结算 ───
type PreSettleHandler struct{ repo *db.PreSettleRepository }

func NewPreSettleHandler(repo *db.PreSettleRepository) *PreSettleHandler {
	return &PreSettleHandler{repo: repo}
}

func (h *PreSettleHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	list, err := h.repo.List(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *PreSettleHandler) Get(c *gin.Context) {
	p, err := h.repo.Get(c.Request.Context(), c.Param("date"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该日预结算"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *PreSettleHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"days": n, "message": "已生成预结算演示数据"})
}
