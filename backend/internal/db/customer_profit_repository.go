// 客户利润仓储。
// 2026-06 自 v1clone_f_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"math/rand"
	"time"
)

// ─────────────── F1 客户利润 ───────────────

type CustomerProfit struct {
	ID             string    `json:"id"`
	CustomerID     string    `json:"customer_id"`
	CustomerName   string    `json:"customer_name,omitempty"`
	OperatingMonth string    `json:"operating_month"`
	Revenue        float64   `json:"revenue"`
	Cost           float64   `json:"cost"`
	GrossProfit    float64   `json:"gross_profit"`
	GrossMargin    float64   `json:"gross_margin"`
	EnergyMWh      float64   `json:"energy_mwh"`
	CreatedAt      time.Time `json:"created_at"`
}

type CustomerProfitRepository struct{ pool *Pool }

func NewCustomerProfitRepository(pool *Pool) *CustomerProfitRepository {
	return &CustomerProfitRepository{pool: pool}
}

func (r *CustomerProfitRepository) List(ctx context.Context, month string, limit int) ([]*CustomerProfit, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args := []any{limit}
	q := `SELECT p.id, p.customer_id::text, c.user_name, p.operating_month,
			p.revenue, p.cost, p.gross_profit, p.gross_margin, p.energy_mwh, p.created_at
		  FROM customer_profit p JOIN customers c ON c.id = p.customer_id`
	if month != "" {
		args = append(args, month)
		q += " WHERE p.operating_month = $2"
	}
	q += " ORDER BY p.operating_month DESC, p.gross_profit DESC LIMIT $1"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*CustomerProfit, 0)
	for rows.Next() {
		var p CustomerProfit
		if err := rows.Scan(&p.ID, &p.CustomerID, &p.CustomerName, &p.OperatingMonth,
			&p.Revenue, &p.Cost, &p.GrossProfit, &p.GrossMargin, &p.EnergyMWh,
			&p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

func (r *CustomerProfitRepository) GenerateDemo(ctx context.Context) (int, error) {
	rows, err := r.pool.Query(ctx, `SELECT id FROM customers LIMIT 50`)
	if err != nil {
		return 0, err
	}
	customers := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		customers = append(customers, id)
	}
	rows.Close()

	cnt := 0
	for _, cid := range customers {
		for i := 0; i < 6; i++ {
			t := time.Now().AddDate(0, -i, 0)
			ym := t.Format("2006-01")
			energy := 5000 + rand.Float64()*30000
			revenue := energy * (420 + rand.Float64()*40)
			cost := energy * (380 + rand.Float64()*30)
			profit := revenue - cost
			margin := 0.0
			if revenue > 0 {
				margin = profit / revenue * 100
			}
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO customer_profit
				   (customer_id, operating_month, revenue, cost, gross_profit, gross_margin, energy_mwh)
				 VALUES ($1,$2,$3,$4,$5,$6,$7)
				 ON CONFLICT (customer_id, operating_month) DO UPDATE SET
				   revenue = EXCLUDED.revenue, cost = EXCLUDED.cost,
				   gross_profit = EXCLUDED.gross_profit, gross_margin = EXCLUDED.gross_margin,
				   energy_mwh = EXCLUDED.energy_mwh`,
				cid, ym, revenue, cost, profit, margin, energy); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
