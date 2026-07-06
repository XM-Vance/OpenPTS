// 金额精度（P4 stage-one）。
//
// 结算/账单域金额从 float64 迁移到 numeric 列 + shopspring/decimal，消除浮点漂移，
// 使"分项之和 == 合计"等不变式精确成立（不再靠 almostEq 容差判定）。
//
// pgx 侧的 numeric↔decimal 编解码在连接建立时注册（见 db.go 的 AfterConnect）。
package db

import "github.com/shopspring/decimal"

func init() {
	// 让 decimal.Decimal 序列化为 JSON 数字（非带引号字符串），
	// 保持对前端的响应形状不变（仍是 number）——后端精确、契约不漂移。
	// PTIS 金额量级（元，百万级）远在 JS 安全整数范围内，无精度风险。
	decimal.MarshalJSONWithoutQuotes = true
}
