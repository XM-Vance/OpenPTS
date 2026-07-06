// 客户负荷分析仓储。
// 2026-06 自 v1clone_e_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"time"
)

// ─────────────── E2 客户负荷分析 ───────────────

type CustomerLoadSummary struct {
	CustomerID      string  `json:"customer_id"`
	CustomerName    string  `json:"customer_name"`
	Days            int     `json:"days"`
	AvgDaily        float64 `json:"avg_daily"`   // MWh
	PeakLoad        float64 `json:"peak_load"`   // MW
	ValleyLoad      float64 `json:"valley_load"` // MW
	PeakValleyRatio float64 `json:"peak_valley_ratio"`
	CV              float64 `json:"cv"`        // 变异系数
	IsIntent        bool    `json:"is_intent"` // Phase 3b：true=意向客户（数据来自诊断聚合，无曲线）
}

type CustomerLoadCurve struct {
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Date         time.Time `json:"date"`
	Curve96      []float64 `json:"curve_96"`
}

type CustomerLoadRepository struct{ pool *Pool }

func NewCustomerLoadRepository(pool *Pool) *CustomerLoadRepository {
	return &CustomerLoadRepository{pool: pool}
}

// Summary 最近 N 天每个客户的统计汇总。
func (r *CustomerLoadRepository) Summary(ctx context.Context, days int) ([]*CustomerLoadSummary, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	// 按活跃组织隔离（总部「全部省」不过滤）：曲线分支按 u.org_id、诊断兜底分支按 cd.org_id。
	args := []any{since}
	orgB1, orgB2 := "", ""
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		orgB1 = " AND u.org_id = $2::uuid"
		orgB2 = " AND cd.org_id = $2::uuid"
	}
	// Phase 3b：服务客户取 user_load_data 曲线统计；意向客户无曲线，用最新一条
	// customer_diagnosis 的聚合(avg_daily_load/coverage_days)兜底进汇总（峰谷留空）。
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT * FROM (
			SELECT u.customer_id::text AS cid, c.user_name AS name, COUNT(DISTINCT u.date) AS days,
			       AVG(u.total_load)::float8 AS avg_d, false AS is_intent,
			       (SELECT MAX(v) FROM user_load_data ul2,
			           LATERAL unnest(ul2.curve_96) AS v
			         WHERE ul2.customer_id = u.customer_id AND ul2.date >= $1::date)::float8 AS peak,
			       (SELECT MIN(v) FROM user_load_data ul3,
			           LATERAL unnest(ul3.curve_96) AS v
			         WHERE ul3.customer_id = u.customer_id AND ul3.date >= $1::date
			           AND v > 0)::float8 AS valley
			FROM user_load_data u
			JOIN customers c ON c.id = u.customer_id
			WHERE u.date >= $1::date%s
			GROUP BY u.customer_id, c.user_name

			UNION ALL

			(SELECT DISTINCT ON (cd.customer_id) cd.customer_id::text AS cid, c.user_name AS name,
			        COALESCE(cd.coverage_days, 0) AS days, cd.avg_daily_load::float8 AS avg_d, true AS is_intent,
			        NULL::float8 AS peak, NULL::float8 AS valley
			 FROM customer_diagnosis cd
			 JOIN customers c ON c.id = cd.customer_id
			 WHERE c.lifecycle_stage = 'intent' AND cd.avg_daily_load IS NOT NULL%s
			   AND NOT EXISTS (SELECT 1 FROM user_load_data u2
			                   WHERE u2.customer_id = cd.customer_id AND u2.date >= $1::date)
			 ORDER BY cd.customer_id, cd.diagnosed_at DESC)
		) s
		ORDER BY avg_d DESC NULLS LAST
		LIMIT 100`, orgB1, orgB2), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*CustomerLoadSummary, 0)
	for rows.Next() {
		var s CustomerLoadSummary
		var avg, peak, valley *float64
		if err := rows.Scan(&s.CustomerID, &s.CustomerName, &s.Days, &avg, &s.IsIntent, &peak, &valley); err != nil {
			return nil, err
		}
		if avg != nil {
			s.AvgDaily = *avg
		}
		if peak != nil {
			s.PeakLoad = *peak
		}
		if valley != nil {
			s.ValleyLoad = *valley
		}
		if s.ValleyLoad > 0 {
			s.PeakValleyRatio = s.PeakLoad / s.ValleyLoad
		}
		// CV 简化为 (peak-valley)/(peak+valley)
		if s.PeakLoad+s.ValleyLoad > 0 {
			s.CV = (s.PeakLoad - s.ValleyLoad) / (s.PeakLoad + s.ValleyLoad)
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

// LatestCurve 指定客户最近一天的 96 点曲线。
func (r *CustomerLoadRepository) LatestCurve(ctx context.Context, customerID string) (*CustomerLoadCurve, error) {
	var c CustomerLoadCurve
	q := `
		SELECT u.customer_id::text, cu.user_name, u.date, u.curve_96
		FROM user_load_data u JOIN customers cu ON cu.id = u.customer_id
		WHERE u.customer_id = $1`
	args := []any{customerID}
	if org, scoped := OrgFilter(ctx); scoped { // 防按 id 读他省客户曲线
		args = append(args, org)
		q += fmt.Sprintf(" AND u.org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY u.date DESC LIMIT 1"
	err := r.pool.QueryRow(ctx, q, args...).
		Scan(&c.CustomerID, &c.CustomerName, &c.Date, &c.Curve96)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
