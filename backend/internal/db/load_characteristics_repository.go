// 负荷特性分析。
// 2026-06 自 new_modules2_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── 负荷特性分析 ───────────────

type LoadCharacteristic struct {
	ID              string    `json:"id"`
	CustomerID      string    `json:"customer_id"`
	CustomerName    string    `json:"customer_name,omitempty"`
	AnalysisMonth   string    `json:"analysis_month"`
	AvgDailyMWh     float64   `json:"avg_daily_mwh"`
	PeakMW          float64   `json:"peak_mw"`
	ValleyMW        float64   `json:"valley_mw"`
	PeakValleyRatio float64   `json:"peak_valley_ratio"`
	LoadFactor      float64   `json:"load_factor"`
	PeakHours       float64   `json:"peak_hours"`
	LoadType        string    `json:"load_type"`
	CreatedAt       time.Time `json:"created_at"`
}

type LoadCharacteristicsRepository struct{ pool *Pool }

func NewLoadCharacteristicsRepository(pool *Pool) *LoadCharacteristicsRepository {
	return &LoadCharacteristicsRepository{pool: pool}
}

func (r *LoadCharacteristicsRepository) List(ctx context.Context, month string, limit int) ([]*LoadCharacteristic, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	org, scoped := OrgFilter(ctx)
	args := []any{}
	q := `SELECT lc.id, lc.customer_id::text, c.user_name, lc.analysis_month,
		      lc.avg_daily_mwh, lc.peak_mw, lc.valley_mw, lc.peak_valley_ratio,
		      lc.load_factor, lc.peak_hours, lc.load_type, lc.created_at
	      FROM load_characteristics lc
	      JOIN customers c ON c.id = lc.customer_id`
	where := ""
	if month != "" {
		args = append(args, month)
		where = fmt.Sprintf(" WHERE lc.analysis_month = $%d", len(args))
	}
	if scoped {
		if where == "" {
			where = " WHERE"
		} else {
			where += " AND"
		}
		args = append(args, org)
		where += fmt.Sprintf(" lc.org_id = $%d::uuid", len(args))
	}
	q += where + " ORDER BY lc.analysis_month DESC, lc.peak_mw DESC LIMIT " + itoaNew(limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*LoadCharacteristic, 0)
	for rows.Next() {
		var l LoadCharacteristic
		if err := rows.Scan(&l.ID, &l.CustomerID, &l.CustomerName, &l.AnalysisMonth,
			&l.AvgDailyMWh, &l.PeakMW, &l.ValleyMW, &l.PeakValleyRatio,
			&l.LoadFactor, &l.PeakHours, &l.LoadType, &l.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &l)
	}
	return list, rows.Err()
}

func (r *LoadCharacteristicsRepository) GenerateDemo(ctx context.Context) (int, error) {
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	args := []any{orgID}
	rows, err := r.pool.Query(ctx, `SELECT id FROM customers WHERE org_id = $1::uuid LIMIT 30`, args...)
	if err != nil {
		return 0, err
	}
	customers := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		customers = append(customers, id)
	}
	rows.Close()
	loadTypes := []string{"industrial", "commercial", "mixed", "residential"}
	cnt := 0
	for _, cid := range customers {
		for i := 0; i < 3; i++ {
			ym := time.Now().AddDate(0, -i, 0).Format("2006-01")
			avgDaily := 50 + rand.Float64()*200
			peak := avgDaily * (1.3 + rand.Float64()*0.5)
			valley := avgDaily * (0.3 + rand.Float64()*0.3)
			ratio := peak / valley
			factor := avgDaily / peak * 100
			peakHours := 6 + rand.Float64()*6
			ltype := loadTypes[rand.Intn(len(loadTypes))]
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO load_characteristics
				   (customer_id, analysis_month, avg_daily_mwh, peak_mw, valley_mw,
				    peak_valley_ratio, load_factor, peak_hours, load_type, org_id)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::uuid)
				 ON CONFLICT (org_id, customer_id, analysis_month) DO UPDATE SET
				   avg_daily_mwh = EXCLUDED.avg_daily_mwh, peak_mw = EXCLUDED.peak_mw,
				   valley_mw = EXCLUDED.valley_mw, peak_valley_ratio = EXCLUDED.peak_valley_ratio,
				   load_factor = EXCLUDED.load_factor, peak_hours = EXCLUDED.peak_hours,
				   load_type = EXCLUDED.load_type`,
				cid, ym, avgDaily, peak, valley, ratio, factor, peakHours, ltype, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
