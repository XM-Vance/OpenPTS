// 调频域：调频出清。走调频(freq_regulation)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerFreq(g *gin.RouterGroup, d *Deps) {
	freqH := handler.NewFreqHandler(d.FreqRepo)

	reqFRRead := middleware.RequirePermission(d.PermSvc, "freq_regulation:read")
	reqFRWrite := middleware.RequirePermission(d.PermSvc, "freq_regulation:write")

	g.GET("/freq/clearing", reqFRRead, freqH.List)
	g.POST("/freq/demo-data", reqFRWrite, freqH.GenerateDemoData)
}
