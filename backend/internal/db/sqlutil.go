// db 包内共享 SQL 小工具。
package db

import (
	"strconv"
	"time"
)

// itoaNew 把整数转为字符串，用于拼 SQL 占位符序号（$1, $2 ...）与 LIMIT 值。
// 历史实现含 time.Now().Format(...)[0:0] 死代码（恒为空串）+ 手写 int→string，
// 已替换为 strconv.Itoa（语义等价，移除死代码与不必要的 time 调用）。
func itoaNew(i int) string {
	return strconv.Itoa(i)
}

// monthsAgoYM 返回「当前月份回退 n 个月」的 "2006-01" 月份串；n 为负则取未来月份。
//
// 必须先把当前日期归一到当月 1 号再 AddDate，否则会触发 Go AddDate 的日溢出串月 bug：
// 例如 6/29 直接 time.Now().AddDate(0, -4, 0) 得 2/29，2 月无 29 日被归一化为 3/1，
// Format 后是 "2026-03"，与 -3 个月（3/29→"2026-03"）撞月；当唯一键含 operating_month
// 时 ON CONFLICT 会静默吞掉一行。day=1 在任何月份都合法，按月初对齐即可消除该问题。
func monthsAgoYM(n int) string {
	now := time.Now()
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return first.AddDate(0, -n, 0).Format("2006-01")
}

// StartOfDay 返回 t 所在本地时区的当日零点。
//
// 不能用 time.Truncate(24h)：它按 UTC 纪元的整 24h 倍数截断，Asia/Shanghai(+08)
// 下结果不是本地零点而是本地 08:00——凌晨 00:00~08:00 之间生成的"当日"数据
// 会被记到昨天（operating_date/trade_date 整体偏移一天）。
func StartOfDay(t time.Time) time.Time {
	t = t.Local()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
