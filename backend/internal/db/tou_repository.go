// TOU 时段规则仓储。
// 2026-06 自 v1clone_e_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ─────────────── E4 TOU 时段规则 ───────────────

type TOURule struct {
	ID            string          `json:"id"`
	RuleName      string          `json:"rule_name"`
	EffectiveFrom time.Time       `json:"effective_from"`
	EffectiveTo   *time.Time      `json:"effective_to,omitempty"`
	Periods       json.RawMessage `json:"periods"`
	CreatedAt     time.Time       `json:"created_at"`
}

type TOURepository struct{ pool *Pool }

func NewTOURepository(pool *Pool) *TOURepository { return &TOURepository{pool: pool} }

func (r *TOURepository) List(ctx context.Context) ([]*TOURule, error) {
	q := `SELECT id, rule_name, effective_from, effective_to, periods, created_at
	 FROM tou_rules`
	args := []any{}
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY effective_from DESC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*TOURule, 0)
	for rows.Next() {
		var t TOURule
		if err := rows.Scan(&t.ID, &t.RuleName, &t.EffectiveFrom, &t.EffectiveTo,
			&t.Periods, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

func (r *TOURepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	// 构造 96 段标签：valley(0-8h) / shoulder(8-10, 12-14, 18-22) / peak(10-12, 14-18) / sharp(无)
	tags := make([]string, 96)
	for i := 0; i < 96; i++ {
		hour := i / 4
		switch {
		case hour < 8:
			tags[i] = "valley"
		case hour < 10:
			tags[i] = "shoulder"
		case hour < 12:
			tags[i] = "peak"
		case hour < 14:
			tags[i] = "shoulder"
		case hour < 18:
			tags[i] = "peak"
		case hour < 22:
			tags[i] = "shoulder"
		default:
			tags[i] = "valley"
		}
	}
	periods, _ := json.Marshal(map[string]any{"tags": tags})
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tou_rules (rule_name, effective_from, effective_to, periods, org_id)
		 VALUES ('广东工商业-2026', $1::date, $2::date, $3::jsonb, $4::uuid)
		 ON CONFLICT DO NOTHING`,
		"2026-01-01", "2026-12-31", periods, orgID)
	if err != nil {
		return 0, err
	}
	return 1, nil
}
