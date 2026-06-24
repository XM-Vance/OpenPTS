package document

import "testing"

func TestNormalizeRow(t *testing.T) {
	cases := []struct {
		name    string
		in      map[string]string
		wantKey string
		wantVal string
	}{
		{"customer→customer_name", map[string]string{"customer": "抚州联创电子"}, "customer_name", "抚州联创电子"},
		{"company→customer_name", map[string]string{"company": "冠星迦南门业"}, "customer_name", "冠星迦南门业"},
		{"month→period", map[string]string{"month": "2026-03"}, "period", "2026-03"},
		{"period→month", map[string]string{"period": "2026-03"}, "month", "2026-03"},
		{"monthly_energy→energy", map[string]string{"monthly_energy": "1580.5"}, "energy", "1580.5"},
		{"unit_price→price", map[string]string{"unit_price": "0.42"}, "price", "0.42"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			normalizeRow(c.in)
			if c.in[c.wantKey] != c.wantVal {
				t.Fatalf("normalizeRow 期望 %s=%q，实际 %q", c.wantKey, c.wantVal, c.in[c.wantKey])
			}
		})
	}
}

// 已有标准 key 时不被别名覆盖。
func TestNormalizeRowNoOverwrite(t *testing.T) {
	row := map[string]string{"customer_name": "正确客户", "customer": "别名客户"}
	normalizeRow(row)
	if row["customer_name"] != "正确客户" {
		t.Fatalf("不应被别名覆盖，实际 %q", row["customer_name"])
	}
}

func TestYmOf(t *testing.T) {
	cases := map[string]string{
		"2026-01-01": "2026-01",
		"2026/1/1":   "2026-01",
		"2026年3月":    "2026-03",
		"2026-12":    "2026-12",
		"没有日期":       "没有日期",
	}
	for in, want := range cases {
		if got := ymOf(in); got != want {
			t.Fatalf("ymOf(%q)=%q, 期望 %q", in, got, want)
		}
	}
}

func TestReviewTargetPermission(t *testing.T) {
	cases := map[string]string{
		"customers":          "customer_management:write",
		"contracts":          "retail_management:write",
		"policy":             "",
		"customer_energy":    "",
		"monthly_settlement": "",
	}
	for target, want := range cases {
		if got := ReviewTargetPermission(target); got != want {
			t.Fatalf("ReviewTargetPermission(%q)=%q, 期望 %q", target, got, want)
		}
		// 与 IsReviewTarget 一致性：需权限 ⇔ 高风险目标
		if (ReviewTargetPermission(target) != "") != IsReviewTarget(target) {
			t.Fatalf("%q：ReviewTargetPermission 与 IsReviewTarget 不一致", target)
		}
	}
}

func TestNeedsReview(t *testing.T) {
	cases := []struct {
		name    string
		target  string
		hasGlm  bool
		minConf float64
		want    bool
	}{
		{"高风险-客户(即便高置信)", "customers", false, 1.0, true},
		{"高风险-合同(即便高置信)", "contracts", true, 0.99, true},
		{"记录型-政策高置信→自动", "policy", true, 0.95, false},
		{"记录型-电量高置信→自动", "customer_energy", false, 1.0, false},
		{"记录型-低置信→人工", "monthly_settlement", true, 0.6, true},
		{"记录型-Excel无GLM高置信→自动", "load", false, 1.0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := NeedsReview(c.target, c.hasGlm, c.minConf); got != c.want {
				t.Fatalf("NeedsReview(%q,%v,%.2f)=%v, 期望 %v", c.target, c.hasGlm, c.minConf, got, c.want)
			}
		})
	}
}
