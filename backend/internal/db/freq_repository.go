// 调频仓储：日级汇总（透视 AGC/AVC 清算 + 关联需求/补偿）。
package db

import (
	"context"
	"fmt"
	"time"
)

type FreqDailySummary struct {
	Date         time.Time `json:"date"`
	AGCVolume    *float64  `json:"agc_volume,omitempty"`
	AGCPrice     *float64  `json:"agc_price,omitempty"`
	AGCRevenue   *float64  `json:"agc_revenue,omitempty"`
	AVCVolume    *float64  `json:"avc_volume,omitempty"`
	AVCPrice     *float64  `json:"avc_price,omitempty"`
	AVCRevenue   *float64  `json:"avc_revenue,omitempty"`
	DemandVolume *float64  `json:"demand_volume,omitempty"`
	DemandPrice  *float64  `json:"demand_price,omitempty"`
	CompFee      *float64  `json:"comp_fee,omitempty"`
}

type FreqRepository struct {
	pool *Pool
}

func NewFreqRepository(pool *Pool) *FreqRepository {
	return &FreqRepository{pool: pool}
}

// ListDailySummary 透视 clearing 按 type 拆为列，左连接 demand，子查询 comp_fee 合计。
func (r *FreqRepository) ListDailySummary(ctx context.Context, limit int) ([]*FreqDailySummary, error) {
	org, scoped := OrgFilter(ctx)
	args := []any{}
	idx := 0
	orgWhere := ""
	if scoped {
		args = append(args, org)
		idx++
		orgWhere = fmt.Sprintf("AND org_id = $%d::uuid", idx)
	}
	idx++
	limitIdx := idx
	args = append(args, limit)

	q := fmt.Sprintf(`
		WITH clearing AS (
			SELECT settlement_date,
				MAX(CASE WHEN regulation_type='AGC' THEN cleared_volume END) AS agc_volume,
				MAX(CASE WHEN regulation_type='AGC' THEN cleared_price END)  AS agc_price,
				MAX(CASE WHEN regulation_type='AGC' THEN revenue END)        AS agc_revenue,
				MAX(CASE WHEN regulation_type='AVC' THEN cleared_volume END) AS avc_volume,
				MAX(CASE WHEN regulation_type='AVC' THEN cleared_price END)  AS avc_price,
				MAX(CASE WHEN regulation_type='AVC' THEN revenue END)        AS avc_revenue
			FROM frequency_regulation_clearing
			WHERE 1=1 %s
			GROUP BY settlement_date
		)
		SELECT c.settlement_date,
			c.agc_volume, c.agc_price, c.agc_revenue,
			c.avc_volume, c.avc_price, c.avc_revenue,
			d.demand_volume, d.demand_price,
			(SELECT SUM(compensation_fee)
			   FROM freq_comp_fee f
			  WHERE f.settlement_date = c.settlement_date %s) AS comp_fee
		FROM clearing c
		LEFT JOIN frequency_regulation_demand d
		  ON d.settlement_date = c.settlement_date %s
		ORDER BY c.settlement_date DESC
		LIMIT $%d`, orgWhere, orgWhere, orgWhere, limitIdx)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*FreqDailySummary, 0, limit)
	for rows.Next() {
		var s FreqDailySummary
		if err := rows.Scan(&s.Date,
			&s.AGCVolume, &s.AGCPrice, &s.AGCRevenue,
			&s.AVCVolume, &s.AVCPrice, &s.AVCRevenue,
			&s.DemandVolume, &s.DemandPrice, &s.CompFee); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

// resolveOrgID 返回活跃组织 org_id；scoped=false 时回退到默认组织。
func (r *FreqRepository) resolveOrgID(ctx context.Context) (string, error) {
	org, scoped := OrgFilter(ctx)
	if scoped {
		return org, nil
	}
	var id string
	if err := r.pool.QueryRow(ctx,
		"SELECT id FROM organizations WHERE code='default'").Scan(&id); err != nil {
		return "", fmt.Errorf("resolve default org: %w", err)
	}
	return id, nil
}

// UpsertClearing 写入/覆盖某日某调频类型的清算数据。
func (r *FreqRepository) UpsertClearing(
	ctx context.Context, d time.Time, regType string, volume, price, revenue float64,
) error {
	orgID, err := r.resolveOrgID(ctx)
	if err != nil {
		return err
	}
	const q = `
		INSERT INTO frequency_regulation_clearing
			(settlement_date, regulation_type, cleared_volume, cleared_price, revenue, org_id)
		VALUES ($1, $2, $3, $4, $5, $6::uuid)
		ON CONFLICT (org_id, settlement_date, regulation_type) DO UPDATE SET
			cleared_volume = EXCLUDED.cleared_volume,
			cleared_price  = EXCLUDED.cleared_price,
			revenue        = EXCLUDED.revenue`
	_, err = r.pool.Exec(ctx, q, d, regType, volume, price, revenue, orgID)
	return err
}

// UpsertDemand 写入/覆盖某日调频需求。
func (r *FreqRepository) UpsertDemand(
	ctx context.Context, d time.Time, volume, price float64, source string,
) error {
	orgID, err := r.resolveOrgID(ctx)
	if err != nil {
		return err
	}
	const q = `
		INSERT INTO frequency_regulation_demand
			(settlement_date, demand_volume, demand_price, source, org_id)
		VALUES ($1, $2, $3, $4, $5::uuid)
		ON CONFLICT (org_id, settlement_date) DO UPDATE SET
			demand_volume = EXCLUDED.demand_volume,
			demand_price  = EXCLUDED.demand_price,
			source        = EXCLUDED.source`
	_, err = r.pool.Exec(ctx, q, d, volume, price, source, orgID)
	return err
}

// UpsertCompFee 写入/覆盖某日某频次补偿费。
func (r *FreqRepository) UpsertCompFee(
	ctx context.Context, d time.Time, freqType string, fee float64,
) error {
	orgID, err := r.resolveOrgID(ctx)
	if err != nil {
		return err
	}
	const q = `
		INSERT INTO freq_comp_fee
			(settlement_date, frequency_type, compensation_fee, org_id)
		VALUES ($1, $2, $3, $4::uuid)
		ON CONFLICT (org_id, settlement_date, frequency_type) DO UPDATE SET
			compensation_fee = EXCLUDED.compensation_fee`
	_, err = r.pool.Exec(ctx, q, d, freqType, fee, orgID)
	return err
}
