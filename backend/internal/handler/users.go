// 用户 CRUD handler。
package handler

import (
	"net/http"
	"strconv"

	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UsersHandler struct {
	repo    *db.UserRepository
	permSvc *auth.PermissionService
}

func NewUsersHandler(repo *db.UserRepository, permSvc *auth.PermissionService) *UsersHandler {
	return &UsersHandler{repo: repo, permSvc: permSvc}
}

// UserView 用户视图（含角色清单）。
type UserView struct {
	*db.User
	Roles []string `json:"roles"`
}

// List GET /api/v1/users?keyword=&is_active=true&limit=20&offset=0
func (h *UsersHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		b := v == "true" || v == "1"
		isActive = &b
	}

	users, total, err := h.repo.List(c.Request.Context(), db.UserListFilter{
		Keyword:  c.Query("keyword"),
		IsActive: isActive,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": users, "total": total})
}

// Get GET /api/v1/users/:id
func (h *UsersHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	u, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	roles, _ := h.repo.ListUserRoles(c.Request.Context(), id)
	c.JSON(http.StatusOK, UserView{User: u, Roles: roles})
}

type CreateUserRequest struct {
	Username    string   `json:"username" binding:"required,min=2,max=64"`
	Password    string   `json:"password" binding:"required,min=4,max=128"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email" binding:"omitempty,email"`
	Phone       string   `json:"phone"`
	Roles       []string `json:"roles"`
}

// Create POST /api/v1/users
func (h *UsersHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "哈希密码失败"})
		return
	}
	u, err := h.repo.Create(c.Request.Context(), req.Username, hash, req.DisplayName)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "创建失败：用户名可能已存在"})
		return
	}

	// 更新 email / phone
	if req.Email != "" || req.Phone != "" {
		var email, phone *string
		if req.Email != "" {
			email = &req.Email
		}
		if req.Phone != "" {
			phone = &req.Phone
		}
		if updated, errU := h.repo.Update(c.Request.Context(), u.ID, nil, email, phone, nil); errU == nil {
			u = updated
		}
	}

	// 角色：未指定时默认赋予 viewer（全模块只读），避免造出「能登录却看不到任何功能」的零权限账号。
	// 仅创建时兜底；编辑用户时清空角色是管理员的显式选择，不在此自动补。
	roles := req.Roles
	if len(roles) == 0 {
		roles = []string{"viewer"}
	}
	_ = h.repo.SetRoles(c.Request.Context(), u.ID, roles)
	h.permSvc.Invalidate(u.ID)
	c.JSON(http.StatusCreated, u)
}

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name"`
	Email       *string `json:"email" binding:"omitempty,email"`
	Phone       *string `json:"phone"`
	IsActive    *bool   `json:"is_active"`
}

// Update PUT /api/v1/users/:id
func (h *UsersHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	u, err := h.repo.Update(c.Request.Context(), id, req.DisplayName, req.Email, req.Phone, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	if req.IsActive != nil {
		h.permSvc.Invalidate(id)
	}
	c.JSON(http.StatusOK, u)
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=4,max=128"`
}

// ResetPassword POST /api/v1/users/:id/password
func (h *UsersHandler) ResetPassword(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "哈希密码失败"})
		return
	}
	if err := h.repo.UpdatePassword(c.Request.Context(), id, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已重置"})
}

type SetUserRolesRequest struct {
	Roles []string `json:"roles"`
}

// SetRoles PUT /api/v1/users/:id/roles
func (h *UsersHandler) SetRoles(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req SetUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.repo.SetRoles(c.Request.Context(), id, req.Roles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置角色失败"})
		return
	}
	h.permSvc.Invalidate(id)
	c.JSON(http.StatusOK, gin.H{"message": "已更新"})
}
