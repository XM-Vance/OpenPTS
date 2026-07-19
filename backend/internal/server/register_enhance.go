// 模块关联增强域：自定义字段 / 标签 / 全局搜索 / 文档→合同自动填充 / 意向客户转正。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
)

func registerEnhance(g *gin.RouterGroup, d *Deps) {
	customFieldH := handler.NewCustomFieldHandler(d.CustomFieldRepo)
	tagH := handler.NewTagHandler(d.TagRepo)
	searchH := handler.NewSearchHandler(d.Pool)
	docRepo := db.NewDocumentRepository(d.Pool)
	integrationH := handler.NewIntegrationHandler(d.Pool, docRepo, d.RetailRepo, d.CustomerRepo, d.IntentRepo)

	reqCMWrite := middleware.RequirePermission(d.PermSvc, "customer_management:write")
	reqCMDelete := middleware.RequirePermission(d.PermSvc, "customer_management:delete")
	reqRMWrite := middleware.RequirePermission(d.PermSvc, "retail_management:write")
	reqSYRead := middleware.RequirePermission(d.PermSvc, "system:read")
	reqSYWrite := middleware.RequirePermission(d.PermSvc, "system:write")
	reqSYDelete := middleware.RequirePermission(d.PermSvc, "system:delete")

	// 自定义字段管理
	g.GET("/custom-fields", reqSYRead, customFieldH.List)
	g.POST("/custom-fields", reqSYWrite, customFieldH.Create)
	g.PUT("/custom-fields/:id", reqSYWrite, customFieldH.Update)
	g.DELETE("/custom-fields/:id", reqSYDelete, customFieldH.Delete)

	// 自定义字段值（实体级，0102 新增）—— 让定义的字段能真正写入客户/合同等实体
	g.GET("/custom-fields/values", reqSYRead, customFieldH.ListValues)
	g.POST("/custom-fields/values", reqCMWrite, customFieldH.UpsertValue)
	g.DELETE("/custom-fields/values", reqCMDelete, customFieldH.DeleteValue)

	// 标签管理
	g.GET("/tags", reqSYRead, tagH.List)
	g.POST("/tags", reqSYWrite, tagH.Create)
	g.PUT("/tags/:id", reqSYWrite, tagH.Update)
	g.DELETE("/tags/:id", reqSYDelete, tagH.Delete)
	g.POST("/tags/batch-apply", reqSYWrite, tagH.BatchApply)
	g.GET("/tags/entity", reqSYRead, tagH.GetEntityTags)

	// 全局搜索
	g.GET("/search", reqSYRead, searchH.Search)

	// 文档→合同自动填充
	g.POST("/documents/:id/apply-to-contract", reqRMWrite, integrationH.ApplyToContract)

	// 意向客户转正
	g.POST("/intent-customers/:id/convert", reqCMWrite, integrationH.ConvertIntentCustomer)
}
