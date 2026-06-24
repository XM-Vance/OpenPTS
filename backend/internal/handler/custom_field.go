// 自定义字段 handler：字段定义 CRUD。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// CustomFieldHandler 自定义字段定义 handler。
type CustomFieldHandler struct{ repo *db.CustomFieldRepository }

// NewCustomFieldHandler 创建自定义字段 handler。
func NewCustomFieldHandler(repo *db.CustomFieldRepository) *CustomFieldHandler {
	return &CustomFieldHandler{repo: repo}
}

// List GET /api/v1/custom-fields?entity_type=customer
func (h *CustomFieldHandler) List(c *gin.Context) {
	entityType := c.Query("entity_type")
	list, err := h.repo.List(c.Request.Context(), entityType)
	if err != nil {
		log.Error().Err(err).Msg("查询自定义字段失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Create POST /api/v1/custom-fields
func (h *CustomFieldHandler) Create(c *gin.Context) {
	var in db.CustomFieldDefInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	result, err := h.repo.Create(c.Request.Context(), &in, claimsUserID(c))
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("创建自定义字段失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": result})
}

// Update PUT /api/v1/custom-fields/:id
func (h *CustomFieldHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var in db.CustomFieldDefInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	result, err := h.repo.Update(c.Request.Context(), id, &in)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("更新自定义字段失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": result})
}

// Delete DELETE /api/v1/custom-fields/:id
func (h *CustomFieldHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("删除自定义字段失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
