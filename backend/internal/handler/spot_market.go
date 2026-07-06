// 现货市场。
// 2026-06 自 new_modules.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── 现货市场 ───

type SpotMarketHandler struct{ repo *db.SpotMarketRepository }

func NewSpotMarketHandler(repo *db.SpotMarketRepository) *SpotMarketHandler {
	return &SpotMarketHandler{repo: repo}
}

func (h *SpotMarketHandler) List(c *gin.Context) {
	start, end := parseDateRange(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	list, err := h.repo.ListByRange(c.Request.Context(), start, end, limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list}) // 对齐前端 da_avg_price/rt_avg_price/max_price 等
}

// Statistics 现货市场统计汇总（对齐前端 max_price/min_price/avg_price/da_avg_price 等）。
func (h *SpotMarketHandler) Statistics(c *gin.Context) {
	start, end := parseDateRange(c)
	s, err := h.repo.StatisticsByRange(c.Request.Context(), start, end)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// PriceCurve 现货价格曲线（对齐前端 day_ahead_prices/realtime_prices 数组）。
func (h *SpotMarketHandler) PriceCurve(c *gin.Context) {
	start, end := parseDateRange(c)
	curve, err := h.repo.PriceCurveSeries(c.Request.Context(), start, end)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, curve)
}

// parseDateRange 解析 start_date/end_date，缺省取近 30 天。
func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	end := time.Now()
	start := end.AddDate(0, 0, -30)
	if s := c.Query("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start = t
		}
	}
	if e := c.Query("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end = t
		}
	}
	return start, end
}

func (h *SpotMarketHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"days": n, "message": "已生成现货市场演示数据"})
}
