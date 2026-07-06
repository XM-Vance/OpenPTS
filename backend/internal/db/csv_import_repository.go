// 通用 CSV 数据导入仓储：支撑 solar/storage/freq 等模块从外部计量系统导入真实数据。
//
// 设计：通用化解析 CSV → 批量写入目标表。每个模块的 CSV 字段映射由 handler 定义，
// 本仓储提供批量 INSERT 的通用能力（COPY 式批量插入）。
//
// 替代各模块 handler 里的 rand.Float64() 造数：导入真实计量数据后，handler 优先读真实表，
// 无数据时回退 demo 并在响应标注 is_demo=true（前端显示「演示数据」水印）。
package db

import (
	"context"
	"fmt"
	"strings"
)

// CSVImportResult 导入结果。
type CSVImportResult struct {
	Table   string `json:"table"`
	Imported int   `json:"imported"`
	IsDemo  bool   `json:"is_demo"`
}

// BatchInsert 通用批量插入。
// table: 目标表名（白名单校验防注入）；columns: 列名列表；rows: 值的二维切片（每行对应 columns）。
// 占位符按 args 数量自动编号。orgCol/orgVal：若目标表有 org_id 列，传入列名与值。
func (p *Pool) BatchInsert(
	ctx context.Context, table string, columns []string, rows [][]any,
	orgCol string, orgVal string,
) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	// 白名单校验表名与列名（防 SQL 注入：只允许字母数字下划线）
	if !isSafeIdent(table) {
		return 0, fmt.Errorf("非法表名: %s", table)
	}
	for _, col := range columns {
		if !isSafeIdent(col) {
			return 0, fmt.Errorf("非法列名: %s", col)
		}
	}

	cols := columns
	if orgCol != "" && orgVal != "" {
		cols = append([]string{orgCol}, cols...)
	}
	colList := strings.Join(cols, ", ")

	// 每行一个 ($p1,$p2,...) 占位符组
	ph := make([]string, len(rows))
	startIdx := 1
	for i, row := range rows {
		n := len(row)
		if orgCol != "" && orgVal != "" {
			n++
		}
		parts := make([]string, n)
		for j := 0; j < n; j++ {
			parts[j] = fmt.Sprintf("$%d", startIdx+j)
		}
		ph[i] = "(" + strings.Join(parts, ",") + ")"
		startIdx += n
	}

	// 展开 args（每行前置 orgVal）
	args := make([]any, 0, startIdx)
	for _, row := range rows {
		if orgCol != "" && orgVal != "" {
			args = append(args, orgVal)
		}
		args = append(args, row...)
	}

	q := fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s`,
		table, colList, strings.Join(ph, ","))
	tag, err := p.Exec(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// isSafeIdent 校验标识符只含字母数字下划线（防注入）。
func isSafeIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}
	return true
}
