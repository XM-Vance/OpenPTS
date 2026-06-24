package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// CarbonHandler 碳交易行情接口（CEA/CCER/EUA）。
type CarbonHandler struct{ repo *db.CarbonRepository }

func NewCarbonHandler(repo *db.CarbonRepository) *CarbonHandler {
	return &CarbonHandler{repo: repo}
}

// Summary GET /api/v1/carbon/summary —— 三个碳产品的最新行情概览。
func (h *CarbonHandler) Summary(c *gin.Context) {
	list, err := h.repo.Summary(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("查询碳行情概览失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// List GET /api/v1/carbon/quotes?product=&days= —— 行情明细。
func (h *CarbonHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "180"))
	product := c.Query("product")
	list, err := h.repo.List(c.Request.Context(), product, days)
	if err != nil {
		log.Error().Err(err).Msg("查询碳行情失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// GenerateDemoData POST /api/v1/carbon/demo-data —— 生成 CEA/CCER/EUA 演示行情。
func (h *CarbonHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("生成碳行情演示数据失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": n, "message": "已生成碳交易演示行情（CEA/CCER/EUA）"})
}
