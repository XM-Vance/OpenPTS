// 系统模块（菜单数据来源）。
package handler

import (
	"net/http"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
)

type ModulesHandler struct {
	repo *db.ModuleRepository
}

func NewModulesHandler(repo *db.ModuleRepository) *ModulesHandler {
	return &ModulesHandler{repo: repo}
}

// List GET /api/v1/modules
func (h *ModulesHandler) List(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}
