// 月度交易复盘仓储。
// 2026-06 自 v1clone_f_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"math/rand"
	"time"
)

// ─────────────── F2 月度交易复盘 ───────────────

type MonthlyTradeReview struct {
	ID              string    `json:"id"`
	OperatingMonth  string    `json:"operating_month"`
	RetailRevenue   float64   `json:"retail_revenue"`
	WholesaleCost   float64   `json:"wholesale_cost"`
	GrossProfit     float64   `json:"gross_profit"`
	GrossMargin     float64   `json:"gross_margin"`
	ActiveCustomers int       `json:"active_customers"`
	TotalEnergyMWh  float64   `json:"total_energy_mwh"`
	Note            *string   `json:"note,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type MonthlyTradeReviewRepository struct{ pool *Pool }

func NewMonthlyTradeReviewRepository(pool *Pool) *MonthlyTradeReviewRepository {
	return &MonthlyTradeReviewRepository{pool: pool}
}

func (r *MonthlyTradeReviewRepository) List(ctx context.Context, limit int) ([]*MonthlyTradeReview, error) {
	if limit <= 0 || limit > 24 {
		limit = 12
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, operating_month, retail_revenue, wholesale_cost, gross_profit,
			gross_margin, active_customers, total_energy_mwh, note, created_at
		 FROM monthly_trade_review ORDER BY operating_month DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*MonthlyTradeReview, 0, limit)
	for rows.Next() {
		var m MonthlyTradeReview
		if err := rows.Scan(&m.ID, &m.OperatingMonth, &m.RetailRevenue, &m.WholesaleCost,
			&m.GrossProfit, &m.GrossMargin, &m.ActiveCustomers, &m.TotalEnergyMWh,
			&m.Note, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

func (r *MonthlyTradeReviewRepository) GenerateDemo(ctx context.Context) (int, error) {
	notes := []string{"稳健增长", "新客户拉动", "电价波动较大", "成本压力", "节假日下滑"}
	cnt := 0
	for i := 0; i < 12; i++ {
		t := time.Now().AddDate(0, -i, 0)
		ym := t.Format("2006-01")
		energy := 80000.0 + rand.Float64()*40000
		revenue := energy * (430 + rand.Float64()*40)
		cost := energy * (385 + rand.Float64()*30)
		profit := revenue - cost
		margin := profit / revenue * 100
		customers := 8 + rand.Intn(8)
		note := notes[rand.Intn(len(notes))]
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO monthly_trade_review
			   (operating_month, retail_revenue, wholesale_cost, gross_profit,
			    gross_margin, active_customers, total_energy_mwh, note)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			 ON CONFLICT (operating_month) DO UPDATE SET
			   retail_revenue = EXCLUDED.retail_revenue,
			   wholesale_cost = EXCLUDED.wholesale_cost,
			   gross_profit = EXCLUDED.gross_profit,
			   gross_margin = EXCLUDED.gross_margin,
			   active_customers = EXCLUDED.active_customers,
			   total_energy_mwh = EXCLUDED.total_energy_mwh,
			   note = EXCLUDED.note`,
			ym, revenue, cost, profit, margin, customers, energy, note); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
