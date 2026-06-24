// 客户分析 handler：异动告警 + 客户特征 + 演示数据生成（自动补足客户数）。
package handler

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type AnalyticsHandler struct {
	repo     *db.AnalyticsRepository
	custRepo *db.CustomerRepository
}

func NewAnalyticsHandler(repo *db.AnalyticsRepository, custRepo *db.CustomerRepository) *AnalyticsHandler {
	return &AnalyticsHandler{repo: repo, custRepo: custRepo}
}

// Stats GET /api/v1/analytics/alerts/stats
func (h *AnalyticsHandler) Stats(c *gin.Context) {
	s, err := h.repo.GetAlertStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// ListAlerts GET /api/v1/analytics/alerts?limit=&include_acked=true
func (h *AnalyticsHandler) ListAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	includeAcked := c.Query("include_acked") == "true"
	list, err := h.repo.ListAlerts(c.Request.Context(), limit, includeAcked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// AckAlert POST /api/v1/analytics/alerts/:id/ack
func (h *AnalyticsHandler) AckAlert(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	uid := currentUserID(c) // 复用 retail.go 中的辅助
	if uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	if err := h.repo.AckAlert(c.Request.Context(), id, *uid); err != nil {
		if errors.Is(err, db.ErrAlertNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "告警不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "确认失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已确认"})
}

// ListCharacteristics GET /api/v1/analytics/characteristics?limit=
func (h *AnalyticsHandler) ListCharacteristics(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	list, err := h.repo.ListLatestCharacteristics(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// GenerateDemoData POST /api/v1/analytics/demo-data
// 若客户数 < 3 则自动补足；为所有客户写最新 characteristic；随机生成 12 条告警。
func (h *AnalyticsHandler) GenerateDemoData(c *gin.Context) {
	ctx := c.Request.Context()

	customers, _, err := h.custRepo.List(ctx, db.CustomerListFilter{Limit: 50})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取客户失败"})
		return
	}

	// 若客户不足 3 个，补两个演示客户（按名字去重）
	if len(customers) < 3 {
		extras := []struct {
			Name, Manager, Location string
			Tags                    []string
		}{
			{"分析示范-钢铁厂", "分析师 A", "上海宝山", []string{"工业", "高峰"}},
			{"分析示范-数据中心", "分析师 B", "广州天河", []string{"稳定型", "新型负荷"}},
		}
		for _, e := range extras {
			existing, _, _ := h.custRepo.List(ctx, db.CustomerListFilter{Keyword: e.Name, Limit: 1})
			if len(existing) > 0 {
				continue
			}
			_, _ = h.custRepo.Create(ctx, db.CustomerInput{
				UserName: e.Name,
				Manager:  e.Manager,
				Location: e.Location,
				Tags:     e.Tags,
			}, nil)
		}
		customers, _, _ = h.custRepo.List(ctx, db.CustomerListFilter{Limit: 50})
	}

	if len(customers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未能创建演示客户"})
		return
	}

	today := time.Now().Truncate(24 * time.Hour)

	tagPool := []string{"工业", "高峰用户", "稳定型", "波动型", "光伏自发", "新型负荷", "重点客户"}
	qualities := []string{"A", "B", "C"}

	// 1. 每客户写一条最新 characteristic
	chars := 0
	for _, cust := range customers {
		shuffled := append([]string(nil), tagPool...)
		rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		n := 2 + rand.IntN(2)
		tags := append([]string(nil), shuffled[:n]...)

		regularity := r4a(0.55 + 0.4*rand.Float64())
		quality := qualities[rand.IntN(len(qualities))]
		longTerm := json.RawMessage(`{"trend":"steady","yoy_growth":` + fjsona(rand.Float64()*0.2) + `}`)
		shortTerm := json.RawMessage(`{"vol_7d":` + fjsona(0.05+rand.Float64()*0.1) + `,"peak_shift_min":` + fjsona(rand.Float64()*30) + `}`)

		if err := h.repo.UpsertCharacteristic(ctx, cust.ID, today,
			longTerm, shortTerm, tags, regularity, quality); err != nil {
			log.Error().Err(err).Msg("写入特征失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}
		chars++
	}

	// 2. 12 条随机告警
	alertTypes := []string{"load_drop", "shape_change", "quality_drop", "spike"}
	severities := []string{"info", "warn", "critical"}
	alerts := 0
	for j := 0; j < 12; j++ {
		cust := customers[rand.IntN(len(customers))]
		d := today.AddDate(0, 0, -1-rand.IntN(20))
		atype := alertTypes[rand.IntN(len(alertTypes))]
		sev := severities[rand.IntN(len(severities))]
		conf := r2a(60 + 35*rand.Float64())
		if err := h.repo.InsertAlert(ctx, cust.ID, d, atype, sev,
			"rule_"+atype, alertReason(atype), conf); err != nil {
			log.Error().Err(err).Msg("写入告警失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}
		alerts++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "已生成演示分析数据",
		"customers":       len(customers),
		"characteristics": chars,
		"alerts":          alerts,
	})
}

func alertReason(atype string) string {
	switch atype {
	case "load_drop":
		return "负荷较基线显著下降"
	case "shape_change":
		return "典型日曲线形态偏移"
	case "quality_drop":
		return "数据完整性下降"
	case "spike":
		return "出现单点尖峰"
	}
	return "未分类异常"
}

func r2a(v float64) float64 { return math.Round(v*100) / 100 }
func r4a(v float64) float64 { return math.Round(v*10000) / 10000 }
func fjsona(v float64) string {
	return strconv.FormatFloat(math.Round(v*10000)/10000, 'f', -1, 64)
}
