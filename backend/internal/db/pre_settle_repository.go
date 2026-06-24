// 预结算明细。
// 2026-06 自 p0_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── U2 预结算明细 ───────────────

type PreSettleDaily struct {
	ID               string    `json:"id"`
	OperatingDate    time.Time `json:"operating_date"`
	DeclaredCurve96  []float64 `json:"declared_curve_96"`
	ClearedCurve96   []float64 `json:"cleared_curve_96"`
	SpotPrice96      []float64 `json:"spot_price_96"`
	TotalDeclared    float64   `json:"total_declared"`
	TotalCleared     float64   `json:"total_cleared"`
	TotalDeviation   float64   `json:"total_deviation"`
	DeviationRatio   float64   `json:"deviation_ratio"`
	EnergyRevenue    float64   `json:"energy_revenue"`
	DeviationPenalty float64   `json:"deviation_penalty"`
	FinalAmount      float64   `json:"final_amount"`
	CreatedAt        time.Time `json:"created_at"`
}

type PreSettleRepository struct{ pool *Pool }

func NewPreSettleRepository(pool *Pool) *PreSettleRepository {
	return &PreSettleRepository{pool: pool}
}

func (r *PreSettleRepository) List(ctx context.Context, days int) ([]*PreSettleDaily, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `
		SELECT id, operating_date, declared_curve_96, cleared_curve_96, spot_price_96,
		       total_declared, total_cleared, total_deviation, deviation_ratio,
		       energy_revenue, deviation_penalty, final_amount, created_at
		FROM pre_settlement_daily
		WHERE operating_date >= $1`
	n := 2
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", n)
		n++
	}
	q += " ORDER BY operating_date DESC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*PreSettleDaily, 0)
	for rows.Next() {
		var p PreSettleDaily
		if err := rows.Scan(&p.ID, &p.OperatingDate, &p.DeclaredCurve96,
			&p.ClearedCurve96, &p.SpotPrice96, &p.TotalDeclared, &p.TotalCleared,
			&p.TotalDeviation, &p.DeviationRatio, &p.EnergyRevenue,
			&p.DeviationPenalty, &p.FinalAmount, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

func (r *PreSettleRepository) Get(ctx context.Context, date string) (*PreSettleDaily, error) {
	args := []any{date}
	q := `
		SELECT id, operating_date, declared_curve_96, cleared_curve_96, spot_price_96,
		       total_declared, total_cleared, total_deviation, deviation_ratio,
		       energy_revenue, deviation_penalty, final_amount, created_at
		FROM pre_settlement_daily WHERE operating_date = $1::date`
	n := 2
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", n)
		n++
	}
	var p PreSettleDaily
	err := r.pool.QueryRow(ctx, q, args...).
		Scan(&p.ID, &p.OperatingDate, &p.DeclaredCurve96, &p.ClearedCurve96,
			&p.SpotPrice96, &p.TotalDeclared, &p.TotalCleared, &p.TotalDeviation,
			&p.DeviationRatio, &p.EnergyRevenue, &p.DeviationPenalty,
			&p.FinalAmount, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PreSettleRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	cnt := 0
	for i := 0; i < 14; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		declared := make([]float64, 96)
		cleared := make([]float64, 96)
		spot := make([]float64, 96)
		totalDec := 0.0
		totalClr := 0.0
		rev := 0.0
		for p := 0; p < 96; p++ {
			hour := p / 4
			base := 800.0 + 600*float64(hour%12)/12 + rand.Float64()*100
			declared[p] = base
			cleared[p] = base * (0.9 + rand.Float64()*0.2)
			// 现货价格：峰段高、谷段低
			if hour >= 8 && hour < 22 {
				spot[p] = 500 + rand.Float64()*200
			} else {
				spot[p] = 250 + rand.Float64()*100
			}
			totalDec += declared[p] * 0.25 // 每点 15min
			totalClr += cleared[p] * 0.25
			rev += cleared[p] * 0.25 * spot[p]
		}
		dev := totalClr - totalDec
		penalty := 0.0
		ratio := 0.0
		if totalDec > 0 {
			ratio = dev / totalDec
			if absF(ratio) > 0.05 {
				penalty = absF(dev) * 100
			}
		}
		if _, err := r.pool.Exec(ctx, `
			INSERT INTO pre_settlement_daily
			  (operating_date, declared_curve_96, cleared_curve_96, spot_price_96,
			   total_declared, total_cleared, total_deviation, deviation_ratio,
			   energy_revenue, deviation_penalty, final_amount, org_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::uuid)
			ON CONFLICT (org_id, operating_date) DO UPDATE SET
			  declared_curve_96 = EXCLUDED.declared_curve_96,
			  cleared_curve_96 = EXCLUDED.cleared_curve_96,
			  spot_price_96 = EXCLUDED.spot_price_96,
			  total_declared = EXCLUDED.total_declared,
			  total_cleared = EXCLUDED.total_cleared,
			  total_deviation = EXCLUDED.total_deviation,
			  deviation_ratio = EXCLUDED.deviation_ratio,
			  energy_revenue = EXCLUDED.energy_revenue,
			  deviation_penalty = EXCLUDED.deviation_penalty,
			  final_amount = EXCLUDED.final_amount`,
			d, declared, cleared, spot, totalDec, totalClr, dev, ratio,
			rev, penalty, rev-penalty, orgID); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}

func absF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
