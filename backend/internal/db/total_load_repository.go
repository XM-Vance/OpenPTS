// 系统总负荷。
// 2026-06 自 p1_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// ─────────────── V1 系统总负荷 ───────────────

type TotalLoadDaily struct {
	ID         string    `json:"id"`
	DataDate   time.Time `json:"data_date"`
	Region     string    `json:"region"`
	Curve96    []float64 `json:"curve_96"`
	PeakLoad   float64   `json:"peak_load"`
	ValleyLoad float64   `json:"valley_load"`
	AvgLoad    float64   `json:"avg_load"`
	TotalMWh   float64   `json:"total_mwh"`
	CreatedAt  time.Time `json:"created_at"`
}

type TotalLoadRepository struct{ pool *Pool }

func NewTotalLoadRepository(pool *Pool) *TotalLoadRepository { return &TotalLoadRepository{pool: pool} }

func (r *TotalLoadRepository) List(ctx context.Context, days int) ([]*TotalLoadDaily, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `
		SELECT id, data_date, region, curve_96, peak_load, valley_load, avg_load, total_mwh, created_at
		FROM total_load_daily WHERE data_date >= $1`
	n := 2
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", n)
		n++
	}
	q += " ORDER BY data_date DESC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*TotalLoadDaily, 0)
	for rows.Next() {
		var t TotalLoadDaily
		if err := rows.Scan(&t.ID, &t.DataDate, &t.Region, &t.Curve96,
			&t.PeakLoad, &t.ValleyLoad, &t.AvgLoad, &t.TotalMWh, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

func (r *TotalLoadRepository) GenerateDemo(ctx context.Context) (int, error) {
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
	for i := 0; i < 14; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		curve := make([]float64, 96)
		peak, valley, sum := 0.0, math.MaxFloat64, 0.0
		for p := 0; p < 96; p++ {
			hour := p / 4
			base := 15000.0 // 15 GW 基础
			// 双峰：上午 9 + 晚上 19
			if hour >= 8 && hour <= 12 {
				base += 4000
			} else if hour >= 17 && hour <= 21 {
				base += 5000
			} else if hour >= 0 && hour <= 5 {
				base -= 3000
			}
			v := base + (rand.Float64()-0.5)*500
			curve[p] = v
			if v > peak {
				peak = v
			}
			if v < valley {
				valley = v
			}
			sum += v
		}
		avg := sum / 96
		totalMWh := sum * 0.25 / 1000 // MW × 15min/60 = MWh
		if _, err := r.pool.Exec(ctx, `
			INSERT INTO total_load_daily (data_date, region, curve_96, peak_load, valley_load, avg_load, total_mwh, org_id)
			VALUES ($1, 'system', $2, $3, $4, $5, $6, $7::uuid)
			ON CONFLICT (org_id, data_date) DO UPDATE SET
			  curve_96 = EXCLUDED.curve_96, peak_load = EXCLUDED.peak_load,
			  valley_load = EXCLUDED.valley_load, avg_load = EXCLUDED.avg_load,
			  total_mwh = EXCLUDED.total_mwh`,
			d, curve, peak, valley, avg, totalMWh, orgID); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
