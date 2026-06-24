// 保函管理 handler：CRUD。
package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BondHandler struct {
	repo *db.BondRepository
}

func NewBondHandler(repo *db.BondRepository) *BondHandler {
	return &BondHandler{repo: repo}
}

// List GET /api/v1/bonds?keyword=&status=&limit=&offset=
func (h *BondHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, total, err := h.repo.List(c.Request.Context(), db.BondListFilter{
		Keyword: c.Query("keyword"),
		Status:  c.Query("status"),
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// Get GET /api/v1/bonds/:id
func (h *BondHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	bond, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "保函不存在"})
		return
	}
	c.JSON(http.StatusOK, bond)
}

type BondRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=200"`
	BondType    string  `json:"bond_type"`
	Amount      float64 `json:"amount"`
	Issuer      string  `json:"issuer"`
	Beneficiary string  `json:"beneficiary"`
	IssueDate   string  `json:"issue_date"`
	ExpireDate  string  `json:"expire_date"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
}

func (req BondRequest) toInput() db.BondInput {
	status := req.Status
	if status == "" {
		status = "active"
	}
	var issueDate, expireDate *time.Time
	if req.IssueDate != "" {
		if t, err := time.Parse("2006-01-02", req.IssueDate); err == nil {
			issueDate = &t
		}
	}
	if req.ExpireDate != "" {
		if t, err := time.Parse("2006-01-02", req.ExpireDate); err == nil {
			expireDate = &t
		}
	}
	return db.BondInput{
		Name:        req.Name,
		BondType:    req.BondType,
		Amount:      req.Amount,
		Issuer:      req.Issuer,
		Beneficiary: req.Beneficiary,
		IssueDate:   issueDate,
		ExpireDate:  expireDate,
		Status:      status,
		Description: req.Description,
	}
}

// Create POST /api/v1/bonds
func (h *BondHandler) Create(c *gin.Context) {
	var req BondRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var createdBy *uuid.UUID
	if claimsAny, ok := c.Get(auth.ClaimsContextKey); ok {
		uid := claimsAny.(*auth.Claims).UserID
		createdBy = &uid
	}
	bond, err := h.repo.Create(c.Request.Context(), req.toInput(), createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, bond)
}

// Update PUT /api/v1/bonds/:id
func (h *BondHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req BondRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	bond, err := h.repo.Update(c.Request.Context(), id, req.toInput())
	if err != nil {
		if errors.Is(err, db.ErrBondNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "保函不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, bond)
}

// Delete DELETE /api/v1/bonds/:id
func (h *BondHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, db.ErrBondNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "保函不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
