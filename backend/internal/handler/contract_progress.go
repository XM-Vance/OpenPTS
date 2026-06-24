// 签约进度跟踪。
// 2026-06 自 new_modules.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── 签约进度跟踪 ───

type ContractProgressHandler struct {
	repo *db.ContractProgressRepository
}

func NewContractProgressHandler(repo *db.ContractProgressRepository) *ContractProgressHandler {
	return &ContractProgressHandler{repo: repo}
}

func (h *ContractProgressHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := h.repo.List(c.Request.Context(), c.Query("month"), c.Query("status"), limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

type createContractProgressReq struct {
	ContractID     string  `json:"contract_id" binding:"required"`
	OperatingMonth string  `json:"operating_month" binding:"required"`
	PlannedEnergy  float64 `json:"planned_energy_mwh"`
	ActualEnergy   float64 `json:"actual_energy_mwh"`
	Status         string  `json:"status"`
	Note           string  `json:"note"`
}

func (h *ContractProgressHandler) Create(c *gin.Context) {
	var req createContractProgressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	if req.Status == "" {
		req.Status = "on_track"
	}
	id, err := h.repo.Create(c.Request.Context(), db.ContractProgressInput{
		ContractID:     req.ContractID,
		OperatingMonth: req.OperatingMonth,
		PlannedEnergy:  req.PlannedEnergy,
		ActualEnergy:   req.ActualEnergy,
		Status:         req.Status,
		Note:           req.Note,
	})
	if err != nil {
		if errors.Is(err, db.ErrOrgRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先选择具体省份"})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *ContractProgressHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成签约进度演示数据"})
}
