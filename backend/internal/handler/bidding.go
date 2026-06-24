// 竞价管理。
// 2026-06 自 new_modules2.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// ─── 竞价管理 ───

type BiddingHandler struct{ repo *db.BiddingRepository }

func NewBiddingHandler(repo *db.BiddingRepository) *BiddingHandler {
	return &BiddingHandler{repo: repo}
}

func (h *BiddingHandler) List(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	list, err := h.repo.List(c.Request.Context(), days)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

type createBiddingReq struct {
	TradeDate      string  `json:"trade_date" binding:"required"`
	BiddingSession string  `json:"bidding_session" binding:"required"`
	DeclaredMW     float64 `json:"declared_mw"`
	DeclaredPrice  float64 `json:"declared_price"`
	Strategy       string  `json:"strategy"`
	Note           string  `json:"note"`
}

func (h *BiddingHandler) Create(c *gin.Context) {
	var req createBiddingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	d, err := parseDate(req.TradeDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误，应为 YYYY-MM-DD"})
		return
	}
	id, err := h.repo.Create(c.Request.Context(), db.BiddingInput{
		TradeDate:      d,
		BiddingSession: req.BiddingSession,
		DeclaredMW:     req.DeclaredMW,
		DeclaredPrice:  req.DeclaredPrice,
		Strategy:       req.Strategy,
		Note:           req.Note,
	})
	if err != nil {
		if errors.Is(err, db.ErrOrgRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先选择具体省份"})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *BiddingHandler) GenerateDemoData(c *gin.Context) {
	n, err := h.repo.GenerateDemo(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"records": n, "message": "已生成竞价演示数据"})
}
