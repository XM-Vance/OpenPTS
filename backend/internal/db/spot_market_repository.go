// 现货市场。
// 2026-06 自 new_modules_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"
)

// ─────────────── 现货市场 ───────────────

type SpotMarketDaily struct {
	ID             string    `json:"id"`
	TradeDate      time.Time `json:"trade_date"`
	DayAheadAvg    float64   `json:"day_ahead_avg"`
	DayAheadHigh   float64   `json:"day_ahead_high"`
	DayAheadLow    float64   `json:"day_ahead_low"`
	RealTimeAvg    float64   `json:"real_time_avg"`
	RealTimeHigh   float64   `json:"real_time_high"`
	RealTimeLow    float64   `json:"real_time_low"`
	TotalVolumeMWh float64   `json:"total_volume_mwh"`
	Spread         float64   `json:"spread"`
	CreatedAt      time.Time `json:"created_at"`
}

// SpotMarketRow 对齐前端列表契约（spot/_view.tsx:122-176）：
// da_avg_price/rt_avg_price/max_price/min_price/volatility/date。
type SpotMarketRow struct {
	Date       string  `json:"date"`
	DaAvgPrice float64 `json:"da_avg_price"`
	RtAvgPrice float64 `json:"rt_avg_price"`
	MaxPrice   float64 `json:"max_price"`
	MinPrice   float64 `json:"min_price"`
	Volatility float64 `json:"volatility"`
	VolumeMWh  float64 `json:"volume_mwh"`
}

type SpotMarketRepository struct{ pool *Pool }

func NewSpotMarketRepository(pool *Pool) *SpotMarketRepository {
	return &SpotMarketRepository{pool: pool}
}

// SpotMarketStatistics 对齐前端契约（spot/_view.tsx:91-99）：max_price/min_price/
// avg_price/volatility/da_avg_price/rt_avg_price。
type SpotMarketStatistics struct {
	MaxPrice    float64 `json:"max_price"`     // 期间最高价（日前最高）
	MinPrice    float64 `json:"min_price"`     // 期间最低价（日前最低）
	AvgPrice    float64 `json:"avg_price"`     // 期间均价（日前均价）
	Volatility  float64 `json:"volatility"`    // 波动率（标准差/均值，无量纲比例）
	DaAvgPrice  float64 `json:"da_avg_price"`  // 日前均价
	RtAvgPrice  float64 `json:"rt_avg_price"`  // 实时均价
	TradeDays   int     `json:"trade_days"`
}

// Statistics 聚合指定日期范围的现货市场统计（对齐前端 start_date/end_date 参数）。
func (r *SpotMarketRepository) StatisticsByRange(ctx context.Context, start, end time.Time) (*SpotMarketStatistics, error) {
	q := `SELECT
			COALESCE(MAX(day_ahead_high),0),
			COALESCE(MIN(day_ahead_low),0),
			COALESCE(AVG(day_ahead_avg),0),
			COALESCE(STDDEV(day_ahead_avg) / NULLIF(AVG(day_ahead_avg),0),0),
			COALESCE(AVG(day_ahead_avg),0),
			COALESCE(AVG(real_time_avg),0),
			COUNT(*)
		  FROM spot_market_daily WHERE trade_date >= $1 AND trade_date <= $2`
	var s SpotMarketStatistics
	err := r.pool.QueryRow(ctx, q, start, end).Scan(
		&s.MaxPrice, &s.MinPrice, &s.AvgPrice, &s.Volatility,
		&s.DaAvgPrice, &s.RtAvgPrice, &s.TradeDays)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SpotPriceCurve 对齐前端契约（spot/_view.tsx:106-108）：day_ahead_prices/
// realtime_prices/forecast_prices 为数组。
type SpotPriceCurve struct {
	DayAheadPrices  []float64 `json:"day_ahead_prices"`
	RealTimePrices  []float64 `json:"realtime_prices"`
	ForecastPrices  []float64 `json:"forecast_prices"` // 暂无预测源，留空
	Dates           []string  `json:"dates"`
}

// PriceCurveSeries 返回按日的现货价格数组（对齐前端数组契约）。
func (r *SpotMarketRepository) PriceCurveSeries(ctx context.Context, start, end time.Time) (*SpotPriceCurve, error) {
	q := `SELECT trade_date::text, day_ahead_avg, real_time_avg
		  FROM spot_market_daily WHERE trade_date >= $1 AND trade_date <= $2
		  ORDER BY trade_date ASC`
	rows, err := r.pool.Query(ctx, q, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	curve := &SpotPriceCurve{
		DayAheadPrices: []float64{}, RealTimePrices: []float64{}, ForecastPrices: []float64{}, Dates: []string{},
	}
	for rows.Next() {
		var d string
		var da, rt float64
		if err := rows.Scan(&d, &da, &rt); err != nil {
			return nil, err
		}
		curve.Dates = append(curve.Dates, d)
		curve.DayAheadPrices = append(curve.DayAheadPrices, da)
		curve.RealTimePrices = append(curve.RealTimePrices, rt)
		curve.ForecastPrices = append(curve.ForecastPrices, 0) // 预测源暂无
	}
	return curve, rows.Err()
}

// PriceCurve 返回现货价格曲线（按日，供 price-curve 端点）。
func (r *SpotMarketRepository) PriceCurve(ctx context.Context, days int) ([]*SpotMarketDaily, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	q := `SELECT id, trade_date, day_ahead_avg, day_ahead_high, day_ahead_low,
			real_time_avg, real_time_high, real_time_low, total_volume_mwh, spread, created_at
		  FROM spot_market_daily WHERE trade_date >= $1 ORDER BY trade_date ASC`
	rows, err := r.pool.Query(ctx, q, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*SpotMarketDaily, 0)
	for rows.Next() {
		var s SpotMarketDaily
		if err := rows.Scan(&s.ID, &s.TradeDate, &s.DayAheadAvg, &s.DayAheadHigh, &s.DayAheadLow,
			&s.RealTimeAvg, &s.RealTimeHigh, &s.RealTimeLow, &s.TotalVolumeMWh, &s.Spread, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

// ListByRange 返回对齐前端列表契约的行（SpotMarketRow），按日期范围。
func (r *SpotMarketRepository) ListByRange(ctx context.Context, start, end time.Time, limit int) ([]*SpotMarketRow, error) {
	if limit <= 0 || limit > 365 {
		limit = 100
	}
	q := `SELECT trade_date::text, day_ahead_avg, real_time_avg,
			day_ahead_high, day_ahead_low,
			COALESCE(STDDEV(day_ahead_avg) OVER (ORDER BY trade_date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW) / NULLIF(day_ahead_avg,0), 0),
			COALESCE(total_volume_mwh, 0)
		  FROM spot_market_daily WHERE trade_date >= $1 AND trade_date <= $2
		  ORDER BY trade_date DESC LIMIT $3`
	rows, err := r.pool.Query(ctx, q, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*SpotMarketRow, 0)
	for rows.Next() {
		var s SpotMarketRow
		if err := rows.Scan(&s.Date, &s.DaAvgPrice, &s.RtAvgPrice,
			&s.MaxPrice, &s.MinPrice, &s.Volatility, &s.VolumeMWh); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

func (r *SpotMarketRepository) List(ctx context.Context, days int) ([]*SpotMarketDaily, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `SELECT id, trade_date, day_ahead_avg, day_ahead_high, day_ahead_low,
	       real_time_avg, real_time_high, real_time_low, total_volume_mwh, spread, created_at
	FROM spot_market_daily WHERE trade_date >= $1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY trade_date DESC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*SpotMarketDaily, 0)
	for rows.Next() {
		var s SpotMarketDaily
		if err := rows.Scan(&s.ID, &s.TradeDate, &s.DayAheadAvg, &s.DayAheadHigh,
			&s.DayAheadLow, &s.RealTimeAvg, &s.RealTimeHigh, &s.RealTimeLow,
			&s.TotalVolumeMWh, &s.Spread, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *SpotMarketRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用 FJ
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve FJ org: %w", err)
		}
	}
	cnt := 0
	for i := 0; i < 30; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		daAvg := 380 + rand.Float64()*120
		daHigh := daAvg * (1.15 + rand.Float64()*0.25)
		daLow := daAvg * (0.6 + rand.Float64()*0.15)
		rtAvg := daAvg * (0.9 + rand.Float64()*0.2)
		rtHigh := rtAvg * (1.2 + rand.Float64()*0.3)
		rtLow := rtAvg * (0.5 + rand.Float64()*0.2)
		volume := 350000 + rand.Float64()*100000
		spread := rtAvg - daAvg
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO spot_market_daily
			   (trade_date, day_ahead_avg, day_ahead_high, day_ahead_low,
			    real_time_avg, real_time_high, real_time_low, total_volume_mwh, spread, org_id)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::uuid)
			 ON CONFLICT (org_id, trade_date) DO UPDATE SET
			   day_ahead_avg = EXCLUDED.day_ahead_avg, day_ahead_high = EXCLUDED.day_ahead_high,
			   day_ahead_low = EXCLUDED.day_ahead_low, real_time_avg = EXCLUDED.real_time_avg,
			   real_time_high = EXCLUDED.real_time_high, real_time_low = EXCLUDED.real_time_low,
			   total_volume_mwh = EXCLUDED.total_volume_mwh, spread = EXCLUDED.spread`,
			d, daAvg, daHigh, daLow, rtAvg, rtHigh, rtLow, volume, spread, orgID); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
