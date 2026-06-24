// 储能管理 handler：站点列表 + 日运营记录 + 演示数据生成（2 站点 × N 天）。
package handler

import (
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type StorageHandler struct {
	repo *db.StorageRepository
}

func NewStorageHandler(repo *db.StorageRepository) *StorageHandler {
	return &StorageHandler{repo: repo}
}

// ListStations GET /api/v1/storage/stations
func (h *StorageHandler) ListStations(c *gin.Context) {
	list, err := h.repo.ListStations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// ListOperations GET /api/v1/storage/stations/:id/operations?limit=
func (h *StorageHandler) ListOperations(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	if limit <= 0 || limit > 200 {
		limit = 30
	}
	list, err := h.repo.ListOperations(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// GenerateDemoData POST /api/v1/storage/demo-data
// 创建/更新 2 个演示站点，并为每个站点生成最近 N 天运营记录。
func (h *StorageHandler) GenerateDemoData(c *gin.Context) {
	var req struct {
		Days int `json:"days"`
	}
	_ = c.ShouldBindJSON(&req)
	days := req.Days
	if days <= 0 || days > 90 {
		days = 30
	}
	ctx := c.Request.Context()

	type spec struct {
		Name        string
		CapacityMWh float64
		MaxPowerMW  float64
		Location    string
	}
	specs := []spec{
		{"储能站A（华东园区）", 50, 25, "浙江绍兴"},
		{"储能站B（华南分布式）", 100, 50, "广东东莞"},
	}

	stations := make([]*db.StorageStation, 0, len(specs))
	for _, sp := range specs {
		s, err := h.repo.UpsertStation(ctx, sp.Name, sp.CapacityMWh, sp.MaxPowerMW, sp.Location, "active")
		if err != nil {
			if respondOrgRequired(c, err) {
				return
			}
			log.Error().Err(err).Msg("写入站点失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}
		stations = append(stations, s)
	}

	today := time.Now().Truncate(24 * time.Hour)
	for _, s := range stations {
		for i := days; i >= 1; i-- {
			d := today.AddDate(0, 0, -i)
			// 充电：~80% 容量（谷价时段填充）
			charge := s.CapacityMWh * (0.75 + 0.15*rand.Float64())
			// 放电 / 充电 比 ~88%（往返效率）
			discharge := charge * (0.85 + 0.06*rand.Float64())
			// 套利收益：放电 × 峰价 - 充电 × 谷价
			peakPrice := 650 + 100*rand.Float64()
			offPeakPrice := 150 + 60*rand.Float64()
			revenue := discharge*peakPrice - charge*offPeakPrice
			soc := 35 + 35*rand.Float64()
			cycles := charge / s.CapacityMWh

			if err := h.repo.UpsertOperation(ctx, s.ID, d,
				r2st(charge), r2st(discharge), r2st(revenue), r2st(soc), r2st(cycles)); err != nil {
				if respondOrgRequired(c, err) {
					return
				}
				log.Error().Err(err).Msg("写入运营失败")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "已生成演示储能数据",
		"stations": len(stations),
		"days":     days,
	})
}

func r2st(v float64) float64 { return math.Round(v*100) / 100 }
