// 分时交易规则 handler：规则 CRUD + Export。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// TradeRuleHandler 交易规则 handler。
type TradeRuleHandler struct{ repo *db.TradeRuleRepository }

// NewTradeRuleHandler 创建交易规则 handler。
func NewTradeRuleHandler(repo *db.TradeRuleRepository) *TradeRuleHandler {
	return &TradeRuleHandler{repo: repo}
}

// List GET /api/v1/trade-rules?category=settlement
func (h *TradeRuleHandler) List(c *gin.Context) {
	category := c.Query("category")
	list, err := h.repo.List(c.Request.Context(), category)
	if err != nil {
		log.Error().Err(err).Msg("查询交易规则失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Create POST /api/v1/trade-rules
func (h *TradeRuleHandler) Create(c *gin.Context) {
	var in db.TradeRuleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	result, err := h.repo.Create(c.Request.Context(), &in, claimsUserID(c))
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("创建交易规则失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": result})
}

// Update PUT /api/v1/trade-rules/:id
func (h *TradeRuleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var in db.TradeRuleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	result, err := h.repo.Update(c.Request.Context(), id, &in)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("更新交易规则失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": result})
}

// Delete DELETE /api/v1/trade-rules/:id
func (h *TradeRuleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("删除交易规则失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// Export GET /api/v1/trade-rules/export — 导出全部规则为 JSON 数组。
func (h *TradeRuleHandler) Export(c *gin.Context) {
	list, err := h.repo.Export(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("导出交易规则失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, list)
}
