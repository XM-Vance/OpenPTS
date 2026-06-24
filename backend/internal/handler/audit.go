// 审计日志 handler：分页查询 + 多维筛选。
package handler

import (
	"net/http"
	"strconv"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type AuditHandler struct {
	repo *db.AuditRepository
}

func NewAuditHandler(repo *db.AuditRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

// List GET /api/v1/audit/logs?username=&method=&resource=&days=7&limit=100&offset=0
// 响应含 total(同条件总行数),供前端分页。
func (h *AuditHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	logs, total, err := h.repo.List(c.Request.Context(), db.AuditFilter{
		Username: c.Query("username"),
		Method:   c.Query("method"),
		Resource: c.Query("resource"),
		Days:     days,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		log.Error().Err(err).Msg("查询失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": logs, "total": total})
}
