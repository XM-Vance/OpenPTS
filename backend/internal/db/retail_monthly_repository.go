// 零售月度结算。
// 2026-06 自 p0_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── U1 零售月度结算 ───────────────

type RetailMonthlySettle struct {
	ID               string    `json:"id"`
	ContractID       string    `json:"contract_id"`
	CustomerName     string    `json:"customer_name,omitempty"`
	OperatingMonth   string    `json:"operating_month"`
	ContractEnergy   float64   `json:"contract_energy_mwh"`
	ActualEnergy     float64   `json:"actual_energy_mwh"`
	WeightedAvgPrice float64   `json:"weighted_avg_price"`
	Receivable       float64   `json:"receivable_amount"`
	Actual           float64   `json:"actual_amount"`
	DeviationEnergy  float64   `json:"deviation_energy_mwh"`
	PenaltyAmount    float64   `json:"penalty_amount"`
	Note             *string   `json:"note,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type RetailMonthlyRepository struct{ pool *Pool }

func NewRetailMonthlyRepository(pool *Pool) *RetailMonthlyRepository {
	return &RetailMonthlyRepository{pool: pool}
}

func (r *RetailMonthlyRepository) List(ctx context.Context, contractID string, limit int) ([]*RetailMonthlySettle, error) {
	if limit <= 0 || limit > 200 {
		limit = 60
	}
	args := []any{limit}
	n := 2
	q := `SELECT s.id, s.contract_id::text, c.user_name, s.operating_month,
		s.contract_energy_mwh, s.actual_energy_mwh, s.weighted_avg_price,
		s.receivable_amount, s.actual_amount, s.deviation_energy_mwh,
		s.penalty_amount, s.note, s.created_at
	  FROM retail_monthly_settlement s
	  JOIN retail_contracts rc ON rc.id = s.contract_id
	  JOIN customers c ON c.id = rc.customer_id`
	where := make([]string, 0, 3)
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		where = append(where, fmt.Sprintf("s.org_id = $%d::uuid", n))
		n++
	}
	if contractID != "" {
		args = append(args, contractID)
		where = append(where, fmt.Sprintf("s.contract_id = $%d", n))
		n++
	}
	if len(where) > 0 {
		q += " WHERE " + joinWhere(where)
	}
	q += " ORDER BY s.operating_month DESC, c.user_name ASC LIMIT $1"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*RetailMonthlySettle, 0)
	for rows.Next() {
		var s RetailMonthlySettle
		if err := rows.Scan(&s.ID, &s.ContractID, &s.CustomerName, &s.OperatingMonth,
			&s.ContractEnergy, &s.ActualEnergy, &s.WeightedAvgPrice,
			&s.Receivable, &s.Actual, &s.DeviationEnergy, &s.PenaltyAmount,
			&s.Note, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *RetailMonthlyRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用 FJ
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='FJ'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve FJ org: %w", err)
		}
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, purchasing_energy_mwh FROM retail_contracts WHERE status = 'active'`)
	if err != nil {
		return 0, err
	}
	contracts := []struct {
		id     string
		energy float64
	}{}
	for rows.Next() {
		var id string
		var e float64
		if err := rows.Scan(&id, &e); err != nil {
			rows.Close()
			return 0, err
		}
		contracts = append(contracts, struct {
			id     string
			energy float64
		}{id, e})
	}
	rows.Close()

	cnt := 0
	for _, c := range contracts {
		monthlyContract := c.energy / 12
		for i := 0; i < 6; i++ {
			t := time.Now().AddDate(0, -i, 0)
			ym := t.Format("2006-01")
			actual := monthlyContract * (0.85 + rand.Float64()*0.3)
			price := 420 + rand.Float64()*40
			receivable := actual * price
			actualAmount := receivable * (0.97 + rand.Float64()*0.03)
			dev := actual - monthlyContract
			penalty := 0.0
			if dev < -monthlyContract*0.05 {
				penalty = -dev * 50
			}
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO retail_monthly_settlement
				 (contract_id, operating_month, contract_energy_mwh, actual_energy_mwh,
				  weighted_avg_price, receivable_amount, actual_amount,
				  deviation_energy_mwh, penalty_amount, org_id)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::uuid)
				 ON CONFLICT (org_id, contract_id, operating_month) DO UPDATE SET
				   contract_energy_mwh = EXCLUDED.contract_energy_mwh,
				   actual_energy_mwh = EXCLUDED.actual_energy_mwh,
				   weighted_avg_price = EXCLUDED.weighted_avg_price,
				   receivable_amount = EXCLUDED.receivable_amount,
				   actual_amount = EXCLUDED.actual_amount,
				   deviation_energy_mwh = EXCLUDED.deviation_energy_mwh,
				   penalty_amount = EXCLUDED.penalty_amount`,
				c.id, ym, monthlyContract, actual, price, receivable, actualAmount, dev, penalty, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
