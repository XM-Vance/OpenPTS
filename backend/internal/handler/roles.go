// 角色 CRUD handler + 权限分配。
package handler

import (
	"net/http"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
)

type RolesHandler struct {
	roleRepo *db.RoleRepository
	permRepo *db.PermissionRepository
}

func NewRolesHandler(roleRepo *db.RoleRepository, permRepo *db.PermissionRepository) *RolesHandler {
	return &RolesHandler{roleRepo: roleRepo, permRepo: permRepo}
}

// RoleView 角色视图（含权限点清单）。
type RoleView struct {
	*db.Role
	Permissions []string `json:"permissions"`
}

// List GET /api/v1/roles
func (h *RolesHandler) List(c *gin.Context) {
	list, err := h.roleRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Get GET /api/v1/roles/:code
func (h *RolesHandler) Get(c *gin.Context) {
	code := c.Param("code")
	role, err := h.roleRepo.Get(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		return
	}
	perms, _ := h.permRepo.ListByRole(c.Request.Context(), code)
	c.JSON(http.StatusOK, RoleView{Role: role, Permissions: perms})
}

type CreateRoleRequest struct {
	Code        string   `json:"code" binding:"required,min=2,max=64"`
	Name        string   `json:"name" binding:"required,min=2,max=128"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// Create POST /api/v1/roles
func (h *RolesHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}
	r, err := h.roleRepo.Create(c.Request.Context(), req.Code, req.Name, desc)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "创建失败：编码可能已存在"})
		return
	}
	if len(req.Permissions) > 0 {
		_ = h.roleRepo.SetPermissions(c.Request.Context(), req.Code, req.Permissions)
	}
	c.JSON(http.StatusCreated, r)
}

type UpdateRoleRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=128"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}

// Update PUT /api/v1/roles/:code
func (h *RolesHandler) Update(c *gin.Context) {
	code := c.Param("code")
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	r, err := h.roleRepo.Update(c.Request.Context(), code, req.Name, req.Description, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, r)
}

// Delete DELETE /api/v1/roles/:code（系统角色不可删）
func (h *RolesHandler) Delete(c *gin.Context) {
	code := c.Param("code")
	if err := h.roleRepo.Delete(c.Request.Context(), code); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

type SetRolePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

// SetPermissions PUT /api/v1/roles/:code/permissions
func (h *RolesHandler) SetPermissions(c *gin.Context) {
	code := c.Param("code")
	var req SetRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.roleRepo.SetPermissions(c.Request.Context(), code, req.Permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置权限失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新"})
}
