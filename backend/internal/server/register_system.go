// 系统管理域：用户 / 角色 / 组织(省份) / 模块 / 权限 / 菜单可见性 / 操作审计 /
// 系统配置 / 安全大屏。统一走用户管理(user_management)与系统(system)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerSystem(g *gin.RouterGroup, d *Deps) {
	usersH := handler.NewUsersHandler(d.UserRepo, d.PermSvc)
	rolesH := handler.NewRolesHandler(d.RoleRepo, d.PermRepo, d.UserRepo)
	orgH := handler.NewOrgHandler(db.NewOrgRepository(d.Pool), d.UserRepo)
	modulesH := handler.NewModulesHandler(d.ModRepo)
	menuH := handler.NewMenuHandler(d.Pool)
	permsH := handler.NewPermissionsHandler(d.PermRepo)
	auditH := handler.NewAuditHandler(d.AuditRepo)
	settingsH := handler.NewSettingsHandler(d.SettingsRepo)
	securityH := handler.NewSecurityHandler(d.Pool)

	reqUMRead := middleware.RequirePermission(d.PermSvc, "user_management:read")
	reqUMWrite := middleware.RequirePermission(d.PermSvc, "user_management:write")
	reqUMDelete := middleware.RequirePermission(d.PermSvc, "user_management:delete")
	reqSYRead := middleware.RequirePermission(d.PermSvc, "system:read")

	// 用户 / 角色 / 模块 / 权限
	g.GET("/users", reqUMRead, usersH.List)
	g.GET("/users/:id", reqUMRead, usersH.Get)
	g.POST("/users", reqUMWrite, usersH.Create)
	g.PUT("/users/:id", reqUMWrite, usersH.Update)
	g.POST("/users/:id/password", reqUMWrite, usersH.ResetPassword)
	g.PUT("/users/:id/roles", reqUMWrite, usersH.SetRoles)
	g.PUT("/users/:id/orgs", reqUMWrite, orgH.SetUserOrgs)
	// 组织（省份）管理
	g.GET("/orgs", reqUMRead, orgH.List)
	g.POST("/orgs", reqUMWrite, orgH.Create)
	g.PATCH("/orgs/:id", reqUMWrite, orgH.Update)
	g.GET("/orgs/:id/members", reqUMRead, orgH.Members)
	g.GET("/roles", reqUMRead, rolesH.List)
	g.GET("/roles/:code", reqUMRead, rolesH.Get)
	g.POST("/roles", reqUMWrite, rolesH.Create)
	g.PUT("/roles/:code", reqUMWrite, rolesH.Update)
	g.DELETE("/roles/:code", reqUMDelete, rolesH.Delete)
	g.PUT("/roles/:code/permissions", reqUMWrite, rolesH.SetPermissions)
	g.GET("/modules", reqUMRead, modulesH.List)
	g.GET("/permissions", reqUMRead, permsH.List)

	// 菜单页面可见性管理
	g.GET("/menu/pages", reqUMRead, menuH.GetAllPages)
	g.GET("/menu/visible", menuH.GetVisiblePages)
	g.GET("/menu/roles", reqUMRead, menuH.GetAllPages) // 列出所有角色-页面分配
	g.GET("/menu/roles/:code", reqUMRead, menuH.GetVisiblePages)
	g.PUT("/menu/roles/:code", reqUMWrite, menuH.UpdateRolePages)

	// 操作审计
	g.GET("/audit/logs", reqSYRead, auditH.List)

	// R6 安全大屏（system:read）
	g.GET("/system/security/overview", reqSYRead, securityH.Overview)

	// W2 系统配置（system:read 查看，system:write 编辑）
	g.GET("/system/settings", reqSYRead, settingsH.List)
	g.PUT("/system/settings/:key", middleware.RequirePermission(d.PermSvc, "system:write"), settingsH.Update)
}
