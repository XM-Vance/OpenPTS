// 偏差结算。
// 2026-06 自 new_modules_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── 偏差结算 ───────────────

type DeviationSettlement struct {
	ID              string    `json:"id"`
	OperatingDate   time.Time `json:"operating_date"`
	DeclaredEnergy  float64   `json:"declared_energy_mwh"`
	ActualEnergy    float64   `json:"actual_energy_mwh"`
	DeviationEnergy float64   `json:"deviation_energy_mwh"`
	DeviationRate   float64   `json:"deviation_rate"`
	DeviationCost   float64   `json:"deviation_cost"`
	PenaltyCost     float64   `json:"penalty_cost"`
	TotalSettlement float64   `json:"total_settlement"`
	Category        string    `json:"category"`
	CreatedAt       time.Time `json:"created_at"`
}

type DeviationRepository struct{ pool *Pool }

func NewDeviationRepository(pool *Pool) *DeviationRepository {
	return &DeviationRepository{pool: pool}
}

func (r *DeviationRepository) List(ctx context.Context, category string, days int) ([]*DeviationSettlement, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `SELECT id, operating_date, declared_energy_mwh, actual_energy_mwh,
			  deviation_energy_mwh, deviation_rate, deviation_cost, penalty_cost,
			  total_settlement, category, created_at
		  FROM deviation_settlement WHERE operating_date >= $1`
	idx := 2
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", idx)
		idx++
	}
	if category != "" {
		args = append(args, category)
		q += fmt.Sprintf(" AND category = $%d", idx)
		idx++
	}
	q += " ORDER BY operating_date DESC LIMIT 200"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*DeviationSettlement, 0)
	for rows.Next() {
		var d DeviationSettlement
		if err := rows.Scan(&d.ID, &d.OperatingDate, &d.DeclaredEnergy, &d.ActualEnergy,
			&d.DeviationEnergy, &d.DeviationRate, &d.DeviationCost, &d.PenaltyCost,
			&d.TotalSettlement, &d.Category, &d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *DeviationRepository) Summary(ctx context.Context, days int) ([]*DeviationSummary, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	org, scoped := OrgFilter(ctx)
	orgCond := ""
	args := []any{since}
	idx := 2
	if scoped {
		args = append(args, org)
		orgCond = fmt.Sprintf(" AND org_id = $%d::uuid", idx)
		idx++
	}
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT category,
		       SUM(deviation_energy_mwh)::float8,
		       SUM(total_settlement)::float8,
		       AVG(deviation_rate)::float8,
		       COUNT(*)
		FROM deviation_settlement
		WHERE operating_date >= $1%s
		GROUP BY category
		ORDER BY category`, orgCond), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*DeviationSummary, 0)
	for rows.Next() {
		var s DeviationSummary
		if err := rows.Scan(&s.Category, &s.TotalDeviationEnergy, &s.TotalCost,
			&s.AvgDeviationRate, &s.Count); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

type DeviationSummary struct {
	Category             string  `json:"category"`
	TotalDeviationEnergy float64 `json:"total_deviation_energy_mwh"`
	TotalCost            float64 `json:"total_cost"`
	AvgDeviationRate     float64 `json:"avg_deviation_rate"`
	Count                int     `json:"count"`
}

func (r *DeviationRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用 FJ
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='FJ'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve FJ org: %w", err)
		}
	}
	categories := []string{"day_ahead", "real_time", "intraday"}
	cnt := 0
	for i := 0; i < 30; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		for _, cat := range categories {
			declared := 2000 + rand.Float64()*1000
			deviation := declared * (0.02 + rand.Float64()*0.08)
			if rand.Float64() < 0.5 {
				deviation = -deviation
			}
			actual := declared + deviation
			rate := deviation / declared * 100
			devCost := deviation * (350 + rand.Float64()*100)
			penalty := 0.0
			if rate > 5 || rate < -5 {
				// 考核费按偏差「绝对值」计,恒非负;此前用带符号 deviation,
				// 负偏差(少发/少用)会算出负考核费=倒贴奖励,与预结算 |dev| 口径不一致。
				// 注:系数 50 维持原值(预结算用 100,两者是否统一属业务规则,另议)。
				penalty = absF(deviation) * 50
			}
			total := devCost + penalty
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO deviation_settlement
				   (operating_date, declared_energy_mwh, actual_energy_mwh, deviation_energy_mwh,
				    deviation_rate, deviation_cost, penalty_cost, total_settlement, category, org_id)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::uuid)
				 ON CONFLICT (org_id, operating_date, category) DO UPDATE SET
				   declared_energy_mwh = EXCLUDED.declared_energy_mwh,
				   actual_energy_mwh = EXCLUDED.actual_energy_mwh,
				   deviation_energy_mwh = EXCLUDED.deviation_energy_mwh,
				   deviation_rate = EXCLUDED.deviation_rate,
				   deviation_cost = EXCLUDED.deviation_cost,
				   penalty_cost = EXCLUDED.penalty_cost,
				   total_settlement = EXCLUDED.total_settlement`,
				d, declared, actual, deviation, rate, devCost, penalty, total, cat, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
