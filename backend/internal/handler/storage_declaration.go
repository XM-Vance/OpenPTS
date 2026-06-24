// 储能申报策略handler。
// 2026-06 自 v1clone_e.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── E6 ───
type StorageDeclHandler struct {
	repo *db.StorageDeclarationRepository
}

func NewStorageDeclHandler(repo *db.StorageDeclarationRepository) *StorageDeclHandler {
	return &StorageDeclHandler{repo: repo}
}

func (h *StorageDeclHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	list, err := h.repo.List(c.Request.Context(), c.Query("station_id"), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *StorageDeclHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"declarations": n, "message": "已生成储能申报演示数据"})
}
