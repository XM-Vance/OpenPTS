// Deps：路由注册所需的全部依赖集合（repo / 服务 / 中间件协作者）。
// 由 cmd/server/main.go 装配后注入 NewRouter；各 register_<域>.go 按需取用。
package server

import (
	"github.com/ptis/backend/internal/approval"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/config"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/docling"
	"github.com/ptis/backend/internal/handler"
	"github.com/ptis/backend/internal/middleware"
	"github.com/ptis/backend/internal/scheduler"
	"github.com/ptis/backend/internal/storage"
)

type Deps struct {
	Config               *config.Config
	Pool                 *db.Pool
	JWT                  *auth.JWTService
	UserRepo             *db.UserRepository
	RoleRepo             *db.RoleRepository
	PermRepo             *db.PermissionRepository
	ModRepo              *db.ModuleRepository
	CustomerRepo         *db.CustomerRepository
	CustomerEnergyRepo   *db.CustomerEnergyRepository
	PolicyRepo           *db.PolicyRepository
	RetailRepo           *db.RetailRepository
	LoadRepo             *db.LoadRepository
	PriceRepo            *db.PriceRepository
	SettlementRepo       *db.SettlementRepository
	FreqRepo             *db.FreqRepository
	StorageRepo          *db.StorageRepository
	AnalyticsRepo        *db.AnalyticsRepository
	DashboardRepo        *db.DashboardRepository
	SchedulerRepo        *db.SchedulerRepository
	Scheduler            *scheduler.Scheduler
	AuditRepo            *db.AuditRepository
	MonthlyRepo          *db.MonthlySettlementRepository
	SpotTrendRepo        *db.SpotTrendRepository
	DAReviewRepo         *db.DATradeReviewRepository
	WeatherRepo          *db.WeatherRepository
	RPARepo              *db.RPARepository
	ContractPriceRepo    *db.ContractPriceRepository
	IntentRepo           *db.IntentCustomerRepository
	CustLoadRepo         *db.CustomerLoadRepository
	LoadDiagRepo         *db.LoadDiagnosisRepository
	TOURepo              *db.TOURepository
	GridAgencyRepo       *db.GridAgencyRepository
	StorageDeclRepo      *db.StorageDeclarationRepository
	CustProfitRepo       *db.CustomerProfitRepository
	MTRReviewRepo        *db.MonthlyTradeReviewRepository
	MatchQuoteRepo       *db.MatchQuoteRepository
	ManualRepo           *db.MonthlyManualRepository
	SSEHub               *handler.SSEHub
	AuditWriter          *middleware.AuditWriter
	AttachmentRepo       *db.AttachmentRepository
	ApprovalRepo         *db.ApprovalRepository
	ApprovalReg          *approval.Registry
	ObjectStore          *storage.ObjectStore
	RetailMonthlyRepo    *db.RetailMonthlyRepository
	PreSettleRepo        *db.PreSettleRepository
	ForecastBaseRepo     *db.ForecastBaseRepository
	TotalLoadRepo        *db.TotalLoadRepository
	MediumForecastRepo   *db.MediumForecastRepository
	AccuracyRepo         *db.AccuracyRepository
	MechEnergyRepo       *db.MechanismEnergyRepository
	MarketAnalysisRepo   *db.MarketAnalysisRepository
	SettingsRepo         *db.SettingsRepository
	ContractProgressRepo *db.ContractProgressRepository
	DeviationRepo        *db.DeviationRepository
	GreenPowerRepo       *db.GreenPowerRepository
	RollingTradeRepo     *db.RollingTradeRepository
	SpotMarketRepo       *db.SpotMarketRepository
	VPPRepo              *db.VPPRepository
	BiddingRepo          *db.BiddingRepository
	LoadCharRepo         *db.LoadCharacteristicsRepository
	LoadCharExtRepo      *db.LoadCharacteristicsExtRepository
	LoadDataRepo         *db.LoadDataRepository
	MeterRepo            *db.MeterRepository
	PriceTrendRepo       *db.PriceTrendRepository
	CustAnalysisRepo     *db.CustomerAnalysisRepository
	TradeStrategyRepo    *db.TradeStrategyRepository
	AgentRepo            *db.AgentRepository
	SolarRepo            *db.SolarRepository
	DASimRepo            *db.DASimulationRepository
	MarketDataRepo       *db.MarketDataRepository
	CarbonRepo           *db.CarbonRepository
	PermSvc              *auth.PermissionService
	Docling              *docling.Client
	CustomFieldRepo      *db.CustomFieldRepository
	TagRepo              *db.TagRepository
}
