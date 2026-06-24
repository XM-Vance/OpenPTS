// 客户历史电量档案 handler：列出客户逐月电量（按活跃省隔离）。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

type CustomerEnergyHandler struct {
	repo *db.CustomerEnergyRepository
}

func NewCustomerEnergyHandler(repo *db.CustomerEnergyRepository) *CustomerEnergyHandler {
	return &CustomerEnergyHandler{repo: repo}
}

// List GET /api/v1/customer-energy?customer_id=&limit=
func (h *CustomerEnergyHandler) List(c *gin.Context) {
	customerID := c.Query("customer_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	list, err := h.repo.List(c.Request.Context(), customerID, limit)
	if err != nil {
		log.Error().Err(err).Msg("查询客户电量失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}
