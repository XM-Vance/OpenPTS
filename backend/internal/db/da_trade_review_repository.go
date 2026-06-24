// 日前交易复盘仓储。
// 2026-06 自 v1clone_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"math/rand"
	"time"
)

// ─────────────── D3 日前交易复盘 ───────────────

type DATradeReview struct {
	ID                string    `json:"id"`
	TradingDate       time.Time `json:"trading_date"`
	DeclaredEnergyMWh float64   `json:"declared_energy_mwh"`
	ClearedEnergyMWh  float64   `json:"cleared_energy_mwh"`
	AvgDeclaredPrice  float64   `json:"avg_declared_price"`
	AvgClearedPrice   float64   `json:"avg_cleared_price"`
	Revenue           float64   `json:"revenue"`
	Note              *string   `json:"note,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type DATradeReviewRepository struct{ pool *Pool }

func NewDATradeReviewRepository(pool *Pool) *DATradeReviewRepository {
	return &DATradeReviewRepository{pool: pool}
}

func (r *DATradeReviewRepository) List(ctx context.Context, limit int) ([]*DATradeReview, error) {
	if limit <= 0 || limit > 90 {
		limit = 30
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, trading_date, declared_energy_mwh, cleared_energy_mwh,
			avg_declared_price, avg_cleared_price, revenue, note, created_at
		 FROM day_ahead_trade_review ORDER BY trading_date DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*DATradeReview, 0, limit)
	for rows.Next() {
		var t DATradeReview
		if err := rows.Scan(&t.ID, &t.TradingDate, &t.DeclaredEnergyMWh,
			&t.ClearedEnergyMWh, &t.AvgDeclaredPrice, &t.AvgClearedPrice,
			&t.Revenue, &t.Note, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

func (r *DATradeReviewRepository) GenerateDemo(ctx context.Context) (int, error) {
	notes := []string{"成交良好", "出清偏低", "现货反弹", "申报激进", "保守策略"}
	cnt := 0
	for i := 0; i < 30; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		declared := 1800 + rand.Float64()*600
		cleared := declared * (0.85 + rand.Float64()*0.15)
		decPrice := 360 + rand.Float64()*120
		clrPrice := decPrice * (0.85 + rand.Float64()*0.25)
		revenue := cleared * clrPrice
		note := notes[rand.Intn(len(notes))]
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO day_ahead_trade_review
			   (trading_date, declared_energy_mwh, cleared_energy_mwh,
			    avg_declared_price, avg_cleared_price, revenue, note)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)
			 ON CONFLICT (trading_date) DO UPDATE SET
			   declared_energy_mwh = EXCLUDED.declared_energy_mwh,
			   cleared_energy_mwh = EXCLUDED.cleared_energy_mwh,
			   avg_declared_price = EXCLUDED.avg_declared_price,
			   avg_cleared_price = EXCLUDED.avg_cleared_price,
			   revenue = EXCLUDED.revenue, note = EXCLUDED.note`,
			d, declared, cleared, decPrice, clrPrice, revenue, note); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
