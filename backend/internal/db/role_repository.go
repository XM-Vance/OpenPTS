// 角色仓储：CRUD + 权限批量分配（事务）。
package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Role struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var ErrRoleNotFound = errors.New("角色不存在")

type RoleRepository struct {
	pool *Pool
}

func NewRoleRepository(pool *Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

const roleColumns = "code, name, description, is_system, is_active, created_at, updated_at"

func (r *RoleRepository) scan(row pgx.Row) (*Role, error) {
	var x Role
	err := row.Scan(&x.Code, &x.Name, &x.Description, &x.IsSystem, &x.IsActive, &x.CreatedAt, &x.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &x, nil
}

func (r *RoleRepository) Get(ctx context.Context, code string) (*Role, error) {
	q := `SELECT ` + roleColumns + ` FROM auth_roles WHERE code = $1`
	return r.scan(r.pool.QueryRow(ctx, q, code))
}

func (r *RoleRepository) List(ctx context.Context) ([]*Role, error) {
	q := `SELECT ` + roleColumns + ` FROM auth_roles ORDER BY is_system DESC, code`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Role, 0, 8)
	for rows.Next() {
		x, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, x)
	}
	return list, rows.Err()
}

func (r *RoleRepository) Create(ctx context.Context, code, name string, description *string) (*Role, error) {
	q := `INSERT INTO auth_roles (code, name, description, is_system) VALUES ($1, $2, $3, FALSE) RETURNING ` + roleColumns
	return r.scan(r.pool.QueryRow(ctx, q, code, name, description))
}

func (r *RoleRepository) Update(ctx context.Context, code, name string, description *string, isActive bool) (*Role, error) {
	q := `UPDATE auth_roles SET name = $2, description = $3, is_active = $4 WHERE code = $1 RETURNING ` + roleColumns
	return r.scan(r.pool.QueryRow(ctx, q, code, name, description, isActive))
}

func (r *RoleRepository) Delete(ctx context.Context, code string) error {
	const q = `DELETE FROM auth_roles WHERE code = $1 AND is_system = FALSE`
	tag, err := r.pool.Exec(ctx, q, code)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("无法删除：角色不存在或为系统角色")
	}
	return nil
}

// SetPermissions 批量替换角色的权限分配（事务）。
func (r *RoleRepository) SetPermissions(ctx context.Context, code string, permCodes []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_code = $1`, code); err != nil {
		return err
	}
	// 单条 set-based 插入替代 N 次循环 Exec：unnest 展开权限码数组，一次往返完成。
	if len(permCodes) > 0 {
		if _, err := tx.Exec(ctx,
			`INSERT INTO role_permissions (role_code, permission_code)
			 SELECT $1, unnest($2::text[]) ON CONFLICT DO NOTHING`,
			code, permCodes,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
