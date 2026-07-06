// 客户域：客户档案 CRUD + 历史电量 + 360 视图 + 意向客户 + 代理商 + 保函。
// 统一走客户管理(customer_management)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerCustomer(g *gin.RouterGroup, d *Deps) {
	customersH := handler.NewCustomersHandler(d.CustomerRepo, d.PermSvc)
	customerEnergyH := handler.NewCustomerEnergyHandler(d.CustomerEnergyRepo)
	intentH := handler.NewIntentCustomerHandler(d.IntentRepo)
	agentH := handler.NewAgentHandler(d.AgentRepo)
	bondH := handler.NewBondHandler(d.BondRepo)

	reqCMRead := middleware.RequirePermission(d.PermSvc, "customer_management:read")
	reqCMWrite := middleware.RequirePermission(d.PermSvc, "customer_management:write")
	reqCMDelete := middleware.RequirePermission(d.PermSvc, "customer_management:delete")

	// 客户档案
	g.GET("/customers", reqCMRead, customersH.List)
	g.GET("/customers/:id", reqCMRead, customersH.Get)
	g.POST("/customers", reqCMWrite, customersH.Create)
	g.PUT("/customers/:id", reqCMWrite, customersH.Update)
	g.DELETE("/customers/:id", reqCMDelete, customersH.Delete)

	// 客户历史电量档案（市场化账单/月度电量入库后在此查看）
	g.GET("/customer-energy", reqCMRead, customerEnergyH.List)

	// 客户 360 视图
	g.GET("/customers/:id/360", reqCMRead, customersH.View360)

	// E1 意向客户（沿用客户模块权限）
	g.GET("/intent-customers", reqCMRead, intentH.List)
	g.GET("/intent-customers/diagnose", reqCMRead, intentH.Diagnose)
	g.POST("/intent-customers/demo-data", reqCMWrite, intentH.GenerateDemoData)

	// H1 代理商管理（沿用客户管理权限）
	g.GET("/agents", reqCMRead, agentH.List)
	g.GET("/agents/:id", reqCMRead, agentH.Get)
	g.POST("/agents", reqCMWrite, agentH.Create)
	g.PUT("/agents/:id", reqCMWrite, agentH.Update)
	g.DELETE("/agents/:id", reqCMDelete, agentH.Delete)
	g.GET("/agents/:id/customers", reqCMRead, agentH.ListCustomers)

	// H2 保函管理（沿用客户管理权限）
	g.GET("/bonds", reqCMRead, bondH.List)
	g.GET("/bonds/:id", reqCMRead, bondH.Get)
	g.POST("/bonds", reqCMWrite, bondH.Create)
	g.PUT("/bonds/:id", reqCMWrite, bondH.Update)
	g.DELETE("/bonds/:id", reqCMDelete, bondH.Delete)
}
