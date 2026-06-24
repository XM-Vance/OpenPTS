// 标签定义 handler：标签 CRUD。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// TagHandler 标签 handler。
type TagHandler struct{ repo *db.TagRepository }

// NewTagHandler 创建标签 handler。
func NewTagHandler(repo *db.TagRepository) *TagHandler {
	return &TagHandler{repo: repo}
}

// List GET /api/v1/tags?entity_type=customer
func (h *TagHandler) List(c *gin.Context) {
	entityType := c.Query("entity_type")
	list, err := h.repo.List(c.Request.Context(), entityType)
	if err != nil {
		log.Error().Err(err).Msg("查询标签失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Create POST /api/v1/tags
func (h *TagHandler) Create(c *gin.Context) {
	var in db.TagInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	result, err := h.repo.Create(c.Request.Context(), &in, claimsUserID(c))
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("创建标签失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": result})
}

// Update PUT /api/v1/tags/:id
func (h *TagHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var in db.TagInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	result, err := h.repo.Update(c.Request.Context(), id, &in)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("更新标签失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": result})
}

// Delete DELETE /api/v1/tags/:id
func (h *TagHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("删除标签失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// BatchApply POST /api/v1/tags/batch-apply
func (h *TagHandler) BatchApply(c *gin.Context) {
	var req struct {
		TagID      uuid.UUID   `json:"tag_id" binding:"required"`
		EntityType string      `json:"entity_type" binding:"required"`
		EntityIDs  []uuid.UUID `json:"entity_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误：需要 tag_id, entity_type, entity_ids"})
		return
	}
	applied, err := h.repo.BatchApply(c.Request.Context(), req.TagID, req.EntityType, req.EntityIDs, claimsUserID(c))
	if err != nil {
		log.Error().Err(err).Msg("批量打标签失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"applied": applied, "total": len(req.EntityIDs)})
}

// GetEntityTags GET /api/v1/tags/entity?entity_type=customer&entity_id=xxx
func (h *TagHandler) GetEntityTags(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")
	if entityType == "" || entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需要 entity_type 和 entity_id"})
		return
	}
	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_id 格式错误"})
		return
	}
	tags, err := h.repo.GetEntityTags(c.Request.Context(), entityType, entityID)
	if err != nil {
		log.Error().Err(err).Msg("查询实体标签失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": tags})
}
