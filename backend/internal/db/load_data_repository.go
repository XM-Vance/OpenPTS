// 负荷数据查询仓储：支撑前端 /load-data/* 的查询端点。
//
// 数据源（均建表于 0005/0006，无 org_id，通过 JOIN customers 过滤活跃省）：
//   - raw_meter_data（0005）：电表原始数据（分区表）
//   - raw_mp_data（0005）：计量点原始数据（分区表）—— mp-missing 导出源
//   - unified_load_curve（0006）：汇聚后的统一曲线（curves 端点）
//   - user_load_data（0006）：96 点规范化（calendar 端点）
//
// 注意：与 load_characteristics_ext_repository.go（客户特征画像）不同——本仓储是计量数据管理。
// 校准（calibration）类端点需算法 + 无校准表，本仓储不实现（留 stub）。
package db

import (
	"errors"
	"math"
	"strings"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LoadDataRepository struct{ pool *Pool }

func NewLoadDataRepository(pool *Pool) *LoadDataRepository {
	return &LoadDataRepository{pool: pool}
}

// ─── Customers：负荷数据客户列表（JOIN 统计量）───

type LoadDataCustomer struct {
	ID            uuid.UUID  `json:"id"`
	CustomerName  string     `json:"customer_name"`
	AccountNo     *string    `json:"account_no,omitempty"`
	LifecycleStage string    `json:"lifecycle_stage"`
	DataPoints    int        `json:"data_points"`    // unified_load_curve 记录数
	LatestDate    *time.Time `json:"latest_date,omitempty"`
	DataQuality   *string    `json:"data_quality,omitempty"`
}

// ListCustomers 列出有负荷数据的客户（带 search/lifecycle 筛选、分页）。
func (r *LoadDataRepository) ListCustomers(
	ctx context.Context, page, pageSize int, search string,
) ([]*LoadDataCustomer, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	base := `FROM customers c
		LEFT JOIN (SELECT customer_id, COUNT(*) AS n, MAX(date) AS latest,
			MAX(data_quality) AS dq FROM unified_load_curve GROUP BY customer_id) u
		ON u.customer_id = c.id`
	args := []any{}
	where := " WHERE 1=1"
	idx := 1
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		where += fmt.Sprintf(" AND c.org_id = $%d::uuid", idx)
		idx++
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND c.user_name ILIKE $%d", idx)
		idx++
	}
	// 总数
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) "+base+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, (page-1)*pageSize)
	q := fmt.Sprintf(`SELECT c.id, c.user_name, c.account_no, c.lifecycle_stage,
			COALESCE(u.n,0), u.latest, u.dq %s%s ORDER BY COALESCE(u.latest,'1970-01-01') DESC
			LIMIT $%d OFFSET $%d`, base, where, idx, idx+1)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*LoadDataCustomer, 0, pageSize)
	for rows.Next() {
		var c LoadDataCustomer
		if err := rows.Scan(&c.ID, &c.CustomerName, &c.AccountNo, &c.LifecycleStage,
			&c.DataPoints, &c.LatestDate, &c.DataQuality); err != nil {
			return nil, 0, err
		}
		list = append(list, &c)
	}
	return list, total, rows.Err()
}

// ListSignedCustomers 仅列出已签约（在服）客户。
func (r *LoadDataRepository) ListSignedCustomers(ctx context.Context) ([]*LoadDataCustomer, error) {
	q := `SELECT c.id, c.user_name, c.account_no, c.lifecycle_stage,
			COALESCE(u.n,0), u.latest, u.dq
		  FROM customers c
		  LEFT JOIN (SELECT customer_id, COUNT(*) AS n, MAX(date) AS latest,
			MAX(data_quality) AS dq FROM unified_load_curve GROUP BY customer_id) u
			ON u.customer_id = c.id
		  WHERE c.lifecycle_stage IN ('service','intention')`
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND c.org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY c.user_name ASC LIMIT 500"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*LoadDataCustomer
	for rows.Next() {
		var c LoadDataCustomer
		if err := rows.Scan(&c.ID, &c.CustomerName, &c.AccountNo, &c.LifecycleStage,
			&c.DataPoints, &c.LatestDate, &c.DataQuality); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}

// ─── CustomerDetail：客户负荷汇总 ───

type LoadDataDetail struct {
	CustomerID   uuid.UUID    `json:"customer_id"`
	CustomerName string       `json:"customer_name"`
	Summary      map[string]any `json:"summary"`
	Recent       []LoadDataDay `json:"recent"`
}

type LoadDataDay struct {
	Date           time.Time    `json:"date"`
	TotalLoad      *float64     `json:"total_load,omitempty"`
	DataQuality    string       `json:"data_quality"`
	CurveData      json.RawMessage `json:"curve_data,omitempty"`
}

// GetCustomerDetail 取客户负荷汇总 + 近期日数据。
func (r *LoadDataRepository) GetCustomerDetail(
	ctx context.Context, customerID uuid.UUID, days int,
) (*LoadDataDetail, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	// 客户基础信息
	var name string
	qC := `SELECT user_name FROM customers WHERE id=$1`
	argsC := []any{customerID}
	if org, scoped := OrgFilter(ctx); scoped {
		argsC = append(argsC, org)
		qC += fmt.Sprintf(" AND org_id = $%d::uuid", len(argsC))
	}
	if err := r.pool.QueryRow(ctx, qC, argsC...).Scan(&name); err != nil {
		return nil, err
	}
	since := time.Now().AddDate(0, 0, -days)
	q := `SELECT date, total_daily_load, data_quality, curve_data
		  FROM unified_load_curve WHERE customer_id=$1 AND date >= $2`
	args := []any{customerID, since}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY date DESC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var recent []LoadDataDay
	var totalLoad float64
	for rows.Next() {
		var d LoadDataDay
		var curve []byte
		if err := rows.Scan(&d.Date, &d.TotalLoad, &d.DataQuality, &curve); err != nil {
			return nil, err
		}
		d.CurveData = json.RawMessage(curve)
		if d.TotalLoad != nil {
			totalLoad += *d.TotalLoad
		}
		recent = append(recent, d)
	}
	return &LoadDataDetail{
		CustomerID:   customerID,
		CustomerName: name,
		Summary: map[string]any{
			"days":       len(recent),
			"total_load": totalLoad,
		},
		Recent: recent,
	}, rows.Err()
}

// ─── Calendar：日历热力图（数据质量按日）───

type CalendarDay struct {
	Date        string `json:"date"`
	QualityFlag string `json:"quality_flag"`
	TotalLoad   *float64 `json:"total_load,omitempty"`
}

// CustomerCalendar 取客户指定月份的数据质量日历（user_load_data 的 quality_flag）。
func (r *LoadDataRepository) CustomerCalendar(
	ctx context.Context, customerID uuid.UUID, month string,
) ([]*CalendarDay, error) {
	// month 形如 2026-06
	q := `SELECT date::text, quality_flag, total_load FROM user_load_data
		  WHERE customer_id=$1 AND date::text LIKE $2`
	args := []any{customerID, month + "-%"}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY date ASC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*CalendarDay
	for rows.Next() {
		var d CalendarDay
		if err := rows.Scan(&d.Date, &d.QualityFlag, &d.TotalLoad); err != nil {
			return nil, err
		}
		out = append(out, &d)
	}
	return out, rows.Err()
}

// ─── Curves：客户曲线（unified_load_curve）───

// CustomerCurves 取客户指定日期范围的统一曲线。
func (r *LoadDataRepository) CustomerCurves(
	ctx context.Context, customerID uuid.UUID, start, end time.Time,
) ([]*LoadDataDay, error) {
	q := `SELECT date, total_daily_load, data_quality, curve_data
		  FROM unified_load_curve WHERE customer_id=$1 AND date >= $2 AND date <= $3`
	args := []any{customerID, start, end}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY date ASC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*LoadDataDay
	for rows.Next() {
		var d LoadDataDay
		var curve []byte
		if err := rows.Scan(&d.Date, &d.TotalLoad, &d.DataQuality, &curve); err != nil {
			return nil, err
		}
		d.CurveData = json.RawMessage(curve)
		out = append(out, &d)
	}
	return out, rows.Err()
}

// ─── MpMissing：缺数计量点导出（raw_mp_data 分区表）───

type MpMissingItem struct {
	MpID         string  `json:"mp_id"`
	CustomerName *string `json:"customer_name,omitempty"`
	MissingDate  string  `json:"missing_date"`
}

// ExportMpMissing 查找指定月份缺失计量点数据的记录（简化：列出该月无 raw_mp_data 的 mp_id）。
// raw_mp_data 无 org_id，通过 customer_id JOIN customers 过滤。
func (r *LoadDataRepository) ExportMpMissing(
	ctx context.Context, month string,
) ([]*MpMissingItem, error) {
	// 找出该月各 mp_id 的数据缺口：对比应有天数 vs 实际记录数
	q := `SELECT m.mp_id, c.user_name, m.missing_date FROM (
			SELECT mp_id, COALESCE(customer_id, uuid_nil()) AS cid,
				generate_series(d::date, (d::date + interval '1 month - 1 day')::date, '1 day')::text AS missing_date
			FROM (
				SELECT mp_id, customer_id, ($1 || '-01')::date AS d
				FROM raw_mp_data
				WHERE date::text LIKE $2
				GROUP BY mp_id, customer_id
			) src
		) m
		LEFT JOIN customers c ON c.id = m.cid
		WHERE NOT EXISTS (
			SELECT 1 FROM raw_mp_data r
			WHERE r.mp_id = m.mp_id AND r.date::text = m.missing_date
		)`
	args := []any{month, month + "-%"}
	if org, scoped := OrgFilter(ctx); scoped {
		// raw_mp_data 无 org_id，只能过滤有 customer 关联且 org 匹配的
		args = append(args, org)
		q += fmt.Sprintf(" AND (c.id IS NULL OR c.org_id = $%d::uuid)", len(args))
	}
	q += " LIMIT 1000"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*MpMissingItem
	for rows.Next() {
		var it MpMissingItem
		if err := rows.Scan(&it.MpID, &it.CustomerName, &it.MissingDate); err != nil {
			return nil, err
		}
		out = append(out, &it)
	}
	return out, rows.Err()
}

// ─── WP6.4：负荷数据校准 ───

// LoadCalibration 校准台账行。
type LoadCalibration struct {
	ID          string     `json:"id"`
	CustomerID  string     `json:"customer_id"`
	CustomerName string    `json:"customer_name,omitempty"`
	PeriodMonth string     `json:"period_month"`
	MeterKWh    float64    `json:"meter_kwh"`
	SystemKWh   float64    `json:"system_kwh"`
	Coefficient float64    `json:"coefficient"`
	Status      string     `json:"status"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
	Note        *string    `json:"note,omitempty"`
}

// ComputeCalibration 计算某客户某月的表计侧 vs 系统侧电量与校准系数（不落库）。
// 表计侧 = raw_meter_data 聚合（Σ曲线×倍率/4）；系统侧 = user_load_data.total_load 合计。
// 任一侧无数据返回错误（不臆造）。
func (r *LoadDataRepository) ComputeCalibration(ctx context.Context, customerID, month string) (*LoadCalibration, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	// 表计侧（JSONB period_data 逐点×倍率/4）
	var meterKWh *float64
	q := `SELECT SUM((elem->>'value')::float8 * m.multiplier) / 4.0
		FROM raw_meter_data m, jsonb_array_elements(m.period_data) AS elem
		WHERE m.customer_id = $1::uuid AND to_char(m.date,'YYYY-MM') = $2
		  AND m.org_id = $3::uuid`
	if err := r.pool.QueryRow(ctx, q, customerID, month, org).Scan(&meterKWh); err != nil {
		return nil, err
	}
	if meterKWh == nil {
		return nil, errors.New("该月无表计数据（先 POST /import/meter 导入 raw_meter_data）")
	}
	// 系统侧
	var sysKWh *float64
	q2 := `SELECT SUM(total_load) FROM user_load_data
		WHERE customer_id = $1::uuid AND to_char(date,'YYYY-MM') = $2
		  AND quality_flag <> 'missing' AND org_id = $3::uuid`
	if err := r.pool.QueryRow(ctx, q2, customerID, month, org).Scan(&sysKWh); err != nil {
		return nil, err
	}
	if sysKWh == nil {
		return nil, errors.New("该月无系统负荷数据（先导入表计并聚合，或检查导入日期范围）")
	}
	coef := 1.0
	if *sysKWh > 0 {
		coef = *meterKWh / *sysKWh
	}
	return &LoadCalibration{
		CustomerID: customerID, PeriodMonth: month,
		MeterKWh: mathRound(*meterKWh, 2), SystemKWh: mathRound(*sysKWh, 2),
		Coefficient: mathRound(coef, 6), Status: "preview",
	}, nil
}

// SaveCalibration 保存/更新校准行（同客户+月幂等；applied 终态拒绝覆盖）。
func (r *LoadDataRepository) SaveCalibration(ctx context.Context, cal *LoadCalibration, note string) (*LoadCalibration, error) {
	org, err := MustScoped(ctx)
	if err != nil {
		return nil, err
	}
	q := `INSERT INTO load_calibrations
		(org_id, customer_id, period_month, meter_kwh, system_kwh, coefficient, status, note)
		VALUES ($1::uuid,$2,$3,$4,$5,$6,'preview',$7)
		ON CONFLICT (org_id, customer_id, period_month) DO UPDATE SET
		  meter_kwh=EXCLUDED.meter_kwh, system_kwh=EXCLUDED.system_kwh,
		  coefficient=EXCLUDED.coefficient, note=EXCLUDED.note,
		  status=CASE WHEN load_calibrations.status='applied' THEN load_calibrations.status ELSE 'preview' END
		RETURNING id::text, customer_id::text, period_month, meter_kwh, system_kwh,
		  coefficient, status, applied_at, note`
	var out LoadCalibration
	err = r.pool.QueryRow(ctx, q, org, cal.CustomerID, cal.PeriodMonth,
		cal.MeterKWh, cal.SystemKWh, cal.Coefficient, nullStr(note)).Scan(
		&out.ID, &out.CustomerID, &out.PeriodMonth, &out.MeterKWh, &out.SystemKWh,
		&out.Coefficient, &out.Status, &out.AppliedAt, &out.Note)
	if err != nil {
		return nil, err
	}
	if out.Status == "applied" {
		return nil, errors.New("该月校准已应用（终态），如需重算请先作废")
	}
	return &out, nil
}

// ApplyCalibration 应用校准：按系数缩放该客户该月 user_load_data（curve×coef、total×coef）。
// 幂等：重复应用按「原始值×系数」语义需要先 void 恢复——本实现直接记录 applied_at
// 并在重复应用时返回错误。
func (r *LoadDataRepository) ApplyCalibration(ctx context.Context, id, actor string) error {
	q := `UPDATE load_calibrations SET status='applied', applied_at=now(), applied_by=$2::uuid
		WHERE id=$1::uuid AND status='preview'
		RETURNING customer_id::text, period_month, coefficient`
	var custID, month string
	var coef float64
	if err := r.pool.QueryRow(ctx, q, id, nullStr(actor)).Scan(&custID, &month, &coef); err != nil {
		return errors.New("校准单不存在或已应用")
	}
	// 缩放系统侧曲线
	tag, err := r.pool.Exec(ctx, `
		UPDATE user_load_data SET
		  curve_96 = (SELECT array_agg(e * $3::numeric ORDER BY ord)::numeric[] 
		              FROM unnest(curve_96) WITH ORDINALITY AS t(e, ord)),
		  total_load = total_load * $3::numeric,
		  updated_at = now()
		WHERE customer_id = $1::uuid AND to_char(date,'YYYY-MM') = $2`,
		custID, month, coef)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("该月无系统负荷数据可校准")
	}
	return nil
}

// VoidCalibration 作废校准（preview/applied → void；不回滚已缩放的数据，
// 提示用户重导入表计并重新聚合恢复）。
func (r *LoadDataRepository) VoidCalibration(ctx context.Context, id string) error {
	q := `UPDATE load_calibrations SET status='void' WHERE id=$1::uuid AND status IN ('preview','applied')`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("校准单不存在或已作废")
	}
	return nil
}

// ListCalibrations 校准台账（月份过滤，org 过滤）。
func (r *LoadDataRepository) ListCalibrations(ctx context.Context, month string, limit int) ([]*LoadCalibration, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := `SELECT lc.id::text, lc.customer_id::text, c.user_name, lc.period_month,
		lc.meter_kwh, lc.system_kwh, lc.coefficient, lc.status, lc.applied_at, lc.note
	  FROM load_calibrations lc JOIN customers c ON c.id = lc.customer_id`
	where := []string{"TRUE"}
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		where = append(where, fmt.Sprintf("lc.org_id = $%d::uuid", len(args)))
	}
	if month != "" {
		args = append(args, month)
		where = append(where, fmt.Sprintf("lc.period_month = $%d", len(args)))
	}
	args = append(args, limit)
	q += " WHERE " + strings.Join(where, " AND ") + " ORDER BY lc.period_month DESC, c.user_name LIMIT $" + fmt.Sprint(len(args))
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*LoadCalibration, 0, 16)
	for rows.Next() {
		var lc LoadCalibration
		if err := rows.Scan(&lc.ID, &lc.CustomerID, &lc.CustomerName, &lc.PeriodMonth,
			&lc.MeterKWh, &lc.SystemKWh, &lc.Coefficient, &lc.Status, &lc.AppliedAt, &lc.Note); err != nil {
			return nil, err
		}
		list = append(list, &lc)
	}
	return list, rows.Err()
}

func mathRound(v float64, digits int) float64 {
	p := math.Pow(10, float64(digits))
	return math.Round(v*p) / p
}
