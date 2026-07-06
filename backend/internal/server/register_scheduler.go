// 调度域：任务调度(jobs/runs/触发/启停) + RPA 监控。
// 统一走任务调度(task_scheduler)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerScheduler(g *gin.RouterGroup, d *Deps) {
	schedulerH := handler.NewSchedulerHandler(d.SchedulerRepo, d.Scheduler)
	rpaH := handler.NewRPAHandler(d.RPARepo)

	reqTSRead := middleware.RequirePermission(d.PermSvc, "task_scheduler:read")
	reqTSWrite := middleware.RequirePermission(d.PermSvc, "task_scheduler:write")

	// 任务调度
	g.GET("/scheduler/jobs", reqTSRead, schedulerH.ListJobs)
	g.GET("/scheduler/runs", reqTSRead, schedulerH.ListRuns)
	g.POST("/scheduler/jobs/:id/trigger", reqTSWrite, schedulerH.Trigger)
	g.PUT("/scheduler/jobs/:id/enabled", reqTSWrite, schedulerH.SetEnabled)

	// RPA 监控（沿用任务调度权限）
	g.GET("/rpa/jobs", reqTSRead, rpaH.ListJobs)
	g.GET("/rpa/runs", reqTSRead, rpaH.ListRuns)
	g.POST("/rpa/demo-data", reqTSWrite, rpaH.GenerateDemoData)
}
