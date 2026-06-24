// 调频管理 handler：日清算汇总列表 + 演示数据生成（写 3 张表）。
package handler

import (
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type FreqHandler struct {
	repo *db.FreqRepository
}

func NewFreqHandler(repo *db.FreqRepository) *FreqHandler {
	return &FreqHandler{repo: repo}
}

// List GET /api/v1/freq/clearing?limit=
func (h *FreqHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	if limit <= 0 || limit > 200 {
		limit = 30
	}
	items, err := h.repo.ListDailySummary(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GenerateDemoData POST /api/v1/freq/demo-data
// 一天生成：AGC 清算 + AVC 清算 + 总需求 + 补偿费用 共 4 条记录。
func (h *FreqHandler) GenerateDemoData(c *gin.Context) {
	var req struct {
		Days int `json:"days"`
	}
	_ = c.ShouldBindJSON(&req)
	days := req.Days
	if days <= 0 || days > 90 {
		days = 30
	}
	ctx := c.Request.Context()
	today := time.Now().Truncate(24 * time.Hour)

	for i := days; i >= 1; i-- {
		d := today.AddDate(0, 0, -i)
		baseMul := 1.0
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			baseMul = 0.7 // 周末调频需求偏低
		}

		// AGC（一次调频）
		agcVol := r2((120 + 20*rand.Float64()) * baseMul)
		agcPrice := r2(18 + 6*rand.Float64())
		agcRev := r2(agcVol * agcPrice * 24) // 假设全天有效，价格 ¥/MW·h
		if err := h.repo.UpsertClearing(ctx, d, "AGC", agcVol, agcPrice, agcRev); err != nil {
			log.Error().Err(err).Msg("写入 AGC 失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}

		// AVC（电压调节）
		avcVol := r2((75 + 15*rand.Float64()) * baseMul)
		avcPrice := r2(12 + 5*rand.Float64())
		avcRev := r2(avcVol * avcPrice * 24)
		if err := h.repo.UpsertClearing(ctx, d, "AVC", avcVol, avcPrice, avcRev); err != nil {
			log.Error().Err(err).Msg("写入 AVC 失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}

		// 需求
		demandVol := r2(agcVol + avcVol + 10*rand.Float64())
		demandPrice := r2(14 + 4*rand.Float64())
		if err := h.repo.UpsertDemand(ctx, d, demandVol, demandPrice, "market"); err != nil {
			log.Error().Err(err).Msg("写入 demand 失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}

		// 补偿费
		comp := r2(4000 + 2000*rand.Float64())
		if err := h.repo.UpsertCompFee(ctx, d, "standard", comp); err != nil {
			log.Error().Err(err).Msg("写入 comp_fee 失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "已生成演示调频数据", "days": days})
}

func r2(v float64) float64 { return math.Round(v*100) / 100 }
