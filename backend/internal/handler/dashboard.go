// 仪表盘 handler：跨模块 KPI 总览 + 时间序列（任何已登录用户可访问）。
// 套了 TTL 内存缓存：summary 30s、series 5min。命中率走 Prometheus 指标。
package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ptis/backend/internal/cache"
	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type DashboardHandler struct {
	repo  *db.DashboardRepository
	cache *cache.TTLCache
}

func NewDashboardHandler(repo *db.DashboardRepository) *DashboardHandler {
	return &DashboardHandler{repo: repo, cache: cache.New("dashboard")}
}

// Summary GET /api/v1/dashboard/summary
func (h *DashboardHandler) Summary(c *gin.Context) {
	s, err := cache.GetOrLoad(h.cache, "summary", 30*time.Second,
		func() (*db.DashboardSummary, error) {
			return h.repo.GetSummary(c.Request.Context())
		})
	if err != nil {
		log.Error().Err(err).Msg("查询失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// SettlementSeries GET /api/v1/dashboard/series/settlement?days=14
func (h *DashboardHandler) SettlementSeries(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	if days <= 0 || days > 90 {
		days = 14
	}
	key := "series:settlement:" + strconv.Itoa(days)
	list, err := cache.GetOrLoad(h.cache, key, 5*time.Minute,
		func() ([]*db.DailySeriesPoint, error) {
			return h.repo.GetSettlementSeries(c.Request.Context(), days)
		})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// FreqSeries GET /api/v1/dashboard/series/freq?days=14
func (h *DashboardHandler) FreqSeries(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	if days <= 0 || days > 90 {
		days = 14
	}
	key := "series:freq:" + strconv.Itoa(days)
	list, err := cache.GetOrLoad(h.cache, key, 5*time.Minute,
		func() ([]*db.DailySeriesPoint, error) {
			return h.repo.GetFreqSeries(c.Request.Context(), days)
		})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// SettlementSummary GET /api/v1/dashboard/settlement-summary
func (h *DashboardHandler) SettlementSummary(c *gin.Context) {
	s, err := cache.GetOrLoad(h.cache, "settlement-summary", 5*time.Minute,
		func() (*db.SettlementSummaryResult, error) {
			return h.repo.GetSettlementSummary(c.Request.Context())
		})
	if err != nil {
		log.Error().Err(err).Msg("查询失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// GetConfig GET /api/v1/dashboard/config
// 返回当前用户的仪表盘布局配置（widget 显示/隐藏 + 排列顺序）。
func (h *DashboardHandler) GetConfig(c *gin.Context) {
	uid := claimsUserID(c)
	if uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	config, err := h.repo.GetUserConfig(c.Request.Context(), *uid)
	if err != nil {
		log.Error().Err(err).Msg("读取仪表盘配置失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	if config == nil {
		c.JSON(http.StatusOK, gin.H{"widgets": nil})
		return
	}
	c.JSON(http.StatusOK, json.RawMessage(config))
}

// SaveConfig PUT /api/v1/dashboard/config
// 保存当前用户的仪表盘布局配置。
func (h *DashboardHandler) SaveConfig(c *gin.Context) {
	uid := claimsUserID(c)
	if uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体读取失败"})
		return
	}
	if !json.Valid(body) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 JSON"})
		return
	}
	if len(body) > 65536 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置过大"})
		return
	}
	if err := h.repo.UpsertUserConfig(c.Request.Context(), *uid, body); err != nil {
		log.Error().Err(err).Msg("保存仪表盘配置失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
