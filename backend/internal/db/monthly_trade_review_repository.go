// 月度交易复盘仓储。
// 2026-06 自 v1clone_f_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/shopspring/decimal"
)

// ─────────────── F2 月度交易复盘 ───────────────

type MonthlyTradeReview struct {
	ID              string          `json:"id"`
	OperatingMonth  string          `json:"operating_month"`
	RetailRevenue   decimal.Decimal `json:"retail_revenue"` // P4: numeric(18,4)
	WholesaleCost   decimal.Decimal `json:"wholesale_cost"` // P4: numeric(18,4)
	GrossProfit     decimal.Decimal `json:"gross_profit"`   // P4: = retail_revenue - wholesale_cost 精确
	GrossMargin     float64         `json:"gross_margin"`   // 毛利率 %（保持 float）
	ActiveCustomers int             `json:"active_customers"`
	TotalEnergyMWh  float64         `json:"total_energy_mwh"`
	Note            *string         `json:"note,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

type MonthlyTradeReviewRepository struct{ pool *Pool }

func NewMonthlyTradeReviewRepository(pool *Pool) *MonthlyTradeReviewRepository {
	return &MonthlyTradeReviewRepository{pool: pool}
}

// GetByMonth 取指定月份的月度复盘记录（供 overview 端点），按活跃组织隔离。
func (r *MonthlyTradeReviewRepository) GetByMonth(ctx context.Context, month string) (*MonthlyTradeReview, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	q := `SELECT id, operating_month, retail_revenue, wholesale_cost, gross_profit,
			gross_margin, active_customers, total_energy_mwh, note, created_at
		  FROM monthly_trade_review WHERE operating_month = $1`
	args := []any{month}
	if org, scoped := OrgFilter(ctx); scoped { // 总部「全部省」不过滤
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	var m MonthlyTradeReview
	err := r.pool.QueryRow(ctx, q, args...).Scan(
		&m.ID, &m.OperatingMonth, &m.RetailRevenue, &m.WholesaleCost,
		&m.GrossProfit, &m.GrossMargin, &m.ActiveCustomers, &m.TotalEnergyMWh,
		&m.Note, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MonthlyTradeReviewRepository) List(ctx context.Context, limit int) ([]*MonthlyTradeReview, error) {
	if limit <= 0 || limit > 24 {
		limit = 12
	}
	q := `SELECT id, operating_month, retail_revenue, wholesale_cost, gross_profit,
			gross_margin, active_customers, total_energy_mwh, note, created_at
		 FROM monthly_trade_review`
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped { // 按活跃组织隔离
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(args))
	}
	args = append(args, limit)
	q += fmt.Sprintf(" ORDER BY operating_month DESC LIMIT $%d", len(args))
	rows, err := r.pool.Query(ctx, q, args...)
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
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped { // 总部/全部省：复盘 demo 落到 FJ（行须带 org_id 以对齐唯一键）
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	notes := []string{"稳健增长", "新客户拉动", "电价波动较大", "成本压力", "节假日下滑"}
	cnt := 0
	for i := 0; i < 12; i++ {
		ym := monthsAgoYM(i)
		energy := 80000.0 + rand.Float64()*40000
		// P4: 金额 decimal（分项舍 4 位再相减），gross_profit = revenue - cost 精确；毛利率仍 float。
		revenue := decimal.NewFromFloat(energy).Mul(decimal.NewFromFloat(430 + rand.Float64()*40)).Round(4)
		cost := decimal.NewFromFloat(energy).Mul(decimal.NewFromFloat(385 + rand.Float64()*30)).Round(4)
		profit := revenue.Sub(cost)
		margin, _ := profit.Div(revenue).Mul(decimal.NewFromInt(100)).Float64()
		customers := 8 + rand.IntN(8)
		note := notes[rand.IntN(len(notes))]
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO monthly_trade_review
			   (org_id, operating_month, retail_revenue, wholesale_cost, gross_profit,
			    gross_margin, active_customers, total_energy_mwh, note, is_demo)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,TRUE)
			 ON CONFLICT (org_id, operating_month) DO UPDATE SET
			   retail_revenue = EXCLUDED.retail_revenue,
			   wholesale_cost = EXCLUDED.wholesale_cost,
			   gross_profit = EXCLUDED.gross_profit,
			   gross_margin = EXCLUDED.gross_margin,
			   active_customers = EXCLUDED.active_customers,
			   total_energy_mwh = EXCLUDED.total_energy_mwh,
			   note = EXCLUDED.note, is_demo = TRUE`,
			orgID, ym, revenue, cost, profit, margin, customers, energy, note); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
