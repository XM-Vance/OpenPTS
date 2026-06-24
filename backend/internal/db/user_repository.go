// 用户仓储：CRUD + 角色分配 + 列表查询。
package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	DisplayName  *string    `json:"display_name,omitempty"`
	Email        *string    `json:"email,omitempty"`
	Phone        *string    `json:"phone,omitempty"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	OrgID        *string    `json:"org_id,omitempty"` // 主/默认活跃省（组织 ID）
	IsHQ         bool       `json:"is_hq"`            // 总部标记
}

const userColumns = "id, username, password_hash, display_name, email, phone, is_active, last_login_at, created_at, updated_at, org_id::text, is_hq"

type UserRepository struct {
	pool *Pool
}

func NewUserRepository(pool *Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) scan(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName,
		&u.Email, &u.Phone, &u.IsActive, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &u.OrgID, &u.IsHQ,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE username = $1 AND is_active = TRUE LIMIT 1`
	return r.scan(r.pool.QueryRow(ctx, q, username))
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE id = $1 LIMIT 1`
	return r.scan(r.pool.QueryRow(ctx, q, id))
}

// IsOrgMember 校验用户是否可访问某省（user_orgs 成员）。orgID 为组织 UUID 字符串。
func (r *UserRepository) IsOrgMember(ctx context.Context, userID uuid.UUID, orgID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_orgs WHERE user_id = $1 AND org_id = $2::uuid)`,
		userID, orgID).Scan(&exists)
	return exists, err
}

// ListUserOrgs 返回用户可访问的省：总部返回全部 active 组织，否则返回 user_orgs 集合。
func (r *UserRepository) ListUserOrgs(ctx context.Context, userID uuid.UUID, isHQ bool) ([]*Org, error) {
	var q string
	var args []any
	if isHQ {
		q = `SELECT id::text, code, name, is_active, created_at FROM organizations
		     WHERE is_active = TRUE ORDER BY (code='default'), code`
	} else {
		q = `SELECT o.id::text, o.code, o.name, o.is_active, o.created_at
		       FROM user_orgs uo JOIN organizations o ON o.id = uo.org_id
		      WHERE uo.user_id = $1 ORDER BY (o.code='default'), o.code`
		args = []any{userID}
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Org, 0)
	for rows.Next() {
		var o Org
		if err := rows.Scan(&o.ID, &o.Code, &o.Name, &o.IsActive, &o.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &o)
	}
	return list, rows.Err()
}

// SetUserOrgs 设置用户可访问的省 + 总部标记 + 主省（事务）。orgIDs/primaryOrgID 为组织 UUID 字符串。
func (r *UserRepository) SetUserOrgs(ctx context.Context, userID uuid.UUID, orgIDs []string, isHQ bool, primaryOrgID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`UPDATE users SET is_hq = $2, org_id = NULLIF($3,'')::uuid WHERE id = $1`,
		userID, isHQ, primaryOrgID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM user_orgs WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if len(orgIDs) > 0 {
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_orgs (user_id, org_id)
			 SELECT $1, unnest($2::text[])::uuid ON CONFLICT DO NOTHING`,
			userID, orgIDs); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *UserRepository) Create(ctx context.Context, username, passwordHash, displayName string) (*User, error) {
	q := `INSERT INTO users (username, password_hash, display_name) VALUES ($1, $2, $3) RETURNING ` + userColumns
	var dn *string
	if displayName != "" {
		dn = &displayName
	}
	return r.scan(r.pool.QueryRow(ctx, q, username, passwordHash, dn))
}

// UserListFilter 用户列表过滤参数。
type UserListFilter struct {
	Keyword  string
	IsActive *bool
	Limit    int
	Offset   int
}

func (r *UserRepository) List(ctx context.Context, f UserListFilter) ([]*User, int, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	where := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if f.Keyword != "" {
		args = append(args, "%"+f.Keyword+"%")
		where = append(where, fmt.Sprintf("(username ILIKE $%d OR display_name ILIKE $%d)", len(args), len(args)))
	}
	if f.IsActive != nil {
		args = append(args, *f.IsActive)
		where = append(where, fmt.Sprintf("is_active = $%d", len(args)))
	}
	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users"+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := "SELECT " + userColumns + " FROM users" + whereClause +
		" ORDER BY created_at DESC LIMIT " + strconv.Itoa(f.Limit) + " OFFSET " + strconv.Itoa(f.Offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*User, 0, f.Limit)
	for rows.Next() {
		u, err := r.scan(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, u)
	}
	return list, total, rows.Err()
}

// Update 部分更新非空字段。
func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, displayName, email, phone *string, isActive *bool) (*User, error) {
	sets := make([]string, 0, 4)
	args := []any{id}
	if displayName != nil {
		args = append(args, *displayName)
		sets = append(sets, fmt.Sprintf("display_name = $%d", len(args)))
	}
	if email != nil {
		args = append(args, *email)
		sets = append(sets, fmt.Sprintf("email = $%d", len(args)))
	}
	if phone != nil {
		args = append(args, *phone)
		sets = append(sets, fmt.Sprintf("phone = $%d", len(args)))
	}
	if isActive != nil {
		args = append(args, *isActive)
		sets = append(sets, fmt.Sprintf("is_active = $%d", len(args)))
	}
	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}
	q := "UPDATE users SET " + strings.Join(sets, ", ") + " WHERE id = $1 RETURNING " + userColumns
	return r.scan(r.pool.QueryRow(ctx, q, args...))
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	const q = `UPDATE users SET password_hash = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id, passwordHash)
	return err
}

// AssignRoleIfRoleExists 分配角色：角色存在返回 (true,nil)；不存在返回 (false,nil)；错误返回 err。
func (r *UserRepository) AssignRoleIfRoleExists(ctx context.Context, userID uuid.UUID, roleCode string) (bool, error) {
	const q = `
        INSERT INTO user_roles (user_id, role_code)
        SELECT $1, code FROM auth_roles WHERE code = $2 AND is_active = TRUE
        ON CONFLICT (user_id, role_code) DO NOTHING
    `
	tag, err := r.pool.Exec(ctx, q, userID, roleCode)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() > 0 {
		return true, nil
	}
	var exists bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM auth_roles WHERE code = $1 AND is_active = TRUE)`,
		roleCode,
	).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	const q = `SELECT role_code FROM user_roles WHERE user_id = $1 ORDER BY role_code`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	codes := make([]string, 0, 4)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

// SetRoles 批量替换用户的角色（事务）。
func (r *UserRepository) SetRoles(ctx context.Context, userID uuid.UUID, roleCodes []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID); err != nil {
		return err
	}
	// 单条 set-based 插入替代 N 次循环 Exec：仅插入仍处于 active 的角色（语义不变）。
	if len(roleCodes) > 0 {
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_roles (user_id, role_code)
             SELECT $1, code FROM auth_roles WHERE code = ANY($2::text[]) AND is_active = TRUE
             ON CONFLICT DO NOTHING`,
			userID, roleCodes,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
