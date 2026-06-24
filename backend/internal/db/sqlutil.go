// db 包内共享 SQL 小工具。
package db

import "strconv"

// itoaNew 把整数转为字符串，用于拼 SQL 占位符序号（$1, $2 ...）与 LIMIT 值。
// 历史实现含 time.Now().Format(...)[0:0] 死代码（恒为空串）+ 手写 int→string，
// 已替换为 strconv.Itoa（语义等价，移除死代码与不必要的 time 调用）。
func itoaNew(i int) string {
	return strconv.Itoa(i)
}
