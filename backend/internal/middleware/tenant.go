// 多租户中间件：从 JWT claims + X-Org-Id 头解析「当前活跃省（组织）」并校验访问权，
// 写入 gin context。业务 handler 通过 OrgFromContext(c) 读取。
//
// 活跃省解析：X-Org-Id 头 → JWT 主组织（claims.OrgID）。
// 访问校验：
//   - 总部（claims.IsHQ）：可访问任意省，含哨兵 "*"（全部省）。
//   - 非总部：活跃省必须等于主省，或在 user_orgs 成员集合内，否则 403。
package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
)

const OrgContextKey = "ptis.org_id"

// OrgAll 是总部「全部省」的哨兵活跃省值。
const OrgAll = "*"

// OrgMembershipFunc 校验用户是否可访问某组织（省）。由 router 注入（通常 UserRepo.IsOrgMember）。
type OrgMembershipFunc func(ctx context.Context, userID uuid.UUID, orgID string) (bool, error)

// Tenant 解析并校验活跃省，写入 context。
func Tenant(isMember OrgMembershipFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsAny, ok := c.Get(auth.ClaimsContextKey)
		if !ok {
			c.Next() // 无 claims（理论上认证组不会到这）；交后续处理
			return
		}
		claims := claimsAny.(*auth.Claims)

		active := c.GetHeader("X-Org-Id")
		if active == "" {
			active = claims.OrgID // 主/默认省：登录时已确定，天然有效
		}

		switch {
		case active == OrgAll:
			if !claims.IsHQ {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权访问全部省"})
				return
			}
		case active != "" && active != claims.OrgID && !claims.IsHQ:
			// 非总部切换到非主省：校验成员资格
			allowed, err := isMember(c.Request.Context(), claims.UserID, active)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "校验组织权限失败"})
				return
			}
			if !allowed {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无该省访问权限"})
				return
			}
		}

		c.Set(OrgContextKey, active)
		// 同时注入到 request context，让 repo 层通过 OrgFromCtx(ctx) 取活跃省
		c.Request = c.Request.WithContext(db.WithOrg(c.Request.Context(), active))
		c.Next()
	}
}

// OrgFromContext 读取当前活跃省（组织 ID）；"" 表示未指定，"*" 表示总部全部省。
func OrgFromContext(c *gin.Context) string {
	if v, ok := c.Get(OrgContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
