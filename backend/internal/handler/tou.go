// TOU 时段规则handler。
// 2026-06 自 v1clone_e.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── E4 ───
type TOUHandler struct{ repo *db.TOURepository }

func NewTOUHandler(repo *db.TOURepository) *TOUHandler { return &TOUHandler{repo: repo} }

func (h *TOUHandler) List(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// TOURequest 创建/更新请求体。Periods 透传 96 段标签等 JSON。
type TOURequest struct {
	RuleName      string          `json:"rule_name" binding:"required,min=1,max=128"`
	EffectiveFrom string          `json:"effective_from" binding:"required"` // YYYY-MM-DD
	EffectiveTo   string          `json:"effective_to,omitempty"`            // YYYY-MM-DD，可空
	Periods       json.RawMessage `json:"periods" binding:"required"`        // jsonb
}

func (req TOURequest) toInput() (db.TOUInput, error) {
	from, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		return db.TOUInput{}, errors.New("effective_from 格式应为 YYYY-MM-DD")
	}
	in := db.TOUInput{
		RuleName:      req.RuleName,
		EffectiveFrom: from,
		Periods:       req.Periods,
	}
	if req.EffectiveTo != "" {
		to, err := time.Parse("2006-01-02", req.EffectiveTo)
		if err != nil {
			return db.TOUInput{}, errors.New("effective_to 格式应为 YYYY-MM-DD")
		}
		in.EffectiveTo = &to
	}
	return in, nil
}

func (h *TOUHandler) Create(c *gin.Context) {
	var req TOURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	in, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.repo.Create(c.Request.Context(), in)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *TOUHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req TOURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	in, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.repo.Update(c.Request.Context(), id, in)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		if errors.Is(err, db.ErrTOUNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "TOU 规则不存在"})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *TOUHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, db.ErrTOUNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "TOU 规则不存在"})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

func (h *TOUHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rules": n, "message": "已生成 TOU 演示规则"})
}
