// 交易策略。
// 2026-06 自 new_modules2.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── 交易策略 ───

type TradeStrategyHandler struct{ repo *db.TradeStrategyRepository }

func NewTradeStrategyHandler(repo *db.TradeStrategyRepository) *TradeStrategyHandler {
	return &TradeStrategyHandler{repo: repo}
}

func (h *TradeStrategyHandler) List(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

type createStrategyReq struct {
	StrategyName string          `json:"strategy_name" binding:"required"`
	StrategyType string          `json:"strategy_type" binding:"required"`
	TargetMarket string          `json:"target_market" binding:"required"`
	Parameters   json.RawMessage `json:"parameters"`
	Status       string          `json:"status"`
	Note         string          `json:"note"`
}

func (h *TradeStrategyHandler) Create(c *gin.Context) {
	var req createStrategyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	if len(req.Parameters) == 0 {
		req.Parameters = json.RawMessage("{}")
	}
	id, err := h.repo.Create(c.Request.Context(), db.TradeStrategyInput{
		StrategyName: req.StrategyName,
		StrategyType: req.StrategyType,
		TargetMarket: req.TargetMarket,
		Parameters:   req.Parameters,
		Status:       req.Status,
		Note:         req.Note,
	})
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *TradeStrategyHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"strategies": n, "message": "已生成交易策略演示数据"})
}
