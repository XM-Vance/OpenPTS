// 系统模块仓储：仅提供列表（菜单结构数据）。
package db

import "context"

type Module struct {
	Code       string   `json:"code"`
	Name       string   `json:"name"`
	MenuGroup  *string  `json:"menu_group,omitempty"`
	RoutePaths []string `json:"route_paths"`
	SortOrder  int      `json:"sort_order"`
	IsActive   bool     `json:"is_active"`
}

type ModuleRepository struct {
	pool *Pool
}

func NewModuleRepository(pool *Pool) *ModuleRepository {
	return &ModuleRepository{pool: pool}
}

func (r *ModuleRepository) List(ctx context.Context) ([]*Module, error) {
	const q = `
        SELECT code, name, menu_group, route_paths, sort_order, is_active
        FROM auth_modules
        WHERE is_active = TRUE
        ORDER BY sort_order, code
    `
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Module, 0, 16)
	for rows.Next() {
		var m Module
		if err := rows.Scan(&m.Code, &m.Name, &m.MenuGroup, &m.RoutePaths, &m.SortOrder, &m.IsActive); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}
