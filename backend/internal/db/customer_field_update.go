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

// UpdateCustomerField 白名单单字段更新（用连接池，非事务）。
func (r *CustomerRepository) UpdateCustomerField(ctx context.Context, id, field, value string) (int64, error) {
	return r.updateCustomerFieldOn(ctx, r.pool, id, field, value)
}

// UpdateCustomerFieldTx 白名单单字段更新；在指定事务上执行（Approve 单事务原子化用，P0-B2）。
func (r *CustomerRepository) UpdateCustomerFieldTx(ctx context.Context, ex Executor, id, field, value string) (int64, error) {
	return r.updateCustomerFieldOn(ctx, ex, id, field, value)
}

func (r *CustomerRepository) updateCustomerFieldOn(ctx context.Context, ex Executor, id, field, value string) (int64, error) {
	col, ok := customerApprovableFields[field]
	if !ok {
		return 0, fmt.Errorf("不允许通过审批修改的字段: %s", field)
	}
	q := fmt.Sprintf(`UPDATE customers SET %s = NULLIF($1, ''), updated_at = now() WHERE id = $2`, col)
	args := []any{value, id}
	if org, scoped := OrgFilter(ctx); scoped { // 防按 id 改他省客户；总部「全部省」不限
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	tag, err := ex.Exec(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
