// 自定义字段 handler：字段定义 CRUD + 实体值读写。
package handler

import (
	"encoding/json"
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

// ─── 自定义字段值（实体级，0102 新增）───

// customFieldValueRequest 字段值写入入参。
type customFieldValueRequest struct {
	DefinitionID string          `json:"definition_id" binding:"required"`
	EntityID     string          `json:"entity_id" binding:"required"`
	Value        json.RawMessage `json:"value"`
}

// ListValues GET /api/v1/custom-fields/values?entity_id=xxx
// 查询某实体的全部自定义字段值（按活跃省过滤）。
func (h *CustomFieldHandler) ListValues(c *gin.Context) {
	entityID, err := uuid.Parse(c.Query("entity_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_id 格式错误"})
		return
	}
	list, err := h.repo.ListValues(c.Request.Context(), entityID)
	if err != nil {
		log.Error().Err(err).Msg("查询自定义字段值失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// UpsertValue POST /api/v1/custom-fields/values
// 新增/更新某实体某字段的值（按 definition_id+entity_id 唯一键 upsert）。
func (h *CustomFieldHandler) UpsertValue(c *gin.Context) {
	var in customFieldValueRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	defID, err := uuid.Parse(in.DefinitionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "definition_id 格式错误"})
		return
	}
	entID, err := uuid.Parse(in.EntityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_id 格式错误"})
		return
	}
	result, err := h.repo.UpsertValue(c.Request.Context(), defID, entID, in.Value, claimsUserID(c))
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("写入自定义字段值失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": result})
}

// DeleteValue DELETE /api/v1/custom-fields/values?definition_id=xxx&entity_id=yyy
func (h *CustomFieldHandler) DeleteValue(c *gin.Context) {
	defID, err := uuid.Parse(c.Query("definition_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "definition_id 格式错误"})
		return
	}
	entID, err := uuid.Parse(c.Query("entity_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_id 格式错误"})
		return
	}
	if err := h.repo.DeleteValue(c.Request.Context(), defID, entID); err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		if err == db.ErrCustomFieldValueNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "字段值不存在"})
			return
		}
		log.Error().Err(err).Msg("删除自定义字段值失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
