// 意向客户仓储。
// 2026-06 自 v1clone_e_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// ─────────────── E1 意向客户 ───────────────

type IntentCustomer struct {
	ID            string          `json:"id"`
	CustomerName  string          `json:"customer_name"`
	Meters        json.RawMessage `json:"meters"`
	CoverageStart *time.Time      `json:"coverage_start,omitempty"`
	CoverageEnd   *time.Time      `json:"coverage_end,omitempty"`
	CoverageDays  *int            `json:"coverage_days,omitempty"`
	Completeness  *float64        `json:"completeness,omitempty"`
	AvgDailyLoad  *float64        `json:"avg_daily_load,omitempty"`
	Status        string          `json:"status"`
	Extra         json.RawMessage `json:"extra"`
	CreatedAt     time.Time       `json:"created_at"`
}

type IntentCustomerDiagnosis struct {
	IntentCustomer
	DataScore      float64 `json:"data_score"` // 0-100
	LoadScore      float64 `json:"load_score"`
	CoverageScore  float64 `json:"coverage_score"`
	OverallScore   float64 `json:"overall_score"`
	MatchedPackage *string `json:"matched_package,omitempty"`
	Recommendation string  `json:"recommendation"`
}

type IntentCustomerRepository struct{ pool *Pool }

func NewIntentCustomerRepository(pool *Pool) *IntentCustomerRepository {
	return &IntentCustomerRepository{pool: pool}
}

// CreateBasic 仅凭名称新建意向客户（文档「确认入库」用），状态 pending。
// 写操作要求具体活跃省。
func (r *IntentCustomerRepository) CreateBasic(ctx context.Context, name string) error {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return ErrOrgRequired
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO intent_customers (customer_name, meters, status, org_id)
		VALUES ($1, '[]'::jsonb, 'pending', $2::uuid)
		ON CONFLICT DO NOTHING`, name, org)
	return err
}

func (r *IntentCustomerRepository) List(ctx context.Context) ([]*IntentCustomer, error) {
	q := `SELECT id, customer_name, meters, coverage_start, coverage_end,
			coverage_days, completeness, avg_daily_load, status, extra, created_at
		 FROM intent_customers`
	args := []any{}
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY created_at DESC LIMIT 200"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*IntentCustomer, 0)
	for rows.Next() {
		var c IntentCustomer
		if err := rows.Scan(&c.ID, &c.CustomerName, &c.Meters, &c.CoverageStart,
			&c.CoverageEnd, &c.CoverageDays, &c.Completeness, &c.AvgDailyLoad,
			&c.Status, &c.Extra, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}

// Diagnose 计算每个意向客户的评分（基于已有字段做加权）。
func (r *IntentCustomerRepository) Diagnose(ctx context.Context) ([]*IntentCustomerDiagnosis, error) {
	list, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*IntentCustomerDiagnosis, 0, len(list))
	for _, c := range list {
		d := IntentCustomerDiagnosis{IntentCustomer: *c}
		// 数据完整度评分
		if c.Completeness != nil {
			d.DataScore = math.Min(100, *c.Completeness)
		}
		// 覆盖时长评分：>=90 天满分
		if c.CoverageDays != nil {
			days := float64(*c.CoverageDays)
			d.CoverageScore = math.Min(100, days/90*100)
		}
		// 负荷规模评分：>5000 kW 满分（按 MW 折算）
		if c.AvgDailyLoad != nil {
			d.LoadScore = math.Min(100, *c.AvgDailyLoad/5000*100)
		}
		d.OverallScore = d.DataScore*0.3 + d.CoverageScore*0.3 + d.LoadScore*0.4
		switch {
		case d.OverallScore >= 80:
			p := "标准工商业月度套餐"
			d.MatchedPackage = &p
			d.Recommendation = "推荐转化为正式客户"
		case d.OverallScore >= 60:
			p := "短期试用套餐"
			d.MatchedPackage = &p
			d.Recommendation = "可签短期试用合同"
		default:
			d.Recommendation = "建议补充负荷数据"
		}
		out = append(out, &d)
	}
	return out, nil
}

func (r *IntentCustomerRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用 FJ
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='FJ'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve FJ org: %w", err)
		}
	}
	cnt := 0
	names := []string{"佛山陶瓷工业园", "东莞机械厂", "广州物流园", "中山纺织集团", "珠海电子科技园"}
	for _, n := range names {
		now := time.Now()
		days := 60 + rand.Intn(90)
		start := now.AddDate(0, 0, -days)
		completeness := 75 + rand.Float64()*25
		avgLoad := 1500 + rand.Float64()*5000
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO intent_customers
			   (customer_name, meters, coverage_start, coverage_end, coverage_days,
			    completeness, avg_daily_load, status, org_id)
			 VALUES ($1, '[]'::jsonb, $2, $3, $4, $5, $6, 'pending', $7::uuid)
			 ON CONFLICT DO NOTHING`,
			n, start, now, days, completeness, avgLoad, orgID); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
