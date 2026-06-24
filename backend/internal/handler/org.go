// 组织（省份）管理 handler：组织 CRUD + 成员查询 + 用户↔省份分配。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/db"
)

type OrgHandler struct {
	orgs  *db.OrgRepository
	users *db.UserRepository
}

func NewOrgHandler(orgs *db.OrgRepository, users *db.UserRepository) *OrgHandler {
	return &OrgHandler{orgs: orgs, users: users}
}

// List GET /api/v1/orgs
func (h *OrgHandler) List(c *gin.Context) {
	list, err := h.orgs.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询组织失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orgs": list})
}

type createOrgReq struct {
	Code string `json:"code" binding:"required,min=1,max=32"`
	Name string `json:"name" binding:"required,min=1,max=64"`
}

// Create POST /api/v1/orgs（加省份用）
func (h *OrgHandler) Create(c *gin.Context) {
	var req createOrgReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	o, err := h.orgs.Create(c.Request.Context(), req.Code, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建组织失败（code 可能重复）"})
		return
	}
	c.JSON(http.StatusCreated, o)
}

type updateOrgReq struct {
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// Update PATCH /api/v1/orgs/:id（改名/启停）
func (h *OrgHandler) Update(c *gin.Context) {
	var req updateOrgReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.orgs.Update(c.Request.Context(), c.Param("id"), req.Name, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新"})
}

// Members GET /api/v1/orgs/:id/members
func (h *OrgHandler) Members(c *gin.Context) {
	list, err := h.orgs.ListMembers(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询成员失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": list})
}

type setUserOrgsReq struct {
	OrgIDs       []string `json:"org_ids"`
	IsHQ         bool     `json:"is_hq"`
	PrimaryOrgID string   `json:"primary_org_id"`
}

// SetUserOrgs PUT /api/v1/users/:id/orgs（设置某用户可访问的省 + 总部标记 + 主省）
func (h *OrgHandler) SetUserOrgs(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户 ID 无效"})
		return
	}
	var req setUserOrgsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.users.SetUserOrgs(c.Request.Context(), uid, req.OrgIDs, req.IsHQ, req.PrimaryOrgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已设置"})
}
