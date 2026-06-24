// forecast_base_ext.go —— 对齐前端 lib/api/forecast-base-data.ts 契约。
// 前端期望「可用性矩阵 + 多曲线」结构，后端现有 holidays/typical_curves 不含该模型，
// 故返回形状正确的空结构（页面优雅空态，不再 404；后续接入真实基础数据后填充）。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /forecast-base-data/availability → { date_range: [], availability_matrix: [] }
func (h *ForecastBaseHandler) Availability(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"date_range":          []string{},
		"availability_matrix": []any{},
	})
}

// POST /forecast-base-data/curves → { curves: [] }
func (h *ForecastBaseHandler) Curves(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"curves": []any{}})
}
