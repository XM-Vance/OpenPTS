// 文档域：文档解析管线（原件/解析件存档 + 结构化提取 + 人工确认入库，按活跃省隔离）
// + 政策文件库。统一走文档管理(document_management)权限码。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/document"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerDocument(g *gin.RouterGroup, d *Deps) {
	// 文档解析管线：仓储 + 异步解析 worker + 确认入库映射器
	docRepo := db.NewDocumentRepository(d.Pool)
	docImporters := &document.Importers{Customers: d.CustomerRepo, Intent: d.IntentRepo, Load: d.LoadRepo, Monthly: d.MonthlyRepo, CustomerEnergy: d.CustomerEnergyRepo, Policy: d.PolicyRepo, Retail: d.RetailRepo}
	docWorker := document.NewWorker(docRepo, d.Docling, d.ObjectStore, docImporters)
	documentH := handler.NewDocumentHandler(docRepo, docWorker, docImporters, d.ObjectStore, d.PermSvc)
	policyH := handler.NewPolicyHandler(d.PolicyRepo)

	reqDocRead := middleware.RequirePermission(d.PermSvc, "document_management:read")
	reqDocWrite := middleware.RequirePermission(d.PermSvc, "document_management:write")
	reqDocDelete := middleware.RequirePermission(d.PermSvc, "document_management:delete")

	// 文档解析管线
	g.POST("/documents", reqDocWrite, documentH.Upload)
	g.GET("/documents", reqDocRead, documentH.List)
	g.GET("/documents/:id", reqDocRead, documentH.Get)
	g.GET("/documents/:id/original", reqDocRead, documentH.Original)
	g.GET("/documents/:id/parsed", reqDocRead, documentH.Parsed)
	g.POST("/documents/:id/extractions", reqDocWrite, documentH.AddExtraction)
	g.PUT("/documents/:id/extractions/:eid", reqDocWrite, documentH.UpdateExtraction)
	g.DELETE("/documents/:id/extractions/:eid", reqDocWrite, documentH.DeleteExtraction)
	g.POST("/documents/:id/reparse", reqDocWrite, documentH.Reparse)
	g.POST("/documents/:id/apply", reqDocWrite, documentH.Apply)
	g.DELETE("/documents/:id", reqDocDelete, documentH.Delete)

	// 政策文件库（文档中心归档：解析「确认入库 → 政策文件」或手动新增）
	g.GET("/policies", reqDocRead, policyH.List)
	g.POST("/policies", reqDocWrite, policyH.Create)
	g.DELETE("/policies/:id", reqDocDelete, policyH.Delete)
}
