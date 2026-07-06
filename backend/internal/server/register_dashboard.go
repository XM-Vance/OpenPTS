// 仪表盘域：跨模块 KPI 总览（任何已登录用户可访问，无独立权限码）。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
)

func registerDashboard(g *gin.RouterGroup, d *Deps) {
	dashboardH := handler.NewDashboardHandler(d.DashboardRepo)

	g.GET("/dashboard/summary", dashboardH.Summary)
	g.GET("/dashboard/settlement-summary", dashboardH.SettlementSummary)
	g.GET("/dashboard/series/settlement", dashboardH.SettlementSeries)
	g.GET("/dashboard/series/freq", dashboardH.FreqSeries)
	g.GET("/dashboard/config", dashboardH.GetConfig)
	g.PUT("/dashboard/config", dashboardH.SaveConfig)
}
