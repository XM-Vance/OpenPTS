// 气象域：气象数据 / 风电场风速 / 水库水文 / 站点·实况·预报。
// 气象是负荷预测的输入，沿用负荷管理(load_management)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerWeather(g *gin.RouterGroup, d *Deps) {
	weatherH := handler.NewWeatherHandler(d.WeatherRepo)

	reqLMRead := middleware.RequirePermission(d.PermSvc, "load_management:read")
	reqLMWrite := middleware.RequirePermission(d.PermSvc, "load_management:write")

	// 气象（沿用负荷模块权限：气象 → 负荷预测的输入）
	g.GET("/weather", reqLMRead, weatherH.List)
	g.POST("/weather/demo-data", reqLMWrite, weatherH.GenerateDemoData)

	// 外部气象观测（风电场风速 / 水库水文，原市场行情，现并入气象数据）
	g.GET("/weather/wind-farm", reqLMRead, weatherH.WindFarm)
	g.GET("/weather/hydrology", reqLMRead, weatherH.Hydrology)
	g.POST("/weather/obs-demo-data", reqLMWrite, weatherH.GenObsDemo)

	// 气象站点/实况/预报（契约对齐前端实际调用路径；空数据时返回形状正确的空态）
	g.GET("/weather/locations", reqLMRead, weatherH.ListLocations)
	g.POST("/weather/locations", reqLMWrite, weatherH.CreateLocation)
	g.PUT("/weather/locations/:id", reqLMWrite, weatherH.UpdateLocation)
	g.DELETE("/weather/locations/:id", reqLMWrite, weatherH.DeleteLocation)
	g.GET("/weather/actuals", reqLMRead, weatherH.Actuals)
	g.GET("/weather/actuals/summary", reqLMRead, weatherH.ActualsSummary)
	g.GET("/weather/forecasts", reqLMRead, weatherH.Forecasts)
	g.GET("/weather/forecasts/summary", reqLMRead, weatherH.ForecastsSummary)
	g.GET("/weather/forecast-dates", reqLMRead, weatherH.ForecastDates)
}
