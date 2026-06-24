// 仪表盘聚合仓储：一条 SQL 出 9 个跨模块 KPI + 两个 14 日趋势序列。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DashboardSummary struct {
	CustomerCount       int      `json:"customer_count"`
	ActiveContracts     int      `json:"active_contracts"`
	ActivePackages      int      `json:"active_packages"`
	PendingAlerts       int      `json:"pending_alerts"`
	CriticalAlerts      int      `json:"critical_alerts"`
	ActiveStations      int      `json:"active_stations"`
	Storage30dRevenue   float64  `json:"storage_30d_revenue"`
	Freq7dRevenue       float64  `json:"freq_7d_revenue"`
	LatestSettlementFee *float64 `json:"latest_settlement_fee,omitempty"`
}

type DailySeriesPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

type DashboardRepository struct {
	pool *Pool
}

func NewDashboardRepository(pool *Pool) *DashboardRepository {
	return &DashboardRepository{pool: pool}
}

// orgClause returns "AND org_id = $N::uuid" and the args slice with org appended,
// or returns the args unchanged if not scoped.
func orgClause(org string, scoped bool, args []any) (string, []any) {
	if !scoped {
		return "", args
	}
	args = append(args, org)
	return fmt.Sprintf(" AND org_id = $%d::uuid", len(args)), args
}

// GetSummary 一条 SQL 跨 7 张表抓 9 个 KPI。
func (r *DashboardRepository) GetSummary(ctx context.Context) (*DashboardSummary, error) {
	org, scoped := OrgFilter(ctx)
	f, args := orgClause(org, scoped, nil)

	q := fmt.Sprintf(`SELECT
		(SELECT COUNT(*) FROM customers WHERE true%s),
		(SELECT COUNT(*) FROM retail_contracts WHERE status = 'active'%s),
		(SELECT COUNT(*) FROM retail_packages WHERE status = 'active'%s),
		(SELECT COUNT(*) FROM customer_anomaly_alerts WHERE NOT acknowledged%s),
		(SELECT COUNT(*) FROM customer_anomaly_alerts
		   WHERE severity = 'critical' AND NOT acknowledged%s),
		(SELECT COUNT(*) FROM storage_stations WHERE status = 'active'%s),
		(SELECT COALESCE(SUM(revenue), 0) FROM storage_daily_operation
		   WHERE operation_date > NOW()::date - INTERVAL '30 days'%s),
		(SELECT COALESCE(SUM(revenue), 0) FROM frequency_regulation_clearing
		   WHERE settlement_date > NOW()::date - INTERVAL '7 days'%s),
		(SELECT total_energy_fee FROM settlement_daily
		   WHERE version = 'PRELIMINARY'%s
		   ORDER BY operating_date DESC LIMIT 1)`,
		f, f, f, f, f, f, f, f, f)

	var s DashboardSummary
	if err := r.pool.QueryRow(ctx, q, args...).Scan(
		&s.CustomerCount, &s.ActiveContracts, &s.ActivePackages,
		&s.PendingAlerts, &s.CriticalAlerts, &s.ActiveStations,
		&s.Storage30dRevenue, &s.Freq7dRevenue, &s.LatestSettlementFee,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSettlementSeries 最近 N 日批发结算总电费（PRELIMINARY 版本）。
func (r *DashboardRepository) GetSettlementSeries(ctx context.Context, days int) ([]*DailySeriesPoint, error) {
	org, scoped := OrgFilter(ctx)
	since := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -days)
	args := []any{since}
	f := ""
	if scoped {
		args = append(args, org)
		f = fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}

	q := fmt.Sprintf(`SELECT operating_date, COALESCE(total_energy_fee, 0)
		FROM settlement_daily
		WHERE operating_date >= $1 AND version = 'PRELIMINARY'%s
		ORDER BY operating_date ASC`, f)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*DailySeriesPoint, 0, days)
	for rows.Next() {
		var p DailySeriesPoint
		if err := rows.Scan(&p.Date, &p.Value); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// GetFreqSeries 最近 N 日调频收益合计（AGC + AVC）。
func (r *DashboardRepository) GetFreqSeries(ctx context.Context, days int) ([]*DailySeriesPoint, error) {
	org, scoped := OrgFilter(ctx)
	since := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -days)
	args := []any{since}
	f := ""
	if scoped {
		args = append(args, org)
		f = fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}

	q := fmt.Sprintf(`SELECT settlement_date, COALESCE(SUM(revenue), 0)
		FROM frequency_regulation_clearing
		WHERE settlement_date >= $1%s
		GROUP BY settlement_date
		ORDER BY settlement_date ASC`, f)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*DailySeriesPoint, 0, days)
	for rows.Next() {
		var p DailySeriesPoint
		if err := rows.Scan(&p.Date, &p.Value); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

/* ── SettlementSummary (仪表盘结算概览) ── */

type KPIResult struct {
	YearlyGrossProfit  float64 `json:"yearly_gross_profit"`
	MonthlyGrossProfit float64 `json:"monthly_gross_profit"`
	PriceSpread        float64 `json:"price_spread"`
	RetailAvgPrice     float64 `json:"retail_avg_price"`
}

type ChartPoint struct {
	Label              string   `json:"label"`
	MonthlyGrossProfit *float64 `json:"monthly_gross_profit,omitempty"`
	YearlyGrossProfit  *float64 `json:"yearly_gross_profit,omitempty"`
	TotalPurchase      *float64 `json:"total_purchase,omitempty"`
	TotalRetail        *float64 `json:"total_retail,omitempty"`
	TotalWholesale     *float64 `json:"total_wholesale,omitempty"`
}

type CustomerOverview struct {
	Total    int            `json:"total"`
	ByType   map[string]int `json:"by_type"`
	ByStatus map[string]int `json:"by_status"`
}

type AlertItem struct {
	ID        string `json:"id"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

type SettlementSummaryResult struct {
	KPI              *KPIResult        `json:"kpi"`
	MonthlyChart     []ChartPoint      `json:"monthly_chart"`
	YearlyChart      []ChartPoint      `json:"yearly_chart"`
	CustomerOverview *CustomerOverview `json:"customer_overview"`
	Alerts           []AlertItem       `json:"alerts"`
}

func (r *DashboardRepository) GetSettlementSummary(ctx context.Context) (*SettlementSummaryResult, error) {
	org, scoped := OrgFilter(ctx)
	result := &SettlementSummaryResult{
		MonthlyChart: make([]ChartPoint, 0),
		YearlyChart:  make([]ChartPoint, 0),
		Alerts:       make([]AlertItem, 0),
	}

	// 1. KPI
	kpiArgs := []any{}
	f := ""
	if scoped {
		kpiArgs = append(kpiArgs, org)
		f = fmt.Sprintf(" AND org_id = $%d::uuid", len(kpiArgs))
	}
	kpiQ := fmt.Sprintf(`SELECT
		COALESCE(SUM(CASE WHEN operating_date >= date_trunc('year', NOW()) THEN total_energy_fee END), 0),
		COALESCE(SUM(CASE WHEN operating_date >= date_trunc('month', NOW()) THEN total_energy_fee END), 0),
		COALESCE(AVG(CASE WHEN total_purchase_fee > 0 AND total_retail_fee > 0
			THEN (total_retail_fee * 1.0 / NULLIF(total_retail_volume, 0) - total_purchase_fee * 1.0 / NULLIF(total_purchase_volume, 0))
			END), 0),
		COALESCE(AVG(CASE WHEN total_retail_fee > 0 THEN total_retail_fee * 1.0 / NULLIF(total_retail_volume, 0) END), 0)
		FROM settlement_daily WHERE version = 'PRELIMINARY'%s`, f)
	var kpi KPIResult
	err := r.pool.QueryRow(ctx, kpiQ, kpiArgs...).Scan(
		&kpi.YearlyGrossProfit, &kpi.MonthlyGrossProfit,
		&kpi.PriceSpread, &kpi.RetailAvgPrice,
	)
	if err != nil {
		_ = err
	}
	if kpi.YearlyGrossProfit != 0 || kpi.MonthlyGrossProfit != 0 {
		result.KPI = &kpi
	}

	// 2. Monthly chart
	monthArgs := []any{}
	fm := ""
	if scoped {
		monthArgs = append(monthArgs, org)
		fm = fmt.Sprintf(" AND org_id = $%d::uuid", len(monthArgs))
	}
	monthQ := fmt.Sprintf(`SELECT
		to_char(operating_date, 'YYYY-MM'),
		COALESCE(SUM(total_energy_fee), 0),
		COALESCE(SUM(total_purchase_fee), 0),
		COALESCE(SUM(total_retail_fee), 0),
		COALESCE(SUM(total_wholesale_fee), 0)
		FROM settlement_daily WHERE version = 'PRELIMINARY'
		AND operating_date >= date_trunc('year', NOW())%s
		GROUP BY to_char(operating_date, 'YYYY-MM')
		ORDER BY 1`, fm)
	mRows, err := r.pool.Query(ctx, monthQ, monthArgs...)
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var cp ChartPoint
			v := 0.0
			if err := mRows.Scan(&cp.Label, &v, &cp.TotalPurchase, &cp.TotalRetail, &cp.TotalWholesale); err == nil {
				cp.MonthlyGrossProfit = &v
				result.MonthlyChart = append(result.MonthlyChart, cp)
			}
		}
	}

	// 3. Yearly chart
	yearArgs := []any{}
	fy := ""
	if scoped {
		yearArgs = append(yearArgs, org)
		fy = fmt.Sprintf(" AND org_id = $%d::uuid", len(yearArgs))
	}
	yearQ := fmt.Sprintf(`SELECT
		to_char(operating_date, 'YYYY'),
		COALESCE(SUM(total_energy_fee), 0),
		COALESCE(SUM(total_purchase_fee), 0),
		COALESCE(SUM(total_retail_fee), 0),
		COALESCE(SUM(total_wholesale_fee), 0)
		FROM settlement_daily WHERE version = 'PRELIMINARY'%s
		GROUP BY to_char(operating_date, 'YYYY')
		ORDER BY 1`, fy)
	yRows, err := r.pool.Query(ctx, yearQ, yearArgs...)
	if err == nil {
		defer yRows.Close()
		for yRows.Next() {
			var cp ChartPoint
			v := 0.0
			if err := yRows.Scan(&cp.Label, &v, &cp.TotalPurchase, &cp.TotalRetail, &cp.TotalWholesale); err == nil {
				cp.YearlyGrossProfit = &v
				result.YearlyChart = append(result.YearlyChart, cp)
			}
		}
	}

	// 4. Customer overview
	custArgs := []any{}
	fc := ""
	if scoped {
		custArgs = append(custArgs, org)
		fc = fmt.Sprintf(" WHERE org_id = $%d::uuid", len(custArgs))
	}
	custQ := fmt.Sprintf(`SELECT COUNT(*), type, status FROM customers%s GROUP BY type, status`, fc)
	cRows, err := r.pool.Query(ctx, custQ, custArgs...)
	if err == nil {
		defer cRows.Close()
		ov := &CustomerOverview{
			Total:    0,
			ByType:   make(map[string]int),
			ByStatus: make(map[string]int),
		}
		for cRows.Next() {
			var count int
			var typ, status string
			if err := cRows.Scan(&count, &typ, &status); err == nil {
				ov.Total += count
				ov.ByType[typ] += count
				ov.ByStatus[status] += count
			}
		}
		if ov.Total > 0 {
			result.CustomerOverview = ov
		}
	}

	// 5. Alerts
	alertArgs := []any{}
	fa := ""
	if scoped {
		alertArgs = append(alertArgs, org)
		fa = fmt.Sprintf(" AND org_id = $%d::uuid", len(alertArgs))
	}
	alertQ := fmt.Sprintf(`SELECT id, severity, message, created_at::text
		FROM customer_anomaly_alerts WHERE NOT acknowledged%s
		ORDER BY created_at DESC LIMIT 5`, fa)
	aRows, err := r.pool.Query(ctx, alertQ, alertArgs...)
	if err == nil {
		defer aRows.Close()
		for aRows.Next() {
			var a AlertItem
			if err := aRows.Scan(&a.ID, &a.Level, &a.Message, &a.CreatedAt); err == nil {
				result.Alerts = append(result.Alerts, a)
			}
		}
	}

	return result, nil
}

// ── 用户仪表盘配置 ──

// GetUserConfig 获取用户的仪表盘布局配置；无记录时返回 nil（前端用默认布局）。
func (r *DashboardRepository) GetUserConfig(ctx context.Context, userID uuid.UUID) (json.RawMessage, error) {
	var config json.RawMessage
	err := r.pool.QueryRow(ctx, `SELECT config FROM user_dashboard_configs WHERE user_id = $1`, userID).Scan(&config)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return config, nil
}

// UpsertUserConfig 保存（插入或更新）用户的仪表盘布局配置。
func (r *DashboardRepository) UpsertUserConfig(ctx context.Context, userID uuid.UUID, config json.RawMessage) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_dashboard_configs (user_id, config, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (user_id) DO UPDATE SET config = $2, updated_at = now()
	`, userID, config)
	return err
}
