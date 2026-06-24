// 路由组装：依赖通过 Deps 注入；业务路由按权限码挂载 RequirePermission 中间件。
package server

import (
	"github.com/ptis/backend/internal/approval"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/config"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/docling"
	"github.com/ptis/backend/internal/document"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
	"github.com/ptis/backend/internal/scheduler"
	"github.com/ptis/backend/internal/storage"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	Config         *config.Config
	Pool           *db.Pool
	JWT            *auth.JWTService
	UserRepo       *db.UserRepository
	RoleRepo       *db.RoleRepository
	PermRepo       *db.PermissionRepository
	ModRepo        *db.ModuleRepository
	CustomerRepo   *db.CustomerRepository
	CustomerEnergyRepo *db.CustomerEnergyRepository
	PolicyRepo     *db.PolicyRepository
	RetailRepo     *db.RetailRepository
	LoadRepo       *db.LoadRepository
	PriceRepo      *db.PriceRepository
	SettlementRepo *db.SettlementRepository
	FreqRepo       *db.FreqRepository
	StorageRepo    *db.StorageRepository
	AnalyticsRepo  *db.AnalyticsRepository
	DashboardRepo  *db.DashboardRepository
	SchedulerRepo  *db.SchedulerRepository
	Scheduler      *scheduler.Scheduler
	AuditRepo      *db.AuditRepository
	MonthlyRepo    *db.MonthlySettlementRepository
	SpotTrendRepo  *db.SpotTrendRepository
	DAReviewRepo   *db.DATradeReviewRepository
	WeatherRepo    *db.WeatherRepository
	RPARepo        *db.RPARepository
	ContractPriceRepo *db.ContractPriceRepository
	IntentRepo     *db.IntentCustomerRepository
	CustLoadRepo   *db.CustomerLoadRepository
	LoadDiagRepo   *db.LoadDiagnosisRepository
	TOURepo        *db.TOURepository
	GridAgencyRepo *db.GridAgencyRepository
	StorageDeclRepo *db.StorageDeclarationRepository
	CustProfitRepo *db.CustomerProfitRepository
	MTRReviewRepo  *db.MonthlyTradeReviewRepository
	MatchQuoteRepo *db.MatchQuoteRepository
	ManualRepo     *db.MonthlyManualRepository
	SSEHub         *handler.SSEHub
	AuditWriter    *middleware.AuditWriter
	AttachmentRepo *db.AttachmentRepository
	ApprovalRepo   *db.ApprovalRepository
	ApprovalReg    *approval.Registry
	ObjectStore    *storage.ObjectStore
	RetailMonthlyRepo *db.RetailMonthlyRepository
	PreSettleRepo     *db.PreSettleRepository
	ForecastBaseRepo  *db.ForecastBaseRepository
	TotalLoadRepo      *db.TotalLoadRepository
	MediumForecastRepo *db.MediumForecastRepository
	AccuracyRepo       *db.AccuracyRepository
	MechEnergyRepo     *db.MechanismEnergyRepository
	MarketAnalysisRepo *db.MarketAnalysisRepository
	SettingsRepo       *db.SettingsRepository
	ContractProgressRepo *db.ContractProgressRepository
	DeviationRepo       *db.DeviationRepository
	GreenPowerRepo      *db.GreenPowerRepository
	RollingTradeRepo    *db.RollingTradeRepository
	SpotMarketRepo      *db.SpotMarketRepository
	VPPRepo             *db.VPPRepository
	BiddingRepo         *db.BiddingRepository
	LoadCharRepo        *db.LoadCharacteristicsRepository
	CustAnalysisRepo    *db.CustomerAnalysisRepository
	TradeStrategyRepo   *db.TradeStrategyRepository
	AgentRepo          *db.AgentRepository
	BondRepo           *db.BondRepository
	SolarRepo          *db.SolarRepository
	DASimRepo          *db.DASimulationRepository
	MarketDataRepo    *db.MarketDataRepository
	CarbonRepo        *db.CarbonRepository
	PermSvc        *auth.PermissionService
	Docling        *docling.Client
	CustomFieldRepo *db.CustomFieldRepository
	TagRepo        *db.TagRepository
}

func NewRouter(d *Deps) *gin.Engine {
	if d.Config.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// 方法不匹配返回 405（而非默认 404），便于契约闸区分"路由不存在"与"方法不对"。
	r.HandleMethodNotAllowed = true
	// R7: 中间件链优化 — 轻量在前、重量在后，未命中路由时少执行
	r.Use(middleware.Recovery())                // 最轻：仅捕获 panic
	r.Use(middleware.RequestID())               // 轻量：注入 trace-id
	r.Use(middleware.CORS(!d.Config.IsProd()))  // 轻量：CORS 头 + OPTIONS 预检放行
	r.Use(middleware.Logger())                  // 轻量：请求日志
	r.Use(middleware.Metrics())                 // Prometheus 指标采集
	r.Use(middleware.RateLimit(60, 120))        // 重量：令牌桶计算
	r.Use(middleware.DemoGate(d.Config.IsProd())) // 生产环境拦截 demo-data 端点

	r.GET("/health", handler.Health(d.Pool))
	r.GET("/metrics", middleware.PrometheusHandler()) // Prometheus 抓取端点（无需鉴权）
	r.GET("/docs", handler.SwaggerUI)                 // OpenAPI 文档 UI
	r.GET("/docs/openapi.yaml", handler.OpenAPISpec)  // OpenAPI 规范源

	authHandler := handler.NewAuthHandler(d.UserRepo, d.JWT, d.PermSvc)
	usersH := handler.NewUsersHandler(d.UserRepo, d.PermSvc)
	rolesH := handler.NewRolesHandler(d.RoleRepo, d.PermRepo)
	orgH := handler.NewOrgHandler(db.NewOrgRepository(d.Pool), d.UserRepo)
	modulesH := handler.NewModulesHandler(d.ModRepo)
	menuH := handler.NewMenuHandler(d.Pool)
	permsH := handler.NewPermissionsHandler(d.PermRepo)
	customersH := handler.NewCustomersHandler(d.CustomerRepo, d.PermSvc)
	retailH := handler.NewRetailHandler(d.RetailRepo)
	loadH := handler.NewLoadHandler(d.LoadRepo, d.Config.DemoMode)
	priceH := handler.NewPriceHandler(d.PriceRepo, d.Config.DemoMode)
	settlementH := handler.NewSettlementHandler(d.SettlementRepo)
	freqH := handler.NewFreqHandler(d.FreqRepo)
	storageH := handler.NewStorageHandler(d.StorageRepo)
	analyticsH := handler.NewAnalyticsHandler(d.AnalyticsRepo, d.CustomerRepo)
	dashboardH := handler.NewDashboardHandler(d.DashboardRepo)
	schedulerH := handler.NewSchedulerHandler(d.SchedulerRepo, d.Scheduler)
	auditH := handler.NewAuditHandler(d.AuditRepo)
	monthlyH := handler.NewMonthlySettlementHandler(d.MonthlyRepo)
	spotTrendH := handler.NewSpotTrendHandler(d.SpotTrendRepo)
	daReviewH := handler.NewDATradeReviewHandler(d.DAReviewRepo)
	weatherH := handler.NewWeatherHandler(d.WeatherRepo)
	rpaH := handler.NewRPAHandler(d.RPARepo)
	contractPriceH := handler.NewContractPriceHandler(d.ContractPriceRepo)
	intentH := handler.NewIntentCustomerHandler(d.IntentRepo)
	custLoadH := handler.NewCustomerLoadHandler(d.CustLoadRepo)
	loadDiagH := handler.NewLoadDiagnosisHandler(d.LoadDiagRepo)
	touH := handler.NewTOUHandler(d.TOURepo)
	gridAgencyH := handler.NewGridAgencyHandler(d.GridAgencyRepo)
	storageDeclH := handler.NewStorageDeclHandler(d.StorageDeclRepo)
	custProfitH := handler.NewCustomerProfitHandler(d.CustProfitRepo)
	mtrReviewH := handler.NewMonthlyTradeReviewHandler(d.MTRReviewRepo)
	matchQuoteH := handler.NewMatchQuoteHandler(d.MatchQuoteRepo)
	manualH := handler.NewMonthlyManualHandler(d.ManualRepo)
	exportH := handler.NewExportHandler(d.CustomerRepo, d.MonthlyRepo, d.ContractPriceRepo)
	importH := handler.NewImportHandler(d.CustomerRepo)

	// SSE 推送中心（全进程单例，挂到 Deps 上让别处 publish）
	sseH := handler.NewSSEHandler(d.SSEHub, d.JWT)
	wsH := handler.NewWebSocketHandler(d.JWT)
	attachmentH := handler.NewAttachmentHandler(d.AttachmentRepo, d.ObjectStore)
	approvalH := handler.NewApprovalHandler(d.ApprovalRepo, d.ApprovalReg, d.SSEHub)
	securityH := handler.NewSecurityHandler(d.Pool)
	retailMonthlyH := handler.NewRetailMonthlyHandler(d.RetailMonthlyRepo)
	preSettleH := handler.NewPreSettleHandler(d.PreSettleRepo)
	forecastBaseH := handler.NewForecastBaseHandler(d.ForecastBaseRepo)
	contractPdfH := handler.NewContractPDFHandler(d.RetailRepo, d.AttachmentRepo, d.ObjectStore)
	totalLoadH := handler.NewTotalLoadHandler(d.TotalLoadRepo)
	mediumH := handler.NewMediumForecastHandler(d.MediumForecastRepo)
	accuracyH := handler.NewAccuracyHandler(d.AccuracyRepo)
	mechH := handler.NewMechanismEnergyHandler(d.MechEnergyRepo)
	marketH := handler.NewMarketAnalysisHandler(d.MarketAnalysisRepo)
	settingsH := handler.NewSettingsHandler(d.SettingsRepo)
	loadImportH := handler.NewLoadImportHandler(d.LoadRepo, d.CustomerRepo)
	// 文档解析管线：仓储 + 异步解析 worker + 确认入库映射器
	docRepo := db.NewDocumentRepository(d.Pool)
	docImporters := &document.Importers{Customers: d.CustomerRepo, Intent: d.IntentRepo, Load: d.LoadRepo, Monthly: d.MonthlyRepo, CustomerEnergy: d.CustomerEnergyRepo, Policy: d.PolicyRepo, Retail: d.RetailRepo}
	docWorker := document.NewWorker(docRepo, d.Docling, d.ObjectStore, docImporters)
	documentH := handler.NewDocumentHandler(docRepo, docWorker, docImporters, d.ObjectStore, d.PermSvc)

	// 新增模块 handler
	contractProgressH := handler.NewContractProgressHandler(d.ContractProgressRepo)
	deviationH := handler.NewDeviationHandler(d.DeviationRepo)
	greenPowerH := handler.NewGreenPowerHandler(d.GreenPowerRepo)
	rollingTradeH := handler.NewRollingTradeHandler(d.RollingTradeRepo)
	spotMarketH := handler.NewSpotMarketHandler(d.SpotMarketRepo)
	vppH := handler.NewVPPHandler(d.VPPRepo)
	biddingH := handler.NewBiddingHandler(d.BiddingRepo)
	bidStubH := handler.NewBidStubHandler()
	tradeStubH := handler.NewTradeStubHandler()
	miscStubH := handler.NewMiscStubHandler()
	loadCharH := handler.NewLoadCharacteristicsHandler(d.LoadCharRepo)
	custAnalysisH := handler.NewCustomerAnalysisHandler(d.CustAnalysisRepo)
	tradeStrategyH := handler.NewTradeStrategyHandler(d.TradeStrategyRepo)
	agentH := handler.NewAgentHandler(d.AgentRepo)
	bondH := handler.NewBondHandler(d.BondRepo)
	solarH := handler.NewSolarHandler(d.SolarRepo)
	daSimH := handler.NewDASimulationHandler(d.DASimRepo)
	marketDataH := handler.NewMarketDataHandler(d.MarketDataRepo)
	carbonH := handler.NewCarbonHandler(d.CarbonRepo)

	// 模块关联增强：自定义字段 / 标签 / 交易规则 / 全局搜索 / 联动操作
	customFieldH := handler.NewCustomFieldHandler(d.CustomFieldRepo)
	tagH := handler.NewTagHandler(d.TagRepo)
	tradeRuleH := handler.NewTradeRuleHandler(db.NewTradeRuleRepository(d.Pool))
	customerEnergyH := handler.NewCustomerEnergyHandler(d.CustomerEnergyRepo)
	policyH := handler.NewPolicyHandler(d.PolicyRepo)
	searchH := handler.NewSearchHandler(d.Pool)
	integrationH := handler.NewIntegrationHandler(d.Pool, docRepo, d.RetailRepo, d.CustomerRepo, d.IntentRepo)

	reqUMRead := middleware.RequirePermission(d.PermSvc, "user_management:read")
	reqUMWrite := middleware.RequirePermission(d.PermSvc, "user_management:write")
	reqUMDelete := middleware.RequirePermission(d.PermSvc, "user_management:delete")
	reqCMRead := middleware.RequirePermission(d.PermSvc, "customer_management:read")
	reqCMWrite := middleware.RequirePermission(d.PermSvc, "customer_management:write")
	reqCMDelete := middleware.RequirePermission(d.PermSvc, "customer_management:delete")
	reqRMRead := middleware.RequirePermission(d.PermSvc, "retail_management:read")
	reqRMWrite := middleware.RequirePermission(d.PermSvc, "retail_management:write")
	reqRMDelete := middleware.RequirePermission(d.PermSvc, "retail_management:delete")
	reqLMRead := middleware.RequirePermission(d.PermSvc, "load_management:read")
	reqLMWrite := middleware.RequirePermission(d.PermSvc, "load_management:write")
	reqPMRead := middleware.RequirePermission(d.PermSvc, "price_management:read")
	reqPMWrite := middleware.RequirePermission(d.PermSvc, "price_management:write")
	reqSMRead := middleware.RequirePermission(d.PermSvc, "settlement_management:read")
	reqSMWrite := middleware.RequirePermission(d.PermSvc, "settlement_management:write")
	reqFRRead := middleware.RequirePermission(d.PermSvc, "freq_regulation:read")
	reqFRWrite := middleware.RequirePermission(d.PermSvc, "freq_regulation:write")
	reqSTRead := middleware.RequirePermission(d.PermSvc, "storage:read")
	reqSTWrite := middleware.RequirePermission(d.PermSvc, "storage:write")
	reqANRead := middleware.RequirePermission(d.PermSvc, "analytics:read")
	reqANWrite := middleware.RequirePermission(d.PermSvc, "analytics:write")
	reqTSRead := middleware.RequirePermission(d.PermSvc, "task_scheduler:read")
	reqTSWrite := middleware.RequirePermission(d.PermSvc, "task_scheduler:write")
	reqSYRead := middleware.RequirePermission(d.PermSvc, "system:read")
	reqSYWrite := middleware.RequirePermission(d.PermSvc, "system:write")
	reqDocRead := middleware.RequirePermission(d.PermSvc, "document_management:read")
	reqDocWrite := middleware.RequirePermission(d.PermSvc, "document_management:write")
	reqDocDelete := middleware.RequirePermission(d.PermSvc, "document_management:delete")

	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/ping", handler.Ping)
		// 登录端点：每个 IP 每分钟最多 30 次，防暴力破解
		apiV1.POST("/auth/login", middleware.LoginRateLimit(30), authHandler.Login)
		// SSE 流：JWT 在 handler 内通过 query 参数解析
		apiV1.GET("/stream/alerts", sseH.Stream)
		// WebSocket 双向（演示用：echo）
		apiV1.GET("/ws/echo", wsH.Echo)

		authed := apiV1.Group("")
		authed.Use(middleware.JWT(d.JWT))
		authed.Use(middleware.Tenant(d.UserRepo.IsOrgMember))
		authed.Use(middleware.Audit(d.AuditWriter))
		{
			authed.GET("/auth/me", authHandler.Me)
			authed.GET("/auth/me/permissions", authHandler.MyPermissions)
			authed.POST("/auth/change-password", authHandler.ChangePassword)
			authed.POST("/auth/logout", authHandler.Logout) // 清除登录 Cookie(P1-8)

			// 仪表盘（任何已登录用户可访问，跨模块 KPI 总览）
			authed.GET("/dashboard/summary", dashboardH.Summary)
			authed.GET("/dashboard/settlement-summary", dashboardH.SettlementSummary)
			authed.GET("/dashboard/series/settlement", dashboardH.SettlementSeries)
			authed.GET("/dashboard/series/freq", dashboardH.FreqSeries)
			authed.GET("/dashboard/config", dashboardH.GetConfig)
			authed.PUT("/dashboard/config", dashboardH.SaveConfig)

			// 用户 / 角色 / 模块 / 权限
			authed.GET("/users", reqUMRead, usersH.List)
			authed.GET("/users/:id", reqUMRead, usersH.Get)
			authed.POST("/users", reqUMWrite, usersH.Create)
			authed.PUT("/users/:id", reqUMWrite, usersH.Update)
			authed.POST("/users/:id/password", reqUMWrite, usersH.ResetPassword)
			authed.PUT("/users/:id/roles", reqUMWrite, usersH.SetRoles)
			authed.PUT("/users/:id/orgs", reqUMWrite, orgH.SetUserOrgs)
			// 组织（省份）管理
			authed.GET("/orgs", reqUMRead, orgH.List)
			authed.POST("/orgs", reqUMWrite, orgH.Create)
			authed.PATCH("/orgs/:id", reqUMWrite, orgH.Update)
			authed.GET("/orgs/:id/members", reqUMRead, orgH.Members)
			authed.GET("/roles", reqUMRead, rolesH.List)
			authed.GET("/roles/:code", reqUMRead, rolesH.Get)
			authed.POST("/roles", reqUMWrite, rolesH.Create)
			authed.PUT("/roles/:code", reqUMWrite, rolesH.Update)
			authed.DELETE("/roles/:code", reqUMDelete, rolesH.Delete)
			authed.PUT("/roles/:code/permissions", reqUMWrite, rolesH.SetPermissions)
			authed.GET("/modules", reqUMRead, modulesH.List)
			authed.GET("/permissions", reqUMRead, permsH.List)

		// 菜单页面可见性管理
		authed.GET("/menu/pages", reqUMRead, menuH.GetAllPages)
		authed.GET("/menu/visible", menuH.GetVisiblePages)
		authed.GET("/menu/roles", reqUMRead, menuH.GetAllPages) // 列出所有角色-页面分配
		authed.GET("/menu/roles/:code", reqUMRead, menuH.GetVisiblePages)
		authed.PUT("/menu/roles/:code", reqUMWrite, menuH.UpdateRolePages)

			// 客户档案
			authed.GET("/customers", reqCMRead, customersH.List)
			authed.GET("/customers/:id", reqCMRead, customersH.Get)
			authed.POST("/customers", reqCMWrite, customersH.Create)
			authed.PUT("/customers/:id", reqCMWrite, customersH.Update)
			authed.DELETE("/customers/:id", reqCMDelete, customersH.Delete)

			// 客户历史电量档案（市场化账单/月度电量入库后在此查看）
			authed.GET("/customer-energy", reqCMRead, customerEnergyH.List)

			// 零售
			authed.GET("/retail/pricing-models", reqRMRead, retailH.ListPricingModels)
			authed.GET("/retail/packages", reqRMRead, retailH.ListPackages)
			authed.POST("/retail/packages", reqRMWrite, retailH.CreatePackage)
			authed.PUT("/retail/packages/:id", reqRMWrite, retailH.UpdatePackage)
			authed.DELETE("/retail/packages/:id", reqRMDelete, retailH.DeletePackage)
			authed.GET("/retail/contracts", reqRMRead, retailH.ListContracts)
			authed.POST("/retail/contracts", reqRMWrite, retailH.CreateContract)
			authed.GET("/retail/contracts/:id", reqRMRead, retailH.GetContract)
			authed.PUT("/retail/contracts/:id", reqRMWrite, retailH.UpdateContract)
			authed.DELETE("/retail/contracts/:id", reqRMDelete, retailH.DeleteContract)

			// 负荷 / 价格 / 结算 / 调频 / 储能
			authed.POST("/load/forecast", reqLMRead, loadH.Forecast)
			authed.POST("/load/demo-data", reqLMWrite, loadH.GenerateDemoData)
			authed.POST("/price/forecast", reqPMRead, priceH.Forecast)
			authed.POST("/price/demo-data", reqPMWrite, priceH.GenerateDemoData)
			authed.GET("/settlement/daily", reqSMRead, settlementH.List)
			authed.GET("/settlement/daily/:date", reqSMRead, settlementH.Get)
			authed.POST("/settlement/demo-data", reqSMWrite, settlementH.GenerateDemoData)
			authed.GET("/freq/clearing", reqFRRead, freqH.List)
			authed.POST("/freq/demo-data", reqFRWrite, freqH.GenerateDemoData)
			authed.GET("/storage/stations", reqSTRead, storageH.ListStations)
			authed.GET("/storage/stations/:id/operations", reqSTRead, storageH.ListOperations)
			authed.POST("/storage/demo-data", reqSTWrite, storageH.GenerateDemoData)

			// 客户分析
			authed.GET("/analytics/alerts/stats", reqANRead, analyticsH.Stats)
			authed.GET("/analytics/alerts", reqANRead, analyticsH.ListAlerts)
			authed.POST("/analytics/alerts/:id/ack", reqANWrite, analyticsH.AckAlert)
			authed.GET("/analytics/characteristics", reqANRead, analyticsH.ListCharacteristics)
			authed.POST("/analytics/demo-data", reqANWrite, analyticsH.GenerateDemoData)

			// 任务调度
			authed.GET("/scheduler/jobs", reqTSRead, schedulerH.ListJobs)
			authed.GET("/scheduler/runs", reqTSRead, schedulerH.ListRuns)
			authed.POST("/scheduler/jobs/:id/trigger", reqTSWrite, schedulerH.Trigger)
			authed.PUT("/scheduler/jobs/:id/enabled", reqTSWrite, schedulerH.SetEnabled)

			// 操作审计
			authed.GET("/audit/logs", reqSYRead, auditH.List)

			// 文档解析管线：原件/解析件存档 + 结构化提取 + 人工确认入库（按活跃省隔离）
			authed.POST("/documents", reqDocWrite, documentH.Upload)
			authed.GET("/documents", reqDocRead, documentH.List)
			authed.GET("/documents/:id", reqDocRead, documentH.Get)
			authed.GET("/documents/:id/original", reqDocRead, documentH.Original)
			authed.GET("/documents/:id/parsed", reqDocRead, documentH.Parsed)
			authed.POST("/documents/:id/extractions", reqDocWrite, documentH.AddExtraction)
			authed.PUT("/documents/:id/extractions/:eid", reqDocWrite, documentH.UpdateExtraction)
			authed.DELETE("/documents/:id/extractions/:eid", reqDocWrite, documentH.DeleteExtraction)
			authed.POST("/documents/:id/reparse", reqDocWrite, documentH.Reparse)
			authed.POST("/documents/:id/apply", reqDocWrite, documentH.Apply)
			authed.DELETE("/documents/:id", reqDocDelete, documentH.Delete)

			// 政策文件库（文档中心归档：解析「确认入库 → 政策文件」或手动新增）
			authed.GET("/policies", reqDocRead, policyH.List)
			authed.POST("/policies", reqDocWrite, policyH.Create)
			authed.DELETE("/policies/:id", reqDocDelete, policyH.Delete)

			// 月度结算（沿用结算模块权限）
			authed.GET("/settlement/monthly", reqSMRead, monthlyH.List)
			authed.POST("/settlement/monthly/demo-data", reqSMWrite, monthlyH.GenerateDemoData)

			// 现货价格趋势（沿用价格模块权限）
			authed.GET("/price/trend/daily", reqPMRead, spotTrendH.DailyAvg)
			authed.GET("/price/trend/hourly", reqPMRead, spotTrendH.HourlyAvg)

			// 日前交易复盘（沿用价格模块权限）
			authed.GET("/trade/da-review", reqPMRead, daReviewH.List)
			authed.POST("/trade/da-review/demo-data", reqPMWrite, daReviewH.GenerateDemoData)

			// 气象（沿用负荷模块权限：气象 → 负荷预测的输入）
			authed.GET("/weather", reqLMRead, weatherH.List)
			authed.POST("/weather/demo-data", reqLMWrite, weatherH.GenerateDemoData)

			// 外部气象观测（风电场风速 / 水库水文，原市场行情，现并入气象数据）
			authed.GET("/weather/wind-farm", reqLMRead, weatherH.WindFarm)
			authed.GET("/weather/hydrology", reqLMRead, weatherH.Hydrology)
			authed.POST("/weather/obs-demo-data", reqLMWrite, weatherH.GenObsDemo)

			// ===== 契约对齐补端点（对齐前端实际调用路径；空数据时返回形状正确的空态）=====
			// 气象站点/实况/预报
			authed.GET("/weather/locations", reqLMRead, weatherH.ListLocations)
			authed.POST("/weather/locations", reqLMWrite, weatherH.CreateLocation)
			authed.PUT("/weather/locations/:id", reqLMWrite, weatherH.UpdateLocation)
			authed.DELETE("/weather/locations/:id", reqLMWrite, weatherH.DeleteLocation)
			authed.GET("/weather/actuals", reqLMRead, weatherH.Actuals)
			authed.GET("/weather/actuals/summary", reqLMRead, weatherH.ActualsSummary)
			authed.GET("/weather/forecasts", reqLMRead, weatherH.Forecasts)
			authed.GET("/weather/forecasts/summary", reqLMRead, weatherH.ForecastsSummary)
			authed.GET("/weather/forecast-dates", reqLMRead, weatherH.ForecastDates)
			// 预测基础数据
			authed.GET("/forecast-base-data/availability", reqLMRead, forecastBaseH.Availability)
			authed.POST("/forecast-base-data/curves", reqLMRead, forecastBaseH.Curves)
			// 日前竞价（空态占位，待功能立项）
			authed.GET("/bid/trade-sources", reqPMRead, bidStubH.TradeSources)
			authed.GET("/bid/trade-sources/:id", reqPMRead, bidStubH.TradeSourceDetail)
			authed.POST("/bid/trade-sources/auto", reqPMWrite, bidStubH.CreateTradeSource)
			authed.POST("/bid/trade-sources/manual", reqPMWrite, bidStubH.CreateTradeSource)
			authed.PUT("/bid/trade-sources/:id", reqPMWrite, bidStubH.UpdateTradeSource)
			authed.POST("/bid/trade-sources/:id/status", reqPMWrite, bidStubH.SetTradeSourceStatus)
			authed.DELETE("/bid/trade-sources/:id", reqPMWrite, bidStubH.DeleteTradeSource)
			authed.GET("/bid/simulations/next-day", reqPMRead, bidStubH.NextDaySimulation)
			authed.POST("/bid/simulations/manual-save", reqPMWrite, bidStubH.ManualSimulation)
			authed.POST("/bid/simulations/manual-reset", reqPMWrite, bidStubH.ManualSimulation)
			authed.GET("/bid/analysis/summary", reqPMRead, bidStubH.ProfitSummary)
			authed.GET("/bid/analysis/profit-curve", reqPMRead, bidStubH.ProfitCurve)
			authed.GET("/bid/analysis/daily", reqPMRead, bidStubH.ProfitDaily)
			authed.GET("/bid/analysis/daily-review", reqPMRead, bidStubH.DailyReview)
			authed.GET("/bid/analysis/daily-review/:date", reqPMRead, bidStubH.DailyReview)
			// 月度复盘 / 滚动交易 / 交易策略 分析子视图（空态占位）
			authed.GET("/trade/monthly-review/overview", reqPMRead, tradeStubH.MonthlyOverview)
			authed.GET("/trade/monthly-review/detail", reqPMRead, tradeStubH.MonthlyDetail)
			authed.GET("/trade/monthly-review/contract-details", reqPMRead, tradeStubH.MonthlyContractDetails)
			authed.GET("/trade/monthly-review/contract-earnings", reqPMRead, tradeStubH.MonthlyContractEarnings)
			authed.POST("/trade/monthly-review/recalculate", reqPMWrite, tradeStubH.MonthlyRecalculate)
			authed.GET("/trade/rolling/list", reqPMRead, tradeStubH.RollingList)
			authed.GET("/trade/rolling/statistics", reqPMRead, tradeStubH.RollingStatistics)
			authed.GET("/trade/strategies/contract-curve", reqPMRead, tradeStubH.StrategyContractCurve)
			authed.GET("/trade/strategies/d2", reqPMRead, tradeStubH.StrategyD2)
			authed.GET("/trade/strategies/monthly", reqPMRead, tradeStubH.StrategyMonthly)
			// 日前复盘子视图 / 批发月结算 / 合同电价趋势 / 调频补偿 / 客户分析视图（空态占位）
			authed.GET("/trade/da-review/overview", reqPMRead, miscStubH.EmptyContainer)
			authed.GET("/trade/da-review/detail", reqPMRead, miscStubH.EmptyContainer)
			authed.GET("/trade/da-review/day-ahead", reqPMRead, miscStubH.EmptyContainer)
			authed.GET("/trade/da-review/operation-detail", reqPMRead, miscStubH.EmptyContainer)
			authed.GET("/trade/da-review/trade-dates", reqPMRead, miscStubH.TradeDates)
			authed.GET("/wholesale-monthly-settlement", reqSMRead, miscStubH.EmptyContainer)
			authed.GET("/wholesale-monthly-settlement/year", reqSMRead, miscStubH.EmptyContainer)
			authed.GET("/wholesale-monthly-settlement/years", reqSMRead, miscStubH.SettlementYears)
			authed.POST("/wholesale-monthly-settlement/import", reqSMWrite, miscStubH.EmptyImport)
			authed.GET("/contract-price-trend/curve-analysis", reqPMRead, miscStubH.EmptyContainer)
			authed.GET("/contract-price-trend/price-trend", reqPMRead, miscStubH.EmptyContainer)
			authed.GET("/contract-price-trend/quantity-structure", reqPMRead, miscStubH.EmptyContainer)
			authed.GET("/contract-price/daily-summary", reqPMRead, miscStubH.EmptyContainer)
			authed.GET("/freq-comp-fee", reqFRRead, miscStubH.EmptyContainer)
			authed.POST("/freq-comp-fee/import", reqFRWrite, miscStubH.EmptyImport)
			authed.GET("/analytics/customer-load", reqANRead, miscStubH.EmptyContainer)
			authed.GET("/load-characteristics/overview/scatter-data", reqLMRead, miscStubH.EmptyContainer)
			authed.GET("/customer-profit-analysis/dashboard", reqANRead, miscStubH.ProfitDashboard)
			// ===== 契约对齐补端点 END =====

			// RPA 监控（沿用任务调度权限）
			authed.GET("/rpa/jobs", reqTSRead, rpaH.ListJobs)
			authed.GET("/rpa/runs", reqTSRead, rpaH.ListRuns)
			authed.POST("/rpa/demo-data", reqTSWrite, rpaH.GenerateDemoData)

			// 合同电价日维度（沿用零售模块权限）
			authed.GET("/retail/price-daily", reqRMRead, contractPriceH.List)
			authed.POST("/retail/price-daily/demo-data", reqRMWrite, contractPriceH.GenerateDemoData)

			// E1 意向客户（沿用客户模块权限）
			authed.GET("/intent-customers", reqCMRead, intentH.List)
			authed.GET("/intent-customers/diagnose", reqCMRead, intentH.Diagnose)
			authed.POST("/intent-customers/demo-data", reqCMWrite, intentH.GenerateDemoData)

			// E2 客户负荷分析（沿用客户分析模块权限）
			authed.GET("/analytics/customer-load/summary", reqANRead, custLoadH.Summary)
			authed.GET("/analytics/customer-load/:id/curve", reqANRead, custLoadH.LatestCurve)

			// E3 负荷数据诊断（沿用负荷模块权限）
			authed.GET("/load/diagnosis", reqLMRead, loadDiagH.List)

			// E4 TOU 时段规则（沿用价格模块权限）
			authed.GET("/price/tou-rules", reqPMRead, touH.List)
			authed.POST("/price/tou-rules/demo-data", reqPMWrite, touH.GenerateDemoData)

			// E5 电网代理价（沿用价格模块权限）
			authed.GET("/price/grid-agency", reqPMRead, gridAgencyH.List)
			authed.POST("/price/grid-agency/demo-data", reqPMWrite, gridAgencyH.GenerateDemoData)

			// E6 储能申报（沿用储能模块权限）
			authed.GET("/storage/declarations", reqSTRead, storageDeclH.List)
			authed.POST("/storage/declarations/demo-data", reqSTWrite, storageDeclH.GenerateDemoData)

			// F1 客户利润（沿用客户分析权限）
			authed.GET("/analytics/customer-profit", reqANRead, custProfitH.List)
			authed.POST("/analytics/customer-profit/demo-data", reqANWrite, custProfitH.GenerateDemoData)

			// F2 月度交易复盘（沿用价格模块权限）
			authed.GET("/trade/monthly-review", reqPMRead, mtrReviewH.List)
			authed.POST("/trade/monthly-review/demo-data", reqPMWrite, mtrReviewH.GenerateDemoData)

			// F3 撮合报价（沿用价格模块权限）
			authed.GET("/trade/match-quotes", reqPMRead, matchQuoteH.List)
			authed.POST("/trade/match-quotes/demo-data", reqPMWrite, matchQuoteH.GenerateDemoData)

			// F4 月度手工数据（沿用结算模块权限）
			authed.GET("/settlement/manual-data", reqSMRead, manualH.List)
			authed.POST("/settlement/manual-data", reqSMWrite, manualH.Create)
			authed.POST("/settlement/manual-data/demo-data", reqSMWrite, manualH.GenerateDemoData)

			// M1 数据导出（需 system:read，防任意用户批量导出敏感数据）
			authed.GET("/export/:resource", reqSYRead, exportH.Resource)

			// M2 SSE 测试推送（仅 system:read 即可触发，便于运维联调）
			authed.POST("/stream/test", reqSYRead, sseH.PublishTest)

			// T4 在线用户（任何已登录用户可查）
			authed.GET("/online", sseH.Online)

			// N1 数据导入（写权限）+ 限流：每用户 30 秒 1 次（rps=1/30=0.033，burst=2）
			authed.POST("/import/customers", reqCMWrite, middleware.StrictRateLimit(0.033, 2), importH.Customers)

			// N2 附件（读需 system:read 防签名 URL 泄漏；写/删需 customer_management 权限）
			// 上传严格限流（防滥用上传大文件）：每用户 5 秒 1 次
			authed.GET("/attachments", reqSYRead, attachmentH.List)
			authed.GET("/attachments/:id/url", reqSYRead, attachmentH.DownloadURL)
			authed.POST("/attachments", reqCMWrite, middleware.StrictRateLimit(0.2, 3), attachmentH.Upload)
			authed.DELETE("/attachments/:id", reqCMDelete, attachmentH.Delete)

			// N3 审批流（读/发起需 system:read；审批权用 system:write）
			authed.GET("/approvals", reqSYRead, approvalH.List)
			authed.GET("/approvals/templates", reqSYRead, approvalH.ListTemplates)
			authed.GET("/approvals/by-resource", reqSYRead, approvalH.ByResource)
			authed.GET("/approvals/:id", reqSYRead, approvalH.Get)
			authed.POST("/approvals", reqSYRead, approvalH.Submit)
			authed.POST("/approvals/:id/approve", middleware.RequirePermission(d.PermSvc, "system:write"), approvalH.Approve)
			authed.POST("/approvals/:id/reject", middleware.RequirePermission(d.PermSvc, "system:write"), approvalH.Reject)
			authed.POST("/approvals/:id/withdraw", reqSYRead, approvalH.Withdraw)

			// R6 安全大屏（system:read）
			authed.GET("/system/security/overview", reqSYRead, securityH.Overview)

			// U1 零售月度结算
			authed.GET("/retail/monthly-settlement", reqRMRead, retailMonthlyH.List)
			authed.POST("/retail/monthly-settlement/demo-data", reqRMWrite, retailMonthlyH.GenerateDemoData)

			// U2 预结算
			authed.GET("/settlement/pre", reqSMRead, preSettleH.List)
			authed.GET("/settlement/pre/:date", reqSMRead, preSettleH.Get)
			authed.POST("/settlement/pre/demo-data", reqSMWrite, preSettleH.GenerateDemoData)

			// U3 预测基础数据
			authed.GET("/load/holidays", reqLMRead, forecastBaseH.ListHolidays)
			authed.GET("/load/typical-curves", reqLMRead, forecastBaseH.ListCurves)
			authed.POST("/load/base-data/demo-data", reqLMWrite, forecastBaseH.GenerateDemoData)

			// U4 合同 PDF 生成
			authed.POST("/retail/contracts/:id/pdf", reqRMWrite, contractPdfH.Generate)

			// V1 系统总负荷
			authed.GET("/load/total", reqLMRead, totalLoadH.List)
			authed.POST("/load/total/demo-data", reqLMWrite, totalLoadH.GenerateDemoData)

			// V2 中期负荷预测
			authed.GET("/load/medium-forecast", reqLMRead, mediumH.List)
			authed.POST("/load/medium-forecast/demo-data", reqLMWrite, mediumH.GenerateDemoData)

			// V3 预测准确率（system:read）
			authed.GET("/forecast/accuracy", reqSYRead, accuracyH.List)
			authed.GET("/forecast/accuracy/summary", reqSYRead, accuracyH.Summary)
			authed.POST("/forecast/accuracy/demo-data", reqLMWrite, accuracyH.GenerateDemoData)

			// V4 机制电量（沿用结算权限）
			authed.GET("/settlement/mechanism-energy", reqSMRead, mechH.List)
			authed.POST("/settlement/mechanism-energy/demo-data", reqSMWrite, mechH.GenerateDemoData)

			// V5 市场分析（沿用价格权限）
			authed.GET("/price/market-analysis", reqPMRead, marketH.List)
			authed.POST("/price/market-analysis/demo-data", reqPMWrite, marketH.GenerateDemoData)

			// W2 系统配置（system:read 查看，system:write 编辑）
			authed.GET("/system/settings", reqSYRead, settingsH.List)
			authed.PUT("/system/settings/:key", middleware.RequirePermission(d.PermSvc, "system:write"), settingsH.Update)

			// W3 电表数据导入
			authed.POST("/import/load", reqLMWrite, middleware.StrictRateLimit(0.033, 2), loadImportH.Load)

			// G1 签约进度跟踪（沿用零售模块权限）
			authed.GET("/retail/contract-progress", reqRMRead, contractProgressH.List)
			authed.POST("/retail/contract-progress", reqRMWrite, contractProgressH.Create)
			authed.POST("/retail/contract-progress/demo-data", reqRMWrite, contractProgressH.GenerateDemoData)

			// G2 偏差结算（沿用结算模块权限）
			authed.GET("/settlement/deviation", reqSMRead, deviationH.List)
			authed.GET("/settlement/deviation/summary", reqSMRead, deviationH.Summary)
			authed.POST("/settlement/deviation/demo-data", reqSMWrite, deviationH.GenerateDemoData)

			// G3 绿电交易（沿用零售模块权限）
			authed.GET("/trade/green-power", reqRMRead, greenPowerH.List)
			authed.POST("/trade/green-power/demo-data", reqRMWrite, greenPowerH.GenerateDemoData)

			// G4 滚动撮合交易（沿用价格模块权限）
			authed.GET("/trade/rolling", reqPMRead, rollingTradeH.List)
			authed.GET("/trade/rolling/rounds", reqPMRead, rollingTradeH.List)
			authed.GET("/trade/rolling/period-history", reqPMRead, rollingTradeH.List)
			authed.POST("/trade/rolling/demo-data", reqPMWrite, rollingTradeH.GenerateDemoData)

			// G4a 撮合报价补充路由
			authed.GET("/trade/match-quotes/days", reqPMRead, matchQuoteH.List)
			authed.GET("/trade/match-quotes/quotes", reqPMRead, matchQuoteH.List)

			// G5 现货市场（沿用价格模块权限）
			authed.GET("/price/spot-market", reqPMRead, spotMarketH.List)
			authed.POST("/price/spot-market/demo-data", reqPMWrite, spotMarketH.GenerateDemoData)

			// G6 虚拟电厂（沿用储能模块权限）
			authed.GET("/vpp/resources", reqSTRead, vppH.ListResources)
			authed.GET("/vpp/dispatches", reqSTRead, vppH.ListDispatches)
			authed.POST("/vpp/demo-data", reqSTWrite, vppH.GenerateDemoData)

			// G7 竞价管理（沿用价格模块权限）
			authed.GET("/trade/bidding", reqPMRead, biddingH.List)
			authed.POST("/trade/bidding", reqPMWrite, biddingH.Create)
			authed.POST("/trade/bidding/demo-data", reqPMWrite, biddingH.GenerateDemoData)

			// G8 负荷特性分析（沿用负荷模块权限）
			authed.GET("/load/characteristics", reqLMRead, loadCharH.List)
			authed.POST("/load/characteristics/demo-data", reqLMWrite, loadCharH.GenerateDemoData)

			// G9 客户分析（沿用客户分析权限）
			authed.GET("/analytics/customer-analysis", reqANRead, custAnalysisH.List)
			authed.POST("/analytics/customer-analysis/demo-data", reqANWrite, custAnalysisH.GenerateDemoData)

			// G10 交易策略（沿用价格模块权限）
			authed.GET("/trade/strategies", reqPMRead, tradeStrategyH.List)
			authed.POST("/trade/strategies", reqPMWrite, tradeStrategyH.Create)
			authed.POST("/trade/strategies/demo-data", reqPMWrite, tradeStrategyH.GenerateDemoData)

			// H1 代理商管理（沿用客户管理权限）
			authed.GET("/agents", reqCMRead, agentH.List)
			authed.GET("/agents/:id", reqCMRead, agentH.Get)
			authed.POST("/agents", reqCMWrite, agentH.Create)
			authed.PUT("/agents/:id", reqCMWrite, agentH.Update)
			authed.DELETE("/agents/:id", reqCMDelete, agentH.Delete)
			authed.GET("/agents/:id/customers", reqCMRead, agentH.ListCustomers)

			// H2 保函管理（沿用客户管理权限）
			authed.GET("/bonds", reqCMRead, bondH.List)
			authed.GET("/bonds/:id", reqCMRead, bondH.Get)
			authed.POST("/bonds", reqCMWrite, bondH.Create)
			authed.PUT("/bonds/:id", reqCMWrite, bondH.Update)
			authed.DELETE("/bonds/:id", reqCMDelete, bondH.Delete)

			// P2 光伏系列（沿用储能模块权限）
			authed.GET("/solar/stations", reqSTRead, solarH.ListStations)
			authed.GET("/solar/stations/:id", reqSTRead, solarH.GetStation)
			authed.POST("/solar/stations", reqSTWrite, solarH.CreateStation)
			authed.PUT("/solar/stations/:id", reqSTWrite, solarH.UpdateStation)
			authed.DELETE("/solar/stations/:id", reqSTWrite, solarH.DeleteStation)
			authed.GET("/solar/forecast", reqSTRead, solarH.ListForecast)
			authed.GET("/solar/revenue", reqSTRead, solarH.ListRevenue)
			authed.POST("/solar/demo-data", reqSTWrite, solarH.GenerateDemoData)

			// P2-4 日前模拟（沿用价格模块权限）
			authed.GET("/trade/da-simulation", reqPMRead, daSimH.List)
			authed.POST("/trade/da-simulation", reqPMWrite, daSimH.Create)
			authed.GET("/trade/da-simulation/:id", reqPMRead, daSimH.Get)
			authed.POST("/trade/da-simulation/:id/run", reqPMWrite, daSimH.RunSimulation)
			authed.DELETE("/trade/da-simulation/:id", reqPMWrite, daSimH.Delete)
			authed.POST("/trade/da-simulation/demo-data", reqPMWrite, daSimH.GenerateDemoData)

			// Q1 市场行情数据（沿用价格模块权限）
			authed.GET("/market-data", reqPMRead, marketDataH.Overview)
			authed.GET("/market-data/overview", reqPMRead, marketDataH.Overview)
			authed.GET("/market-data/tables", reqPMRead, marketDataH.ListTables)
			authed.GET("/market-data/:table", reqPMRead, marketDataH.Query)
			// 碳交易行情（CEA/CCER/EUA，全国统一行情，不分省）
			authed.GET("/carbon/summary", reqPMRead, carbonH.Summary)
			authed.GET("/carbon/quotes", reqPMRead, carbonH.List)
				authed.POST("/carbon/demo-data", reqPMWrite, carbonH.GenerateDemoData)

				// ─── 模块关联增强路由 ──────────────────────────────────
				// 自定义字段管理
				authed.GET("/custom-fields", reqSYRead, customFieldH.List)
				authed.POST("/custom-fields", reqSYWrite, customFieldH.Create)
				authed.PUT("/custom-fields/:id", reqSYWrite, customFieldH.Update)
				authed.DELETE("/custom-fields/:id", reqSYWrite, customFieldH.Delete)

				// 标签管理
				authed.GET("/tags", reqSYRead, tagH.List)
				authed.POST("/tags", reqSYWrite, tagH.Create)
				authed.PUT("/tags/:id", reqSYWrite, tagH.Update)
				authed.DELETE("/tags/:id", reqSYWrite, tagH.Delete)
				authed.POST("/tags/batch-apply", reqSYWrite, tagH.BatchApply)
				authed.GET("/tags/entity", reqSYRead, tagH.GetEntityTags)

				// 交易规则
				authed.GET("/trade-rules", reqSMRead, tradeRuleH.List)
				authed.GET("/trade-rules/export", reqSMRead, tradeRuleH.Export)
				authed.POST("/trade-rules", reqSMWrite, tradeRuleH.Create)
				authed.PUT("/trade-rules/:id", reqSMWrite, tradeRuleH.Update)
				authed.DELETE("/trade-rules/:id", reqSMWrite, tradeRuleH.Delete)

				// 全局搜索
				authed.GET("/search", reqSYRead, searchH.Search)

				// 客户 360 视图
				authed.GET("/customers/:id/360", reqCMRead, customersH.View360)

				// 文档→合同自动填充
				authed.POST("/documents/:id/apply-to-contract", reqRMWrite, integrationH.ApplyToContract)

				// 意向客户转正
				authed.POST("/intent-customers/:id/convert", reqCMWrite, integrationH.ConvertIntentCustomer)
				}
				}

				return r
				}
