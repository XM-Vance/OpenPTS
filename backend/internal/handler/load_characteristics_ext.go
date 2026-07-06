// 负荷特性分析端点 handler：支撑前端 /load/characteristics/* 的 12 个分析子端点。
// 与 load_characteristics.go（基础 List，读 load_characteristics 表）并存：
// 本 handler 处理 overview/customers/customer 系列深度分析，读 customer_characteristics 表。
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

type LoadCharExtHandler struct{ repo *db.LoadCharacteristicsExtRepository }

func NewLoadCharExtHandler(repo *db.LoadCharacteristicsExtRepository) *LoadCharExtHandler {
	return &LoadCharExtHandler{repo: repo}
}

func (h *LoadCharExtHandler) Overview(c *gin.Context) {
	o, err := h.repo.Overview(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *LoadCharExtHandler) Distribution(c *gin.Context) {
	d, err := h.repo.TagDistributionGrouped(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, d) // 返回 {categories:[...]}，对齐前端 EnhancedTagDistribution
}

func (h *LoadCharExtHandler) TagChanges(c *gin.Context) {
	date := c.Query("date")
	list, err := h.repo.TagChanges(c.Request.Context(), date)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"changes": list})
}

func (h *LoadCharExtHandler) ScatterData(c *gin.Context) {
	list, err := h.repo.ScatterData(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list}) // 对齐前端 ScatterDataResponse {items}
}

func (h *LoadCharExtHandler) ListCustomers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	tag := c.Query("tag")
	sortBy := c.DefaultQuery("sort_by", "avg_daily_load")
	order := c.DefaultQuery("order", "desc")
	list, total, err := h.repo.ListCustomers(c.Request.Context(), page, pageSize, search, tag, sortBy, order)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list, "total": total})
}

func (h *LoadCharExtHandler) GetCustomer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户 ID"})
		return
	}
	ch, err := h.repo.GetCustomer(c.Request.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该客户的负荷特性"})
		return
	}
	c.JSON(http.StatusOK, ch)
}

func (h *LoadCharExtHandler) CustomerHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户 ID"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	month := c.Query("month")
	list, err := h.repo.ListHistory(c.Request.Context(), id, limit, month)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *LoadCharExtHandler) CustomerAlerts(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户 ID"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := h.repo.ListCustomerAlerts(c.Request.Context(), id, limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *LoadCharExtHandler) AckAlert(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的告警 ID"})
		return
	}
	var req struct {
		Acknowledged bool   `json:"acknowledged"`
		Notes        string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&req)
	uid := currentUserID(c) // *uuid.UUID（见 retail.go）
	var userID uuid.UUID
	if uid != nil {
		userID = *uid
	}
	if err := h.repo.AckAlertWithNote(c.Request.Context(), id, userID, req.Acknowledged, req.Notes); err != nil {
		if err == db.ErrAlertNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "告警不存在"})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *LoadCharExtHandler) DailyTrend(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户 ID"})
		return
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 start_date 或 end_date"})
		return
	}
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date 格式错误"})
		return
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date 格式错误"})
		return
	}
	list, err := h.repo.DailyTrend(c.Request.Context(), id, start, end)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *LoadCharExtHandler) MonthlyEnergy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户 ID"})
		return
	}
	startMonth := c.Query("start_month")
	endMonth := c.Query("end_month")
	if startMonth == "" || endMonth == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 start_month 或 end_month"})
		return
	}
	list, err := h.repo.MonthlyEnergySeries(c.Request.Context(), id, startMonth, endMonth)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}
