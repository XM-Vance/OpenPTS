// 网关入口：加载配置 → 初始化日志 → 连接 Postgres → 装配依赖 → 启动 HTTP 服务 → 优雅退出。
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/ptis/backend/internal/docling"
	"github.com/ptis/backend/internal/approval"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/config"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
	"github.com/ptis/backend/internal/scheduler"
	"github.com/ptis/backend/internal/server"
	"github.com/ptis/backend/internal/storage"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("加载配置失败")
	}

	middleware.InitLogger(cfg.LogLevel)
	log.Info().
		Str("env", cfg.Environment).
		Str("port", cfg.Port).
		Msg("PTIS 后端启动中")

	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("连接 Postgres 失败")
	}
	defer pool.Close()
	if cfg.DatabaseReplicaURL != "" {
		if err := pool.WithReplica(ctx, cfg.DatabaseReplicaURL); err != nil {
			log.Warn().Err(err).Msg("只读副本连接失败，回退到主库")
		} else {
			log.Info().Msg("Postgres 主+副本已连接（读分流启用）")
		}
	} else {
		log.Info().Msg("Postgres 已连接（单实例）")
	}

	userRepo := db.NewUserRepository(pool)
	roleRepo := db.NewRoleRepository(pool)
	permRepo := db.NewPermissionRepository(pool)
	modRepo := db.NewModuleRepository(pool)
	customerRepo := db.NewCustomerRepository(pool)
	customerEnergyRepo := db.NewCustomerEnergyRepository(pool)
	policyRepo := db.NewPolicyRepository(pool)
	retailRepo := db.NewRetailRepository(pool)
	loadRepo := db.NewLoadRepository(pool)
	priceRepo := db.NewPriceRepository(pool)
	settlementRepo := db.NewSettlementRepository(pool)
	freqRepo := db.NewFreqRepository(pool)
	storageRepo := db.NewStorageRepository(pool)
	analyticsRepo := db.NewAnalyticsRepository(pool)
	dashboardRepo := db.NewDashboardRepository(pool)
	schedulerRepo := db.NewSchedulerRepository(pool)
	auditRepo := db.NewAuditRepository(pool)
	auditWriter := middleware.NewAuditWriter(auditRepo) // P2-15: 批量异步写审计,关闭时 flush
	monthlyRepo := db.NewMonthlySettlementRepository(pool)
	spotTrendRepo := db.NewSpotTrendRepository(pool)
	daReviewRepo := db.NewDATradeReviewRepository(pool)
	weatherRepo := db.NewWeatherRepository(pool)
	rpaRepo := db.NewRPARepository(pool)
	contractPriceRepo := db.NewContractPriceRepository(pool)
	intentRepo := db.NewIntentCustomerRepository(pool)
	custLoadRepo := db.NewCustomerLoadRepository(pool)
	loadDiagRepo := db.NewLoadDiagnosisRepository(pool)
	touRepo := db.NewTOURepository(pool)
	gridAgencyRepo := db.NewGridAgencyRepository(pool)
	storageDeclRepo := db.NewStorageDeclarationRepository(pool)
	custProfitRepo := db.NewCustomerProfitRepository(pool)
	mtrReviewRepo := db.NewMonthlyTradeReviewRepository(pool)
	matchQuoteRepo := db.NewMatchQuoteRepository(pool)
	manualRepo := db.NewMonthlyManualRepository(pool)
	sseHub := handler.NewSSEHub()
	attachmentRepo := db.NewAttachmentRepository(pool)
	approvalRepo := db.NewApprovalRepository(pool)
	retailMonthlyRepo := db.NewRetailMonthlyRepository(pool)
	preSettleRepo := db.NewPreSettleRepository(pool)
	forecastBaseRepo := db.NewForecastBaseRepository(pool)
	totalLoadRepo := db.NewTotalLoadRepository(pool)
	mediumForecastRepo := db.NewMediumForecastRepository(pool)
	accuracyRepo := db.NewAccuracyRepository(pool)
	mechEnergyRepo := db.NewMechanismEnergyRepository(pool)
	marketAnalysisRepo := db.NewMarketAnalysisRepository(pool)
	settingsRepo := db.NewSettingsRepository(pool)

	// 新增模块仓储（11 个）
	contractProgressRepo := db.NewContractProgressRepository(pool)
	deviationRepo := db.NewDeviationRepository(pool)
	greenPowerRepo := db.NewGreenPowerRepository(pool)
	rollingTradeRepo := db.NewRollingTradeRepository(pool)
	spotMarketRepo := db.NewSpotMarketRepository(pool)
	vppRepo := db.NewVPPRepository(pool)
	biddingRepo := db.NewBiddingRepository(pool)
	loadCharRepo := db.NewLoadCharacteristicsRepository(pool)
	custAnalysisRepo := db.NewCustomerAnalysisRepository(pool)
	tradeStrategyRepo := db.NewTradeStrategyRepository(pool)
	solarRepo := db.NewSolarRepository(pool)
	marketDataRepo := db.NewMarketDataRepository(pool)
	daSimRepo := db.NewDASimulationRepository(pool)
	carbonRepo := db.NewCarbonRepository(pool)

	// 模块关联增强仓储
	customFieldRepo := db.NewCustomFieldRepository(pool)
	tagRepo := db.NewTagRepository(pool)

	// 审批 applier 注册：approve 通过后自动落库到目标表。
	approvalRegistry := approval.NewRegistry()
	// 通用 payload 解码器
	type fieldPayload struct {
		Field string `json:"field"`
		New   string `json:"new"`
	}
	parseFP := func(p json.RawMessage) (fieldPayload, error) {
		var fp fieldPayload
		err := json.Unmarshal(p, &fp)
		return fp, err
	}

	approvalRegistry.Register("retail_contracts", func(ctx context.Context, resourceID string, payload json.RawMessage) error {
		fp, err := parseFP(payload)
		if err != nil || fp.Field == "" {
			return err
		}
		_, err = retailRepo.UpdateContractField(ctx, resourceID, fp.Field, fp.New)
		return err
	})

	// T1 customers applier
	approvalRegistry.Register("customers", func(ctx context.Context, resourceID string, payload json.RawMessage) error {
		fp, err := parseFP(payload)
		if err != nil || fp.Field == "" {
			return err
		}
		_, err = customerRepo.UpdateCustomerField(ctx, resourceID, fp.Field, fp.New)
		return err
	})

	// T2 pricing_models applier（resource_id 是 code，不是 uuid）
	approvalRegistry.Register("pricing_models", func(ctx context.Context, resourceID string, payload json.RawMessage) error {
		fp, err := parseFP(payload)
		if err != nil || fp.Field == "" {
			return err
		}
		_, err = retailRepo.UpdatePricingField(ctx, resourceID, fp.Field, fp.New)
		return err
	})
	objStore, err := storage.New()
	if err != nil {
		log.Warn().Err(err).Msg("MinIO 初始化失败，文件功能将不可用")
	}
	permSvc := auth.NewPermissionService(permRepo)

	// 进程内调度器：注册 4 个内置 handler 后启动。
	sched := scheduler.New(schedulerRepo, pool)
	sched.SetPublisher(sseHub) // 任务执行后广播 SSE
	sched.Register("cleanup_tokens", scheduler.CleanupTokens)
	sched.Register("aggregate_daily_active", scheduler.AggregateDailyActive)
	sched.Register("refresh_dashboard_kpi", scheduler.RefreshDashboardKPI)
	sched.Register("expire_contracts", scheduler.ExpireContracts)
	if err := sched.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("启动调度器失败")
	}

	deps := &server.Deps{
		Config:               cfg,
		Pool:                 pool,
		JWT:                  auth.NewJWTService(cfg.JWTSecret, cfg.JWTTTL),
		UserRepo:             userRepo,
		RoleRepo:             roleRepo,
		PermRepo:             permRepo,
		ModRepo:              modRepo,
		CustomerRepo:         customerRepo,
		CustomerEnergyRepo:   customerEnergyRepo,
		PolicyRepo:           policyRepo,
		RetailRepo:           retailRepo,
		LoadRepo:             loadRepo,
		PriceRepo:            priceRepo,
		SettlementRepo:       settlementRepo,
		FreqRepo:             freqRepo,
		StorageRepo:          storageRepo,
		AnalyticsRepo:        analyticsRepo,
		DashboardRepo:        dashboardRepo,
		SchedulerRepo:        schedulerRepo,
		Scheduler:            sched,
		AuditRepo:            auditRepo,
		AuditWriter:          auditWriter,
		MonthlyRepo:          monthlyRepo,
		SpotTrendRepo:        spotTrendRepo,
		DAReviewRepo:         daReviewRepo,
		WeatherRepo:          weatherRepo,
		RPARepo:              rpaRepo,
		ContractPriceRepo:    contractPriceRepo,
		IntentRepo:           intentRepo,
		CustLoadRepo:         custLoadRepo,
		LoadDiagRepo:         loadDiagRepo,
		TOURepo:              touRepo,
		GridAgencyRepo:       gridAgencyRepo,
		StorageDeclRepo:      storageDeclRepo,
		CustProfitRepo:       custProfitRepo,
		MTRReviewRepo:        mtrReviewRepo,
		MatchQuoteRepo:       matchQuoteRepo,
		ManualRepo:           manualRepo,
		SSEHub:               sseHub,
		AttachmentRepo:       attachmentRepo,
		ApprovalRepo:         approvalRepo,
		ApprovalReg:          approvalRegistry,
		RetailMonthlyRepo:    retailMonthlyRepo,
		PreSettleRepo:        preSettleRepo,
		ForecastBaseRepo:     forecastBaseRepo,
		TotalLoadRepo:        totalLoadRepo,
		MediumForecastRepo:   mediumForecastRepo,
		AccuracyRepo:         accuracyRepo,
		MechEnergyRepo:       mechEnergyRepo,
		MarketAnalysisRepo:   marketAnalysisRepo,
		SettingsRepo:         settingsRepo,
		ContractProgressRepo: contractProgressRepo,
		DeviationRepo:        deviationRepo,
		GreenPowerRepo:       greenPowerRepo,
		RollingTradeRepo:     rollingTradeRepo,
		SpotMarketRepo:       spotMarketRepo,
		VPPRepo:              vppRepo,
		BiddingRepo:          biddingRepo,
		LoadCharRepo:         loadCharRepo,
		CustAnalysisRepo:     custAnalysisRepo,
		TradeStrategyRepo:    tradeStrategyRepo,
		ObjectStore:          objStore,
		PermSvc:              permSvc,
		Docling:              docling.New(cfg.DoclingServiceURL),
		SolarRepo:            solarRepo,
		MarketDataRepo:       marketDataRepo,
		DASimRepo:            daSimRepo,
		CarbonRepo:           carbonRepo,
		CustomFieldRepo:      customFieldRepo,
		TagRepo:              tagRepo,
		AgentRepo:            db.NewAgentRepository(pool),
		BondRepo:             db.NewBondRepository(pool),
	}

	r := server.NewRouter(deps)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Msgf("goroutine panic: %v", r)
			}
		}()
		log.Info().Str("port", cfg.Port).Msg("HTTP 服务监听中")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP 服务异常退出")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("收到关闭信号，准备退出")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sched.Stop()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("强制退出")
	}
	// 先 Shutdown(在途请求已全部投递审计)再 flush 写入器,排空缓冲后退出。
	auditWriter.Stop(shutdownCtx)
	log.Info().Msg("已退出")
}
