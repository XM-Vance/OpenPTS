package document

import "testing"

// PDF OCR 输出的 markdown 表格 → 行数据 → 模板识别（批量入库的核心链路）。
func TestMarkdownTableToRowsAndTemplate(t *testing.T) {
	md := `| 月份 | 电量 | 电费 | 合计 |
| --- | --- | --- | --- |
| 2026-04 | 12345.6 | 4567890 | 4691346 |
| 2026年5月 | 13456.7 | 4998877 | 5128877 |`

	rows := markdownTableToRows(md)
	if len(rows) != 3 {
		t.Fatalf("期望 3 行（表头+2 数据行），实际 %d", len(rows))
	}
	if rows[0][0] != "月份" || rows[1][0] != "2026-04" {
		t.Fatalf("行内容解析错误: %v", rows[:2])
	}

	docType, exts := detectTemplate(rows)
	if docType != "结算单" {
		t.Fatalf("期望识别为结算单，实际 %s", docType)
	}
	// 2 行 × (period+energy+energy_fee+total_fee) = 8 个字段
	if len(exts) != 8 {
		t.Fatalf("期望 8 个行级字段，实际 %d: %+v", len(exts), exts)
	}
	// 中文月份归一化
	var foundMay bool
	for _, e := range exts {
		if e.FieldKey == "period" && e.ValueText == "2026-05" {
			foundMay = true
		}
	}
	if !foundMay {
		t.Fatal("「2026年5月」未归一化为 2026-05")
	}
}

func TestMarkdownTableSkipsSeparatorAndEscapes(t *testing.T) {
	md := "| A | B |\n|---|:---:|\n| x\\|y | z |"
	rows := markdownTableToRows(md)
	if len(rows) != 2 {
		t.Fatalf("期望 2 行，实际 %d", len(rows))
	}
	if rows[1][0] != "x|y" {
		t.Fatalf("转义竖线还原错误: %q", rows[1][0])
	}
}
