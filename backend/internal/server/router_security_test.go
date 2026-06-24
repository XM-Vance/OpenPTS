// 路由安全一致性测试(P1-9):遍历真实 router 的全部 /api 路由,
// 断言除显式公开端点外,匿名请求一律 401——防止"忘挂鉴权中间件"的回归。
// (CI 契约闸只验证路由存在且接受任何非 404,拦不住裸奔的 200。)
// 不需要数据库:鉴权中间件在 handler 之前拒绝请求,nil pool 永不触达。
package server

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ptis/backend/internal/docling"
	"github.com/ptis/backend/internal/approval"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/config"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/scheduler"
	"github.com/gin-gonic/gin"
)

// minimalDeps 按 cmd/server/main.go 的装配方式构造 Deps,但不连库、不启调度器。
func minimalDeps() *Deps {
	var pool *db.Pool // nil:路由注册与鉴权拒绝路径都不触库
	schedRepo := db.NewSchedulerRepository(pool)
	return &Deps{
		Config:               &config.Config{Environment: "test"},
		Pool:                 pool,
		JWT:                  auth.NewJWTService("router-test-secret-at-least-32-chars!!", time.Hour),
		UserRepo:             db.NewUserRepository(pool),
		RoleRepo:             db.NewRoleRepository(pool),
		PermRepo:             db.NewPermissionRepository(pool),
		ModRepo:              db.NewModuleRepository(pool),
		CustomerRepo:         db.NewCustomerRepository(pool),
		CustomerEnergyRepo:   db.NewCustomerEnergyRepository(pool),
		PolicyRepo:           db.NewPolicyRepository(pool),
		RetailRepo:           db.NewRetailRepository(pool),
		LoadRepo:             db.NewLoadRepository(pool),
		PriceRepo:            db.NewPriceRepository(pool),
		SettlementRepo:       db.NewSettlementRepository(pool),
		FreqRepo:             db.NewFreqRepository(pool),
		StorageRepo:          db.NewStorageRepository(pool),
		AnalyticsRepo:        db.NewAnalyticsRepository(pool),
		DashboardRepo:        db.NewDashboardRepository(pool),
		SchedulerRepo:        schedRepo,
		Scheduler:            scheduler.New(schedRepo, pool),
		AuditRepo:            db.NewAuditRepository(pool),
		MonthlyRepo:          db.NewMonthlySettlementRepository(pool),
		SpotTrendRepo:        db.NewSpotTrendRepository(pool),
		DAReviewRepo:         db.NewDATradeReviewRepository(pool),
		WeatherRepo:          db.NewWeatherRepository(pool),
		RPARepo:              db.NewRPARepository(pool),
		ContractPriceRepo:    db.NewContractPriceRepository(pool),
		IntentRepo:           db.NewIntentCustomerRepository(pool),
		CustLoadRepo:         db.NewCustomerLoadRepository(pool),
		LoadDiagRepo:         db.NewLoadDiagnosisRepository(pool),
		TOURepo:              db.NewTOURepository(pool),
		GridAgencyRepo:       db.NewGridAgencyRepository(pool),
		StorageDeclRepo:      db.NewStorageDeclarationRepository(pool),
		CustProfitRepo:       db.NewCustomerProfitRepository(pool),
		MTRReviewRepo:        db.NewMonthlyTradeReviewRepository(pool),
		MatchQuoteRepo:       db.NewMatchQuoteRepository(pool),
		ManualRepo:           db.NewMonthlyManualRepository(pool),
		SSEHub:               handler.NewSSEHub(),
		AttachmentRepo:       db.NewAttachmentRepository(pool),
		ApprovalRepo:         db.NewApprovalRepository(pool),
		ApprovalReg:          approval.NewRegistry(),
		ObjectStore:          nil,
		RetailMonthlyRepo:    db.NewRetailMonthlyRepository(pool),
		PreSettleRepo:        db.NewPreSettleRepository(pool),
		ForecastBaseRepo:     db.NewForecastBaseRepository(pool),
		TotalLoadRepo:        db.NewTotalLoadRepository(pool),
		MediumForecastRepo:   db.NewMediumForecastRepository(pool),
		AccuracyRepo:         db.NewAccuracyRepository(pool),
		MechEnergyRepo:       db.NewMechanismEnergyRepository(pool),
		MarketAnalysisRepo:   db.NewMarketAnalysisRepository(pool),
		SettingsRepo:         db.NewSettingsRepository(pool),
		ContractProgressRepo: db.NewContractProgressRepository(pool),
		DeviationRepo:        db.NewDeviationRepository(pool),
		GreenPowerRepo:       db.NewGreenPowerRepository(pool),
		RollingTradeRepo:     db.NewRollingTradeRepository(pool),
		SpotMarketRepo:       db.NewSpotMarketRepository(pool),
		VPPRepo:              db.NewVPPRepository(pool),
		BiddingRepo:          db.NewBiddingRepository(pool),
		LoadCharRepo:         db.NewLoadCharacteristicsRepository(pool),
		CustAnalysisRepo:     db.NewCustomerAnalysisRepository(pool),
		TradeStrategyRepo:    db.NewTradeStrategyRepository(pool),
		AgentRepo:            db.NewAgentRepository(pool),
		BondRepo:             db.NewBondRepository(pool),
		SolarRepo:            db.NewSolarRepository(pool),
		DASimRepo:            db.NewDASimulationRepository(pool),
		MarketDataRepo:       db.NewMarketDataRepository(pool),
		CarbonRepo:           db.NewCarbonRepository(pool),
		PermSvc:              auth.NewPermissionService(db.NewPermissionRepository(pool)),
		Docling:              docling.New("http://127.0.0.1:1"),
	}
}

// fillParams 把 :param / *wildcard 段替换为占位值(鉴权先于参数解析,任意值均可)。
func fillParams(path string) string {
	segs := strings.Split(path, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, ":") || strings.HasPrefix(s, "*") {
			segs[i] = "test-id"
		}
	}
	return strings.Join(segs, "/")
}

func TestAllAPIRoutesRejectAnonymous(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(minimalDeps())

	// 设计上允许匿名访问的端点(新增公开端点必须在此显式登记并说明理由)
	public := map[string]bool{
		"POST /api/v1/auth/login": true, // 登录本身,受独立限流保护
		"GET /api/v1/ping":        true, // 联通性探活,无敏感信息(router.go 与 login 同段注册)
	}

	routes := r.Routes()
	if len(routes) < 200 {
		t.Fatalf("路由数异常(%d 条),router 注册可能不完整", len(routes))
	}
	checked := 0
	for i, rt := range routes {
		if !strings.HasPrefix(rt.Path, "/api/") {
			continue // /health /metrics /docs 等根路由按设计公开
		}
		key := rt.Method + " " + rt.Path
		if public[key] {
			continue
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(rt.Method, fillParams(rt.Path), nil)
		// 每条路由换一个客户端 IP,避开 per-IP 限流(60rps/120burst)
		req.RemoteAddr = fmt.Sprintf("10.7.%d.%d:5555", i/250, i%250+1)
		r.ServeHTTP(w, req)
		if w.Code != 401 {
			t.Errorf("%-7s %-55s 匿名访问返回 %d,预期 401(是否漏挂鉴权中间件?)", rt.Method, rt.Path, w.Code)
		}
		checked++
	}
	t.Logf("已校验 %d 条 /api 路由的匿名拒绝行为", checked)
}
