// 意向客户仓储。
// 2026-06 自 v1clone_e_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
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

// CreateBasic 仅凭名称新建意向客户（文档「确认入库」用）。
// Phase 1b 起：意向客户即统一 customers 表中 lifecycle_stage='intent' 的行；写操作要求具体活跃组织。
func (r *IntentCustomerRepository) CreateBasic(ctx context.Context, name string) error {
	org, err := MustScoped(ctx)
	if err != nil {
		return err
	}
	// 同省同名意向去重（原 intent_customers 无唯一约束，此处更稳）。
	_, err = r.pool.Exec(ctx, `
		INSERT INTO customers (user_name, source, lifecycle_stage, org_id, extra)
		SELECT $1, '意向客户', 'intent', $2::uuid, jsonb_build_object('_phase1_intent', true)
		WHERE NOT EXISTS (
			SELECT 1 FROM customers WHERE user_name = $1 AND lifecycle_stage = 'intent' AND org_id = $2::uuid)`,
		name, org)
	return err
}

// List 返回意向客户（统一 customers 表 stage='intent' + 最新一条 customer_diagnosis）。
func (r *IntentCustomerRepository) List(ctx context.Context) ([]*IntentCustomer, error) {
	q := `SELECT c.id::text, c.user_name, COALESCE(d.meters, '[]'::jsonb),
			d.coverage_start, d.coverage_end, d.coverage_days, d.completeness, d.avg_daily_load,
			c.extra, c.created_at
		 FROM customers c
		 LEFT JOIN LATERAL (
			SELECT meters, coverage_start, coverage_end, coverage_days, completeness, avg_daily_load
			FROM customer_diagnosis cd WHERE cd.customer_id = c.id
			ORDER BY diagnosed_at DESC LIMIT 1
		 ) d ON true
		 WHERE c.lifecycle_stage = 'intent'`
	args := []any{}
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND c.org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY c.created_at DESC LIMIT 200"
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
			&c.Extra, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Status = "pending" // stage='intent' 统一映射为 pending
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
	// 确定 org_id：scoped 用活跃组织，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	cnt := 0
	names := []string{"佛山陶瓷工业园", "东莞机械厂", "广州物流园", "中山纺织集团", "珠海电子科技园"}
	for _, n := range names {
		now := time.Now()
		days := 60 + rand.IntN(90)
		start := now.AddDate(0, 0, -days)
		completeness := 75 + rand.Float64()*25
		avgLoad := 1500 + rand.Float64()*5000
		// 意向客户即 customers(stage='intent')；同省同名去重；新建则带回诊断快照。
		var cid string
		err := r.pool.QueryRow(ctx, `
			INSERT INTO customers (user_name, source, lifecycle_stage, org_id, extra)
			SELECT $1, '意向客户', 'intent', $2::uuid, jsonb_build_object('_phase1_intent', true, '_demo', true)
			WHERE NOT EXISTS (
				SELECT 1 FROM customers WHERE user_name = $1 AND lifecycle_stage = 'intent' AND org_id = $2::uuid)
			RETURNING id`, n, orgID).Scan(&cid)
		if errors.Is(err, pgx.ErrNoRows) {
			continue // 已存在，跳过
		}
		if err != nil {
			return cnt, err
		}
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO customer_diagnosis
			   (customer_id, org_id, coverage_start, coverage_end, coverage_days, completeness, avg_daily_load)
			 VALUES ($1, $2::uuid, $3, $4, $5, $6, $7)`,
			cid, orgID, start, now, days, completeness, avgLoad); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
