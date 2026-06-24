// 定价模型单字段更新：审批通过后由 applier 调用（主键为 code 字符串）。
package db

import (
	"context"
	"fmt"
	"strings"
)

var pricingApprovableFields = map[string]string{
	"display_name": "display_name",
	"enabled":      "enabled",
	"sort_order":   "sort_order",
}

// UpdatePricingField 白名单单字段更新。
// id 是定价模型的 code。
func (r *RetailRepository) UpdatePricingField(ctx context.Context, code, field, value string) (int64, error) {
	col, ok := pricingApprovableFields[field]
	if !ok {
		return 0, fmt.Errorf("不允许通过审批修改的字段: %s", field)
	}
	var q string
	var arg any
	switch field {
	case "enabled":
		arg = strings.EqualFold(strings.TrimSpace(value), "true") ||
			strings.EqualFold(strings.TrimSpace(value), "1")
		q = fmt.Sprintf(`UPDATE pricing_models SET %s = $1, updated_at = now() WHERE code = $2`, col)
	case "sort_order":
		var n int
		_, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &n)
		if err != nil {
			return 0, fmt.Errorf("sort_order 必须为整数: %v", err)
		}
		arg = n
		q = fmt.Sprintf(`UPDATE pricing_models SET %s = $1, updated_at = now() WHERE code = $2`, col)
	default:
		arg = value
		q = fmt.Sprintf(`UPDATE pricing_models SET %s = $1, updated_at = now() WHERE code = $2`, col)
	}
	tag, err := r.pool.Exec(ctx, q, arg, code)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
