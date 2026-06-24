// 零售合同单字段更新：用于审批通过后由 applier 调用，无需走完整 ContractInput。
// 字段白名单防注入。
package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrFieldNotAllowed = errors.New("不允许通过审批修改的字段")
	ErrInvalidValue    = errors.New("字段值非法")
)

// 允许审批回流的字段白名单：列名 → 类型转换函数。
type fieldSpec struct {
	column string
	parse  func(string) (any, error)
}

var contractApprovableFields = map[string]fieldSpec{
	"purchasing_energy_mwh": {"purchasing_energy_mwh", parseFloat},
	"green_power_ratio":     {"green_power_ratio", parseFloat},
	"purchase_end_month":    {"purchase_end_month", parseMonthString},
	"status":                {"status", parseStatus},
}

func parseFloat(s string) (any, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %s 不是数字", ErrInvalidValue, s)
	}
	return v, nil
}

func parseMonthString(s string) (any, error) {
	// YYYY-MM 或 YYYY-MM-DD 都接受，存储为 YYYY-MM
	s = strings.TrimSpace(s)
	if len(s) < 7 {
		return nil, fmt.Errorf("%w: %s 不是月份", ErrInvalidValue, s)
	}
	return s[:7], nil
}

func parseStatus(s string) (any, error) {
	s = strings.TrimSpace(s)
	switch s {
	case "active", "expired", "terminated":
		return s, nil
	}
	return nil, fmt.Errorf("%w: status 必须是 active/expired/terminated", ErrInvalidValue)
}

// UpdateContractField 单字段更新；返回行影响数。
func (r *RetailRepository) UpdateContractField(ctx context.Context, id, field, value string) (int64, error) {
	spec, ok := contractApprovableFields[field]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrFieldNotAllowed, field)
	}
	parsed, err := spec.parse(value)
	if err != nil {
		return 0, err
	}
	q := fmt.Sprintf(`UPDATE retail_contracts SET %s = $1, updated_at = now() WHERE id = $2`, spec.column)
	tag, err := r.pool.Exec(ctx, q, parsed, id)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
