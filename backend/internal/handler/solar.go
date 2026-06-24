// 光伏管理 handler：站点 CRUD + 发电预测 + 收益结算 + 演示数据生成。
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

type SolarHandler struct {
	repo *db.SolarRepository
}

func NewSolarHandler(repo *db.SolarRepository) *SolarHandler {
	return &SolarHandler{repo: repo}
}

// ── 站点 CRUD ──

// ListStations GET /api/v1/solar/stations
func (h *SolarHandler) ListStations(c *gin.Context) {
	list, err := h.repo.ListStations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// GetStation GET /api/v1/solar/stations/:id
func (h *SolarHandler) GetStation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	s, err := h.repo.GetStation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "站点不存在"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// CreateStation POST /api/v1/solar/stations
func (h *SolarHandler) CreateStation(c *gin.Context) {
	var req struct {
		StationName   string   `json:"station_name" binding:"required"`
		Location      string   `json:"location"`
		CapacityKW    float64  `json:"capacity_kw" binding:"required"`
		Status        string   `json:"status"`
		InstalledDate *string  `json:"installed_date"`
		Latitude      *float64 `json:"latitude"`
		Longitude     *float64 `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	var installedDate *time.Time
	if req.InstalledDate != nil {
		t, err := time.Parse("2006-01-02", *req.InstalledDate)
		if err == nil {
			installedDate = &t
		}
	}
	s, err := h.repo.CreateStation(c.Request.Context(), req.StationName, req.Location, req.CapacityKW, status, installedDate, req.Latitude, req.Longitude)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("创建失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, s)
}

// UpdateStation PUT /api/v1/solar/stations/:id
func (h *SolarHandler) UpdateStation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req struct {
		StationName   string   `json:"station_name" binding:"required"`
		Location      string   `json:"location"`
		CapacityKW    float64  `json:"capacity_kw" binding:"required"`
		Status        string   `json:"status"`
		InstalledDate *string  `json:"installed_date"`
		Latitude      *float64 `json:"latitude"`
		Longitude     *float64 `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	var installedDate *time.Time
	if req.InstalledDate != nil {
		t, err := time.Parse("2006-01-02", *req.InstalledDate)
		if err == nil {
			installedDate = &t
		}
	}
	s, err := h.repo.UpdateStation(c.Request.Context(), id, req.StationName, req.Location, req.CapacityKW, status, installedDate, req.Latitude, req.Longitude)
	if err != nil {
		log.Error().Err(err).Msg("更新失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// DeleteStation DELETE /api/v1/solar/stations/:id
func (h *SolarHandler) DeleteStation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.DeleteStation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// ── 发电预测 ──

// ListForecast GET /api/v1/solar/forecast?station_id=&limit=
func (h *SolarHandler) ListForecast(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	var stationID *uuid.UUID
	if sid := c.Query("station_id"); sid != "" {
		parsed, err := uuid.Parse(sid)
		if err == nil {
			stationID = &parsed
		}
	}
	items, err := h.repo.ListForecast(c.Request.Context(), stationID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// ── 收益结算 ──

// ListRevenue GET /api/v1/solar/revenue?station_id=&limit=
func (h *SolarHandler) ListRevenue(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	var stationID *uuid.UUID
	if sid := c.Query("station_id"); sid != "" {
		parsed, err := uuid.Parse(sid)
		if err == nil {
			stationID = &parsed
		}
	}
	items, err := h.repo.ListRevenue(c.Request.Context(), stationID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// ── 演示数据 ──

// GenerateDemoData POST /api/v1/solar/demo-data
// 创建 3 个光伏站点，每个站点生成最近 N 天 96 时段预测数据 + 12 个月结算数据。
func (h *SolarHandler) GenerateDemoData(c *gin.Context) {
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
		Name       string
		Location   string
		CapKW      float64
		Lat        float64
		Lng        float64
		InstallDate string
	}
	specs := []spec{
		{"光伏电站A（华东园区）", "浙江杭州", 5000, 30.27, 120.15, "2022-06-15"},
		{"光伏电站B（华南分布式）", "广东深圳", 8000, 22.54, 114.06, "2023-03-01"},
		{"光伏电站C（西北集中式）", "甘肃兰州", 15000, 36.06, 103.83, "2021-09-20"},
	}

	stations := make([]*db.SolarStation, 0, len(specs))
	for _, sp := range specs {
		installDate, _ := time.Parse("2006-01-02", sp.InstallDate)
		s, err := h.repo.CreateStation(ctx, sp.Name, sp.Location, sp.CapKW, "active", &installDate, &sp.Lat, &sp.Lng)
		if err != nil {
			if respondOrgRequired(c, err) {
				return
			}
			// 可能已存在，尝试用 name 查找
			list, listErr := h.repo.ListStations(ctx)
			if listErr != nil {
				log.Error().Err(err).Msg("写入站点失败")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
				return
			}
			found := false
			for _, st := range list {
				if st.StationName == sp.Name {
					stations = append(stations, st)
					found = true
					break
				}
			}
			if !found {
				log.Error().Err(err).Msg("写入站点失败")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
				return
			}
			continue
		}
		stations = append(stations, s)
	}

	today := time.Now().Truncate(24 * time.Hour)

	// 生成预测数据
	for _, s := range stations {
		for i := days; i >= 1; i-- {
			d := today.AddDate(0, 0, -i)
			for period := 1; period <= 96; period++ {
				// 模拟光伏出力曲线：白天（period ~25-70）有出力，夜间为 0
				hour := float64(period) / 4.0
				var forecastKW float64
				if hour >= 6 && hour <= 18 {
					// 钟形曲线
					center := 12.0
					sigma := 3.0
					amplitude := s.CapacityKW * (0.6 + 0.3*rand.Float64())
					forecastKW = amplitude * math.Exp(-math.Pow(hour-center, 2)/(2*sigma*sigma))
					forecastKW *= (0.9 + 0.2*rand.Float64()) // 随机波动
					if forecastKW < 0 {
						forecastKW = 0
					}
				}
				forecastKW = r2(forecastKW)
				actualKW := r2(forecastKW * (0.85 + 0.3*rand.Float64()))
				var deviation *float64
				if forecastKW > 0 {
					d := r2((actualKW - forecastKW) / forecastKW * 100)
					deviation = &d
				}
				fPtr := forecastKW
				aPtr := actualKW
				if err := h.repo.UpsertForecast(ctx, s.ID, d, period, &fPtr, &aPtr, deviation); err != nil {
					if respondOrgRequired(c, err) {
						return
					}
					log.Error().Err(err).Msg("写入预测数据失败")
					c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
					return
				}
			}
		}
	}

	// 生成收益结算数据（最近 12 个月）
	for _, s := range stations {
		for m := 11; m >= 0; m-- {
			month := today.AddDate(0, -m, 0)
			monthStr := month.Format("2006-01")
			// 发电量：容量 × 日照小时 × 天数 × 系数
			daysInMonth := 30
			energyKWh := r2(s.CapacityKW * (3 + 2*rand.Float64()) * float64(daysInMonth))
			avgPrice := r2(0.35 + 0.15*rand.Float64())
			revenue := r2(energyKWh * avgPrice)
			subsidy := r2(energyKWh * 0.05)
			netIncome := r2(revenue + subsidy)
			if err := h.repo.UpsertRevenue(ctx, s.ID, monthStr, energyKWh, revenue, avgPrice, subsidy, netIncome); err != nil {
				if respondOrgRequired(c, err) {
					return
				}
				log.Error().Err(err).Msg("写入结算数据失败")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "已生成演示光伏数据",
		"stations": len(stations),
		"days":     days,
	})
}
