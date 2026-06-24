// 系统配置 handler。
package handler

import (
	"net/http"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type SettingsHandler struct{ repo *db.SettingsRepository }

func NewSettingsHandler(repo *db.SettingsRepository) *SettingsHandler {
	return &SettingsHandler{repo: repo}
}

// List GET /api/v1/system/settings?category=
func (h *SettingsHandler) List(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context(), c.Query("category"))
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

type updateSettingReq struct {
	Value string `json:"value" binding:"required"`
}

// Update PUT /api/v1/system/settings/:key
func (h *SettingsHandler) Update(c *gin.Context) {
	var req updateSettingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.repo.Update(c.Request.Context(), c.Param("key"), req.Value, claimsUsername(c)); err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新"})
}
