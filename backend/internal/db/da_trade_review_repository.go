// 日前交易复盘仓储。
// 2026-06 自 v1clone_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/shopspring/decimal"
)

// ─────────────── D3 日前交易复盘 ───────────────

type DATradeReview struct {
	ID                string          `json:"id"`
	TradingDate       time.Time       `json:"trading_date"`
	DeclaredEnergyMWh float64         `json:"declared_energy_mwh"`
	ClearedEnergyMWh  float64         `json:"cleared_energy_mwh"`
	AvgDeclaredPrice  decimal.Decimal `json:"avg_declared_price"` // P4: numeric(18,4)
	AvgClearedPrice   decimal.Decimal `json:"avg_cleared_price"`  // P4: numeric(18,4)
	Revenue           decimal.Decimal `json:"revenue"`            // P4: numeric(18,4)
	Note              *string         `json:"note,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}

type DATradeReviewRepository struct{ pool *Pool }

func NewDATradeReviewRepository(pool *Pool) *DATradeReviewRepository {
	return &DATradeReviewRepository{pool: pool}
}

func (r *DATradeReviewRepository) List(ctx context.Context, limit int) ([]*DATradeReview, error) {
	if limit <= 0 || limit > 90 {
		limit = 30
	}
	q := `SELECT id, trading_date, declared_energy_mwh, cleared_energy_mwh,
			avg_declared_price, avg_cleared_price, revenue, note, created_at
		 FROM day_ahead_trade_review`
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped { // 按活跃组织隔离
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(args))
	}
	args = append(args, limit)
	q += fmt.Sprintf(" ORDER BY trading_date DESC LIMIT $%d", len(args))
	rows, err := r.pool.Query(ctx, q, args...)
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
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped { // 总部/全部省：复盘 demo 落到 FJ（行须带 org_id 以对齐唯一键）
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	notes := []string{"成交良好", "出清偏低", "现货反弹", "申报激进", "保守策略"}
	cnt := 0
	for i := 0; i < 30; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		declared := 1800 + rand.Float64()*600
		cleared := declared * (0.85 + rand.Float64()*0.15)
		// P4: 价格/收入 decimal（revenue = cleared*clrPrice）；电量保持 float。
		decPrice := decimal.NewFromFloat(360 + rand.Float64()*120).Round(4)
		clrPrice := decPrice.Mul(decimal.NewFromFloat(0.85 + rand.Float64()*0.25)).Round(4)
		revenue := decimal.NewFromFloat(cleared).Mul(clrPrice).Round(4)
		note := notes[rand.IntN(len(notes))]
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO day_ahead_trade_review
			   (org_id, trading_date, declared_energy_mwh, cleared_energy_mwh,
			    avg_declared_price, avg_cleared_price, revenue, note, is_demo)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,TRUE)
			 ON CONFLICT (org_id, trading_date) DO UPDATE SET
			   declared_energy_mwh = EXCLUDED.declared_energy_mwh,
			   cleared_energy_mwh = EXCLUDED.cleared_energy_mwh,
			   avg_declared_price = EXCLUDED.avg_declared_price,
			   avg_cleared_price = EXCLUDED.avg_cleared_price,
			   revenue = EXCLUDED.revenue, note = EXCLUDED.note, is_demo = TRUE`,
			orgID, d, declared, cleared, decPrice, clrPrice, revenue, note); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
