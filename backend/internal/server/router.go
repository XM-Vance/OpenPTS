// 路由组装：依赖通过 Deps 注入；业务路由按权限码挂载 RequirePermission 中间件。
//
// 结构（R8 按域拆分，替代原 737 行单函数）：
//   - deps.go              依赖集合 Deps
//   - router.go            本文件：gin 引擎 + 全局中间件 + 公共端点 + 调度各域注册
//   - register_<domain>.go 每个业务域一个，自建 handler + 权限码 + 路由
//
// 新增一个域 = 新增一个 register_<域>.go + 在下方 NewRouter 里加一行 register<Domain>(authed, d)。
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func NewRouter(d *Deps) *gin.Engine {
	if d.Config.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// 方法不匹配返回 405（而非默认 404），便于契约闸区分"路由不存在"与"方法不对"。
	r.HandleMethodNotAllowed = true
	// R7: 中间件链优化 — 轻量在前、重量在后，未命中路由时少执行
	// OTel 分布式追踪：最外层，包住 request_id/metrics/鉴权全链（连不上 Tempo 静默降级）。
	r.Use(otelgin.Middleware("openpts-backend"))
	r.Use(middleware.Recovery())                  // 最轻：仅捕获 panic
	r.Use(middleware.RequestID())                 // 轻量：注入 trace-id
	r.Use(middleware.CORS(!d.Config.IsProd()))    // 轻量：CORS 头 + OPTIONS 预检放行
	r.Use(middleware.Logger())                    // 轻量：请求日志
	r.Use(middleware.Metrics())                   // Prometheus 指标采集
	r.Use(middleware.RateLimit(60, 120))          // 重量：令牌桶计算
	r.Use(middleware.DemoGate(d.Config.IsProd())) // 生产环境拦截 demo-data 端点

	r.GET("/health", handler.Health(d.Pool))
	r.GET("/metrics", middleware.PrometheusHandler()) // Prometheus 抓取端点（无需鉴权）
	r.GET("/docs", handler.SwaggerUI)                 // OpenAPI 文档 UI
	r.GET("/docs/openapi.yaml", handler.OpenAPISpec)  // OpenAPI 规范源

	apiV1 := r.Group("/api/v1")
	apiV1.GET("/ping", handler.Ping) // 联通性探活（公开，与 auth/login 同段语义）

	// 鉴权子组：JWT + 多租户 + 审计。绝大多数业务路由挂在此组下。
	authed := apiV1.Group("")
	authed.Use(middleware.JWT(d.JWT))
	authed.Use(middleware.Tenant(d.UserRepo.IsOrgMember))
	authed.Use(middleware.Audit(d.AuditWriter))

	// ─── 含公开端点的域（需要 apiV1 与 authed 两个组）───
	registerAuth(apiV1, authed, d)     // /auth/login(公开) + /auth/me 等
	registerRealtime(apiV1, authed, d) // /stream/alerts、/ws/echo(公开) + /stream/test、/online

	// ─── 纯鉴权业务域 ───
	registerDashboard(authed, d)
	registerSystem(authed, d)        // 用户/角色/组织/模块/权限/菜单/审计/系统配置/安全
	registerCustomer(authed, d)      // 客户/客户电量/360/意向客户/代理商/保函
	registerRetail(authed, d)        // 零售套餐/合同/日价/月结/签约进度/合同PDF/绿电
	registerLoad(authed, d)          // 负荷/诊断/特性/总负荷/中期/基础数据/准确率/电表导入
	registerWeather(authed, d)       // 气象站点/实况/预报/风水文
	registerPrice(authed, d)         // 价格预测/趋势/TOU/电网代理价/现货/市场分析
	registerSettlement(authed, d)    // 日结/月结/手工/预结算/偏差/机制电量/交易规则
	registerFreq(authed, d)          // 调频
	registerStorage(authed, d)       // 储能/申报/VPP/光伏
	registerAnalytics(authed, d)     // 告警/客户负荷/客户利润/客户分析
	registerTrade(authed, d)         // 日前复盘/月度复盘/撮合/滚动/竞价/策略/日前模拟
	registerScheduler(authed, d)     // 任务调度/RPA
	registerDocument(authed, d)      // 文档解析管线/政策文件
	registerAttachment(authed, d)    // 附件
	registerApproval(authed, d)      // 审批流
	registerMarketData(authed, d)    // 市场行情/碳交易
	registerEnhance(authed, d)       // 自定义字段/标签/全局搜索/文档→合同/意向转正
	registerExport(authed, d)        // 数据导出/客户导入
	registerContractStubs(authed, d) // 契约对齐补端点（空态占位）

	return r
}
