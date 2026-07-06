// 电网代理价handler。
// 2026-06 自 v1clone_e.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── E5 ───
type GridAgencyHandler struct{ repo *db.GridAgencyRepository }

func NewGridAgencyHandler(repo *db.GridAgencyRepository) *GridAgencyHandler {
	return &GridAgencyHandler{repo: repo}
}

func (h *GridAgencyHandler) List(c *gin.Context) {
	months, _ := strconv.Atoi(c.DefaultQuery("months", "12"))
	list, err := h.repo.List(c.Request.Context(), c.Query("voltage"), months)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// GridAgencyRequest 创建/更新请求体。
type GridAgencyRequest struct {
	OperatingMonth string  `json:"operating_month" binding:"required"` // YYYY-MM
	VoltageLevel   string  `json:"voltage_level" binding:"required"`   // 380V / 10kV / 35kV / 110kV
	AvgPrice       float64 `json:"avg_price"`
	PeakPrice      float64 `json:"peak_price"`
	FlatPrice      float64 `json:"flat_price"`
	ValleyPrice    float64 `json:"valley_price"`
}

func (req GridAgencyRequest) toInput() db.GridAgencyInput {
	return db.GridAgencyInput{
		OperatingMonth: req.OperatingMonth,
		VoltageLevel:   req.VoltageLevel,
		AvgPrice:       req.AvgPrice,
		PeakPrice:      req.PeakPrice,
		FlatPrice:      req.FlatPrice,
		ValleyPrice:    req.ValleyPrice,
	}
}

func (h *GridAgencyHandler) Create(c *gin.Context) {
	var req GridAgencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	row, err := h.repo.Create(c.Request.Context(), req.toInput())
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, row)
}

func (h *GridAgencyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req GridAgencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	row, err := h.repo.Update(c.Request.Context(), id, req.toInput())
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		if errors.Is(err, db.ErrGridAgencyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "电网代理价记录不存在"})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *GridAgencyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, db.ErrGridAgencyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "电网代理价记录不存在"})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

func (h *GridAgencyHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成电网代理价演示数据"})
}
