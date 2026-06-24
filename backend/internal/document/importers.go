// 「确认入库」映射器：把文档的提取字段写进对应业务表。
// 人工在前端核对/修正字段后触发；按文档归属省写入（与多租户口径一致）。
//
// 支持的 target：
//
//	customers           客户清单 → 客户档案
//	load                负荷数据 → user_load_data（按客户名模糊匹配）
//	monthly_settlement  结算单 → batch_monthly_settlement（按月 Upsert）
//	customer_energy     账单/市场化账单/月度电量 → customer_monthly_energy（按客户名匹配 + 按月 Upsert）
//	policy              政策/规则 → policy_documents（归纳为结构化条目）
//	contracts           合同 → retail_contracts 草稿(status=draft)（按名匹配客户/套餐，缺则提示补全）
package document

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ptis/backend/internal/db"
)

type Importers struct {
	Customers      *db.CustomerRepository
	Intent         *db.IntentCustomerRepository
	Load           *db.LoadRepository
	Monthly        *db.MonthlySettlementRepository
	CustomerEnergy *db.CustomerEnergyRepository
	Policy         *db.PolicyRepository
	Retail         *db.RetailRepository
}

// applyRowResult 行级入库结果（写进 document_applies.detail）。
type applyRowResult struct {
	GroupNo int    `json:"group_no"`
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
}

// SuggestTarget 按文档类型推荐入库目标；空串表示无明确目标（前端让用户选）。
func SuggestTarget(docType string) string {
	switch docType {
	case "客户清单":
		return "customers"
	case "负荷数据":
		return "load"
	case "结算单":
		return "monthly_settlement"
	case "账单":
		return "customer_energy" // 市场化账单/电费账单 → 客户历史电量
	case "政策", "规则":
		return "policy"
	}
	return ""
}

// reviewTargets：高风险入库目标——一律人工确认，绝不自动提交。
//   - customers：创建主数据，错了下游全引用、难清理
//   - contracts：涉及价格/金额，难撤销
//
// 其余（policy/customer_energy/load/monthly_settlement/intent_customers）属记录型、
// 易撤，高置信度可自动提交。
var reviewTargets = map[string]bool{
	"customers": true,
	"contracts": true,
}

// NeedsReview 判断该入库是否必须人工确认（不自动提交）：
// 命中高风险目标，或 GLM 提取最低置信度不足 0.85。
// 命中时 worker 不自动写库，留到文档详情由人核对后「一键归入」。
func NeedsReview(target string, hasGlm bool, minConf float64) bool {
	if reviewTargets[target] {
		return true
	}
	if hasGlm && minConf < 0.85 {
		return true
	}
	return false
}

// IsReviewTarget 是否为高风险入库目标（建客户/合同）。
func IsReviewTarget(target string) bool {
	return reviewTargets[target]
}

// ReviewTargetPermission 返回高风险目标手动入库所需的业务写权限码；
// 非高风险目标返回空串。用于在手动 /apply 时硬性校验：bot(只读)无此权限，
// 即便有 document_management:write 也不能直接建客户/合同（只能由人或走草稿/审批）。
func ReviewTargetPermission(target string) string {
	switch target {
	case "customers":
		return "customer_management:write"
	case "contracts":
		return "retail_management:write"
	}
	return ""
}

// Apply 执行入库。ctx 必须已带文档归属省（调用方负责 db.WithOrg）。
// docID 为来源文档 id（政策入库时用于关联来源；其余目标忽略）。
func (im *Importers) Apply(ctx context.Context, target string, docID string,
	exts []*db.DocumentExtraction) (applied int, detail json.RawMessage, err error) {

	rows := groupExtractions(exts)
	if len(rows) == 0 {
		return 0, nil, fmt.Errorf("无可入库的提取字段（请先解析或核对字段）")
	}
	// 既有行级数据(group>0)又有文档级字段(group 0)时，批量入库以行级为准，
	// 丢弃 group 0（避免把合同摘要字段误当一行结算数据）。
	if len(rows) > 1 {
		if _, ok := rows[0]; ok {
			delete(rows, 0)
		}
	}

	var results []applyRowResult
	switch target {
	case "customers":
		results, applied, err = im.applyCustomers(ctx, rows)
	case "intent_customers":
		results, applied, err = im.applyIntentCustomers(ctx, rows)
	case "load":
		results, applied, err = im.applyLoad(ctx, rows)
	case "monthly_settlement":
		results, applied, err = im.applyMonthly(ctx, rows)
	case "customer_energy":
		results, applied, err = im.applyCustomerEnergy(ctx, rows)
	case "policy":
		results, applied, err = im.applyPolicy(ctx, docID, rows)
	case "contracts":
		results, applied, err = im.applyContracts(ctx, rows)
	default:
		return 0, nil, fmt.Errorf("不支持的入库目标: %s", target)
	}
	if err != nil {
		return 0, nil, err
	}
	detail, _ = json.Marshal(map[string]any{"rows": results})
	return applied, detail, nil
}

// groupExtractions 按 group_no 聚合为「行」：map[field_key]value。
// group 0（文档级字段，PDF 合同/结算单）也作为一行参与。
func groupExtractions(exts []*db.DocumentExtraction) map[int]map[string]string {
	rows := map[int]map[string]string{}
	for _, e := range exts {
		if e.ValueText == nil || strings.TrimSpace(*e.ValueText) == "" {
			continue
		}
		if rows[e.GroupNo] == nil {
			rows[e.GroupNo] = map[string]string{}
		}
		rows[e.GroupNo][e.FieldKey] = strings.TrimSpace(*e.ValueText)
	}
	for _, row := range rows {
		normalizeRow(row)
	}
	return rows
}

// fieldAliases：把不同文档类型/来源的提取字段名归一到入库器认的标准 key。
// 仅在标准 key 缺失时用别名补，绝不覆盖已有值。
// 注意：甲方/乙方(party_a/party_b)谁是客户因合同而异（甲方常为售电公司），
// 不在此归一，交由合同入库（草稿 + 人工复核）处理，避免把售电方当成客户。
var fieldAliases = map[string][]string{
	"customer_name": {"customer", "company", "unit_name", "customer_company"},
	"period":        {"month", "settle_month", "bill_month"},
	"month":         {"period"},
	"energy":        {"monthly_energy", "settled_energy", "electricity", "power_consumption"},
	"price":         {"unit_price", "electricity_price", "avg_price"},
}

// normalizeRow 用别名补齐标准 key（不覆盖已有值）。
// 修复如 applyLoad 只认 customer_name、applyMonthly 只认 period/energy 等"漏认"问题。
func normalizeRow(row map[string]string) {
	for canonical, aliases := range fieldAliases {
		if strings.TrimSpace(row[canonical]) != "" {
			continue
		}
		for _, a := range aliases {
			if v := strings.TrimSpace(row[a]); v != "" {
				row[canonical] = v
				break
			}
		}
	}
}

func (im *Importers) applyCustomers(ctx context.Context, rows map[int]map[string]string) ([]applyRowResult, int, error) {
	results := []applyRowResult{}
	applied := 0
	for g, row := range rows {
		name := row["customer_name"]
		if name == "" {
			name = row["customer"] // PDF 提取的字段名
		}
		if name == "" {
			results = append(results, applyRowResult{GroupNo: g, Error: "缺少客户名称"})
			continue
		}
		in := db.CustomerInput{
			UserName:  name,
			ShortName: row["short_name"],
			Location:  row["location"],
			Manager:   row["manager"],
			Tags:      splitList(row["tags"]),
			Accounts:  json.RawMessage("[]"),
		}
		if _, err := im.Customers.Create(ctx, in, nil); err != nil {
			results = append(results, applyRowResult{GroupNo: g, Error: shortErr(err)})
			continue
		}
		applied++
		results = append(results, applyRowResult{GroupNo: g, OK: true})
	}
	return results, applied, nil
}

func (im *Importers) applyIntentCustomers(ctx context.Context, rows map[int]map[string]string) ([]applyRowResult, int, error) {
	results := []applyRowResult{}
	applied := 0
	for g, row := range rows {
		name := row["customer_name"]
		if name == "" {
			name = row["customer"]
		}
		if name == "" {
			name = row["company"] // 资质类提取的企业名
		}
		if name == "" {
			results = append(results, applyRowResult{GroupNo: g, Error: "缺少客户名称"})
			continue
		}
		if err := im.Intent.CreateBasic(ctx, name); err != nil {
			results = append(results, applyRowResult{GroupNo: g, Error: shortErr(err)})
			continue
		}
		applied++
		results = append(results, applyRowResult{GroupNo: g, OK: true})
	}
	return results, applied, nil
}

func (im *Importers) applyLoad(ctx context.Context, rows map[int]map[string]string) ([]applyRowResult, int, error) {
	results := []applyRowResult{}
	applied := 0
	for g, row := range rows {
		name, dateStr, curveStr := row["customer_name"], row["date"], row["curve"]
		if name == "" || dateStr == "" || curveStr == "" {
			results = append(results, applyRowResult{GroupNo: g, Error: "缺少客户/日期/曲线"})
			continue
		}
		date, err := parseDateLoose(dateStr)
		if err != nil {
			results = append(results, applyRowResult{GroupNo: g, Error: "日期格式无效: " + dateStr})
			continue
		}
		curve, total, err := parseCurve(curveStr)
		if err != nil {
			results = append(results, applyRowResult{GroupNo: g, Error: shortErr(err)})
			continue
		}
		list, _, err := im.Customers.List(ctx, db.CustomerListFilter{Keyword: name, Limit: 1})
		if err != nil || len(list) == 0 {
			results = append(results, applyRowResult{GroupNo: g, Error: "客户不存在: " + name})
			continue
		}
		if err := im.Load.UpsertCurve(ctx, list[0].ID, date, curve, total); err != nil {
			results = append(results, applyRowResult{GroupNo: g, Error: shortErr(err)})
			continue
		}
		applied++
		results = append(results, applyRowResult{GroupNo: g, OK: true})
	}
	return results, applied, nil
}

func (im *Importers) applyMonthly(ctx context.Context, rows map[int]map[string]string) ([]applyRowResult, int, error) {
	results := []applyRowResult{}
	applied := 0
	for g, row := range rows {
		period := normalizeMonth(row["period"])
		if period == "" {
			results = append(results, applyRowResult{GroupNo: g, Error: "缺少结算月份"})
			continue
		}
		energy := num(row["energy"])
		energyFee := num(row["energy_fee"])
		if energyFee == 0 {
			energyFee = num(row["amount"]) // PDF 提取常用 amount
		}
		capacity := num(row["capacity_fee"])
		ancillary := num(row["ancillary_fee"])
		subsidy := num(row["subsidy"])
		total := num(row["total_fee"])
		if total == 0 {
			total = energyFee + capacity + ancillary
		}
		if energy == 0 && total == 0 {
			results = append(results, applyRowResult{GroupNo: g, Error: "电量与金额均为空"})
			continue
		}
		if err := im.Monthly.Upsert(ctx, period, energy, energyFee, capacity, ancillary, subsidy, total, "IMPORTED"); err != nil {
			results = append(results, applyRowResult{GroupNo: g, Error: shortErr(err)})
			continue
		}
		applied++
		results = append(results, applyRowResult{GroupNo: g, OK: true})
	}
	return results, applied, nil
}

// applyCustomerEnergy 市场化账单/月度电量 → customer_monthly_energy。
// 按客户名模糊匹配到客户档案，按（客户,月份）Upsert 月度电量。
func (im *Importers) applyCustomerEnergy(ctx context.Context, rows map[int]map[string]string) ([]applyRowResult, int, error) {
	results := []applyRowResult{}
	applied := 0
	for g, row := range rows {
		name := row["customer_name"]
		if name == "" {
			name = row["customer"]
		}
		month := normalizeMonth(row["month"])
		if month == "" {
			month = normalizeMonth(row["period"])
		}
		energy := num(row["energy"])
		if energy == 0 {
			energy = num(row["monthly_energy"])
		}
		if name == "" || month == "" {
			results = append(results, applyRowResult{GroupNo: g, Error: "缺少客户名称或月份"})
			continue
		}
		if energy == 0 {
			results = append(results, applyRowResult{GroupNo: g, Error: "电量为空"})
			continue
		}
		list, _, err := im.Customers.List(ctx, db.CustomerListFilter{Keyword: name, Limit: 1})
		if err != nil || len(list) == 0 {
			results = append(results, applyRowResult{GroupNo: g, Error: "客户不存在: " + name})
			continue
		}
		if err := im.CustomerEnergy.Upsert(ctx, list[0].ID.String(), month, energy); err != nil {
			results = append(results, applyRowResult{GroupNo: g, Error: shortErr(err)})
			continue
		}
		applied++
		results = append(results, applyRowResult{GroupNo: g, OK: true})
	}
	return results, applied, nil
}

// applyPolicy 政策文件 → policy_documents（归纳为结构化条目，关联来源文档）。
// 政策类提取字段较松散，尽量从多个候选 key 取标题/文号/生效日期/摘要。
func (im *Importers) applyPolicy(ctx context.Context, docID string, rows map[int]map[string]string) ([]applyRowResult, int, error) {
	results := []applyRowResult{}
	applied := 0
	for g, row := range rows {
		title := firstNonEmpty(row["title"], row["policy_name"], row["name"], row["subject"])
		summary := firstNonEmpty(row["summary"], row["abstract"], row["content"])
		if title == "" {
			title = summary // 退而求其次：用摘要作标题
		}
		if rs := []rune(title); len(rs) > 80 {
			title = string(rs[:80])
		}
		if title == "" {
			results = append(results, applyRowResult{GroupNo: g, Error: "缺少政策标题/摘要"})
			continue
		}
		in := db.PolicyInput{
			DocumentID:    docID,
			Title:         title,
			DocNo:         firstNonEmpty(row["doc_no"], row["document_no"], row["no"]),
			Category:      firstNonEmpty(row["category"], row["type"]),
			EffectiveDate: normalizeDate(firstNonEmpty(row["effective_date"], row["date"], row["start_date"])),
			Summary:       summary,
			Source:        "document",
		}
		if _, err := im.Policy.Create(ctx, in); err != nil {
			results = append(results, applyRowResult{GroupNo: g, Error: shortErr(err)})
			continue
		}
		applied++
		results = append(results, applyRowResult{GroupNo: g, OK: true})
	}
	return results, applied, nil
}

// applyContracts 合同文档 → 零售合同草稿(status=draft)。
// 尽力按名匹配客户与套餐；匹配不到不报错崩溃，而是给出清晰的待补全提示。
// 高风险目标(reviewTargets)：worker 不会自动调用，仅由人工/显式 apply 触发，建出的也是草稿待复核。
func (im *Importers) applyContracts(ctx context.Context, rows map[int]map[string]string) ([]applyRowResult, int, error) {
	results := []applyRowResult{}
	applied := 0
	for g, row := range rows {
		// 客户：乙方通常是客户；甲方常为售电公司，作兜底候选
		name := firstNonEmpty(row["customer_name"], row["customer"], row["party_b"], row["company"], row["party_a"])
		if name == "" {
			results = append(results, applyRowResult{GroupNo: g, Error: "未识别到客户名称"})
			continue
		}
		list, _, err := im.Customers.List(ctx, db.CustomerListFilter{Keyword: name, Limit: 1})
		if err != nil || len(list) == 0 {
			results = append(results, applyRowResult{GroupNo: g, Error: "未找到客户「" + name + "」，请先归档客户档案"})
			continue
		}
		pkgName := firstNonEmpty(row["package_name"], row["package"])
		if pkgName == "" {
			results = append(results, applyRowResult{GroupNo: g, Error: "缺少套餐名称，请在合同表单补全"})
			continue
		}
		pkgs, err := im.Retail.ListPackages(ctx, pkgName, "")
		if err != nil || len(pkgs) == 0 {
			results = append(results, applyRowResult{GroupNo: g, Error: "未匹配到套餐「" + pkgName + "」，请先创建套餐或在合同表单补全"})
			continue
		}
		in := db.ContractInput{
			CustomerID:          list[0].ID,
			PackageID:           pkgs[0].ID,
			PurchasingEnergyMWH: num(row["energy"]),
			PurchaseStartMonth:  ymOf(row["start_date"]),
			PurchaseEndMonth:    ymOf(row["end_date"]),
			Status:              "draft",
		}
		if r := num(row["green_power_ratio"]); r > 0 {
			ratio := r / 100
			in.GreenPowerRatio = &ratio
		}
		if _, err := im.Retail.CreateContract(ctx, in, nil); err != nil {
			results = append(results, applyRowResult{GroupNo: g, Error: shortErr(err)})
			continue
		}
		applied++
		results = append(results, applyRowResult{GroupNo: g, OK: true})
	}
	return results, applied, nil
}

// ─── 工具 ───

// ymOf 从日期类文本取 YYYY-MM（容忍 2026-01-01 / 2026/1/1 / 2026年1月）。
var ymRe = regexp.MustCompile(`(\d{4})\D+(\d{1,2})`)

func ymOf(s string) string {
	m := ymRe.FindStringSubmatch(s)
	if m == nil {
		return strings.TrimSpace(s)
	}
	mo := m[2]
	if len(mo) == 1 {
		mo = "0" + mo
	}
	return m[1] + "-" + mo
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func normalizeDate(s string) string {
	if s == "" {
		return ""
	}
	d, err := parseDateLoose(s)
	if err != nil {
		return ""
	}
	return d.Format("2006-01-02")
}

func splitList(s string) []string {
	if s == "" {
		return nil
	}
	out := []string{}
	for _, t := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '，' || r == '、' || r == ';' || r == '；'
	}) {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func num(s string) float64 {
	if s == "" {
		return 0
	}
	clean := strings.NewReplacer(",", "", "，", "", " ", "", "元", "", "万", "").Replace(s)
	f, _ := strconv.ParseFloat(clean, 64)
	if strings.Contains(s, "万") {
		f *= 10000
	}
	return f
}

func parseDateLoose(s string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02", "2006/01/02", "2006-1-2", "2006/1/2", "2006年01月02日", "2006年1月2日"} {
		if d, err := time.Parse(layout, strings.TrimSpace(s)); err == nil {
			return d, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析日期: %s", s)
}

func parseCurve(s string) ([]float64, float64, error) {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == '，' || r == ';' })
	if len(parts) != 96 {
		return nil, 0, fmt.Errorf("曲线需 96 点，实际 %d 点", len(parts))
	}
	curve := make([]float64, 96)
	total := 0.0
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, 0, fmt.Errorf("第 %d 点非数值: %s", i+1, p)
		}
		curve[i] = f
		total += f / 4 // 15min 功率 → 电量
	}
	return curve, total, nil
}

func shortErr(err error) string {
	s := err.Error()
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}
