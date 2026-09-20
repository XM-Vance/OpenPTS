// handler 包内共享小工具（parseDate 等）。
// 2026-06 自 new_modules.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// requireNonNegFloat 校验 float64 非负，负值返回带字段名的 error。
// 用于交易/导入接口拦截负电量、负价格等脏数据（避免负数落库后污染 SUM/结算）。
func requireNonNegFloat(name string, v float64) error {
	if v < 0 {
		return fmt.Errorf("%s 不能为负数", name)
	}
	return nil
}

// requireNonNegDecimal 同上，decimal 版本（金额/单价）。
func requireNonNegDecimal(name string, v decimal.Decimal) error {
	if v.IsNegative() {
		return fmt.Errorf("%s 不能为负数", name)
	}
	return nil
}

// claimsUserIDString 当前用户 ID（字符串，未登录返回空）。
func claimsUserIDString(c *gin.Context) string {
	if id := claimsUserID(c); id != nil {
		return id.String()
	}
	return ""
}
