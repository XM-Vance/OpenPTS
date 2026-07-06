// 绿电交易。
// 2026-06 自 new_modules.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// ─── 绿电交易 ───

type GreenPowerHandler struct{ repo *db.GreenPowerRepository }

func NewGreenPowerHandler(repo *db.GreenPowerRepository) *GreenPowerHandler {
	return &GreenPowerHandler{repo: repo}
}

func (h *GreenPowerHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	list, err := h.repo.List(c.Request.Context(), c.Query("status"), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// GreenPowerTradeRequest 创建请求体。
type GreenPowerTradeRequest struct {
	TradeDate      string          `json:"trade_date" binding:"required"` // YYYY-MM-DD
	ProductName    string          `json:"product_name" binding:"required,min=1,max=50"`
	EnergyMWh      float64         `json:"energy_mwh"`
	Price          decimal.Decimal `json:"price"`  // numeric(18,4)
	Amount         decimal.Decimal `json:"amount"` // numeric(18,4)
	GreenCertCount int             `json:"green_cert_count"`
	Status         string          `json:"status"`
	Counterparty   string          `json:"counterparty"`
}

func (req GreenPowerTradeRequest) toInput() (db.GreenPowerTradeInput, error) {
	d, err := time.Parse("2006-01-02", req.TradeDate)
	if err != nil {
		return db.GreenPowerTradeInput{}, errors.New("trade_date 格式应为 YYYY-MM-DD")
	}
	return db.GreenPowerTradeInput{
		TradeDate:      d,
		ProductName:    req.ProductName,
		EnergyMWh:      req.EnergyMWh,
		Price:          req.Price,
		Amount:         req.Amount,
		GreenCertCount: req.GreenCertCount,
		Status:         req.Status,
		Counterparty:   req.Counterparty,
	}, nil
}

func (h *GreenPowerHandler) Create(c *gin.Context) {
	var req GreenPowerTradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	in, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	trade, err := h.repo.Create(c.Request.Context(), in)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, trade)
}

func (h *GreenPowerHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, db.ErrGreenPowerTradeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "绿电交易记录不存在"})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

func (h *GreenPowerHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"trades": n, "message": "已生成绿电交易演示数据"})
}
