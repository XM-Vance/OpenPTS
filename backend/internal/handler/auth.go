// 鉴权 handler：登录、当前用户信息、当前用户权限。
package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
)

type AuthHandler struct {
	users   *db.UserRepository
	jwt     *auth.JWTService
	permSvc *auth.PermissionService
}

func NewAuthHandler(users *db.UserRepository, jwt *auth.JWTService, permSvc *auth.PermissionService) *AuthHandler {
	return &AuthHandler{users: users, jwt: jwt, permSvc: permSvc}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=4,max=128"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	u, err := h.users.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
		return
	}

	if !auth.CheckPassword(req.Password, u.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	orgID := ""
	if u.OrgID != nil {
		orgID = *u.OrgID
	}
	tok, err := h.jwt.Sign(u.ID, u.Username, orgID, u.IsHQ)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "签发 token 失败"})
		return
	}

	// P1-8:登录态走 httpOnly Cookie(前端 JS 不可读,防 XSS 窃取);
	// body 中保留 token 供脚本/工具与过渡期旧客户端使用(中间件双读)。
	setAuthCookie(c, tok, int(h.jwt.TTL().Seconds()))

	c.JSON(http.StatusOK, LoginResponse{
		Token:    tok,
		UserID:   u.ID.String(),
		Username: u.Username,
	})
}

// setAuthCookie 写入登录 Cookie。Secure 按请求实际协议判定:
// 直连 TLS 或经反代(X-Forwarded-Proto=https)时置位;
// 纯 HTTP 部署(内网默认编排)不置位,否则浏览器拒存 Cookie。
func setAuthCookie(c *gin.Context, token string, maxAge int) {
	secure := c.Request.TLS != nil ||
		strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(auth.TokenCookieName, token, maxAge, "/", "", secure, true)
}

// Logout POST /api/v1/auth/logout(需要 JWT):清除登录 Cookie。
// 幂等;localStorage 时代的旧客户端调用也无副作用。
func (h *AuthHandler) Logout(c *gin.Context) {
	setAuthCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"message": "已退出登录"})
}

// Me GET /api/v1/auth/me（需要 JWT）
func (h *AuthHandler) Me(c *gin.Context) {
	claimsAny, ok := c.Get(auth.ClaimsContextKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	claims := claimsAny.(*auth.Claims)

	u, err := h.users.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已禁用"})
		return
	}

	orgs, err := h.users.ListUserOrgs(c.Request.Context(), claims.UserID, u.IsHQ)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询可访问组织失败"})
		return
	}
	activeOrg := ""
	if u.OrgID != nil {
		activeOrg = *u.OrgID
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":      u.ID.String(),
		"username":     u.Username,
		"display_name": u.DisplayName,
		"is_hq":        u.IsHQ,
		"org_active":   activeOrg,
		"orgs":         orgs,
	})
}

// MyPermissions GET /api/v1/auth/me/permissions
// 返回当前用户拥有的权限码集合，前端用于按钮/菜单可见性控制。
func (h *AuthHandler) MyPermissions(c *gin.Context) {
	claimsAny, ok := c.Get(auth.ClaimsContextKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	claims := claimsAny.(*auth.Claims)

	codes, err := h.permSvc.GetUserPermissions(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询权限失败"})
		return
	}

	list := make([]string, 0, len(codes))
	for code := range codes {
		list = append(list, code)
	}

	c.JSON(http.StatusOK, gin.H{"permissions": list})
}

// ChangePasswordRequest 修改密码请求体。
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword POST /api/v1/auth/change-password
// 用户自助修改密码：校验旧密码 → bcrypt 哈希新密码 → 更新；不影响当前 JWT。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误（新密码至少 8 位）"})
		return
	}

	claimsAny, ok := c.Get(auth.ClaimsContextKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	claims := claimsAny.(*auth.Claims)

	u, err := h.users.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
		return
	}

	if !auth.CheckPassword(req.OldPassword, u.PasswordHash) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原密码错误"})
		return
	}
	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码不能与原密码相同"})
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码哈希失败"})
		return
	}
	if err := h.users.UpdatePassword(c.Request.Context(), claims.UserID, newHash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码已更新"})
}
