// 价格趋势与电价汇总 handler：/price/trend/price-trend 与 /retail/price-daily/daily-summary。
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

type PriceTrendHandler struct{ repo *db.PriceTrendRepository }

func NewPriceTrendHandler(repo *db.PriceTrendRepository) *PriceTrendHandler {
	return &PriceTrendHandler{repo: repo}
}

// PriceTrend 现货-合同价差趋势。
func (h *PriceTrendHandler) PriceTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	resp, err := h.repo.PriceTrend(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DailySummary 合同电价日汇总。
func (h *PriceTrendHandler) DailySummary(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	resp, err := h.repo.DailySummary(c.Request.Context(), date)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
