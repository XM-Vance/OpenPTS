// 计量数据底座仓储（WP0.3）：raw_meter_data 批量幂等导入 + 多表计聚合。
// 聚合口径：同一 (customer, date) 下 Σ(表计 96 点曲线 × 倍率) → user_load_data.curve_96
//（kW/15min）+ total_load（kWh = Σcurve/4），镜像写 unified_load_curve（method='meter_sum'，
// 区别于单源直写的 'direct'）。设计出处：docs/calculations.md §4.4、迁移 0006 注释。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// RawMeterRow 单表计·单日原始读数（curve96 为表计示值曲线，聚合时 ×倍率）。
type RawMeterRow struct {
	MeterID    string    `json:"meter_id"`
	CustomerID string    `json:"customer_id"`
	Date       time.Time `json:"date"`
	Multiplier float64   `json:"multiplier"` // CT/PT 倍率，缺省 1
	Curve96    []float64 `json:"curve_96"`   // 表计示值 kW/15min
}

type MeterRepository struct {
	pool *Pool
}

func NewMeterRepository(pool *Pool) *MeterRepository {
	return &MeterRepository{pool: pool}
}

// periodPoint period_data JSONB 的单元素。
type periodPoint struct {
	Period int     `json:"period"`
	Value  float64 `json:"value"`
}

// curveToPeriodJSON 96 点数组 → period_data JSONB [{"period":1,"value":...},...]。
func curveToPeriodJSON(curve []float64) ([]byte, error) {
	pts := make([]periodPoint, len(curve))
	for i, v := range curve {
		pts[i] = periodPoint{Period: i + 1, Value: v}
	}
	return json.Marshal(pts)
}

// periodJSONToCurve 反向解析（缺时段置 NaN）。
func periodJSONToCurve(data []byte, n int) ([]float64, error) {
	var pts []periodPoint
	if err := json.Unmarshal(data, &pts); err != nil {
		return nil, err
	}
	curve := make([]float64, n)
	for i := range curve {
		curve[i] = math.NaN()
	}
	for _, p := range pts {
		if p.Period >= 1 && p.Period <= n {
			curve[p.Period-1] = p.Value
		}
	}
	return curve, nil
}

// BulkUpsertRaw 批量幂等写 raw_meter_data（(meter_id, date) 冲突覆盖；迁移 0138 唯一索引）。
func (r *MeterRepository) BulkUpsertRaw(ctx context.Context, rows []*RawMeterRow) (int, error) {
	org, err := MustScoped(ctx)
	if err != nil {
		return 0, err
	}
	const q = `
		INSERT INTO raw_meter_data (meter_id, customer_id, org_id, multiplier, date, period_data, data_source)
		VALUES ($1, $2::uuid, $3::uuid, $4, $5, $6::jsonb, 'import')
		ON CONFLICT (meter_id, date) DO UPDATE SET
			customer_id = EXCLUDED.customer_id,
			org_id      = EXCLUDED.org_id,
			multiplier  = EXCLUDED.multiplier,
			period_data = EXCLUDED.period_data,
			data_source = 'import',
			imported_at = now()`
	written := 0
	for _, row := range rows {
		pd, err := curveToPeriodJSON(row.Curve96)
		if err != nil {
			return written, fmt.Errorf("表计 %s 曲线序列化失败: %w", row.MeterID, err)
		}
		var cust *string
		if row.CustomerID != "" {
			cust = &row.CustomerID
		}
		if _, err := r.pool.Exec(ctx, q, row.MeterID, cust, org, row.Multiplier, row.Date, string(pd)); err != nil {
			return written, err
		}
		written++
	}
	return written, nil
}

// AggregatedCurve 聚合结果（单客户·单日）。
type AggregatedCurve struct {
	CustomerID uuid.UUID
	Date       time.Time
	Curve96    []float64 // kW/15min（已 ×倍率求和）
	TotalKWh   float64   // Σcurve/4
	Meters     int       // 参与聚合的表计数
}

// AggregateMeters 把指定日期区间的 raw_meter_data 按 (customer, date) 聚合
//（Σ 表计曲线×倍率），事务写 user_load_data + unified_load_curve（method='meter_sum'）。
// 返回写入的 (customer, date) 数。无 customer 关联的表计行跳过（计入 skipped）。
func (r *MeterRepository) AggregateMeters(ctx context.Context, start, end time.Time) (written, skipped int, err error) {
	org, scoped := OrgFilter(ctx)
	q := `SELECT customer_id::text, date, multiplier, period_data
		FROM raw_meter_data
		WHERE date >= $1 AND date <= $2 AND customer_id IS NOT NULL`
	args := []any{start, end}
	if scoped {
		args = append(args, org)
		q += ` AND org_id = $3::uuid`
	}
	q += ` ORDER BY date, customer_id`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	type key struct {
		customer uuid.UUID
		date     time.Time
	}
	acc := map[key]*AggregatedCurve{}
	order := make([]key, 0, 32)
	for rows.Next() {
		var custID string
		var d time.Time
		var mult float64
		var pd []byte
		if err := rows.Scan(&custID, &d, &mult, &pd); err != nil {
			return written, skipped, err
		}
		curve, err := periodJSONToCurve(pd, 96)
		if err != nil {
			skipped++
			continue
		}
		cid, err := uuid.Parse(custID)
		if err != nil {
			skipped++
			continue
		}
		k := key{cid, d}
		agg, ok := acc[k]
		if !ok {
			agg = &AggregatedCurve{CustomerID: cid, Date: d, Curve96: make([]float64, 96)}
			for i := range agg.Curve96 {
				agg.Curve96[i] = 0
			}
			acc[k] = agg
			order = append(order, k)
		}
		for i, v := range curve {
			agg.Curve96[i] += v * mult
		}
		agg.Meters++
	}
	if err := rows.Err(); err != nil {
		return written, skipped, err
	}

	for _, k := range order {
		agg := acc[k]
		total := 0.0
		for _, v := range agg.Curve96 {
			total += v
		}
		agg.TotalKWh = total / 4
		if err := r.upsertMeterAggregated(ctx, agg); err != nil {
			return written, skipped, err
		}
		written++
	}
	return written, skipped, nil
}

// upsertMeterAggregated 双表事务写（对齐 load_repository.UpsertCurve 的两表一致性模式，
// 但 aggregation_method='meter_sum' 标记多表计汇聚来源）。
func (r *MeterRepository) upsertMeterAggregated(ctx context.Context, agg *AggregatedCurve) error {
	org, err := MustScoped(ctx)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const qUser = `
		INSERT INTO user_load_data (customer_id, date, curve_96, total_load, quality_flag, org_id)
		VALUES ($1, $2, $3, $4, 'ok', $5::uuid)
		ON CONFLICT (org_id, customer_id, date)
		DO UPDATE SET curve_96 = EXCLUDED.curve_96, total_load = EXCLUDED.total_load`
	if _, err := tx.Exec(ctx, qUser, agg.CustomerID, agg.Date, agg.Curve96, agg.TotalKWh, org); err != nil {
		return err
	}

	curveJSON, err := curveToJSONB(agg.Curve96)
	if err != nil {
		return err
	}
	const qUnified = `
		INSERT INTO unified_load_curve (org_id, customer_id, date, curve_data, aggregation_method, total_daily_load, data_quality, extra)
		VALUES ($1::uuid, $2, $3, $4, 'meter_sum', $5, 'normal', $6::jsonb)
		ON CONFLICT (org_id, customer_id, date)
		DO UPDATE SET curve_data = EXCLUDED.curve_data,
		              aggregation_method = EXCLUDED.aggregation_method,
		              total_daily_load = EXCLUDED.total_daily_load,
		              data_quality = EXCLUDED.data_quality,
		              extra = EXCLUDED.extra`
	if _, err := tx.Exec(ctx, qUnified, org, agg.CustomerID, agg.Date, curveJSON, agg.TotalKWh,
		fmt.Sprintf(`{"meters":%d}`, agg.Meters)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
