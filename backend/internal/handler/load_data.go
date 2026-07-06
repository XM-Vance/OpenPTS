// 负荷数据查询端点 handler：支撑前端 /load-data/* 的查询端点。
// 与 load_characteristics_ext.go（客户特征画像）不同——本 handler 是计量数据管理。
// 校准（calibration）/ reaggregate 需算法，暂不实现（契约闸白名单跟踪）。
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

type LoadDataHandler struct{ repo *db.LoadDataRepository }

func NewLoadDataHandler(repo *db.LoadDataRepository) *LoadDataHandler {
	return &LoadDataHandler{repo: repo}
}

func (h *LoadDataHandler) ListCustomers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	list, total, err := h.repo.ListCustomers(c.Request.Context(), page, pageSize, search)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list, "total": total})
}

func (h *LoadDataHandler) SignedCustomers(c *gin.Context) {
	list, err := h.repo.ListSignedCustomers(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *LoadDataHandler) CustomerDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户 ID"})
		return
	}
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	d, err := h.repo.GetCustomerDetail(c.Request.Context(), id, days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该客户的负荷数据"})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *LoadDataHandler) CustomerCalendar(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户 ID"})
		return
	}
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))
	list, err := h.repo.CustomerCalendar(c.Request.Context(), id, month)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *LoadDataHandler) CustomerCurves(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户 ID"})
		return
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		// 默认近 7 天
		endDate = time.Now().Format("2006-01-02")
		startDate = time.Now().AddDate(0, 0, -7).Format("2006-01-02")
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
	list, err := h.repo.CustomerCurves(c.Request.Context(), id, start, end)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *LoadDataHandler) ExportMpMissing(c *gin.Context) {
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))
	list, err := h.repo.ExportMpMissing(c.Request.Context(), month)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list, "total": len(list)})
}
