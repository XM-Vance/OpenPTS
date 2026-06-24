// 权限仓储：用户权限聚合查询、权限点列表、角色权限列表。
package db

import (
	"context"

	"github.com/google/uuid"
)

type Permission struct {
	Code           string `json:"code"`
	Name           string `json:"name"`
	ModuleCode     string `json:"module_code"`
	Action         string `json:"action"`
	PermissionType string `json:"permission_type"`
	IsActive       bool   `json:"is_active"`
}

type PermissionRepository struct {
	pool *Pool
}

func NewPermissionRepository(pool *Pool) *PermissionRepository {
	return &PermissionRepository{pool: pool}
}

// ListUserPermissionCodes 返回用户通过角色拥有的所有权限码（已去重 + 过滤掉禁用项）。
func (r *PermissionRepository) ListUserPermissionCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	const q = `
        SELECT DISTINCT rp.permission_code
        FROM user_roles ur
        JOIN role_permissions rp ON rp.role_code = ur.role_code
        JOIN auth_permissions ap ON ap.code = rp.permission_code
        JOIN auth_roles ar       ON ar.code = ur.role_code
        WHERE ur.user_id = $1
          AND ap.is_active = TRUE
          AND ar.is_active = TRUE
        ORDER BY rp.permission_code
    `
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	codes := make([]string, 0, 32)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

// ListAll 列出所有激活的权限点（前端权限配置 UI 用）。
func (r *PermissionRepository) ListAll(ctx context.Context) ([]*Permission, error) {
	const q = `
        SELECT code, name, module_code, action, permission_type, is_active
        FROM auth_permissions
        WHERE is_active = TRUE
        ORDER BY module_code, action
    `
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Permission, 0, 64)
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.Code, &p.Name, &p.ModuleCode, &p.Action, &p.PermissionType, &p.IsActive); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// ListByRole 返回某角色拥有的权限码列表。
func (r *PermissionRepository) ListByRole(ctx context.Context, roleCode string) ([]string, error) {
	const q = `SELECT permission_code FROM role_permissions WHERE role_code = $1 ORDER BY permission_code`
	rows, err := r.pool.Query(ctx, q, roleCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	codes := make([]string, 0, 16)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}
