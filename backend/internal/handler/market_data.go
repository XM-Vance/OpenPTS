package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

type MarketDataHandler struct {
	repo *db.MarketDataRepository
}

func NewMarketDataHandler(repo *db.MarketDataRepository) *MarketDataHandler {
	return &MarketDataHandler{repo: repo}
}

// Overview 市场数据总览
// GET /api/v1/market-data/overview
func (h *MarketDataHandler) Overview(c *gin.Context) {
	overview, err := h.repo.Overview(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("market_data overview 查询失败")
		respondInternalErr(c, "市场数据总览查询失败")
		return
	}
	c.JSON(http.StatusOK, overview)
}

// ListTables 列出所有数据表
// GET /api/v1/market-data/tables
func (h *MarketDataHandler) ListTables(c *gin.Context) {
	tables, err := h.repo.ListTables(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("market_data ListTables 查询失败")
		respondInternalErr(c, "数据表列表查询失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"tables": tables})
}

// Query 查询指定表数据
// GET /api/v1/market-data/:table?days=30&location_code=PT&page=1&page_size=500
func (h *MarketDataHandler) Query(c *gin.Context) {
	tableName := c.Param("table")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	locationCode := c.Query("location_code")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "0")) // 0 → repo 默认 500

	data, total, err := h.repo.QueryTable(c.Request.Context(), tableName, days, locationCode, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 判断数据范围
	scope := "national"
	if meta, ok := h.repo.GetTableMeta(tableName); ok && meta.Scope != "" {
		scope = meta.Scope
	}
	scopeLabel := "全国数据"
	if scope == "provincial" {
		scopeLabel = "分省数据"
	}
	// 分页信息：page/page_size/total/has_more，前端据此决定是否加载下一页
	effectivePageSize := pageSize
	if effectivePageSize <= 0 || effectivePageSize > 5000 {
		effectivePageSize = 5000
	}
	c.JSON(http.StatusOK, gin.H{
		"table":       tableName,
		"count":       len(data),
		"data":        data,
		"scope":       scope,
		"scope_label": scopeLabel,
		"page":        page,
		"page_size":   effectivePageSize,
		"total":       total,
		"has_more":    page*effectivePageSize < total,
	})
}
