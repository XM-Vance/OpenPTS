// 日前模拟 handler：创建场景、列表、详情、运行模拟、删除。
package handler

import (
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ─── 日前模拟 ───

type DASimulationHandler struct {
	repo *db.DASimulationRepository
}

func NewDASimulationHandler(repo *db.DASimulationRepository) *DASimulationHandler {
	return &DASimulationHandler{repo: repo}
}

// List 列出模拟场景
func (h *DASimulationHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := h.repo.ListScenarios(c.Request.Context(), c.Query("status"), limit)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Get 获取场景详情（含时段结果）
func (h *DASimulationHandler) Get(c *gin.Context) {
	id := c.Param("id")
	scenario, err := h.repo.GetScenario(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "场景不存在"})
		return
	}
	results, err := h.repo.GetPeriodResults(c.Request.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"scenario": scenario, "period_results": results})
}

type createScenarioReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	SimDate     string `json:"sim_date" binding:"required"`
}

// Create 创建模拟场景（草稿）
func (h *DASimulationHandler) Create(c *gin.Context) {
	var req createScenarioReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	simDate, err := parseDate(req.SimDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误，应为 YYYY-MM-DD"})
		return
	}
	userID := ""
	if uid, ok := c.Get("user_id"); ok {
		if s, _ := uid.(string); s != "" {
			userID = s
		}
	}
	id, err := h.repo.CreateScenario(c.Request.Context(), req.Name, req.Description, simDate, userID)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

type runSimulationReq struct {
	Volumes []float64 `json:"volumes" binding:"required"`
}

// RunSimulation 运行模拟计算
// 本骨架使用本地模拟算法生成结果；二次开发可在此接入更精细的出清/撮合模型。
func (h *DASimulationHandler) RunSimulation(c *gin.Context) {
	id := c.Param("id")
	scenario, err := h.repo.GetScenario(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "场景不存在"})
		return
	}

	var req runSimulationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 构建 96 时段申报量
	volumes := req.Volumes
	if len(volumes) == 0 {
		volumes = make([]float64, 96)
		for i := range volumes {
			volumes[i] = 20
		}
	}
	// 若用户只提供 24 个点，插值到 96
	if len(volumes) == 24 {
		expanded := make([]float64, 96)
		for i := 0; i < 24; i++ {
			expanded[i*4] = volumes[i]
			expanded[i*4+1] = volumes[i]
			expanded[i*4+2] = volumes[i]
			expanded[i*4+3] = volumes[i]
		}
		volumes = expanded
	}
	// 截断到 96
	if len(volumes) > 96 {
		volumes = volumes[:96]
	}

	// 模拟计算：基于典型日内价格曲线生成模拟价格
	results := make([]*db.DASimulationPeriodResult, 0, 96)
	var totalVol, totalCost float64
	for i, vol := range volumes {
		period := i + 1
		hour := float64(period-1) / 4.0

		// 典型日价格曲线：谷→平→峰→平→谷
		price := 300.0
		price += 100 * math.Sin((hour-6)*math.Pi/12)  // 基础形状
		price += 50 * math.Sin((hour-14)*math.Pi/4)    // 午高峰
		price += (rand.Float64() - 0.5) * 40            // 随机波动
		if price < 0 {
			price = 50
		}

		cost := vol * price

		// 若场景日期已过，模拟实际价格
		var spotPrice *float64
		var settlement float64
		if scenario.SimDate.Before(time.Now().Truncate(24 * time.Hour)) {
			sp := price * (0.85 + rand.Float64()*0.3)
			spotPrice = &sp
			settlement = vol * (sp - price)
		}

		results = append(results, &db.DASimulationPeriodResult{
			ScenarioID:        id,
			Period:            period,
			DeclaredVolumeMWh: vol,
			SimulatedPrice:    math.Round(price*100) / 100,
			SimulatedCost:     math.Round(cost*100) / 100,
			SpotActualPrice:   spotPrice,
			SettlementAmount:  math.Round(settlement*100) / 100,
		})
		totalVol += vol
		totalCost += cost
	}

	avgPrice := 0.0
	if totalVol > 0 {
		avgPrice = totalCost / totalVol
	}
	profit := 0.0
	for _, r := range results {
		profit += r.SettlementAmount
	}

	if err := h.repo.SaveSimulationResults(c.Request.Context(), id, results, avgPrice, totalCost, profit); err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "模拟计算完成",
		"period_count": len(results),
		"total_volume_mwh": math.Round(totalVol*100) / 100,
		"avg_price":        math.Round(avgPrice*100) / 100,
		"total_cost":       math.Round(totalCost*100) / 100,
		"profit":           math.Round(profit*100) / 100,
	})
}

// Delete 删除场景
func (h *DASimulationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.DeleteScenario(c.Request.Context(), id); err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// GenerateDemoData 生成演示数据
func (h *DASimulationHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"scenarios": n, "message": "已生成日前模拟演示数据"})
}
