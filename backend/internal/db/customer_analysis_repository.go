// 客户分析。
// 2026-06 自 new_modules2_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── 客户分析 ───────────────

type CustomerAnalysisItem struct {
	ID             string          `json:"id"`
	CustomerID     string          `json:"customer_id"`
	CustomerName   string          `json:"customer_name,omitempty"`
	AnalysisMonth  string          `json:"analysis_month"`
	EnergyConsumed float64         `json:"energy_consumed_mwh"`
	BillAmount     float64         `json:"bill_amount"`
	AvgUnitPrice   float64         `json:"avg_unit_price"`
	PeakRatio      float64         `json:"peak_ratio"`
	Score          float64         `json:"score"`
	RiskLevel      string          `json:"risk_level"`
	Tags           []string        `json:"tags"`
	Extra          json.RawMessage `json:"extra"`
	CreatedAt      time.Time       `json:"created_at"`
}

type CustomerAnalysisRepository struct{ pool *Pool }

func NewCustomerAnalysisRepository(pool *Pool) *CustomerAnalysisRepository {
	return &CustomerAnalysisRepository{pool: pool}
}

func (r *CustomerAnalysisRepository) List(ctx context.Context, month string, limit int) ([]*CustomerAnalysisItem, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	org, scoped := OrgFilter(ctx)
	args := []any{}
	q := `SELECT ca.id, ca.customer_id::text, c.user_name, ca.analysis_month,
		      ca.energy_consumed_mwh, ca.bill_amount, ca.avg_unit_price,
		      ca.peak_ratio, ca.score, ca.risk_level, ca.tags, ca.extra, ca.created_at
	      FROM customer_analysis ca
	      JOIN customers c ON c.id = ca.customer_id`
	where := ""
	if month != "" {
		args = append(args, month)
		where = fmt.Sprintf(" WHERE ca.analysis_month = $%d", len(args))
	}
	if scoped {
		if where == "" {
			where = " WHERE"
		} else {
			where += " AND"
		}
		args = append(args, org)
		where += fmt.Sprintf(" ca.org_id = $%d::uuid", len(args))
	}
	q += where + " ORDER BY ca.analysis_month DESC, ca.score DESC LIMIT " + itoaNew(limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*CustomerAnalysisItem, 0)
	for rows.Next() {
		var a CustomerAnalysisItem
		if err := rows.Scan(&a.ID, &a.CustomerID, &a.CustomerName, &a.AnalysisMonth,
			&a.EnergyConsumed, &a.BillAmount, &a.AvgUnitPrice,
			&a.PeakRatio, &a.Score, &a.RiskLevel, &a.Tags, &a.Extra,
			&a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

func (r *CustomerAnalysisRepository) GenerateDemo(ctx context.Context) (int, error) {
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	rows, err := r.pool.Query(ctx, `SELECT id FROM customers WHERE org_id = $1::uuid LIMIT 30`, orgID)
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
	risks := []string{"low", "low", "medium", "medium", "high"}
	cnt := 0
	for _, cid := range customers {
		for i := 0; i < 3; i++ {
			ym := time.Now().AddDate(0, -i, 0).Format("2006-01")
			energy := 3000 + rand.Float64()*20000
			price := 400 + rand.Float64()*60
			bill := energy * price
			peakRatio := 0.3 + rand.Float64()*0.3
			score := 60 + rand.Float64()*40
			risk := risks[rand.Intn(len(risks))]
			tags := []string{"active"}
			if score > 85 {
				tags = append(tags, "high_value")
			}
			if risk == "high" {
				tags = append(tags, "attention")
			}
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO customer_analysis
				   (customer_id, analysis_month, energy_consumed_mwh, bill_amount,
				    avg_unit_price, peak_ratio, score, risk_level, tags, org_id)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::uuid)
				 ON CONFLICT (org_id, customer_id, analysis_month) DO UPDATE SET
				   energy_consumed_mwh = EXCLUDED.energy_consumed_mwh,
				   bill_amount = EXCLUDED.bill_amount,
				   avg_unit_price = EXCLUDED.avg_unit_price,
				   peak_ratio = EXCLUDED.peak_ratio, score = EXCLUDED.score,
				   risk_level = EXCLUDED.risk_level, tags = EXCLUDED.tags`,
				cid, ym, energy, bill, price, peakRatio, score, risk, tags, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
