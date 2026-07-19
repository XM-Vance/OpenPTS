// 提权防护：限制高风险的用户/角色/组织变更操作，防止普通用户管理员自我提权。
//
// 背景：SetRoles / SetPermissions / SetUserOrgs 三个接口仅挂了 user_management:write，
// 但请求体里的角色码、权限码、is_hq 全部直接信任，无层级校验。导致拥有 user_management:write
// 的普通用户管理员可：
//   - 把自己分配 super_admin 系统角色；
//   - 给任意角色注入 system:write 等高权权限码；
//   - 把自己 is_hq 置 true，绕过全部省份租户隔离。
//
// 修复：这些操作要求调用方自身已是 super_admin（系统内置最高权角色），且禁止自我赋权
// （改自己的角色/组织必须由另一名超管操作，避免一人完成提权闭环）。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
)

// superAdminRole 系统内置最高权角色码（seed 0015，is_system=true，拥有全部权限）。
const superAdminRole = "super_admin"

// requireSuperAdmin 校验当前调用方是 super_admin；否则 403。
// 用 PermissionService 查调用方角色集合（经 ListUserRoles），不依赖 Claims.Roles（Claims 未带角色）。
func requireSuperAdmin(c *gin.Context, users *db.UserRepository) (*auth.Claims, bool) {
	claimsAny, ok := c.Get(auth.ClaimsContextKey)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return nil, false
	}
	claims := claimsAny.(*auth.Claims)
	codes, err := users.ListUserRoles(c.Request.Context(), claims.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "权限校验失败"})
		return nil, false
	}
	for _, code := range codes {
		if code == superAdminRole {
			return claims, true
		}
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "仅超级管理员可执行此操作"})
	return nil, false
}

// forbidSelfAssign 校验目标用户不是调用方自己（自我赋权需另一名超管确认）。
func forbidSelfAssign(c *gin.Context, claims *auth.Claims, targetID uuid.UUID) bool {
	if claims.UserID == targetID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "不可修改自己的角色或组织授权，请联系另一名超级管理员"})
		return true
	}
	return false
}

// requireSystemRoleTarget 校验被操作的角色码不是系统内置角色（除非调用方要管理它）。
// 返回 true 表示是系统角色。
func isSystemRoleCode(code string) bool {
	return code == superAdminRole
}
