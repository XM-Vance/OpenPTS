// 客户负荷分析handler。
// 2026-06 自 v1clone_e.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── E2 ───
type CustomerLoadHandler struct{ repo *db.CustomerLoadRepository }

func NewCustomerLoadHandler(repo *db.CustomerLoadRepository) *CustomerLoadHandler {
	return &CustomerLoadHandler{repo: repo}
}

func (h *CustomerLoadHandler) Summary(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	list, err := h.repo.Summary(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *CustomerLoadHandler) LatestCurve(c *gin.Context) {
	id := c.Param("id")
	curve, err := h.repo.LatestCurve(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到负荷数据"})
		return
	}
	c.JSON(http.StatusOK, curve)
}
