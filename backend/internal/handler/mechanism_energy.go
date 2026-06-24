// 机制电量。
// 2026-06 自 p1.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── V4 机制电量 ───
type MechanismEnergyHandler struct{ repo *db.MechanismEnergyRepository }

func NewMechanismEnergyHandler(repo *db.MechanismEnergyRepository) *MechanismEnergyHandler {
	return &MechanismEnergyHandler{repo: repo}
}

func (h *MechanismEnergyHandler) List(c *gin.Context) {
	months, _ := strconv.Atoi(c.DefaultQuery("months", "12"))
	list, err := h.repo.List(c.Request.Context(), c.Query("voltage"), months)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *MechanismEnergyHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成机制电量演示数据"})
}
