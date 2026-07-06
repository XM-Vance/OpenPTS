// 负荷与预测域：短期预测 / 诊断 / 特性 / 系统总负荷 / 中期预测 /
// 预测基础数据(节假日·典型曲线) / 预测准确率 / 电表数据导入。
// 多数走负荷管理(load_management)权限；准确率查看沿用 system:read。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerLoad(g *gin.RouterGroup, d *Deps) {
	loadH := handler.NewLoadHandler(d.LoadRepo, d.Config.DemoMode)
	loadDiagH := handler.NewLoadDiagnosisHandler(d.LoadDiagRepo)
	loadCharH := handler.NewLoadCharacteristicsHandler(d.LoadCharRepo)
	loadCharExtH := handler.NewLoadCharExtHandler(d.LoadCharExtRepo)
	loadDataH := handler.NewLoadDataHandler(d.LoadDataRepo)
	totalLoadH := handler.NewTotalLoadHandler(d.TotalLoadRepo)
	mediumH := handler.NewMediumForecastHandler(d.MediumForecastRepo)
	forecastBaseH := handler.NewForecastBaseHandler(d.ForecastBaseRepo)
	accuracyH := handler.NewAccuracyHandler(d.AccuracyRepo)
	loadImportH := handler.NewLoadImportHandler(d.LoadRepo, d.CustomerRepo)

	reqLMRead := middleware.RequirePermission(d.PermSvc, "load_management:read")
	reqLMWrite := middleware.RequirePermission(d.PermSvc, "load_management:write")
	reqSYRead := middleware.RequirePermission(d.PermSvc, "system:read")

	// 负荷预测
	g.POST("/load/forecast", reqLMRead, loadH.Forecast)
	g.POST("/load/demo-data", reqLMWrite, loadH.GenerateDemoData)

	// 预测基础数据
	g.GET("/forecast-base-data/availability", reqLMRead, forecastBaseH.Availability)
	g.POST("/forecast-base-data/curves", reqLMRead, forecastBaseH.Curves)

	// E3 负荷数据诊断（沿用负荷模块权限）
	g.GET("/load/diagnosis", reqLMRead, loadDiagH.List)

	// U3 预测基础数据
	g.GET("/load/holidays", reqLMRead, forecastBaseH.ListHolidays)
	g.GET("/load/typical-curves", reqLMRead, forecastBaseH.ListCurves)
	g.POST("/load/base-data/demo-data", reqLMWrite, forecastBaseH.GenerateDemoData)

	// V1 系统总负荷
	g.GET("/load/total", reqLMRead, totalLoadH.List)
	g.POST("/load/total/demo-data", reqLMWrite, totalLoadH.GenerateDemoData)

	// V2 中期负荷预测
	g.GET("/load/medium-forecast", reqLMRead, mediumH.List)
	g.POST("/load/medium-forecast/demo-data", reqLMWrite, mediumH.GenerateDemoData)

	// V3 预测准确率（system:read）
	g.GET("/forecast/accuracy", reqSYRead, accuracyH.List)
	g.GET("/forecast/accuracy/summary", reqSYRead, accuracyH.Summary)
	g.POST("/forecast/accuracy/demo-data", reqLMWrite, accuracyH.GenerateDemoData)

	// G8 负荷特性分析（沿用负荷模块权限）
	g.GET("/load/characteristics", reqLMRead, loadCharH.List)
	g.POST("/load/characteristics/demo-data", reqLMWrite, loadCharH.GenerateDemoData)
	// G8 负荷特性深度分析（11 个子端点，读 customer_characteristics 表）；
	// analyze/batch/all 需算法，暂不实现（前端已知保持调用，后端 404 由契约闸白名单跟踪）。
	g.GET("/load/characteristics/overview", reqLMRead, loadCharExtH.Overview)
	g.GET("/load/characteristics/overview/distribution", reqLMRead, loadCharExtH.Distribution)
	g.GET("/load/characteristics/overview/tag-changes", reqLMRead, loadCharExtH.TagChanges)
	g.GET("/load/characteristics/overview/scatter-data", reqLMRead, loadCharExtH.ScatterData)
	g.GET("/load/characteristics/customers", reqLMRead, loadCharExtH.ListCustomers)
	g.GET("/load/characteristics/customer/:id", reqLMRead, loadCharExtH.GetCustomer)
	g.GET("/load/characteristics/customer/:id/history", reqLMRead, loadCharExtH.CustomerHistory)
	g.GET("/load/characteristics/customer/:id/alerts", reqLMRead, loadCharExtH.CustomerAlerts)
	g.POST("/load/characteristics/alerts/:id/acknowledge", reqLMRead, loadCharExtH.AckAlert)
	g.GET("/load/characteristics/customer/:id/daily-trend", reqLMRead, loadCharExtH.DailyTrend)
	g.GET("/load/characteristics/customer/:id/monthly-energy", reqLMRead, loadCharExtH.MonthlyEnergy)

	// W3 电表数据导入
	g.POST("/import/load", reqLMWrite, middleware.StrictRateLimit(0.033, 2), loadImportH.Load)

	// 负荷数据查询（计量数据管理，6 个查询端点）；
	// reaggregate / calibration 系列（5 个）需算法，暂不实现（契约闸白名单跟踪）。
	g.GET("/load-data/customers", reqLMRead, loadDataH.ListCustomers)
	g.GET("/load-data/signed-customers", reqLMRead, loadDataH.SignedCustomers)
	g.GET("/load-data/customers/:id", reqLMRead, loadDataH.CustomerDetail)
	g.GET("/load-data/customers/:id/calendar", reqLMRead, loadDataH.CustomerCalendar)
	g.GET("/load-data/customers/:id/curves", reqLMRead, loadDataH.CustomerCurves)
	g.GET("/load-data/export/mp-missing", reqLMRead, loadDataH.ExportMpMissing)
}
