// 合同文档生成（WP6.2）：自包含 HTML（中文零依赖，浏览器「打印为 PDF」即正式文件）。
// 取代 501 占位（gofpdf 2026-08 移除后无中文能力）。带 print CSS + 签章位。
package contractpdf

import (
	"bytes"
	"fmt"
	"html/template"
)

// HTMLTemplate 自包含合同模板（打印 A4、内联样式、签章位）。
var HTMLTemplate = template.Must(template.New("contract").Parse(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>零售合同 {{.ContractID}}</title>
<style>
  @page { size: A4; margin: 2cm; }
  body { font-family: "Songti SC","SimSun","Noto Serif CJK SC",serif; color:#111; line-height:1.8; margin:0; }
  h1 { text-align:center; font-size:22pt; letter-spacing:.3em; margin-bottom:.2em; }
  .no { text-align:center; color:#555; font-size:10pt; margin-bottom:2em; }
  table { width:100%; border-collapse:collapse; margin:1em 0; font-size:11pt; }
  td, th { border:1px solid #333; padding:8px 10px; text-align:left; }
  th { background:#f5f5f5; width:22%; font-weight:600; }
  .clauses p { text-indent:2em; margin:.6em 0; font-size:11pt; }
  .sign { margin-top:3em; display:flex; justify-content:space-between; }
  .sign div { width:45%; }
  .seal { margin-top:4em; color:#999; font-size:10pt; border:1px dashed #999; padding:2em 1em; text-align:center; }
  @media print { .noprint { display:none; } body { font-size:11pt; } }
  .noprint { position:fixed; top:0; right:0; background:#2563eb; color:#fff; padding:10px 18px;
             border-radius:0 0 0 8px; cursor:pointer; border:none; font-size:14px; }
</style>
</head>
<body>
<button class="noprint" onclick="window.print()">打印 / 保存为 PDF</button>
<h1>电力零售合同</h1>
<div class="no">合同编号：{{.ContractID}}</div>
<table>
  <tr><th>客户名称</th><td>{{.CustomerName}}</td></tr>
  <tr><th>套餐名称</th><td>{{.PackageName}}</td></tr>
  <tr><th>购电量</th><td>{{printf "%.2f" .PurchasingEnergy}} MWh</td></tr>
  {{if gt .GreenPowerRatio 0.0}}<tr><th>绿电比例</th><td>{{printf "%.1f" .GreenPowerRatio}}%</td></tr>{{end}}
  <tr><th>购电期间</th><td>{{.PurchaseStartMonth}} 至 {{.PurchaseEndMonth}}</td></tr>
  <tr><th>合同状态</th><td>{{.Status}}</td></tr>
</table>
<div class="clauses">
  <p>一、甲乙双方经协商一致，就电力零售购售电事宜订立本合同。甲方按本合同约定向乙方供电，乙方按约定支付电费。</p>
  <p>二、电价与结算方式按所选用套餐的定价模型执行，具体以平台结算明细为准。</p>
  <p>三、任何一方违反本合同约定给对方造成损失的，应承担相应违约责任。</p>
  <p>四、本合同自双方签署（或平台审批通过）之日起生效，至购电期间届满时终止。</p>
</div>
<div class="sign">
  <div>甲方（售电方）：<br><br>签字：＿＿＿＿＿＿＿＿<br>日期：＿＿＿＿＿＿＿＿</div>
  <div>乙方（购电方）：{{.CustomerName}}<br><br>签字：＿＿＿＿＿＿＿＿<br>日期：＿＿＿＿＿＿＿＿</div>
</div>
<div class="seal">（此处加盖公章 / 电子签章位）</div>
</body>
</html>
`))

// GenerateHTML 渲染自包含合同 HTML（浏览器打印即 PDF；中文零外部依赖）。
func GenerateHTML(d ContractData) ([]byte, error) {
	var buf bytes.Buffer
	if err := HTMLTemplate.Execute(&buf, d); err != nil {
		return nil, fmt.Errorf("合同模板渲染失败: %w", err)
	}
	return buf.Bytes(), nil
}
