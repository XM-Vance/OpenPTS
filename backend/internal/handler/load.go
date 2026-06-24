// 负荷管理 handler：演示数据生成 + 短期负荷预测接口（算法未内置，见 Forecast）。
package handler

import (
	"errors"
	"math"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type LoadHandler struct {
	repo     *db.LoadRepository
	demoMode bool
}

func NewLoadHandler(repo *db.LoadRepository, demoMode bool) *LoadHandler {
	return &LoadHandler{repo: repo, demoMode: demoMode}
}

// ─── 演示数据生成 ───

type demoDataRequest struct {
	CustomerID string `json:"customer_id" binding:"required,uuid"`
	Days       int    `json:"days"`
}

// GenerateDemoData POST /api/v1/load/demo-data
// 为指定客户生成最近若干天的合成 96 点负荷曲线，便于演示与预测。
func (h *LoadHandler) GenerateDemoData(c *gin.Context) {
	var req demoDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	cid, _ := uuid.Parse(req.CustomerID)

	days := req.Days
	if days <= 0 || days > 180 {
		days = 45
	}

	today := time.Now().Truncate(24 * time.Hour)
	count := 0
	for i := days; i >= 1; i-- {
		d := today.AddDate(0, 0, -i)
		curve, total := synthLoadCurve(d)
		if err := h.repo.UpsertCurve(c.Request.Context(), cid, d, curve, total); err != nil {
			if errors.Is(err, db.ErrOrgRequired) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "请先选择具体省份"})
				return
			}
			log.Error().Err(err).Msg("写入失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}
		count++
	}
	c.JSON(http.StatusOK, gin.H{"message": "已生成演示负荷数据", "days": count})
}

// synthLoadCurve 生成单日 96 点合成负荷曲线（含早/午/晚三个高峰 + 随机噪声）。
func synthLoadCurve(d time.Time) ([]float64, float64) {
	weekend := d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
	base, peak := 800.0, 2400.0
	if weekend {
		peak = 1500.0
	}

	curve := make([]float64, 96)
	var total float64
	for i := 0; i < 96; i++ {
		hour := float64(i) / 4.0
		shape := 0.30
		shape += 0.50 * gaussian(hour, 10.5, 3.2) // 上午峰
		shape += 0.42 * gaussian(hour, 15.0, 2.8) // 下午峰
		shape += 0.30 * gaussian(hour, 20.0, 2.0) // 晚峰
		if shape > 1 {
			shape = 1
		}
		v := base + (peak-base)*shape
		v *= 0.94 + 0.12*rand.Float64() // ±6% 噪声
		v = math.Round(v*100) / 100
		curve[i] = v
		total += v
	}
	// 总电量（MWh）= 各点功率均值 × 24h，此处 96 点 × 0.25h
	return curve, math.Round(total/4*100) / 100
}

func gaussian(x, mu, sigma float64) float64 {
	z := (x - mu) / sigma
	return math.Exp(-0.5 * z * z)
}

// ─── 短期负荷预测 ───

type forecastRequest struct {
	CustomerID string `json:"customer_id" binding:"required,uuid"`
	TargetDate string `json:"target_date" binding:"required"`
}

type algoHistoryItem struct {
	Date    string    `json:"date"`
	Weekday int       `json:"weekday"`
	Curve   []float64 `json:"curve"`
}

type algoForecastRequest struct {
	History    []algoHistoryItem `json:"history"`
	TargetDate string            `json:"target_date"`
}

type algoForecastResponse struct {
	Forecast      []float64 `json:"forecast"`
	Lower         []float64 `json:"lower"`
	Upper         []float64 `json:"upper"`
	Total         float64   `json:"total"`
	Peak          float64   `json:"peak"`
	Valley        float64   `json:"valley"`
	Method        string    `json:"method"`
	SampleDays    int       `json:"sample_days"`
	TargetWeekday int       `json:"target_weekday"`
}

// Forecast POST /api/v1/load/forecast
// 短期负荷预测接口。
//
// - DEMO_MODE=true（演示模式）：返回合成预测曲线，便于开箱体验，不代表真实算法效果。
// - 默认（未接入算法）：返回 501，等待二次开发者接入自有预测服务。
//
// 接入指引（二次开发）：
//   1. 用 h.repo.GetRecentCurves(ctx, customerID, targetDate, 30) 取历史 96 点曲线；
//   2. 调用你自有的预测服务（相似日/时序/机器学习），返回 algoForecastResponse 结构；
//   3. 将结果按下面的响应格式返回。请求/响应 DTO 已在本文件下方定义好。
func (h *LoadHandler) Forecast(c *gin.Context) {
	var req forecastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	targetDate, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_date 格式应为 YYYY-MM-DD"})
		return
	}

	// 演示模式：合成一条 96 点预测曲线（复用演示数据生成器）。
	if h.demoMode {
		curve, total := synthLoadCurve(targetDate)
		peak, valley := curve[0], curve[0]
		lower := make([]float64, len(curve))
		upper := make([]float64, len(curve))
		for i, v := range curve {
			if v > peak {
				peak = v
			}
			if v < valley {
				valley = v
			}
			lower[i] = math.Round((v*0.92)*100) / 100
			upper[i] = math.Round((v*1.08)*100) / 100
		}
		resp := algoForecastResponse{
			Forecast: curve, Lower: lower, Upper: upper,
			Total: total, Peak: math.Round(peak*100) / 100, Valley: math.Round(valley*100) / 100,
			Method: "demo-synthetic", SampleDays: 0,
			TargetWeekday: (int(targetDate.Weekday()) + 6) % 7,
		}
		c.JSON(http.StatusOK, gin.H{"target_date": req.TargetDate, "history_days": 0, "forecast": resp})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "负荷预测算法未配置。本开源骨架仅提供数据接口与请求/响应契约，请接入你的预测服务后实现该端点。",
	})
}
