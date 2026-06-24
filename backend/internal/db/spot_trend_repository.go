// 现货价格趋势仓储。
// 2026-06 自 v1clone_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"time"
)

// ─────────────── D2 现货价格趋势（基于已有 day_ahead_spot_price） ───────────────

type SpotDailyAvg struct {
	Date     time.Time `json:"date"`
	AvgPrice float64   `json:"avg_price"`
	PeakAvg  *float64  `json:"peak_avg"`     // 8:00-22:00 平均，指针支持 NULL
	OffPeak  *float64  `json:"off_peak_avg"` // 0:00-8:00 + 22:00-24:00 平均
}

type SpotTrendRepository struct{ pool *Pool }

func NewSpotTrendRepository(pool *Pool) *SpotTrendRepository { return &SpotTrendRepository{pool: pool} }

// DailyAvg 最近 N 天的现货日均价（与峰平谷拆分），自适应 period 粒度。
func (r *SpotTrendRepository) DailyAvg(ctx context.Context, days int) ([]*SpotDailyAvg, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	org, scoped := OrgFilter(ctx)
	orgWhere := ""
	args := []any{since}
	if scoped {
		args = append(args, org)
		orgWhere = fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT date,
		       AVG(price_da)::float8 AS avg_p,
		       COALESCE(AVG(CASE WHEN period >= (SELECT MIN(period) + ROUND((MAX(period)-MIN(period)+1)*8/24)
		                              FROM day_ahead_spot_price WHERE date >= $1::date%s)
		                          AND period <= (SELECT MIN(period) + ROUND((MAX(period)-MIN(period)+1)*22/24)
		                              FROM day_ahead_spot_price WHERE date >= $1::date%s)
		                 THEN price_da END), 0)::float8 AS peak_p,
		       COALESCE(AVG(CASE WHEN period < (SELECT MIN(period) + ROUND((MAX(period)-MIN(period)+1)*8/24)
		                             FROM day_ahead_spot_price WHERE date >= $1::date%s)
		                          OR period > (SELECT MIN(period) + ROUND((MAX(period)-MIN(period)+1)*22/24)
		                              FROM day_ahead_spot_price WHERE date >= $1::date%s)
		                 THEN price_da END), 0)::float8 AS off_p
		FROM day_ahead_spot_price
		WHERE date >= $1::date%s
		GROUP BY date
		ORDER BY date ASC`, orgWhere, orgWhere, orgWhere, orgWhere, orgWhere), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*SpotDailyAvg, 0)
	for rows.Next() {
		var p SpotDailyAvg
		if err := rows.Scan(&p.Date, &p.AvgPrice, &p.PeakAvg, &p.OffPeak); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// HourlyAvg 最近 N 天的小时平均价格分布（24 个点），自适应 period 粒度。
func (r *SpotTrendRepository) HourlyAvg(ctx context.Context, days int) ([]float64, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	org, scoped := OrgFilter(ctx)
	orgWhere := ""
	args := []any{since}
	if scoped {
		args = append(args, org)
		orgWhere = fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT (period - 1) / GREATEST(1,
		    ((SELECT MAX(period)-MIN(period)+1 FROM day_ahead_spot_price WHERE date >= $1::date%s) / 24)
		) AS hour, AVG(price_da)::float8
		FROM day_ahead_spot_price
		WHERE date >= $1::date%s
		GROUP BY hour
		ORDER BY hour`, orgWhere, orgWhere), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]float64, 24)
	for rows.Next() {
		var hour int
		var avg float64
		if err := rows.Scan(&hour, &avg); err != nil {
			return nil, err
		}
		if hour >= 0 && hour < 24 {
			out[hour] = avg
		}
	}
	return out, rows.Err()
}
