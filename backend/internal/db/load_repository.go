// 负荷数据仓储：历史负荷曲线查询 + 单日曲线写入（演示数据用）。
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DailyLoadCurve 单日 96 点负荷曲线。
type DailyLoadCurve struct {
	Date    time.Time `json:"date"`
	Curve96 []float64 `json:"curve_96"`
	Total   *float64  `json:"total_load,omitempty"`
}

type LoadRepository struct {
	pool *Pool
}

func NewLoadRepository(pool *Pool) *LoadRepository {
	return &LoadRepository{pool: pool}
}

// GetRecentCurves 返回某客户 before 之前最近 limit 天的负荷曲线，按日期升序返回
// （最旧在前，便于算法做时间衰减加权）。
func (r *LoadRepository) GetRecentCurves(
	ctx context.Context, customerID uuid.UUID, before time.Time, limit int,
) ([]*DailyLoadCurve, error) {
	q := `
		SELECT date, curve_96, total_load
		FROM user_load_data
		WHERE customer_id = $1 AND date < $2`
	args := []any{customerID, before}
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += `
		ORDER BY date DESC
		LIMIT $3`
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*DailyLoadCurve, 0, limit)
	for rows.Next() {
		var d DailyLoadCurve
		if err := rows.Scan(&d.Date, &d.Curve96, &d.Total); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 反转为升序（最旧在前）
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}

// UpsertCurve 写入或更新某客户某天的负荷曲线。
func (r *LoadRepository) UpsertCurve(
	ctx context.Context, customerID uuid.UUID, d time.Time, curve []float64, total float64,
) error {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return ErrOrgRequired
	}
	const q = `
		INSERT INTO user_load_data (customer_id, date, curve_96, total_load, quality_flag, org_id)
		VALUES ($1, $2, $3, $4, 'ok', $5::uuid)
		ON CONFLICT (org_id, customer_id, date)
		DO UPDATE SET curve_96 = EXCLUDED.curve_96, total_load = EXCLUDED.total_load`
	_, err := r.pool.Exec(ctx, q, customerID, d, curve, total, org)
	return err
}
