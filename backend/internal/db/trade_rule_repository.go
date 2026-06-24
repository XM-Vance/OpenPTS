// 分时交易规则仓储：trade_rules CRUD + Export。
package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrTradeRuleNotFound = errors.New("交易规则不存在")

// TradeRule 分时交易规则（对应 trade_rules 表）。
type TradeRule struct {
	ID            uuid.UUID  `json:"id"`
	OrgID         *uuid.UUID `json:"org_id,omitempty"`
	RuleKey       string     `json:"rule_key"`
	RuleValue     string     `json:"rule_value"`
	RuleCategory  string     `json:"rule_category"`
	Description   *string    `json:"description,omitempty"`
	EffectiveDate *time.Time `json:"effective_date,omitempty"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty"`
	SourceDocID   *uuid.UUID `json:"source_doc_id,omitempty"`
	IsActive      bool       `json:"is_active"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TradeRuleInput 交易规则写入参数。
type TradeRuleInput struct {
	RuleKey       string     `json:"rule_key"`
	RuleValue     string     `json:"rule_value"`
	RuleCategory  string     `json:"rule_category"`
	Description   string     `json:"description"`
	EffectiveDate *time.Time `json:"effective_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	SourceDocID   string     `json:"source_doc_id"`
	IsActive      bool       `json:"is_active"`
}

// TradeRuleRepository 交易规则仓储。
type TradeRuleRepository struct{ pool *Pool }

// NewTradeRuleRepository 创建交易规则仓储。
func NewTradeRuleRepository(pool *Pool) *TradeRuleRepository {
	return &TradeRuleRepository{pool: pool}
}

const tradeRuleColumns = `id, org_id, rule_key, rule_value, rule_category, description,
	effective_date, expiry_date, source_doc_id, is_active,
	created_by, created_at, updated_at`

func scanTradeRule(row pgx.Row) (*TradeRule, error) {
	var r TradeRule
	err := row.Scan(&r.ID, &r.OrgID, &r.RuleKey, &r.RuleValue, &r.RuleCategory,
		&r.Description, &r.EffectiveDate, &r.ExpiryDate, &r.SourceDocID,
		&r.IsActive, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTradeRuleNotFound
		}
		return nil, err
	}
	return &r, nil
}

// List 按类别列出交易规则；org_id 由 OrgFilter 自动过滤。
func (r *TradeRuleRepository) List(ctx context.Context, category string) ([]*TradeRule, error) {
	where := make([]string, 0, 2)
	args := make([]any, 0, 2)
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		where = append(where, fmt.Sprintf("org_id = $%d::uuid", len(args)))
	}
	if category != "" {
		args = append(args, category)
		where = append(where, fmt.Sprintf("rule_category = $%d", len(args)))
	}
	q := "SELECT " + tradeRuleColumns + " FROM trade_rules"
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY effective_date DESC, rule_key"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*TradeRule, 0, 32)
	for rows.Next() {
		var ru TradeRule
		if err := rows.Scan(&ru.ID, &ru.OrgID, &ru.RuleKey, &ru.RuleValue, &ru.RuleCategory,
			&ru.Description, &ru.EffectiveDate, &ru.ExpiryDate, &ru.SourceDocID,
			&ru.IsActive, &ru.CreatedBy, &ru.CreatedAt, &ru.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &ru)
	}
	return list, rows.Err()
}

// GetByID 按 ID 获取单条交易规则。
func (r *TradeRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*TradeRule, error) {
	q := "SELECT " + tradeRuleColumns + " FROM trade_rules WHERE id = $1"
	args := []any{id}
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	return scanTradeRule(r.pool.QueryRow(ctx, q, args...))
}

// Create 新建交易规则；写操作要求具体活跃租户。
func (r *TradeRuleRepository) Create(ctx context.Context, in *TradeRuleInput, createdBy *uuid.UUID) (*TradeRule, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	var sourceDocID *uuid.UUID
	if in.SourceDocID != "" {
		sd, err := uuid.Parse(in.SourceDocID)
		if err == nil {
			sourceDocID = &sd
		}
	}
	q := `INSERT INTO trade_rules
		(org_id, rule_key, rule_value, rule_category, description,
		 effective_date, expiry_date, source_doc_id, is_active, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING ` + tradeRuleColumns
	return scanTradeRule(r.pool.QueryRow(ctx, q,
		org, in.RuleKey, in.RuleValue, in.RuleCategory, nullStr(in.Description),
		in.EffectiveDate, in.ExpiryDate, sourceDocID, in.IsActive, createdBy))
}

// Update 更新交易规则。
func (r *TradeRuleRepository) Update(ctx context.Context, id uuid.UUID, in *TradeRuleInput) (*TradeRule, error) {
	var sourceDocID *uuid.UUID
	if in.SourceDocID != "" {
		sd, err := uuid.Parse(in.SourceDocID)
		if err == nil {
			sourceDocID = &sd
		}
	}
	q := `UPDATE trade_rules SET
		rule_key=$2, rule_value=$3, rule_category=$4, description=$5,
		effective_date=$6, expiry_date=$7, source_doc_id=$8,
		is_active=$9, updated_at=now()
		WHERE id=$1`
	args := []any{id, in.RuleKey, in.RuleValue, in.RuleCategory, nullStr(in.Description),
		in.EffectiveDate, in.ExpiryDate, sourceDocID, in.IsActive}
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += " RETURNING " + tradeRuleColumns
	return scanTradeRule(r.pool.QueryRow(ctx, q, args...))
}

// Delete 删除交易规则。
func (r *TradeRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM trade_rules WHERE id = $1`
	args := []any{id}
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTradeRuleNotFound
	}
	return nil
}

// Export 导出全部交易规则（用于 JSON 导出）；org_id 由 OrgFilter 自动过滤。
func (r *TradeRuleRepository) Export(ctx context.Context) ([]*TradeRule, error) {
	// Export 返回当前作用域下的全部规则，不过滤 category
	return r.List(ctx, "")
}
