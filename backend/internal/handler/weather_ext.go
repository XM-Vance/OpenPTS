// weather_ext.go —— 站点/实况/预报 端点，对齐前端 lib/api/weather.ts 契约。
// 逐小时数据（WeatherHourlyData）库内不存在，返回形状正确的空数组（页面优雅空态）。
package handler

import (
	"errors"
	"net/http"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// 由降水与气温推导天气类型/图标（与前端 getWeatherType 规则一致；无云量数据时默认晴）。
func weatherTypeIcon(precip, temp float64) (string, string) {
	if precip > 0 {
		switch {
		case temp < 0 && precip > 5:
			return "大雪", "❄️"
		case temp < 0:
			return "小雪", "🌨️"
		case temp <= 2:
			return "雨夹雪", "🌨️"
		case precip > 8:
			return "大雨", "🌧️"
		case precip > 2.5:
			return "中雨", "🌧️"
		default:
			return "小雨", "🌦️"
		}
	}
	return "晴", "☀️"
}

// GET /weather/locations
func (h *WeatherHandler) ListLocations(c *gin.Context) {
	list, err := h.repo.ListLocations(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, list) // 前端期望裸数组
}

// POST /weather/locations
func (h *WeatherHandler) CreateLocation(c *gin.Context) {
	var body struct {
		LocationID string  `json:"location_id"`
		Name       string  `json:"name"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := body.Name
	if name == "" {
		name = body.LocationID
	}
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少站点名称"})
		return
	}
	if err := h.repo.CreateLocation(c.Request.Context(), name, body.Latitude, body.Longitude); err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, &db.WeatherLocationRow{
		LocationID: name, Name: name, Latitude: body.Latitude, Longitude: body.Longitude, Enabled: true,
	})
}

// PUT /weather/locations/:id   （:id 即 location_id，取站点名）
func (h *WeatherHandler) UpdateLocation(c *gin.Context) {
	name := c.Param("id")
	var body struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.UpdateLocation(c.Request.Context(), name, body.Latitude, body.Longitude); err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, &db.WeatherLocationRow{
		LocationID: name, Name: name, Latitude: body.Latitude, Longitude: body.Longitude, Enabled: true,
	})
}

// DELETE /weather/locations/:id
func (h *WeatherHandler) DeleteLocation(c *gin.Context) {
	if err := h.repo.DeleteLocation(c.Request.Context(), c.Param("id")); err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /weather/actuals?location_id=&date=   逐小时实况：库内仅日聚合，返回空数组。
func (h *WeatherHandler) Actuals(c *gin.Context) {
	c.JSON(http.StatusOK, []any{})
}

// GET /weather/actuals/summary?location_id=&date=
func (h *WeatherHandler) ActualsSummary(c *gin.Context) {
	loc := c.Query("location_id")
	date := c.Query("date")
	s, err := h.repo.ActualsSummary(c.Request.Context(), loc, date)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusOK, &db.WeatherDailySummary{Date: date}) // 无数据返回空摘要，前端不崩
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	s.WeatherType, s.WeatherIcon = weatherTypeIcon(s.AvgPrecipitation, s.MaxTemp)
	c.JSON(http.StatusOK, s)
}

// GET /weather/forecasts?location_id=&forecast_date=&target_date=   逐小时预报：返回空数组。
func (h *WeatherHandler) Forecasts(c *gin.Context) {
	c.JSON(http.StatusOK, []any{})
}

// GET /weather/forecasts/summary?location_id=&forecast_date=
func (h *WeatherHandler) ForecastsSummary(c *gin.Context) {
	loc := c.Query("location_id")
	fd := c.Query("forecast_date")
	list, err := h.repo.ForecastsSummary(c.Request.Context(), loc, fd)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	for _, s := range list {
		s.WeatherType, s.WeatherIcon = weatherTypeIcon(s.AvgPrecipitation, s.MaxTemp)
	}
	c.JSON(http.StatusOK, list)
}

// GET /weather/forecast-dates?location_id=&target_date=
func (h *WeatherHandler) ForecastDates(c *gin.Context) {
	dates, err := h.repo.ForecastDates(c.Request.Context(), c.Query("location_id"), c.Query("target_date"))
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, dates)
}
