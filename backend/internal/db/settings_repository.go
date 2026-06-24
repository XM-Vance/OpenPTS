// 系统配置仓储：key/value 全局参数。
package db

import (
	"context"
	"time"
)

type SystemSetting struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	ValueType   string    `json:"value_type"`
	Category    string    `json:"category"`
	Description *string   `json:"description,omitempty"`
	IsEditable  bool      `json:"is_editable"`
	IsSensitive bool      `json:"is_sensitive"`
	UpdatedBy   *string   `json:"updated_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SettingsRepository struct{ pool *Pool }

func NewSettingsRepository(pool *Pool) *SettingsRepository { return &SettingsRepository{pool: pool} }

func (r *SettingsRepository) List(ctx context.Context, category string) ([]*SystemSetting, error) {
	args := []any{}
	q := `SELECT key, value, value_type, category, description,
	             is_editable, is_sensitive, updated_by, created_at, updated_at
	      FROM system_settings`
	if category != "" {
		args = append(args, category)
		q += " WHERE category = $1"
	}
	q += " ORDER BY category, key"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*SystemSetting, 0)
	for rows.Next() {
		var s SystemSetting
		if err := rows.Scan(&s.Key, &s.Value, &s.ValueType, &s.Category,
			&s.Description, &s.IsEditable, &s.IsSensitive, &s.UpdatedBy,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		// 敏感值打码
		if s.IsSensitive && len(s.Value) > 4 {
			s.Value = s.Value[:2] + "****"
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *SettingsRepository) Update(ctx context.Context, key, value, updatedBy string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE system_settings SET value = $1, updated_by = $2, updated_at = now()
		WHERE key = $3 AND is_editable = true`, value, updatedBy, key)
	return err
}
