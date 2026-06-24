// 月度手工数据handler。
// 2026-06 自 v1clone_f.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── F4 月度手工数据 ───
type MonthlyManualHandler struct{ repo *db.MonthlyManualRepository }

func NewMonthlyManualHandler(repo *db.MonthlyManualRepository) *MonthlyManualHandler {
	return &MonthlyManualHandler{repo: repo}
}

func (h *MonthlyManualHandler) List(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context(), c.Query("month"), c.Query("category"))
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

type createManualReq struct {
	OperatingMonth string  `json:"operating_month" binding:"required"`
	Category       string  `json:"category" binding:"required"`
	ItemName       string  `json:"item_name" binding:"required"`
	Value          float64 `json:"value"`
	Unit           string  `json:"unit"`
	Source         string  `json:"source"`
	Note           string  `json:"note"`
}

func (h *MonthlyManualHandler) Create(c *gin.Context) {
	var req createManualReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	if req.Unit == "" {
		req.Unit = "元"
	}
	createdBy := ""
	if v, ok := c.Get(auth.ClaimsContextKey); ok {
		if claims, ok := v.(*auth.Claims); ok {
			createdBy = claims.Username
		}
	}
	id, err := h.repo.Create(c.Request.Context(), db.ManualItemInput{
		OperatingMonth: req.OperatingMonth,
		Category:       req.Category,
		ItemName:       req.ItemName,
		Value:          req.Value,
		Unit:           req.Unit,
		Source:         req.Source,
		Note:           req.Note,
		CreatedBy:      createdBy,
	})
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *MonthlyManualHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成月度手工数据"})
}
