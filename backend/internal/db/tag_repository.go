// 标签定义仓储：tag_definitions CRUD。
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

var ErrTagNotFound = errors.New("标签不存在")

// TagDef 标签定义（对应 tag_definitions 表）。
type TagDef struct {
	ID         uuid.UUID  `json:"id"`
	OrgID      *uuid.UUID `json:"org_id,omitempty"`
	Name       string     `json:"name"`
	Color      string     `json:"color"`
	EntityType string     `json:"entity_type"`
	IsActive   bool       `json:"is_active"`
	SortOrder  int        `json:"sort_order"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// TagInput 标签写入参数。
type TagInput struct {
	Name       string `json:"name"`
	Color      string `json:"color"`
	EntityType string `json:"entity_type"`
	IsActive   bool   `json:"is_active"`
	SortOrder  int    `json:"sort_order"`
}

// TagRepository 标签仓储。
type TagRepository struct{ pool *Pool }

// NewTagRepository 创建标签仓储。
func NewTagRepository(pool *Pool) *TagRepository {
	return &TagRepository{pool: pool}
}

const tagColumns = `id, org_id, name, color, entity_type, is_active, sort_order, created_by, created_at`

func scanTag(row pgx.Row) (*TagDef, error) {
	var t TagDef
	err := row.Scan(&t.ID, &t.OrgID, &t.Name, &t.Color, &t.EntityType,
		&t.IsActive, &t.SortOrder, &t.CreatedBy, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	return &t, nil
}

// List 按实体类型列出标签；org_id 由 OrgFilter 自动过滤。
func (r *TagRepository) List(ctx context.Context, entityType string) ([]*TagDef, error) {
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
	q := "SELECT " + tagColumns + " FROM tag_definitions"
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY sort_order, created_at"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*TagDef, 0, 32)
	for rows.Next() {
		var t TagDef
		if err := rows.Scan(&t.ID, &t.OrgID, &t.Name, &t.Color, &t.EntityType,
			&t.IsActive, &t.SortOrder, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

// Create 新建标签；写操作要求具体活跃省。
func (r *TagRepository) Create(ctx context.Context, in *TagInput, createdBy *uuid.UUID) (*TagDef, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	color := in.Color
	if color == "" {
		color = "#3B82F6"
	}
	entityType := in.EntityType
	if entityType == "" {
		entityType = "customer"
	}
	q := `INSERT INTO tag_definitions
		(org_id, name, color, entity_type, is_active, sort_order, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7)
		RETURNING ` + tagColumns
	return scanTag(r.pool.QueryRow(ctx, q,
		org, in.Name, color, entityType, in.IsActive, in.SortOrder, createdBy))
}

// Update 更新标签。
func (r *TagRepository) Update(ctx context.Context, id uuid.UUID, in *TagInput) (*TagDef, error) {
	color := in.Color
	if color == "" {
		color = "#3B82F6"
	}
	entityType := in.EntityType
	if entityType == "" {
		entityType = "customer"
	}
	q := `UPDATE tag_definitions SET
		name=$2, color=$3, entity_type=$4, is_active=$5, sort_order=$6
		WHERE id=$1`
	args := []any{id, in.Name, color, entityType, in.IsActive, in.SortOrder}
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += " RETURNING " + tagColumns
	return scanTag(r.pool.QueryRow(ctx, q, args...))
}

// Delete 删除标签。
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM tag_definitions WHERE id = $1`
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
		return ErrTagNotFound
	}
	return nil
}

// BatchApply 给多个实体批量打标签（ON CONFLICT DO NOTHING 幂等）。
func (r *TagRepository) BatchApply(ctx context.Context, tagID uuid.UUID, entityType string, entityIDs []uuid.UUID, createdBy *uuid.UUID) (int, error) {
	if len(entityIDs) == 0 {
		return 0, nil
	}
	batch := &pgx.Batch{}
	for _, eid := range entityIDs {
		batch.Queue(
			`INSERT INTO entity_tags (entity_type, entity_id, tag_id, created_by)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (entity_type, entity_id, tag_id) DO NOTHING`,
			entityType, eid, tagID, createdBy,
		)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	applied := 0
	for range entityIDs {
		ct, err := br.Exec()
		if err != nil {
			return applied, err
		}
		applied += int(ct.RowsAffected())
	}
	return applied, nil
}

// GetEntityTags 查询某实体上的所有标签。
func (r *TagRepository) GetEntityTags(ctx context.Context, entityType string, entityID uuid.UUID) ([]*TagDef, error) {
	q := `SELECT t.` + strings.ReplaceAll(tagColumns, ", ", ", t.") + `
	      FROM entity_tags et
	      JOIN tag_definitions t ON t.id = et.tag_id
	      WHERE et.entity_type = $1 AND et.entity_id = $2
	      ORDER BY t.sort_order, t.created_at`
	rows, err := r.pool.Query(ctx, q, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*TagDef, 0, 8)
	for rows.Next() {
		var t TagDef
		if err := rows.Scan(&t.ID, &t.OrgID, &t.Name, &t.Color, &t.EntityType,
			&t.IsActive, &t.SortOrder, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

// RemoveEntityTag 移除某实体的一个标签。
func (r *TagRepository) RemoveEntityTag(ctx context.Context, entityType string, entityID uuid.UUID, tagID uuid.UUID) error {
	ct, err := r.pool.Exec(ctx,
		`DELETE FROM entity_tags WHERE entity_type=$1 AND entity_id=$2 AND tag_id=$3`,
		entityType, entityID, tagID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrTagNotFound
	}
	return nil
}
