// 结算管理 handler：列表 / 详情 / 演示数据生成。
// 复用 load.go 中的 gaussian。
package handler

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const settlementPoints = 48

type SettlementHandler struct {
	repo *db.SettlementRepository
}

func NewSettlementHandler(repo *db.SettlementRepository) *SettlementHandler {
	return &SettlementHandler{repo: repo}
}

// periodDetail 单时段明细（JSONB 内部结构，仅 backend 知晓）
type periodDetail struct {
	Period    int     `json:"period"`
	VolumeMWh float64 `json:"volume_mwh"`
	Price     float64 `json:"price"`
	Fee       float64 `json:"fee"`
}

// List GET /api/v1/settlement/daily?limit=
func (h *SettlementHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	if limit <= 0 || limit > 200 {
		limit = 30
	}
	items, err := h.repo.ListRecent(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Get GET /api/v1/settlement/daily/:date?version=PRELIMINARY
func (h *SettlementHandler) Get(c *gin.Context) {
	d, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date 格式应为 YYYY-MM-DD"})
		return
	}
	version := c.DefaultQuery("version", "PRELIMINARY")
	s, err := h.repo.GetByDate(c.Request.Context(), d, version)
	if err != nil {
		if errors.Is(err, db.ErrSettlementNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "结算记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// GenerateDemoData POST /api/v1/settlement/demo-data
func (h *SettlementHandler) GenerateDemoData(c *gin.Context) {
	var req struct {
		Days int `json:"days"`
	}
	_ = c.ShouldBindJSON(&req)
	days := req.Days
	if days <= 0 || days > 90 {
		days = 30
	}
	today := time.Now().Truncate(24 * time.Hour)
	for i := days; i >= 1; i-- {
		d := today.AddDate(0, 0, -i)
		s := buildDemoSettlement(d)
		if err := h.repo.Upsert(c.Request.Context(), s); err != nil {
			if respondOrgRequired(c, err) {
				return
			}
			log.Error().Err(err).Msg("写入失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "已生成演示结算数据", "days": days})
}

// buildDemoSettlement 合成单日 48 点结算（沿用价格双峰形态）。
func buildDemoSettlement(d time.Time) *db.SettlementDaily {
	weekend := d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
	basePrice, peakPrice := 380.0, 850.0
	if weekend {
		peakPrice = 620.0
	}
	baseVolume := 95.0 // MW

	details := make([]periodDetail, settlementPoints)
	var totalFee, totalVolumeMWh float64

	for p := 0; p < settlementPoints; p++ {
		hour := float64(p) / 2.0
		shape := 0.2
		shape += 0.45 * gaussian(hour, 9.5, 2.5)
		shape += 0.85 * gaussian(hour, 19.5, 2.2)
		if shape > 1 {
			shape = 1
		}

		price := basePrice + (peakPrice-basePrice)*shape
		price *= 0.94 + 0.12*rand.Float64()
		price = math.Round(price*100) / 100

		volume := baseVolume * (0.6 + 0.5*shape)
		volume *= 0.95 + 0.10*rand.Float64()
		volume = math.Round(volume*100) / 100

		// 30 分钟时段：MWh = MW × 0.5h
		mwh := volume * 0.5
		fee := math.Round(mwh*price*100) / 100

		details[p] = periodDetail{Period: p + 1, VolumeMWh: mwh, Price: price, Fee: fee}
		totalFee += fee
		totalVolumeMWh += mwh
	}

	avgPrice := totalFee / totalVolumeMWh
	contractFee := totalFee * 0.60 // 简化拆分：合同 60% / 日前 30% / 实时 10%
	dayAheadFee := totalFee * 0.30
	realTimeFee := totalFee * 0.10
	devRecovery := totalFee * 0.005

	pd, _ := json.Marshal(details)
	round2 := func(x float64) *float64 { v := math.Round(x*100) / 100; return &v }
	round4 := func(x float64) *float64 { v := math.Round(x*10000) / 10000; return &v }

	return &db.SettlementDaily{
		OperatingDate:        d,
		Version:              "PRELIMINARY",
		PeriodDetails:        pd,
		ContractFee:          round2(contractFee),
		DayAheadFee:          round2(dayAheadFee),
		RealTimeFee:          round2(realTimeFee),
		TotalEnergyFee:       round2(totalFee),
		EnergyAvgPrice:       round4(avgPrice),
		DeviationRecoveryFee: round2(devRecovery),
	}
}
