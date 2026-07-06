// 负荷数据仓储：历史负荷曲线查询 + 单日曲线写入（演示数据用）。
package db

import (
	"context"
	"encoding/json"
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
//
// 同时写两份：
//   - user_load_data：96 点规范化数组（curve_96 NUMERIC[]），供 calendar 端点
//   - unified_load_curve：JSONB 曲线（curve_data），供 curves 端点（此前是死表，唯一写入是 demo）
//
// 镜像写 unified_load_curve 让"统一负荷曲线"下游消费者（load/curves、MpMissing 导出）
// 不再永远为空。aggregation_method='direct'（单计量点直写，非多计量点汇聚）。
func (r *LoadRepository) UpsertCurve(
	ctx context.Context, customerID uuid.UUID, d time.Time, curve []float64, total float64,
) error {
	org, err := MustScoped(ctx)
	if err != nil {
		return err
	}
	const q = `
		INSERT INTO user_load_data (customer_id, date, curve_96, total_load, quality_flag, org_id)
		VALUES ($1, $2, $3, $4, 'ok', $5::uuid)
		ON CONFLICT (org_id, customer_id, date)
		DO UPDATE SET curve_96 = EXCLUDED.curve_96, total_load = EXCLUDED.total_load`
	if _, err = r.pool.Exec(ctx, q, customerID, d, curve, total, org); err != nil {
		return err
	}

	// 镜像写 unified_load_curve：curve []float64 → JSONB [[period, value], ...]
	curveJSON, err := curveToJSONB(curve)
	if err != nil {
		return fmt.Errorf("序列化曲线失败: %w", err)
	}
	const qUnified = `
		INSERT INTO unified_load_curve (org_id, customer_id, date, curve_data, aggregation_method, total_daily_load, data_quality)
		VALUES ($1::uuid, $2, $3, $4, 'direct', $5, 'normal')
		ON CONFLICT (org_id, customer_id, date)
		DO UPDATE SET curve_data = EXCLUDED.curve_data,
		              aggregation_method = EXCLUDED.aggregation_method,
		              total_daily_load = EXCLUDED.total_daily_load,
		              data_quality = EXCLUDED.data_quality`
	_, err = r.pool.Exec(ctx, qUnified, org, customerID, d, curveJSON, total)
	return err
}

// curveToJSONB 把 96 点 float64 数组转成紧凑 JSON 数组 [[period, value], ...]。
// 下游 CustomerCurves 把 curve_data 当 json.RawMessage 透传给前端，格式灵活。
func curveToJSONB(curve []float64) ([]byte, error) {
	pairs := make([][2]float64, len(curve))
	for i, v := range curve {
		pairs[i] = [2]float64{float64(i + 1), v}
	}
	return json.Marshal(pairs)
}
