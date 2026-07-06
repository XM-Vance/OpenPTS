// TOU 时段规则仓储。
// 2026-06 自 v1clone_e_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ─────────────── E4 TOU 时段规则 ───────────────

type TOURule struct {
	ID            string          `json:"id"`
	OrgID         string          `json:"org_id,omitempty"`
	RuleName      string          `json:"rule_name"`
	EffectiveFrom time.Time       `json:"effective_from"`
	EffectiveTo   *time.Time      `json:"effective_to,omitempty"`
	Periods       json.RawMessage `json:"periods"`
	CreatedAt     time.Time       `json:"created_at"`
}

var ErrTOUNotFound = errors.New("TOU 规则不存在")

// TOUInput 创建/更新入参。Periods 为 96 段标签等 JSON（透传 jsonb）。
type TOUInput struct {
	RuleName      string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	Periods       json.RawMessage
}

type TOURepository struct{ pool *Pool }

func NewTOURepository(pool *Pool) *TOURepository { return &TOURepository{pool: pool} }

const touColumns = "id, org_id, rule_name, effective_from, effective_to, periods, created_at"

func (r *TOURepository) scan(row pgx.Row) (*TOURule, error) {
	var t TOURule
	err := row.Scan(&t.ID, &t.OrgID, &t.RuleName, &t.EffectiveFrom, &t.EffectiveTo,
		&t.Periods, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTOUNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TOURepository) List(ctx context.Context) ([]*TOURule, error) {
	q := `SELECT ` + touColumns + ` FROM tou_rules`
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
		t, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *TOURepository) Create(ctx context.Context, in TOUInput) (*TOURule, error) {
	org, err := MustScoped(ctx)
	if err != nil {
		return nil, err
	}
	q := `INSERT INTO tou_rules
		(org_id, rule_name, effective_from, effective_to, periods)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING ` + touColumns
	return r.scan(r.pool.QueryRow(ctx, q,
		org, in.RuleName, in.EffectiveFrom, in.EffectiveTo, in.Periods))
}

func (r *TOURepository) Update(ctx context.Context, id string, in TOUInput) (*TOURule, error) {
	q := `UPDATE tou_rules SET
		rule_name = $2, effective_from = $3, effective_to = $4, periods = $5
		WHERE id = $1::uuid`
	args := []any{id, in.RuleName, in.EffectiveFrom, in.EffectiveTo, in.Periods}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += ` RETURNING ` + touColumns
	return r.scan(r.pool.QueryRow(ctx, q, args...))
}

func (r *TOURepository) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM tou_rules WHERE id = $1::uuid`
	args := []any{id}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTOUNotFound
	}
	return nil
}

func (r *TOURepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃组织，否则用默认组织
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
