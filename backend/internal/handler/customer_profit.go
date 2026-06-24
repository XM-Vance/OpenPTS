// 客户利润handler。
// 2026-06 自 v1clone_f.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── F1 客户利润 ───
type CustomerProfitHandler struct{ repo *db.CustomerProfitRepository }

func NewCustomerProfitHandler(repo *db.CustomerProfitRepository) *CustomerProfitHandler {
	return &CustomerProfitHandler{repo: repo}
}

func (h *CustomerProfitHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := h.repo.List(c.Request.Context(), c.Query("month"), limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *CustomerProfitHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成客户利润演示数据"})
}
