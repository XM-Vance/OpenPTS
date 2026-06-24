// 负荷数据诊断handler。
// 2026-06 自 v1clone_e.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── E3 ───
type LoadDiagnosisHandler struct{ repo *db.LoadDiagnosisRepository }

func NewLoadDiagnosisHandler(repo *db.LoadDiagnosisRepository) *LoadDiagnosisHandler {
	return &LoadDiagnosisHandler{repo: repo}
}

func (h *LoadDiagnosisHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	list, err := h.repo.List(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}
