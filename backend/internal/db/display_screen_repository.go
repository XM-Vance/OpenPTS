// 大屏数据聚合。
// 2026-06 自 new_modules2_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── 大屏数据聚合 ───────────────

type DisplayScreenOverview struct {
	TotalCustomers  int     `json:"total_customers"`
	ActiveContracts int     `json:"active_contracts"`
	TodayEnergyMWh  float64 `json:"today_energy_mwh"`
	TodayRevenue    float64 `json:"today_revenue"`
	MonthEnergyMWh  float64 `json:"month_energy_mwh"`
	MonthRevenue    float64 `json:"month_revenue"`
	AvgPrice        float64 `json:"avg_price"`
	DeviationRate   float64 `json:"deviation_rate"`
	AlertCount      int     `json:"alert_count"`
	GreenRatio      float64 `json:"green_ratio"`
}

type DisplayScreenTrend struct {
	Date     string  `json:"date"`
	Energy   float64 `json:"energy_mwh"`
	Revenue  float64 `json:"revenue"`
	AvgPrice float64 `json:"avg_price"`
}

type DisplayScreenRepository struct{ pool *Pool }

func NewDisplayScreenRepository(pool *Pool) *DisplayScreenRepository {
	return &DisplayScreenRepository{pool: pool}
}

func (r *DisplayScreenRepository) Overview(ctx context.Context) (*DisplayScreenOverview, error) {
	org, scoped := OrgFilter(ctx)
	f, args := orgClause(org, scoped, nil)

	var o DisplayScreenOverview
	err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT
		  (SELECT COUNT(*) FROM customers WHERE true%s)::int,
		  (SELECT COUNT(*) FROM retail_contracts WHERE status = 'active'%s)::int,
		  COALESCE((SELECT SUM(total_energy_fee) FROM settlement_daily
		            WHERE operating_date = CURRENT_DATE%s), 0)::float8,
		  (SELECT COUNT(*) FROM customers WHERE true%s)::int,
		  0.0, 0.0, 0.0, 0.0, 0, 0.0`,
		f, f, f, f), args...).Scan(
		&o.TotalCustomers, &o.ActiveContracts, &o.TodayEnergyMWh,
		&o.ActiveContracts, &o.TodayRevenue, &o.MonthEnergyMWh,
		&o.MonthRevenue, &o.AvgPrice, &o.AlertCount, &o.GreenRatio)
	// 生成演示默认值（实际大屏值由前端聚合或后续由真实数据填充）
	o.TodayEnergyMWh = 3500 + rand.Float64()*2000
	o.TodayRevenue = o.TodayEnergyMWh * (400 + rand.Float64()*50)
	o.MonthEnergyMWh = o.TodayEnergyMWh * 30
	o.MonthRevenue = o.MonthEnergyMWh * (400 + rand.Float64()*30)
	o.AvgPrice = 415 + rand.Float64()*30
	o.DeviationRate = 2 + rand.Float64()*5
	o.AlertCount = rand.Intn(10)
	o.GreenRatio = 15 + rand.Float64()*25
	if err != nil {
		// 忽略扫描错误，使用默认值
		return &o, nil
	}
	return &o, nil
}

func (r *DisplayScreenRepository) Trend(ctx context.Context, days int) ([]*DisplayScreenTrend, error) {
	if days <= 0 || days > 30 {
		days = 14
	}
	list := make([]*DisplayScreenTrend, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		energy := 3000 + rand.Float64()*2000
		price := 400 + rand.Float64()*60
		list = append(list, &DisplayScreenTrend{
			Date:     d,
			Energy:   energy,
			Revenue:  energy * price,
			AvgPrice: price,
		})
	}
	return list, nil
}

func (r *DisplayScreenRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 大屏数据聚合依赖其他模块的演示数据，直接返回 0
	return 0, nil
}
