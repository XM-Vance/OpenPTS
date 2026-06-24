// 客户历史电量档案仓储：客户逐月电量（customer_monthly_energy）。
// 数据来源：文档解析「确认入库」选「客户电量档案」（市场化账单/月度电量），按活跃省隔离。
package db

import (
	"context"
	"fmt"
	"time"
)

type CustomerMonthlyEnergy struct {
	ID             string    `json:"id"`
	CustomerID     string    `json:"customer_id"`
	CustomerName   string    `json:"customer_name"`
	Month          string    `json:"month"`           // YYYY-MM
	MonthlyEnergy  float64   `json:"monthly_energy"`  // 月度电量
	AvgDailyEnergy *float64  `json:"avg_daily_energy"`
	VariationCV    *float64  `json:"variation_cv"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CustomerEnergyRepository struct{ pool *Pool }

func NewCustomerEnergyRepository(pool *Pool) *CustomerEnergyRepository {
	return &CustomerEnergyRepository{pool: pool}
}

// List 列出客户逐月电量；customerID 为空时返回本省全部客户。
func (r *CustomerEnergyRepository) List(ctx context.Context, customerID string, limit int) ([]*CustomerMonthlyEnergy, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []any{}
	q := `SELECT cme.id, cme.customer_id, c.user_name, cme.month, cme.monthly_energy,
			cme.avg_daily_energy, cme.variation_cv, cme.created_at, cme.updated_at
		 FROM customer_monthly_energy cme
		 JOIN customers c ON c.id = cme.customer_id
		 WHERE 1=1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND cme.org_id = $%d::uuid", len(args))
	}
	if customerID != "" {
		args = append(args, customerID)
		q += fmt.Sprintf(" AND cme.customer_id = $%d::uuid", len(args))
	}
	args = append(args, limit)
	q += fmt.Sprintf(" ORDER BY cme.month DESC, c.user_name LIMIT $%d", len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*CustomerMonthlyEnergy, 0, limit)
	for rows.Next() {
		var m CustomerMonthlyEnergy
		if err := rows.Scan(&m.ID, &m.CustomerID, &m.CustomerName, &m.Month,
			&m.MonthlyEnergy, &m.AvgDailyEnergy, &m.VariationCV,
			&m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

// Upsert 写入/更新某客户某月电量（文档「确认入库」使用）。
// 写操作要求具体活跃省；同省同客户同月覆盖更新。
func (r *CustomerEnergyRepository) Upsert(ctx context.Context, customerID, month string, monthlyEnergy float64) error {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return ErrOrgRequired
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO customer_monthly_energy (customer_id, month, monthly_energy, org_id)
		 VALUES ($1::uuid,$2,$3,$4::uuid)
		 ON CONFLICT (org_id, customer_id, month) DO UPDATE SET
		   monthly_energy = EXCLUDED.monthly_energy,
		   updated_at = now()`,
		customerID, month, monthlyEnergy, org)
	return err
}
