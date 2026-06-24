// 现货市场。
// 2026-06 自 new_modules_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
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

type SpotMarketRepository struct{ pool *Pool }

func NewSpotMarketRepository(pool *Pool) *SpotMarketRepository {
	return &SpotMarketRepository{pool: pool}
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
			"SELECT id FROM organizations WHERE code='FJ'").Scan(&orgID); err != nil {
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
