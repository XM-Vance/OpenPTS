// 表计导入解析单测（纯函数，不碰 DB）。
package handler

import (
	"strings"
	"testing"
)

func mRows(s string) [][]string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	rows := make([][]string, 0, len(lines))
	for _, l := range lines {
		rows = append(rows, strings.Split(l, ","))
	}
	return rows
}

func TestNormalizeCurve96(t *testing.T) {
	// 48 → 96 重复
	c, est := normalizeCurve96(make([]float64, 48))
	if len(c) != 96 || est {
		t.Errorf("48 点应升 96 且非估算，got len=%d est=%v", len(c), est)
	}
	// 单值 → 均匀 96 点（估算）：日总电量 kWh ÷ 24h = 平均 kW，聚合 Σcurve/4 应还原日电量
	c, est = normalizeCurve96([]float64{2400})
	if len(c) != 96 || !est || c[0] != 100 { // 2400kWh/24h = 100kW
		t.Errorf("日电量 2400 应铺 100kW×96 估算曲线，got len=%d v0=%v est=%v", len(c), c[0], est)
	}
	if c2, _ := normalizeCurve96([]float64{1, 2, 3}); c2 != nil {
		t.Error("非 96/48/1 点应返回 nil")
	}
}

func TestParseMeterFile(t *testing.T) {
	resolve := func(name string) (string, bool) {
		if name == "测试客户" {
			return "00000000-0000-0000-0000-000000000001", true
		}
		return "", false
	}

	t.Run("紧凑 96 点曲线 + 倍率", func(t *testing.T) {
		curve96 := strings.Repeat("100,", 96)
		rows := mRows("客户,表号,日期,倍率,曲线\n测试客户,M-001,2026-08-10,2," + strings.TrimSuffix(curve96, ","))
		valid, previews, invalid := parseMeterFile(rows, resolve)
		if invalid != 0 || len(valid) != 1 {
			t.Fatalf("应 1 行有效，got valid=%d invalid=%d errs=%v", len(valid), invalid, previews[0].Errors)
		}
		if valid[0].Multiplier != 2 || len(valid[0].Curve96) != 96 {
			t.Errorf("倍率/曲线长度不符: %v/%d", valid[0].Multiplier, len(valid[0].Curve96))
		}
		if previews[0].TotalKWh != 2400 { // Σ100×96/4
			t.Errorf("日电量应为 2400 kWh，got %v", previews[0].TotalKWh)
		}
	})

	t.Run("展开 48 列自动升 96 + 日电量单值估算", func(t *testing.T) {
		day := "测试客户,M-002,2026-08-11,1," + strings.TrimSuffix(strings.Repeat("50,", 48), ",")
		flat := "测试客户,M-003,2026-08-11,1,960"
		header := "客户,表号,日期,倍率," + strings.TrimSuffix(strings.Repeat("c,", 48), ",")
		rows := mRows(header + "\n" + day + "\n" + flat)
		valid, previews, _ := parseMeterFile(rows, resolve)
		if len(valid) != 2 {
			t.Fatalf("应 2 行有效，got %d（errs=%v / %v）", len(valid), previews[0].Errors, previews[1].Errors)
		}
		if len(valid[0].Curve96) != 96 || previews[0].Estimated {
			t.Error("48 列应升 96 非估算")
		}
		if !previews[1].Estimated || previews[1].TotalKWh != 960 {
			t.Errorf("单值应按日电量估算 960，got est=%v total=%v", previews[1].Estimated, previews[1].TotalKWh)
		}
	})

	t.Run("客户不存在/坏日期/负值曲线报错", func(t *testing.T) {
		rows := mRows("客户,表号,日期,倍率,曲线\n路人甲,M-1,2026-08-10,1," + strings.TrimSuffix(strings.Repeat("1,", 96), ","))
		_, _, invalid := parseMeterFile(rows, resolve)
		if invalid != 1 {
			t.Errorf("客户不存在应报错，got invalid=%d", invalid)
		}
		rows = mRows("客户,表号,日期,倍率,曲线\n测试客户,M-1,2026/08/10,1," + strings.TrimSuffix(strings.Repeat("1,", 96), ","))
		_, _, invalid = parseMeterFile(rows, resolve)
		if invalid != 1 {
			t.Errorf("坏日期应报错，got invalid=%d", invalid)
		}
		neg := strings.Repeat("-5,", 96)
		rows = mRows("客户,表号,日期,倍率,曲线\n测试客户,M-1,2026-08-10,1," + strings.TrimSuffix(neg, ","))
		_, _, invalid = parseMeterFile(rows, resolve)
		if invalid != 1 {
			t.Errorf("负值曲线应报错，got invalid=%d", invalid)
		}
	})
}
