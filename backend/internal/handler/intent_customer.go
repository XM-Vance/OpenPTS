// 意向客户handler。
// 2026-06 自 v1clone_e.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── E1 ───
type IntentCustomerHandler struct{ repo *db.IntentCustomerRepository }

func NewIntentCustomerHandler(repo *db.IntentCustomerRepository) *IntentCustomerHandler {
	return &IntentCustomerHandler{repo: repo}
}

func (h *IntentCustomerHandler) List(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *IntentCustomerHandler) Diagnose(c *gin.Context) {
	list, err := h.repo.Diagnose(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *IntentCustomerHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"customers": n, "message": "已生成意向客户演示数据"})
}
