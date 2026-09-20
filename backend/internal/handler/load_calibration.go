// 负荷数据校准 handler（WP6.4）：恢复被 YAGNI 删除的校准功能。
// 表计侧（raw_meter_data 逐点×倍率）vs 系统侧（user_load_data）→ 系数 → 预览 → 应用。
//
// 端点：
//	POST /load-data/calibrations          计算并保存预览 {customer_id, month}
//	GET  /load-data/calibrations?month=   台账
//	POST /load-data/calibrations/:id/apply    应用（缩放系统侧曲线/合计）
//	POST /load-data/calibrations/:id/void     作废
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

type LoadCalibrationHandler struct {
	repo *db.LoadDataRepository
}

func NewLoadCalibrationHandler(repo *db.LoadDataRepository) *LoadCalibrationHandler {
	return &LoadCalibrationHandler{repo: repo}
}

// PreviewAndSave POST /api/v1/load-data/calibrations {customer_id, month}
func (h *LoadCalibrationHandler) PreviewAndSave(c *gin.Context) {
	var req struct {
		CustomerID string `json:"customer_id" binding:"required,uuid"`
		Month      string `json:"month" binding:"required"`
		Note       string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: customer_id 与 month 必填"})
		return
	}
	if _, err := time.Parse("2006-01", req.Month); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month 格式应为 YYYY-MM"})
		return
	}
	ctx := c.Request.Context()
	cal, err := h.repo.ComputeCalibration(ctx, req.CustomerID, req.Month)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Warn().Err(err).Msg("校准计算失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saved, err := h.repo.SaveCalibration(ctx, cal, req.Note)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"calibration": saved,
		"message": "已保存预览（应用后将按系数缩放该月系统侧曲线与合计）"})
}

// List GET /api/v1/load-data/calibrations?month=
func (h *LoadCalibrationHandler) List(c *gin.Context) {
	list, err := h.repo.ListCalibrations(c.Request.Context(), c.Query("month"), 0)
	if err != nil {
		log.Error().Err(err).Msg("校准台账查询失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Apply POST /api/v1/load-data/calibrations/:id/apply
func (h *LoadCalibrationHandler) Apply(c *gin.Context) {
	if err := h.repo.ApplyCalibration(c.Request.Context(), c.Param("id"), claimsUserIDString(c)); err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "校准已应用：该月系统侧曲线/合计已按系数缩放"})
}

// Void POST /api/v1/load-data/calibrations/:id/void
func (h *LoadCalibrationHandler) Void(c *gin.Context) {
	if err := h.repo.VoidCalibration(c.Request.Context(), c.Param("id")); err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已作废（已缩放的数据不自动回滚，可重导表计并重新聚合恢复）"})
}
