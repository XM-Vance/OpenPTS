// 自定义字段定义仓储：custom_field_definitions CRUD。
package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrCustomFieldNotFound = errors.New("自定义字段不存在")

// CustomFieldDef 自定义字段定义（对应 custom_field_definitions 表）。
type CustomFieldDef struct {
	ID           uuid.UUID       `json:"id"`
	OrgID        *uuid.UUID      `json:"org_id,omitempty"`
	EntityType   string          `json:"entity_type"`
	FieldKey     string          `json:"field_key"`
	FieldLabel   string          `json:"field_label"`
	FieldType    string          `json:"field_type"`
	Options      json.RawMessage `json:"options,omitempty"`
	DefaultValue *string         `json:"default_value,omitempty"`
	IsRequired   bool            `json:"is_required"`
	IsSearchable bool            `json:"is_searchable"`
	SortOrder    int             `json:"sort_order"`
	CreatedBy    *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CustomFieldDefInput 自定义字段写入参数。
type CustomFieldDefInput struct {
	EntityType   string          `json:"entity_type"`
	FieldKey     string          `json:"field_key"`
	FieldLabel   string          `json:"field_label"`
	FieldType    string          `json:"field_type"`
	Options      json.RawMessage `json:"options,omitempty"`
	DefaultValue string          `json:"default_value"`
	IsRequired   bool            `json:"is_required"`
	IsSearchable bool            `json:"is_searchable"`
	SortOrder    int             `json:"sort_order"`
}

// CustomFieldRepository 自定义字段仓储。
type CustomFieldRepository struct{ pool *Pool }

// NewCustomFieldRepository 创建自定义字段仓储。
func NewCustomFieldRepository(pool *Pool) *CustomFieldRepository {
	return &CustomFieldRepository{pool: pool}
}

const customFieldColumns = `id, org_id, entity_type, field_key, field_label, field_type,
	options, default_value, is_required, is_searchable, sort_order,
	created_by, created_at, updated_at`

func scanCustomField(row pgx.Row) (*CustomFieldDef, error) {
	var d CustomFieldDef
	err := row.Scan(&d.ID, &d.OrgID, &d.EntityType, &d.FieldKey, &d.FieldLabel,
		&d.FieldType, &d.Options, &d.DefaultValue, &d.IsRequired, &d.IsSearchable,
		&d.SortOrder, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCustomFieldNotFound
		}
		return nil, err
	}
	return &d, nil
}

// List 按实体类型列出字段定义；org_id 由 OrgFilter 自动过滤。
func (r *CustomFieldRepository) List(ctx context.Context, entityType string) ([]*CustomFieldDef, error) {
	where := make([]string, 0, 2)
	args := make([]any, 0, 2)
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		where = append(where, fmt.Sprintf("org_id = $%d::uuid", len(args)))
	}
	if entityType != "" {
		args = append(args, entityType)
		where = append(where, fmt.Sprintf("entity_type = $%d", len(args)))
	}
	q := "SELECT " + customFieldColumns + " FROM custom_field_definitions"
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY sort_order, created_at"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*CustomFieldDef, 0, 16)
	for rows.Next() {
		var d CustomFieldDef
		if err := rows.Scan(&d.ID, &d.OrgID, &d.EntityType, &d.FieldKey, &d.FieldLabel,
			&d.FieldType, &d.Options, &d.DefaultValue, &d.IsRequired, &d.IsSearchable,
			&d.SortOrder, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

// Create 新建自定义字段定义；写操作要求具体活跃省。
func (r *CustomFieldRepository) Create(ctx context.Context, in *CustomFieldDefInput, createdBy *uuid.UUID) (*CustomFieldDef, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	fieldType := in.FieldType
	if fieldType == "" {
		fieldType = "text"
	}
	q := `INSERT INTO custom_field_definitions
		(org_id, entity_type, field_key, field_label, field_type,
		 options, default_value, is_required, is_searchable, sort_order, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING ` + customFieldColumns
	return scanCustomField(r.pool.QueryRow(ctx, q,
		org, in.EntityType, in.FieldKey, in.FieldLabel, fieldType,
		nullJSON(in.Options), nullStr(in.DefaultValue), in.IsRequired, in.IsSearchable,
		in.SortOrder, createdBy))
}

// Update 更新自定义字段定义。
func (r *CustomFieldRepository) Update(ctx context.Context, id uuid.UUID, in *CustomFieldDefInput) (*CustomFieldDef, error) {
	fieldType := in.FieldType
	if fieldType == "" {
		fieldType = "text"
	}
	q := `UPDATE custom_field_definitions SET
		entity_type=$2, field_key=$3, field_label=$4, field_type=$5,
		options=$6, default_value=$7, is_required=$8, is_searchable=$9,
		sort_order=$10, updated_at=now()
		WHERE id=$1`
	args := []any{id, in.EntityType, in.FieldKey, in.FieldLabel, fieldType,
		nullJSON(in.Options), nullStr(in.DefaultValue), in.IsRequired, in.IsSearchable,
		in.SortOrder}
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += " RETURNING " + customFieldColumns
	return scanCustomField(r.pool.QueryRow(ctx, q, args...))
}

// Delete 删除自定义字段定义。
func (r *CustomFieldRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM custom_field_definitions WHERE id = $1`
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
		return ErrCustomFieldNotFound
	}
	return nil
}

// nullJSON 对空 json.RawMessage 返回 nil，让 pgx 写入 NULL。
func nullJSON(v json.RawMessage) any {
	if len(v) == 0 {
		return nil
	}
	return v
}
