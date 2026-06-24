// 客户档案单字段更新：审批通过后由 applier 调用。
package db

import (
	"context"
	"fmt"
)

var customerApprovableFields = map[string]string{
	"short_name": "short_name",
	"manager":    "manager",
	"location":   "location",
	"source":     "source",
}

// UpdateCustomerField 白名单单字段更新。
func (r *CustomerRepository) UpdateCustomerField(ctx context.Context, id, field, value string) (int64, error) {
	col, ok := customerApprovableFields[field]
	if !ok {
		return 0, fmt.Errorf("不允许通过审批修改的字段: %s", field)
	}
	q := fmt.Sprintf(`UPDATE customers SET %s = NULLIF($1, ''), updated_at = now() WHERE id = $2`, col)
	tag, err := r.pool.Exec(ctx, q, value, id)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
