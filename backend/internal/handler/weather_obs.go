package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// 外部气象观测接口：风电场风速 / 水库水文（原市场行情，现并入气象数据模块）。

// WindFarm GET /api/v1/weather/wind-farm?station=&hours= —— 风电场逐时风速 + 站点列表。
func (h *WeatherHandler) WindFarm(c *gin.Context) {
	station := c.Query("station")
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "72"))
	stations, err := h.repo.WindFarmStations(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("查询风电场站点失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	// 未指定站点时默认取第一个，避免一次拉回所有站点的海量逐时数据。
	if station == "" && len(stations) > 0 {
		station = stations[0].Code
	}
	items, err := h.repo.WindFarmHourly(c.Request.Context(), station, hours)
	if err != nil {
		log.Error().Err(err).Msg("查询风电场风速失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stations": stations, "station": station, "items": items})
}

// Hydrology GET /api/v1/weather/hydrology?station=&days= —— 水库水文逐日 + 站点列表。
func (h *WeatherHandler) Hydrology(c *gin.Context) {
	station := c.Query("station")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	stations, err := h.repo.HydrologyStations(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("查询水文站点失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	if station == "" && len(stations) > 0 {
		station = stations[0].Code
	}
	items, err := h.repo.HydrologyDaily(c.Request.Context(), station, days)
	if err != nil {
		log.Error().Err(err).Msg("查询水库水文失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stations": stations, "station": station, "items": items})
}

// GenObsDemo POST /api/v1/weather/obs-demo-data —— 生成风电场风速 + 水库水文演示观测数据。
func (h *WeatherHandler) GenObsDemo(c *gin.Context) {
	n, err := h.repo.GenerateObsDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("生成气象观测演示数据失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成风电场风速与水库水文演示数据"})
}
