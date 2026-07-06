// 储能域：储能电站·运行 / 储能申报 / 虚拟电厂(VPP) / 光伏电站·预测·收益。
// 统一走储能(storage)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerStorage(g *gin.RouterGroup, d *Deps) {
	storageH := handler.NewStorageHandler(d.StorageRepo)
	storageDeclH := handler.NewStorageDeclHandler(d.StorageDeclRepo)
	vppH := handler.NewVPPHandler(d.VPPRepo)
	solarH := handler.NewSolarHandler(d.SolarRepo)

	reqSTRead := middleware.RequirePermission(d.PermSvc, "storage:read")
	reqSTWrite := middleware.RequirePermission(d.PermSvc, "storage:write")

	// 储能电站 / 运行
	g.GET("/storage/stations", reqSTRead, storageH.ListStations)
	g.GET("/storage/stations/:id/operations", reqSTRead, storageH.ListOperations)
	g.POST("/storage/demo-data", reqSTWrite, storageH.GenerateDemoData)

	// E6 储能申报（沿用储能模块权限）
	g.GET("/storage/declarations", reqSTRead, storageDeclH.List)
	g.POST("/storage/declarations/demo-data", reqSTWrite, storageDeclH.GenerateDemoData)

	// G6 虚拟电厂（沿用储能模块权限）
	g.GET("/vpp/resources", reqSTRead, vppH.ListResources)
	g.GET("/vpp/dispatches", reqSTRead, vppH.ListDispatches)
	g.POST("/vpp/demo-data", reqSTWrite, vppH.GenerateDemoData)

	// P2 光伏系列（沿用储能模块权限）
	g.GET("/solar/stations", reqSTRead, solarH.ListStations)
	g.GET("/solar/stations/:id", reqSTRead, solarH.GetStation)
	g.POST("/solar/stations", reqSTWrite, solarH.CreateStation)
	g.PUT("/solar/stations/:id", reqSTWrite, solarH.UpdateStation)
	g.DELETE("/solar/stations/:id", reqSTWrite, solarH.DeleteStation)
	g.GET("/solar/forecast", reqSTRead, solarH.ListForecast)
	g.GET("/solar/revenue", reqSTRead, solarH.ListRevenue)
	g.POST("/solar/demo-data", reqSTWrite, solarH.GenerateDemoData)
}
