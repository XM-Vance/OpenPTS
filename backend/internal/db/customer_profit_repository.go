// 客户利润仓储。
// 2026-06 自 v1clone_f_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/shopspring/decimal"
)

// ─────────────── F1 客户利润 ───────────────

type CustomerProfit struct {
	ID             string          `json:"id"`
	CustomerID     string          `json:"customer_id"`
	CustomerName   string          `json:"customer_name,omitempty"`
	OperatingMonth string          `json:"operating_month"`
	Revenue        decimal.Decimal `json:"revenue"`      // P4: numeric(18,4)
	Cost           decimal.Decimal `json:"cost"`         // P4: numeric(18,4)
	GrossProfit    decimal.Decimal `json:"gross_profit"` // P4: numeric(18,4)，= revenue - cost 精确
	GrossMargin    float64         `json:"gross_margin"` // 毛利率 %（比率，保持 float）
	EnergyMWh      float64         `json:"energy_mwh"`
	IsEstimate     bool            `json:"is_estimate"` // Phase 3：true=签约前测算，false=签约后实际结算
	CreatedAt      time.Time       `json:"created_at"`
}

type CustomerProfitRepository struct{ pool *Pool }

func NewCustomerProfitRepository(pool *Pool) *CustomerProfitRepository {
	return &CustomerProfitRepository{pool: pool}
}

// Upsert 写入/更新一条客户利润（P0 结算落库用）。写操作要求具体活跃组织；
// org 唯一键含 is_estimate，故同省·同客户·同月的测算/实际各自 upsert、互不覆盖。
func (r *CustomerProfitRepository) Upsert(ctx context.Context, customerID, operatingMonth string,
	revenue, cost, grossProfit decimal.Decimal, grossMargin, energyMWh float64, isEstimate bool) error {
	org, err := MustScoped(ctx)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO customer_profit
		  (org_id, customer_id, operating_month, revenue, cost, gross_profit, gross_margin, energy_mwh, is_estimate)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (org_id, customer_id, operating_month, is_estimate) DO UPDATE SET
		  revenue = EXCLUDED.revenue, cost = EXCLUDED.cost,
		  gross_profit = EXCLUDED.gross_profit, gross_margin = EXCLUDED.gross_margin,
		  energy_mwh = EXCLUDED.energy_mwh`,
		org, customerID, operatingMonth, revenue, cost, grossProfit, grossMargin, energyMWh, isEstimate)
	return err
}

// List 返回客户利润。isEstimate 区分测算(true)与实际(false)，默认查实际，避免两类混在一张表里互相污染。
func (r *CustomerProfitRepository) List(ctx context.Context, month string, limit int, isEstimate bool) ([]*CustomerProfit, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := `SELECT p.id, p.customer_id::text, c.user_name, p.operating_month,
			p.revenue, p.cost, p.gross_profit, p.gross_margin, p.energy_mwh, p.is_estimate, p.created_at
		  FROM customer_profit p JOIN customers c ON c.id = p.customer_id
		  WHERE p.is_estimate = $1`
	args := []any{isEstimate}
	if org, scoped := OrgFilter(ctx); scoped { // 按活跃组织隔离（总部「全部省」不过滤）
		args = append(args, org)
		q += fmt.Sprintf(" AND p.org_id = $%d::uuid", len(args))
	}
	if month != "" {
		args = append(args, month)
		q += fmt.Sprintf(" AND p.operating_month = $%d", len(args))
	}
	args = append(args, limit)
	q += fmt.Sprintf(" ORDER BY p.operating_month DESC, p.gross_profit DESC LIMIT $%d", len(args))
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
			&p.IsEstimate, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

func (r *CustomerProfitRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 解析活跃组织（HQ/全部省时回退 FJ）：利润行须带 org_id 才能对齐 org 唯一键。
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}

	// service 客户 → 实际结算(is_estimate=false)；intent 客户 → 签约前测算(true)。
	actualIDs, err := r.pickCustomers(ctx, orgID, "lifecycle_stage NOT IN ('intent','lead')", 50)
	if err != nil {
		return 0, err
	}
	estimateIDs, err := r.pickCustomers(ctx, orgID, "lifecycle_stage = 'intent'", 50)
	if err != nil {
		return 0, err
	}

	cnt := 0
	for _, batch := range []struct {
		ids      []string
		estimate bool
	}{{actualIDs, false}, {estimateIDs, true}} {
		n, err := r.genProfitRows(ctx, orgID, batch.ids, batch.estimate)
		cnt += n
		if err != nil {
			return cnt, err
		}
	}
	return cnt, nil
}

// pickCustomers 取指定省、指定生命周期阶段的客户 id。
func (r *CustomerProfitRepository) pickCustomers(ctx context.Context, orgID, stageCond string, limit int) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT id FROM customers WHERE org_id = $1::uuid AND %s LIMIT %d`, stageCond, limit), orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// genProfitRows 为一批客户生成近 6 个月利润行（is_estimate 决定测算/实际）。
func (r *CustomerProfitRepository) genProfitRows(ctx context.Context, orgID string, customerIDs []string, isEstimate bool) (int, error) {
	cnt := 0
	for _, cid := range customerIDs {
		for i := 0; i < 6; i++ {
			ym := monthsAgoYM(i)
			energy := 5000 + rand.Float64()*30000
			// P4: 金额 decimal（分项舍 4 位再相减），gross_profit = revenue - cost 精确；毛利率仍 float。
			revenue := decimal.NewFromFloat(energy).Mul(decimal.NewFromFloat(420 + rand.Float64()*40)).Round(4)
			cost := decimal.NewFromFloat(energy).Mul(decimal.NewFromFloat(380 + rand.Float64()*30)).Round(4)
			profit := revenue.Sub(cost)
			margin := 0.0
			if revenue.IsPositive() {
				margin, _ = profit.Div(revenue).Mul(decimal.NewFromInt(100)).Float64()
			}
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO customer_profit
				   (org_id, customer_id, operating_month, revenue, cost, gross_profit, gross_margin, energy_mwh, is_estimate, is_demo)
				 VALUES ($1::uuid,$2,$3,$4,$5,$6,$7,$8,$9,TRUE)
				 ON CONFLICT (org_id, customer_id, operating_month, is_estimate) DO UPDATE SET
				   revenue = EXCLUDED.revenue, cost = EXCLUDED.cost,
				   gross_profit = EXCLUDED.gross_profit, gross_margin = EXCLUDED.gross_margin,
				   energy_mwh = EXCLUDED.energy_mwh, is_demo = TRUE`,
				orgID, cid, ym, revenue, cost, profit, margin, energy, isEstimate); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
