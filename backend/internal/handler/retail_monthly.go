// 零售月度结算。
// 2026-06 自 p0.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── U1 零售月度结算 ───
type RetailMonthlyHandler struct{ repo *db.RetailMonthlyRepository }

func NewRetailMonthlyHandler(repo *db.RetailMonthlyRepository) *RetailMonthlyHandler {
	return &RetailMonthlyHandler{repo: repo}
}

func (h *RetailMonthlyHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "60"))
	list, err := h.repo.List(c.Request.Context(), c.Query("contract_id"), limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *RetailMonthlyHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成零售月结演示数据"})
}
