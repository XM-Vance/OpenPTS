// handler 包内共享小工具（parseDate 等）。
// 2026-06 自 new_modules.go 按域拆分迁移（纯移动，无逻辑变更）。
package handler

import (
	"time"
)

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
