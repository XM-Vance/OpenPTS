// 权限点列表（前端配置角色 UI 用）。
package handler

import (
	"net/http"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
)

type PermissionsHandler struct {
	repo *db.PermissionRepository
}

func NewPermissionsHandler(repo *db.PermissionRepository) *PermissionsHandler {
	return &PermissionsHandler{repo: repo}
}

// List GET /api/v1/permissions
func (h *PermissionsHandler) List(c *gin.Context) {
	list, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}
