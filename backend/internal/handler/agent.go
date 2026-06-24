// 代理商管理 handler：CRUD + 关联客户列表。
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AgentHandler struct {
	repo *db.AgentRepository
}

func NewAgentHandler(repo *db.AgentRepository) *AgentHandler {
	return &AgentHandler{repo: repo}
}

// List GET /api/v1/agents?keyword=&status=&limit=&offset=
func (h *AgentHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, total, err := h.repo.List(c.Request.Context(), db.AgentListFilter{
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

// Get GET /api/v1/agents/:id
func (h *AgentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	agent, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

type AgentRequest struct {
	AgentName      string  `json:"agent_name" binding:"required,min=1,max=100"`
	ContactPerson  string  `json:"contact_person"`
	Phone          string  `json:"phone"`
	Email          string  `json:"email"`
	Region         string  `json:"region"`
	CommissionRate float64 `json:"commission_rate"`
	Status         string  `json:"status"`
	Description    string  `json:"description"`
}

func (req AgentRequest) toInput() db.AgentInput {
	status := req.Status
	if status == "" {
		status = "active"
	}
	return db.AgentInput{
		AgentName:      req.AgentName,
		ContactPerson:  req.ContactPerson,
		Phone:          req.Phone,
		Email:          req.Email,
		Region:         req.Region,
		CommissionRate: req.CommissionRate,
		Status:         status,
		Description:    req.Description,
	}
}

// Create POST /api/v1/agents
func (h *AgentHandler) Create(c *gin.Context) {
	var req AgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var createdBy *uuid.UUID
	if claimsAny, ok := c.Get(auth.ClaimsContextKey); ok {
		uid := claimsAny.(*auth.Claims).UserID
		createdBy = &uid
	}
	agent, err := h.repo.Create(c.Request.Context(), req.toInput(), createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, agent)
}

// Update PUT /api/v1/agents/:id
func (h *AgentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req AgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	agent, err := h.repo.Update(c.Request.Context(), id, req.toInput())
	if err != nil {
		if errors.Is(err, db.ErrAgentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

// Delete DELETE /api/v1/agents/:id
func (h *AgentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, db.ErrAgentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// ListCustomers GET /api/v1/agents/:id/customers
func (h *AgentHandler) ListCustomers(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	customers, err := h.repo.GetCustomers(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrAgentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "代理商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": customers})
}
