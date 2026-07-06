// 价格趋势与电价汇总仓储：支撑前端 /price/trend/price-trend 与 /retail/price-daily/daily-summary。
//
// 数据源（均为已有表）：
//   - day_ahead_spot_price（0007）：日前现货价（48 点，price_da）
//   - real_time_spot_price（0007）：实时现货价
//   - contract_price_daily（0024）：合同日价（unit_price/daily_energy/daily_amount）
//   - retail_contracts：合同元信息（contract_type 等）
//
// price-trend：JOIN 现货日均价 + 合同加权均价，算价差。
// daily-summary：聚合 contract_price_daily 的 KPI + 按合同类型汇总。
// curve-analysis / quantity-structure 需跨多表 + 粒度转换，暂不实现（契约闸白名单跟踪）。
package db

import (
	"context"
	"fmt"
	"time"
)

type PriceTrendRepository struct{ pool *Pool }

func NewPriceTrendRepository(pool *Pool) *PriceTrendRepository {
	return &PriceTrendRepository{pool: pool}
}

// PriceTrendPoint 价差趋势单日点。
type PriceTrendPoint struct {
	Date         string  `json:"date"`
	SpotPrice    float64 `json:"spot_price"`    // 日前现货均价
	ContractPrice float64 `json:"contract_price"` // 合同加权均价
	Spread       float64 `json:"spread"`        // 合同 - 现货
}

// PriceTrendResponse 价差趋势响应。
type PriceTrendResponse struct {
	DailyTrends   []PriceTrendPoint `json:"daily_trends"`
	SpreadStats   map[string]float64 `json:"spread_stats"` // mean/max/min
}

// PriceTrend 取现货-合同价差趋势（指定天数）。
// 现货侧：day_ahead_spot_price 按日均价；合同侧：contract_price_daily 按 daily_energy 加权。
func (r *PriceTrendRepository) PriceTrend(ctx context.Context, days int) (*PriceTrendResponse, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	// 现货日均价（共享参考表，不过滤省）
	q := `WITH spot AS (
			SELECT trade_date::text AS d, AVG(price_da) AS sp
			FROM day_ahead_spot_price
			WHERE trade_date >= $1 GROUP BY trade_date
		), ctr AS (
			SELECT price_date::text AS d,
				SUM(unit_price * daily_energy) / NULLIF(SUM(daily_energy), 0) AS cp
			FROM contract_price_daily cpd
			JOIN retail_contracts rc ON rc.id = cpd.contract_id
			WHERE price_date >= $1`
	args := []any{since}
	idx := 2
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND rc.org_id = $%d::uuid", idx)
		idx++
	}
	q += fmt.Sprintf(` GROUP BY price_date
		)
		SELECT s.d, COALESCE(s.sp,0), COALESCE(c.cp,0), COALESCE(c.cp,0) - COALESCE(s.sp,0)
		FROM spot s LEFT JOIN ctr c ON c.d = s.d
		ORDER BY s.d ASC`)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []PriceTrendPoint
	var sumSpread, maxSpread, minSpread float64
	first := true
	for rows.Next() {
		var p PriceTrendPoint
		if err := rows.Scan(&p.Date, &p.SpotPrice, &p.ContractPrice, &p.Spread); err != nil {
			return nil, err
		}
		points = append(points, p)
		sumSpread += p.Spread
		if first || p.Spread > maxSpread {
			maxSpread = p.Spread
		}
		if first || p.Spread < minSpread {
			minSpread = p.Spread
		}
		first = false
	}
	stats := map[string]float64{"max": maxSpread, "min": minSpread}
	if len(points) > 0 {
		stats["mean"] = sumSpread / float64(len(points))
	}
	return &PriceTrendResponse{DailyTrends: points, SpreadStats: stats}, rows.Err()
}

// ─── DailySummary：合同电价日汇总 ───

type DailySummaryResponse struct {
	KPIs        map[string]float64        `json:"kpis"`
	TypeSummary map[string]map[string]float64 `json:"type_summary"` // contract_type → {qty, avg_price}
}

// DailySummary 聚合指定日期的合同电价 KPI + 按合同类型汇总。
func (r *PriceTrendRepository) DailySummary(
	ctx context.Context, date string,
) (*DailySummaryResponse, error) {
	// KPI：总电量、加权均价
	q := `SELECT COALESCE(SUM(daily_energy),0),
			SUM(unit_price * daily_energy) / NULLIF(SUM(daily_energy), 0)
		  FROM contract_price_daily cpd
		  JOIN retail_contracts rc ON rc.id = cpd.contract_id
		  WHERE price_date = $1`
	args := []any{date}
	idx := 2
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND rc.org_id = $%d::uuid", idx)
		idx++
	}
	resp := &DailySummaryResponse{KPIs: map[string]float64{}, TypeSummary: map[string]map[string]float64{}}
	var totalQty, avgPrice *float64
	if err := r.pool.QueryRow(ctx, q, args...).Scan(&totalQty, &avgPrice); err != nil {
		return nil, err
	}
	if totalQty != nil {
		resp.KPIs["total_quantity"] = *totalQty
	}
	if avgPrice != nil {
		resp.KPIs["overall_avg_price"] = *avgPrice
	}

	// 按合同类型汇总
	q2 := `SELECT rc.contract_type,
			COALESCE(SUM(cpd.daily_energy),0),
			SUM(cpd.unit_price * cpd.daily_energy) / NULLIF(SUM(cpd.daily_energy), 0)
		  FROM contract_price_daily cpd
		  JOIN retail_contracts rc ON rc.id = cpd.contract_id
		  WHERE cpd.price_date = $1`
	args2 := []any{date}
	if org, scoped := OrgFilter(ctx); scoped {
		args2 = append(args2, org)
		q2 += fmt.Sprintf(" AND rc.org_id = $%d::uuid", len(args2))
	}
	q2 += " GROUP BY rc.contract_type"
	rows, err := r.pool.Query(ctx, q2, args2...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ct string
		var qty, price float64
		if err := rows.Scan(&ct, &qty, &price); err != nil {
			return nil, err
		}
		resp.TypeSummary[ct] = map[string]float64{"qty": qty, "avg_price": price}
	}
	return resp, rows.Err()
}
