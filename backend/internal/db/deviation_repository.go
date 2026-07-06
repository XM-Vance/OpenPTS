// 偏差结算。
// 2026-06 自 new_modules_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/shopspring/decimal"
)

// ─────────────── 偏差结算 ───────────────

type DeviationSettlement struct {
	ID              string          `json:"id"`
	OperatingDate   time.Time       `json:"operating_date"`
	DeclaredEnergy  float64         `json:"declared_energy_mwh"`
	ActualEnergy    float64         `json:"actual_energy_mwh"`
	DeviationEnergy float64         `json:"deviation_energy_mwh"`
	DeviationRate   float64         `json:"deviation_rate"`
	DeviationCost   decimal.Decimal `json:"deviation_cost"`   // P4: numeric(18,4)，金额精确
	PenaltyCost     decimal.Decimal `json:"penalty_cost"`     // P4: numeric(18,4)
	TotalSettlement decimal.Decimal `json:"total_settlement"` // P4: numeric(18,4)
	Category        string          `json:"category"`
	CreatedAt       time.Time       `json:"created_at"`
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
		       SUM(total_settlement)::numeric,
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
	Category             string          `json:"category"`
	TotalDeviationEnergy float64         `json:"total_deviation_energy_mwh"`
	TotalCost            decimal.Decimal `json:"total_cost"` // P4: SUM(numeric) 精确，不再 ::float8
	AvgDeviationRate     float64         `json:"avg_deviation_rate"`
	Count                int             `json:"count"`
}

func (r *DeviationRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃组织，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
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
			// P4: 金额用 decimal 运算，分项先舍入到列精度(4 位)，再求和——
			// 使 total == devCost + penalty 在 numeric(18,4) 下精确成立（账务口径：先舍分项后汇总）。
			devCost := decimal.NewFromFloat(deviation).Mul(decimal.NewFromFloat(350 + rand.Float64()*100)).Round(4)
			penalty := decimal.Zero
			if rate > 5 || rate < -5 {
				// 考核费按偏差「绝对值」计,恒非负;此前用带符号 deviation,
				// 负偏差(少发/少用)会算出负考核费=倒贴奖励,与预结算 |dev| 口径不一致。
				// 注:系数 50 维持原值(预结算用 100,两者是否统一属业务规则,另议)。
				penalty = decimal.NewFromFloat(absF(deviation)).Mul(decimal.NewFromInt(50)).Round(4)
			}
			total := devCost.Add(penalty)
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO deviation_settlement
				   (operating_date, declared_energy_mwh, actual_energy_mwh, deviation_energy_mwh,
				    deviation_rate, deviation_cost, penalty_cost, total_settlement, category, org_id, is_demo)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::uuid,TRUE)
				 ON CONFLICT (org_id, operating_date, category) DO UPDATE SET
				   declared_energy_mwh = EXCLUDED.declared_energy_mwh,
				   actual_energy_mwh = EXCLUDED.actual_energy_mwh,
				   deviation_energy_mwh = EXCLUDED.deviation_energy_mwh,
				   deviation_rate = EXCLUDED.deviation_rate,
				   deviation_cost = EXCLUDED.deviation_cost,
				   penalty_cost = EXCLUDED.penalty_cost,
				   total_settlement = EXCLUDED.total_settlement, is_demo = TRUE`,
				d, declared, actual, deviation, rate, devCost, penalty, total, cat, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
