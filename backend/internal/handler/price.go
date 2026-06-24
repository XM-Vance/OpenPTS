// 价格管理 handler：演示数据生成 + 日前价格预测接口（算法未内置，见 Forecast）。
// 复用 load.go 中的算法请求/响应结构（同包共享）与 gaussian 函数。
package handler

import (
	"math"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const pricePoints = 48

type PriceHandler struct {
	repo     *db.PriceRepository
	demoMode bool
}

func NewPriceHandler(repo *db.PriceRepository, demoMode bool) *PriceHandler {
	return &PriceHandler{repo: repo, demoMode: demoMode}
}

// ─── 演示数据 ───

type priceDemoRequest struct {
	Days int `json:"days"`
}

// GenerateDemoData POST /api/v1/price/demo-data
// 生成最近若干天的合成 48 点日前价（一日双峰：早 9:30 + 晚 19:30）。
func (h *PriceHandler) GenerateDemoData(c *gin.Context) {
	var req priceDemoRequest
	_ = c.ShouldBindJSON(&req) // 允许空 body
	days := req.Days
	if days <= 0 || days > 180 {
		days = 45
	}

	today := time.Now().Truncate(24 * time.Hour)
	for i := days; i >= 1; i-- {
		d := today.AddDate(0, 0, -i)
		for p := 1; p <= pricePoints; p++ {
			v := synthPriceValue(d, p)
			if err := h.repo.UpsertDayAheadPrice(c.Request.Context(), d, p, v); err != nil {
				log.Error().Err(err).Msg("写入失败")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
				return
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "已生成演示日前价数据", "days": days})
}

// synthPriceValue 合成单点 30 分钟日前价（¥/MWh）。
func synthPriceValue(d time.Time, period int) float64 {
	weekend := d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
	base, peak := 180.0, 850.0
	if weekend {
		peak = 600.0
	}
	hour := float64(period-1) / 2.0 // period 1..48 → 0..23.5 时
	shape := 0.2
	shape += 0.45 * gaussian(hour, 9.5, 2.5)  // 早峰
	shape += 0.85 * gaussian(hour, 19.5, 2.2) // 晚峰
	if shape > 1 {
		shape = 1
	}
	v := base + (peak-base)*shape
	v *= 0.92 + 0.16*rand.Float64() // ±8% 噪声
	return math.Round(v*100) / 100
}

// ─── 日前价格预测 ───

type priceForecastRequest struct {
	TargetDate string `json:"target_date" binding:"required"`
}

// Forecast POST /api/v1/price/forecast
// 日前价格预测接口。
//
// - DEMO_MODE=true（演示模式）：返回合成预测曲线，便于开箱体验，不代表真实算法效果。
// - 默认（未接入算法）：返回 501，等待二次开发者接入自有预测服务。
//
// 接入指引（二次开发）：
//   1. 用 h.repo.GetRecentDayAheadCurves(ctx, targetDate, 30) 取历史 48 点价格曲线；
//   2. 调用你自有的预测服务（相似日加权/时序/机器学习），返回 algoForecastResponse 结构；
//   3. 将结果按下面的响应格式返回。请求/响应 DTO 已在 load.go 中定义好。
func (h *PriceHandler) Forecast(c *gin.Context) {
	var req priceForecastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	targetDate, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_date 格式应为 YYYY-MM-DD"})
		return
	}

	// 演示模式：合成一条 48 点日前价预测曲线。
	if h.demoMode {
		curve := make([]float64, pricePoints)
		var total, peak, valley float64
		lower := make([]float64, pricePoints)
		upper := make([]float64, pricePoints)
		for p := 1; p <= pricePoints; p++ {
			v := synthPriceValue(targetDate, p)
			curve[p-1] = v
			total += v
			if p == 1 || v > peak {
				peak = v
			}
			if p == 1 || v < valley {
				valley = v
			}
			lower[p-1] = math.Round((v*0.93)*100) / 100
			upper[p-1] = math.Round((v*1.07)*100) / 100
		}
		resp := algoForecastResponse{
			Forecast: curve, Lower: lower, Upper: upper,
			Total: math.Round(total/2*100) / 100, Peak: math.Round(peak*100) / 100, Valley: math.Round(valley*100) / 100,
			Method: "demo-synthetic", SampleDays: 0,
			TargetWeekday: (int(targetDate.Weekday()) + 6) % 7,
		}
		c.JSON(http.StatusOK, gin.H{"target_date": req.TargetDate, "history_days": 0, "forecast": resp})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "价格预测算法未配置。本开源骨架仅提供数据接口与请求/响应契约，请接入你的预测服务后实现该端点。",
	})
}
