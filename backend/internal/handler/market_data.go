package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, overview)
}

// ListTables 列出所有数据表
// GET /api/v1/market-data/tables
func (h *MarketDataHandler) ListTables(c *gin.Context) {
	tables, err := h.repo.ListTables(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tables": tables})
}

// Query 查询指定表数据
// GET /api/v1/market-data/:table?days=30&location_code=PT
func (h *MarketDataHandler) Query(c *gin.Context) {
	tableName := c.Param("table")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	locationCode := c.Query("location_code")

	data, err := h.repo.QueryTable(c.Request.Context(), tableName, days, locationCode)
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
	c.JSON(http.StatusOK, gin.H{
		"table":      tableName,
		"count":      len(data),
		"data":       data,
		"scope":      scope,
		"scope_label": scopeLabel,
	})
}
