package contractpdf

import (
	"strings"
	"testing"
)

func TestGenerateHTML(t *testing.T) {
	doc, err := GenerateHTML(ContractData{
		ContractID: "CT-001", CustomerName: "测试客户", PackageName: "固定套餐",
		PurchasingEnergy: 1200.5, GreenPowerRatio: 30, PurchaseStartMonth: "2026-07",
		PurchaseEndMonth: "2026-12", Status: "active",
	})
	if err != nil || len(doc) == 0 {
		t.Fatalf("生成失败: %v", err)
	}
	html := string(doc)
	for _, want := range []string{"电力零售合同", "测试客户", "固定套餐", "1200.50", "2026-07", "打印 / 保存为 PDF", "签章位"} {
		if !strings.Contains(html, want) {
			t.Errorf("文档缺少关键内容: %q", want)
		}
	}
	// 绿电比例为 0 时不出该行
	doc2, _ := GenerateHTML(ContractData{ContractID: "x", CustomerName: "c", PackageName: "p", PurchasingEnergy: 1, PurchaseStartMonth: "2026-01", PurchaseEndMonth: "2026-02", Status: "active"})
	if strings.Contains(string(doc2), "绿电比例") {
		t.Error("绿电比例 0 时不应渲染该行")
	}
}
