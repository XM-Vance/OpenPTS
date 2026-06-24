// Excel/CSV 直读解析：不走 OCR，直接读表格结构 → markdown + 行级提取字段。
// 模板识别（按表头列名）：客户清单 / 负荷数据 / 结算单；其余按通用表格处理。
package document

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ptis/backend/internal/db"
	"github.com/xuri/excelize/v2"
)

// tabularResult Excel/CSV 解析产物。
type tabularResult struct {
	DocType     string
	TextContent string
	Tables      json.RawMessage
	Summary     string
	SheetCount  int
	Extractions []db.ExtractionInput
}

// 单 sheet 进 markdown / tables JSON 的行数上限，防超大表撑爆字段。
const maxRowsPerSheet = 500

// parseTabular 解析 xlsx/csv 内容。
func parseTabular(filename string, content []byte) (*tabularResult, error) {
	lower := strings.ToLower(filename)
	var sheets map[string][][]string
	var order []string

	switch {
	case strings.HasSuffix(lower, ".xlsx") || strings.HasSuffix(lower, ".xlsm") || strings.HasSuffix(lower, ".xls"):
		x, err := excelize.OpenReader(bytes.NewReader(content))
		if err != nil {
			return nil, fmt.Errorf("XLSX 解析失败: %w", err)
		}
		defer x.Close()
		sheets = map[string][][]string{}
		for _, name := range x.GetSheetList() {
			rows, err := x.GetRows(name)
			if err != nil {
				continue
			}
			if len(rows) > 0 {
				sheets[name] = rows
				order = append(order, name)
			}
		}
	default: // csv / txt
		buf := content
		if len(buf) >= 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
			buf = buf[3:]
		}
		r := csv.NewReader(bytes.NewReader(buf))
		r.FieldsPerRecord = -1
		rows, err := r.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("CSV 解析失败: %w", err)
		}
		sheets = map[string][][]string{"Sheet1": rows}
		order = []string{"Sheet1"}
	}

	if len(order) == 0 {
		return nil, fmt.Errorf("文件无有效数据")
	}

	// 全文 markdown + tables JSON
	var md strings.Builder
	type tableJSON struct {
		Sheet    string     `json:"sheet"`
		Header   []string   `json:"header"`
		Rows     [][]string `json:"rows"`
		RowCount int        `json:"row_count"`
	}
	tables := make([]tableJSON, 0, len(order))
	for _, name := range order {
		rows := sheets[name]
		md.WriteString(fmt.Sprintf("## 工作表：%s（%d 行）\n\n", name, len(rows)))
		md.WriteString(rowsToMarkdown(rows, maxRowsPerSheet))
		md.WriteString("\n\n")
		t := tableJSON{Sheet: name, RowCount: len(rows)}
		if len(rows) > 0 {
			t.Header = rows[0]
			end := len(rows)
			if end > maxRowsPerSheet {
				end = maxRowsPerSheet
			}
			t.Rows = rows[1:end]
		}
		tables = append(tables, t)
	}
	tablesJSON, _ := json.Marshal(tables)

	// 模板识别基于第一个 sheet
	first := sheets[order[0]]
	docType, exts := detectTemplate(first)

	summary := fmt.Sprintf("Excel/CSV：%d 个工作表，首表 %d 行。识别类型：%s。",
		len(order), len(first), docType)

	return &tabularResult{
		DocType:     docType,
		TextContent: md.String(),
		Tables:      tablesJSON,
		Summary:     summary,
		SheetCount:  len(order),
		Extractions: exts,
	}, nil
}

// rowsToMarkdown 行数据 → markdown 表（限行）。
func rowsToMarkdown(rows [][]string, limit int) string {
	if len(rows) == 0 {
		return "（空表）"
	}
	var b strings.Builder
	n := len(rows)
	if n > limit {
		n = limit
	}
	// 取最大列数对齐
	cols := 0
	for _, r := range rows[:n] {
		if len(r) > cols {
			cols = len(r)
		}
	}
	if cols > 120 {
		cols = 120
	}
	cell := func(r []string, i int) string {
		if i < len(r) {
			return strings.ReplaceAll(strings.TrimSpace(r[i]), "|", "\\|")
		}
		return ""
	}
	for ri := 0; ri < n; ri++ {
		b.WriteString("|")
		for ci := 0; ci < cols; ci++ {
			b.WriteString(" " + cell(rows[ri], ci) + " |")
		}
		b.WriteString("\n")
		if ri == 0 {
			b.WriteString("|" + strings.Repeat(" --- |", cols) + "\n")
		}
	}
	if len(rows) > limit {
		b.WriteString(fmt.Sprintf("\n*（仅展示前 %d 行，共 %d 行）*\n", limit, len(rows)))
	}
	return b.String()
}

// markdownTableToRows 把 OCR 输出的 markdown 表格转回行数据（跳过分隔行）。
// PDF 中的表格数据借此复用 Excel 的模板识别，实现批量入库。
func markdownTableToRows(md string) [][]string {
	rows := [][]string{}
	const pipeEsc = "\x00" // 转义竖线占位：先替换再按 | 切分，避免切开 \| 内容
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		line = strings.ReplaceAll(line, "\\|", pipeEsc)
		cells := strings.Split(strings.Trim(line, "|"), "|")
		// 分隔行（|---|---|）跳过
		sep := true
		for _, c := range cells {
			t := strings.TrimSpace(c)
			if t != "" && strings.Trim(t, "-: ") != "" {
				sep = false
				break
			}
		}
		if sep {
			continue
		}
		row := make([]string, len(cells))
		for i, c := range cells {
			row[i] = strings.TrimSpace(strings.ReplaceAll(c, pipeEsc, "|"))
		}
		rows = append(rows, row)
	}
	return rows
}

// ─── 模板识别 ───

// headerIndex 表头 → 列号（去空格）。
func headerIndex(header []string) map[string]int {
	idx := map[string]int{}
	for i, h := range header {
		idx[strings.TrimSpace(h)] = i
	}
	return idx
}

// pick 从行取列值（任一候选列名）。
func pick(row []string, idx map[string]int, names ...string) string {
	for _, n := range names {
		if i, ok := idx[n]; ok && i < len(row) {
			if v := strings.TrimSpace(row[i]); v != "" {
				return v
			}
		}
	}
	return ""
}

func hasAny(idx map[string]int, names ...string) bool {
	for _, n := range names {
		if _, ok := idx[n]; ok {
			return true
		}
	}
	return false
}

// detectTemplate 表头识别已知业务模板，产出行级提取字段。
// 返回 doc_type 与 extractions（group_no = 数据行序号，从 1 起）。
func detectTemplate(rows [][]string) (string, []db.ExtractionInput) {
	if len(rows) < 2 {
		return "其他", nil
	}
	idx := headerIndex(rows[0])

	switch {
	// 负荷数据：客户 + 日期 +（96点 或 P1..）
	case hasAny(idx, "客户名称", "客户", "户名") && hasAny(idx, "日期", "数据日期") &&
		(hasAny(idx, "P1", "96点", "曲线") || len(rows[0]) >= 90):
		return "负荷数据", extractLoadRows(rows, idx)

	// 结算单：月份 + 电量 + 费用
	case hasAny(idx, "月份", "结算月份", "账期", "运营月份") &&
		hasAny(idx, "电量", "结算电量", "电量(MWh)", "结算电量(MWh)", "用电量") &&
		hasAny(idx, "电费", "金额", "电费金额", "总费用", "电能量费"):
		return "结算单", extractSettlementRows(rows, idx)

	// 客户清单：有客户名称列（且不是负荷格式）
	case hasAny(idx, "客户名称", "客户全称", "用户名称"):
		return "客户清单", extractCustomerRows(rows, idx)
	}
	return "其他", nil
}

func extractCustomerRows(rows [][]string, idx map[string]int) []db.ExtractionInput {
	out := []db.ExtractionInput{}
	for ri, row := range rows[1:] {
		g := ri + 1
		name := pick(row, idx, "客户名称", "客户全称", "用户名称")
		if name == "" {
			continue
		}
		add := func(key, label, val string) {
			if val != "" {
				out = append(out, db.ExtractionInput{GroupNo: g, FieldKey: key, FieldLabel: label,
					ValueText: val, Confidence: 1, Source: "excel"})
			}
		}
		add("customer_name", "客户名称", name)
		add("short_name", "简称", pick(row, idx, "简称"))
		add("location", "所在地", pick(row, idx, "所在地", "地址", "地区"))
		add("manager", "客户经理", pick(row, idx, "客户经理", "经理", "负责人"))
		add("tags", "标签", pick(row, idx, "标签"))
		add("contact", "联系人", pick(row, idx, "联系人"))
		add("phone", "联系电话", pick(row, idx, "联系电话", "电话", "手机"))
	}
	return out
}

func extractLoadRows(rows [][]string, idx map[string]int) []db.ExtractionInput {
	out := []db.ExtractionInput{}
	header := rows[0]
	// P1..P96 展开列
	pCols := []int{}
	for i, h := range header {
		h = strings.TrimSpace(h)
		if len(h) >= 2 && (h[0] == 'P' || h[0] == 'p') {
			pCols = append(pCols, i)
		}
	}
	for ri, row := range rows[1:] {
		g := ri + 1
		name := pick(row, idx, "客户名称", "客户", "户名")
		date := pick(row, idx, "日期", "数据日期")
		if name == "" || date == "" {
			continue
		}
		var curve string
		if len(pCols) >= 90 {
			vals := make([]string, 0, len(pCols))
			for _, ci := range pCols {
				v := ""
				if ci < len(row) {
					v = strings.TrimSpace(row[ci])
				}
				if v == "" {
					v = "0"
				}
				vals = append(vals, v)
			}
			curve = strings.Join(vals, ",")
		} else {
			curve = pick(row, idx, "96点", "曲线")
		}
		if curve == "" {
			continue
		}
		out = append(out,
			db.ExtractionInput{GroupNo: g, FieldKey: "customer_name", FieldLabel: "客户名称", ValueText: name, Confidence: 1, Source: "excel"},
			db.ExtractionInput{GroupNo: g, FieldKey: "date", FieldLabel: "日期", ValueText: date, Confidence: 1, Source: "excel"},
			db.ExtractionInput{GroupNo: g, FieldKey: "curve", FieldLabel: "96点曲线", ValueText: curve, Confidence: 1, Source: "excel"},
		)
	}
	return out
}

func extractSettlementRows(rows [][]string, idx map[string]int) []db.ExtractionInput {
	out := []db.ExtractionInput{}
	for ri, row := range rows[1:] {
		g := ri + 1
		period := pick(row, idx, "月份", "结算月份", "账期", "运营月份")
		if period == "" {
			continue
		}
		add := func(key, label string, names ...string) {
			if v := pick(row, idx, names...); v != "" {
				out = append(out, db.ExtractionInput{GroupNo: g, FieldKey: key, FieldLabel: label,
					ValueText: v, Confidence: 1, Source: "excel"})
			}
		}
		out = append(out, db.ExtractionInput{GroupNo: g, FieldKey: "period", FieldLabel: "结算月份",
			ValueText: normalizeMonth(period), Confidence: 1, Source: "excel"})
		add("energy", "结算电量", "电量", "结算电量", "电量(MWh)", "结算电量(MWh)", "用电量")
		add("energy_fee", "电能量费", "电费", "电费金额", "电能量费")
		add("capacity_fee", "容量费", "容量费", "输配电费", "容量电费")
		add("ancillary_fee", "辅助服务费", "辅助服务费", "辅助费用")
		add("subsidy", "政策补贴", "政策补贴", "补贴")
		add("total_fee", "总费用", "总费用", "金额", "合计")
	}
	return out
}

// normalizeMonth 「2026年1月 / 2026-1 / 202601」→ 2026-01。
func normalizeMonth(s string) string {
	s = strings.TrimSpace(s)
	s = strings.NewReplacer("年", "-", "月", "", "/", "-", ".", "-").Replace(s)
	parts := strings.SplitN(s, "-", 2)
	if len(parts) == 2 && len(parts[1]) == 1 {
		return parts[0] + "-0" + parts[1]
	}
	if len(s) == 6 { // 202601
		return s[:4] + "-" + s[4:]
	}
	return s
}
