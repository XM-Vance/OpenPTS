// 负荷特性扩展仓储：支撑前端 /load/characteristics/* 的 12 个分析端点。
//
// 数据源（注意命名陷阱）：
//   - customer_characteristics（迁移 0010）：日度特征快照，含 long_term/short_term JSONB、
//     tags、regularity_score、baseline_curve。前端 load-characteristics 要的就是这张表。
//   - customer_anomaly_alerts（0010）：异动告警。
//   - analysis_history_log（0010）：分析执行历史。
//   - customer_monthly_energy（0006）：月度电量。
//   - user_load_data（0006）：96 点日曲线（daily-trend 端点）。
//
// 与 load_characteristics_repository.go（读 load_characteristics 表 0037）无关——那是另一张表。
//
// 多租户：customer_characteristics 有 org_id（0010 建表时已含）；analysis_history_log /
// customer_monthly_energy 无 org_id，通过 JOIN customers 过滤 org。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
)

// LoadCharacteristicsExtRepository 负荷特性分析查询。
type LoadCharacteristicsExtRepository struct {
	pool *Pool
}

func NewLoadCharacteristicsExtRepository(pool *Pool) *LoadCharacteristicsExtRepository {
	return &LoadCharacteristicsExtRepository{pool: pool}
}

// ─── 1. Overview：KPI + 分布 + 告警统计 ───

// LoadCharOverview 对齐前端 CharacteristicsOverview 契约（load-characteristics.ts:142）。
type LoadCharOverview struct {
	KPI          OverviewKPI                  `json:"kpi"`
	Distribution CharDistribution             `json:"distribution"`
	Anomalies    []AnomalySummaryItem         `json:"anomalies"`
}

// OverviewKPI 对齐前端 OverviewKpi（load-characteristics.ts:78）。
type OverviewKPI struct {
	CoverageRate            float64 `json:"coverage_rate"`
	CoverageCount           int     `json:"coverage_count"`
	TotalCustomers          int     `json:"total_customers"`
	DominantTag             string  `json:"dominant_tag,omitempty"`
	DominantTagPercentage   float64 `json:"dominant_tag_percentage"`
	LatestDataDate          string  `json:"latest_data_date,omitempty"`
	AnomalyCountToday       int     `json:"anomaly_count_today"`
	AvgRegularityScore      float64 `json:"avg_regularity_score"`
}

// CharDistribution 对齐前端 {by_shift,by_facility}（load-characteristics.ts:144）。
type CharDistribution struct {
	ByShift    []TagDistributionItem `json:"by_shift"`
	ByFacility []TagDistributionItem `json:"by_facility"`
}

// TagDistributionItem 对齐前端（load-characteristics.ts:89）。
type TagDistributionItem struct {
	Name       string  `json:"name"`
	Value      int     `json:"value"`
	Percentage float64 `json:"percentage"`
}

// AnomalySummaryItem 对齐前端（load-characteristics.ts:132）。
type AnomalySummaryItem struct {
	ID           string `json:"id"`
	CustomerID   string `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	Severity     string `json:"severity"`
	Type         string `json:"type"`
	Description  string `json:"description"`
	Time         string `json:"time"`
}

// Overview 汇总负荷特性概览（结构对齐前端 CharacteristicsOverview）。
func (r *LoadCharacteristicsExtRepository) Overview(ctx context.Context) (*LoadCharOverview, error) {
	org, scoped := OrgFilter(ctx)

	// 有特征的客户数 + 平均规律性 + 最新数据日期
	var kpi OverviewKPI
	var latestDate *time.Time
	q := `SELECT COUNT(DISTINCT cc.customer_id),
			COALESCE(AVG(cc.regularity_score), 0),
			MAX(cc.data_date)
		  FROM customer_characteristics cc`
	args := []any{}
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE cc.org_id = $%d::uuid", len(args))
	}
	if err := r.pool.QueryRow(ctx, q, args...).Scan(&kpi.CoverageCount, &kpi.AvgRegularityScore, &latestDate); err != nil {
		return nil, err
	}
	if latestDate != nil {
		kpi.LatestDataDate = latestDate.Format("2006-01-02")
	}

	// 总客户数（算覆盖率）
	var totalCust int
	qT := `SELECT COUNT(*) FROM customers`
	argsT := []any{}
	if scoped {
		argsT = append(argsT, org)
		qT += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(argsT))
	}
	if err := r.pool.QueryRow(ctx, qT, argsT...).Scan(&totalCust); err != nil {
		return nil, err
	}
	kpi.TotalCustomers = totalCust
	if totalCust > 0 {
		kpi.CoverageRate = float64(kpi.CoverageCount) / float64(totalCust)
	}

	// 今日告警数
	today := time.Now().Format("2006-01-02")
	qA := `SELECT COUNT(*) FROM customer_anomaly_alerts WHERE alert_date::date = $1::date`
	argsA := []any{today}
	if scoped {
		argsA = append(argsA, org)
		qA += fmt.Sprintf(" AND org_id = $%d::uuid", len(argsA))
	}
	_ = r.pool.QueryRow(ctx, qA, argsA...).Scan(&kpi.AnomalyCountToday)

	// tag 分布（归入 by_shift / by_facility 两个类别）+ dominant tag
	byShift, byFacility, dominantTag, dominantPct := r.classifyTags(ctx)
	kpi.DominantTag = dominantTag
	kpi.DominantTagPercentage = dominantPct

	// anomalies 列表（最近 20 条告警，对齐前端 AnomalySummaryItem[]）
	anomalies := r.recentAnomalies(ctx, 20)

	return &LoadCharOverview{
		KPI:          kpi,
		Distribution: CharDistribution{ByShift: byShift, ByFacility: byFacility},
		Anomalies:    anomalies,
	}, nil
}

// classifyTags 按关键词把 tags 归类为 by_shift/by_facility，并返回 dominant tag。
// tags 是 TEXT[] 无 category，故用约定关键词归类（移峰/填谷/尖峰/峰谷平→shift；
// 工业/商业/居民/园区/楼宇→facility）；其余并入 by_facility 兜底。
func (r *LoadCharacteristicsExtRepository) classifyTags(ctx context.Context) ([]TagDistributionItem, []TagDistributionItem, string, float64) {
	raw, err := r.TagDistribution(ctx)
	if err != nil || len(raw) == 0 {
		return []TagDistributionItem{}, []TagDistributionItem{}, "", 0
	}
	total := 0
	for _, n := range raw {
		total += n
	}
	var byShift, byFacility []TagDistributionItem
	var dominantTag string
	var dominantN int
	for tag, n := range raw {
		if n > dominantN {
			dominantN = n
			dominantTag = tag
		}
		pct := 0.0
		if total > 0 {
			pct = float64(n) / float64(total) * 100
		}
		item := TagDistributionItem{Name: tag, Value: n, Percentage: pct}
		if isShiftTag(tag) {
			byShift = append(byShift, item)
		} else {
			byFacility = append(byFacility, item)
		}
	}
	dominantPct := 0.0
	if total > 0 {
		dominantPct = float64(dominantN) / float64(total) * 100
	}
	return byShift, byFacility, dominantTag, dominantPct
}

// isShiftTag 判断 tag 是否属于「移峰填谷/负荷时段」类别。
func isShiftTag(tag string) bool {
	for _, kw := range []string{"移峰", "填谷", "尖峰", "峰谷", "峰平谷", "时段", "可调"} {
		if contains(tag, kw) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// recentAnomalies 取最近 N 条告警，映射为 AnomalySummaryItem。
func (r *LoadCharacteristicsExtRepository) recentAnomalies(ctx context.Context, limit int) []AnomalySummaryItem {
	q := `SELECT a.id::text, a.customer_id::text, cust.user_name,
			a.severity, a.alert_type, COALESCE(a.reason,''), a.alert_date
		  FROM customer_anomaly_alerts a
		  JOIN customers cust ON cust.id = a.customer_id`
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE a.org_id = $%d::uuid", len(args))
	}
	q += fmt.Sprintf(" ORDER BY a.alert_date DESC LIMIT %d", limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return []AnomalySummaryItem{}
	}
	defer rows.Close()
	out := []AnomalySummaryItem{}
	for rows.Next() {
		var a AnomalySummaryItem
		var t time.Time
		if err := rows.Scan(&a.ID, &a.CustomerID, &a.CustomerName,
			&a.Severity, &a.Type, &a.Description, &t); err != nil {
			continue
		}
		a.Time = t.Format("2006-01-02 15:04")
		out = append(out, a)
	}
	return out
}

// ─── 2. Distribution：tag 分布（独立端点 /overview/distribution）───

// TagDistributionResponse 对齐前端 EnhancedTagDistribution（categories 结构）。
type TagDistributionResponse struct {
	Categories []TagCategory `json:"categories"`
}

type TagCategory struct {
	Category     string               `json:"category"`
	CategoryName string               `json:"category_name"`
	Items        []TagDistributionItem `json:"items"`
}

// TagDistribution 统计各 tag 覆盖的客户数（基于每客户最新快照，内部复用）。
func (r *LoadCharacteristicsExtRepository) TagDistribution(ctx context.Context) (map[string]int, error) {
	q := `SELECT tag, COUNT(*) FROM (
			SELECT DISTINCT ON (cc.customer_id) cc.customer_id, UNNEST(cc.tags) AS tag
			FROM customer_characteristics cc`
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE cc.org_id = $%d::uuid", len(args))
	}
	q += ` ORDER BY cc.customer_id, cc.data_date DESC
		) t WHERE tag IS NOT NULL AND tag <> '' GROUP BY tag ORDER BY COUNT(*) DESC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dist := make(map[string]int)
	for rows.Next() {
		var tag string
		var n int
		if err := rows.Scan(&tag, &n); err != nil {
			return nil, err
		}
		dist[tag] = n
	}
	return dist, rows.Err()
}

// TagDistributionGrouped 返回前端 categories 结构（供 /overview/distribution 端点）。
func (r *LoadCharacteristicsExtRepository) TagDistributionGrouped(ctx context.Context) (*TagDistributionResponse, error) {
	raw, err := r.TagDistribution(ctx)
	if err != nil {
		return nil, err
	}
	total := 0
	for _, n := range raw {
		total += n
	}
	var shiftItems, facilityItems []TagDistributionItem
	for tag, n := range raw {
		pct := 0.0
		if total > 0 {
			pct = float64(n) / float64(total) * 100
		}
		item := TagDistributionItem{Name: tag, Value: n, Percentage: pct}
		if isShiftTag(tag) {
			shiftItems = append(shiftItems, item)
		} else {
			facilityItems = append(facilityItems, item)
		}
	}
	cats := []TagCategory{
		{Category: "shift", CategoryName: "负荷时段", Items: shiftItems},
		{Category: "facility", CategoryName: "行业类型", Items: facilityItems},
	}
	return &TagDistributionResponse{Categories: cats}, nil
}

// ─── 3. ScatterData：散点图（对齐前端 ScatterDataResponse {items}）───

type ScatterPoint struct {
	CustomerID   string   `json:"customer_id"`
	CustomerName string   `json:"customer_name"`
	AvgDailyLoad float64  `json:"avg_daily_load"`
	CV           float64  `json:"cv"`
	Regularity   *float64 `json:"regularity_score"`
	Tags         []string `json:"tags"`
}

// ScatterData 提取每客户最新特征的关键指标（avg_daily_load/cv 从 long_term JSONB 提取）。
func (r *LoadCharacteristicsExtRepository) ScatterData(ctx context.Context) ([]ScatterPoint, error) {
	q := `SELECT DISTINCT ON (cc.customer_id)
			cc.customer_id::text, cust.user_name,
			COALESCE((cc.long_term->>'avg_daily_load')::float8, 0),
			COALESCE((cc.long_term->>'cv')::float8, 0),
			cc.regularity_score,
			COALESCE(cc.tags, '{}')
		  FROM customer_characteristics cc
		  JOIN customers cust ON cust.id = cc.customer_id`
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE cc.org_id = $%d::uuid", len(args))
	}
	q += ` ORDER BY cc.customer_id, cc.data_date DESC LIMIT 500`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ScatterPoint
	for rows.Next() {
		var p ScatterPoint
		if err := rows.Scan(&p.CustomerID, &p.CustomerName, &p.AvgDailyLoad,
			&p.CV, &p.Regularity, &p.Tags); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ─── 4. Customers：分页列表（带 search/tag/sort）───

type LoadCharCustomer struct {
	ID              string          `json:"id"`
	CustomerName    string          `json:"customer_name"`
	DataDate        time.Time       `json:"data_date"`
	Tags            []string        `json:"tags"`
	RegularityScore *float64        `json:"regularity_score"`
	AvgDailyLoad    float64         `json:"avg_daily_load"`
	LongTerm        json.RawMessage `json:"long_term"`
	ShortTerm       json.RawMessage `json:"short_term"`
}

// ListCustomers 分页列出有特征的客户（支持 search/tag 筛选、sort 排序）。
func (r *LoadCharacteristicsExtRepository) ListCustomers(
	ctx context.Context, page, pageSize int, search, tag, sortBy, order string,
) ([]*LoadCharCustomer, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	// 白名单排序字段（防注入）
	sortCol := map[string]string{
		"avg_daily_load":   "(cc.long_term->>'avg_daily_load')::float8",
		"regularity_score": "cc.regularity_score",
		"data_date":        "cc.data_date",
		"customer_name":    "cust.user_name",
	}[sortBy]
	if sortCol == "" {
		sortCol = "(cc.long_term->>'avg_daily_load')::float8"
	}
	if order != "asc" {
		order = "desc"
	}

	// 子查询：每客户最新快照
	base := `FROM (
			SELECT DISTINCT ON (cc.customer_id)
				cc.id, cc.customer_id, cust.user_name, cc.data_date, cc.tags,
				cc.regularity_score, cc.long_term, cc.short_term
			FROM customer_characteristics cc
			JOIN customers cust ON cust.id = cc.customer_id`
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		base += fmt.Sprintf(" WHERE cc.org_id = $%d::uuid", len(args))
	}
	base += ` ORDER BY cc.customer_id, cc.data_date DESC
		) cc`
	idx := len(args) + 1

	// 外层筛选
	where := " WHERE 1=1"
	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND cc.user_name ILIKE $%d", idx)
		idx++
	}
	if tag != "" {
		args = append(args, tag)
		where += fmt.Sprintf(" AND $%d = ANY(cc.tags)", idx)
		idx++
	}

	// 总数
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) "+base+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 分页
	args = append(args, pageSize, (page-1)*pageSize)
	q := fmt.Sprintf(`SELECT cc.customer_id::text, cc.user_name, cc.data_date, cc.tags,
			cc.regularity_score, COALESCE((cc.long_term->>'avg_daily_load')::float8,0),
			cc.long_term, cc.short_term
		%s%s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		base, where, sortCol, order, idx, idx+1)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*LoadCharCustomer, 0, pageSize)
	for rows.Next() {
		var c LoadCharCustomer
		var lt, st []byte
		if err := rows.Scan(&c.ID, &c.CustomerName, &c.DataDate, &c.Tags,
			&c.RegularityScore, &c.AvgDailyLoad, &lt, &st); err != nil {
			return nil, 0, err
		}
		c.LongTerm = json.RawMessage(lt)
		c.ShortTerm = json.RawMessage(st)
		list = append(list, &c)
	}
	return list, total, rows.Err()
}

// ─── 5. CustomerDetail：单客户最新特征 ───

// GetCustomer 取单客户最新特征快照。
func (r *LoadCharacteristicsExtRepository) GetCustomer(ctx context.Context, customerID uuid.UUID) (*CustomerCharacteristic, error) {
	q := `SELECT cc.id, cc.customer_id, cust.user_name, cc.data_date,
			cc.long_term, cc.short_term, cc.tags, cc.regularity_score, cc.quality_rating
		  FROM customer_characteristics cc
		  JOIN customers cust ON cust.id = cc.customer_id
		  WHERE cc.customer_id = $1`
	args := []any{customerID}
	idx := 2
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND cc.org_id = $%d::uuid", idx)
		idx++
	}
	q += " ORDER BY cc.data_date DESC LIMIT 1"
	var c CustomerCharacteristic
	var lt, st []byte
	err := r.pool.QueryRow(ctx, q, args...).Scan(
		&c.ID, &c.CustomerID, &c.CustomerName, &c.DataDate,
		&lt, &st, &c.Tags, &c.RegularityScore, &c.QualityRating)
	if err != nil {
		return nil, err
	}
	c.LongTerm = json.RawMessage(lt)
	c.ShortTerm = json.RawMessage(st)
	return &c, nil
}

// ─── 6. CustomerHistory：分析执行历史 ───

type AnalysisHistoryItem struct {
	ID           uuid.UUID       `json:"id"`
	Date         time.Time       `json:"date"`
	Tags         []string        `json:"tags"`
	RuleIDs      []string        `json:"rule_ids"`
	Metrics      json.RawMessage `json:"metrics"`
	ExecutionTime *float64       `json:"execution_time,omitempty"`
	Operator     *string         `json:"operator,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// ListHistory 列出客户的分析执行历史（analysis_history_log 无 org_id，通过 customer JOIN 过滤）。
func (r *LoadCharacteristicsExtRepository) ListHistory(
	ctx context.Context, customerID uuid.UUID, limit int, month string,
) ([]*AnalysisHistoryItem, error) {
	if limit <= 0 || limit > 200 {
		limit = 30
	}
	q := `SELECT h.id, h.date, h.tags, h.rule_ids, h.metrics, h.execution_time, h.operator, h.created_at
		  FROM analysis_history_log h
		  JOIN customers c ON c.id = h.customer_id
		  WHERE h.customer_id = $1`
	args := []any{customerID, limit}
	idx := 3
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND c.org_id = $%d::uuid", idx)
		idx++
	}
	if month != "" {
		args = append(args, month+"%")
		q += fmt.Sprintf(" AND h.date::text LIKE $%d", idx)
		idx++
	}
	q += fmt.Sprintf(" ORDER BY h.date DESC LIMIT $2")
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*AnalysisHistoryItem
	for rows.Next() {
		var h AnalysisHistoryItem
		var metrics []byte
		if err := rows.Scan(&h.ID, &h.Date, &h.Tags, &h.RuleIDs, &metrics,
			&h.ExecutionTime, &h.Operator, &h.CreatedAt); err != nil {
			return nil, err
		}
		h.Metrics = json.RawMessage(metrics)
		out = append(out, &h)
	}
	return out, rows.Err()
}

// ─── 7. CustomerAlerts：客户告警列表 ───

// ListCustomerAlerts 列出某客户的告警（复用 customer_anomaly_alerts）。
func (r *LoadCharacteristicsExtRepository) ListCustomerAlerts(
	ctx context.Context, customerID uuid.UUID, limit int,
) ([]*CustomerAlert, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := alertSelect + ` WHERE a.customer_id = $1`
	args := []any{customerID, limit}
	idx := 3
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND a.org_id = $%d::uuid", idx)
		idx++
	}
	q += fmt.Sprintf(" ORDER BY a.alert_date DESC LIMIT $2")
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*CustomerAlert, 0, limit)
	for rows.Next() {
		var a CustomerAlert
		var metrics []byte
		if err := rows.Scan(&a.ID, &a.CustomerID, &a.CustomerName,
			&a.AlertDate, &a.AlertType, &a.Severity,
			&a.Confidence, &a.Reason, &metrics, &a.RuleID,
			&a.Acknowledged, &a.AckBy, &a.AckAt, &a.Note, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Metrics = json.RawMessage(metrics)
		list = append(list, &a)
	}
	return list, rows.Err()
}

// ─── 8. AckAlertWithNote：确认告警（带备注，扩 analytics 的 AckAlert）───

// AckAlertWithNote 确认告警并写入备注。
func (r *LoadCharacteristicsExtRepository) AckAlertWithNote(
	ctx context.Context, id, userID uuid.UUID, acknowledged bool, note string,
) error {
	if !acknowledged {
		// 取消确认
		q := `UPDATE customer_anomaly_alerts
			SET acknowledged = FALSE, acknowledged_by = NULL, acknowledged_at = NULL, note = NULLIF($3,'')
			WHERE id = $1`
		args := []any{id, userID, note}
		if org, scoped := OrgFilter(ctx); scoped {
			args = append(args, org)
			q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
		}
		_, err := r.pool.Exec(ctx, q, args...)
		return err
	}
	q := `UPDATE customer_anomaly_alerts
		SET acknowledged = TRUE, acknowledged_by = $2, acknowledged_at = NOW(), note = NULLIF($3,'')
		WHERE id = $1`
	args := []any{id, userID, note}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAlertNotFound
	}
	return nil
}

// ─── 9. DailyTrend：日趋势（复用 user_load_data 96 点曲线）───

type DailyTrendPoint struct {
	Date        string    `json:"date"`
	Curve       []float64 `json:"curve"`
	TotalLoad   float64   `json:"total_load"`
	QualityFlag string    `json:"quality_flag"`
}

// DailyTrend 取客户指定日期范围的日负荷曲线（user_load_data 的 curve_96）。
func (r *LoadCharacteristicsExtRepository) DailyTrend(
	ctx context.Context, customerID uuid.UUID, start, end time.Time,
) ([]*DailyTrendPoint, error) {
	q := `SELECT date::text, curve_96, total_load, quality_flag
		  FROM user_load_data
		  WHERE customer_id = $1 AND date >= $2 AND date <= $3`
	args := []any{customerID, start, end}
	idx := 4
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", idx)
		idx++
	}
	q += " ORDER BY date ASC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*DailyTrendPoint
	for rows.Next() {
		var p DailyTrendPoint
		if err := rows.Scan(&p.Date, &p.Curve, &p.TotalLoad, &p.QualityFlag); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	return out, rows.Err()
}

// ─── 10. MonthlyEnergy：月度电量（customer_monthly_energy 无 org_id，JOIN 过滤）───

type MonthlyEnergy struct {
	Month          string   `json:"month"`
	MonthlyEnergy  float64  `json:"monthly_energy"`
	AvgDailyEnergy float64  `json:"avg_daily_energy"`
	VariationCV    *float64 `json:"variation_cv,omitempty"`
}

// MonthlyEnergySeries 取客户月度电量序列。
func (r *LoadCharacteristicsExtRepository) MonthlyEnergySeries(
	ctx context.Context, customerID uuid.UUID, startMonth, endMonth string,
) ([]*MonthlyEnergy, error) {
	q := `SELECT e.month, e.monthly_energy, e.avg_daily_energy, e.variation_cv
		  FROM customer_monthly_energy e
		  JOIN customers c ON c.id = e.customer_id
		  WHERE e.customer_id = $1 AND e.month >= $2 AND e.month <= $3`
	args := []any{customerID, startMonth, endMonth}
	idx := 4
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND c.org_id = $%d::uuid", idx)
		idx++
	}
	q += " ORDER BY e.month ASC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*MonthlyEnergy
	for rows.Next() {
		var m MonthlyEnergy
		if err := rows.Scan(&m.Month, &m.MonthlyEnergy, &m.AvgDailyEnergy, &m.VariationCV); err != nil {
			return nil, err
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

// ─── Demo 种子：填充 analysis_history_log 与 customer_monthly_energy ───

// SeedHistory 为客户写入一条分析执行历史（让 /customer/:id/history 有数据）。
// analysis_history_log 无 org_id（通过 customer JOIN 过滤），但 customer 有 org_id。
func (r *LoadCharacteristicsExtRepository) SeedHistory(
	ctx context.Context, customerID uuid.UUID, date time.Time, tags, ruleIDs []string,
) error {
	q := `INSERT INTO analysis_history_log (customer_id, date, tags, rule_ids, metrics, execution_time, operator, created_at)
		VALUES ($1, $2, $3, $4,
			'{"total_points": 96, "quality": "good"}'::jsonb,
			$5, 'demo', now())`
	execTime := 0.5 + randFloat()*2.0
	_, err := r.pool.Exec(ctx, q, customerID, date, tags, ruleIDs, execTime)
	return err
}

// SeedMonthlyEnergy 为客户写入近 12 个月的月度电量（让 monthly-energy 端点有数据）。
func (r *LoadCharacteristicsExtRepository) SeedMonthlyEnergy(
	ctx context.Context, customerID uuid.UUID, baseEnergy float64,
) error {
	for i := 11; i >= 0; i-- {
		month := monthsAgoYM(i)
		// 加季节波动：夏冬高、春秋低
		seasonFactor := 0.85 + randFloat()*0.3
		energy := baseEnergy * seasonFactor
		avgDaily := energy / 30
		cv := 0.05 + randFloat()*0.1
		q := `INSERT INTO customer_monthly_energy (customer_id, month, monthly_energy, avg_daily_energy, variation_cv)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (customer_id, month) DO UPDATE SET
				monthly_energy = EXCLUDED.monthly_energy,
				avg_daily_energy = EXCLUDED.avg_daily_energy,
				variation_cv = EXCLUDED.variation_cv`
		if _, err := r.pool.Exec(ctx, q, customerID, month, energy, avgDaily, cv); err != nil {
			return err
		}
	}
	return nil
}

// randFloat 进程内简单随机（math/rand/v2 顶层函数已自动随机种子，无需手动 NewSource）。
func randFloat() float64 { return rand.Float64() }

// ─── 11. TagChanges：标签变化（相邻快照 diff）───

type TagChangeItem struct {
	CustomerID   string   `json:"customer_id"`
	CustomerName string   `json:"customer_name"`
	Date         string   `json:"date"`
	AddedTags    []string `json:"added_tags"`
	RemovedTags  []string `json:"removed_tags"`
}

// TagChanges 取指定日期与前一快照相比的标签变化。
func (r *LoadCharacteristicsExtRepository) TagChanges(ctx context.Context, date string) ([]*TagChangeItem, error) {
	// 取 date 当天与每客户前一快照，用 LAG 对比 tags
	q := `WITH ranked AS (
			SELECT cc.customer_id, cust.user_name, cc.data_date::date::text, cc.tags,
				LAG(cc.tags) OVER (PARTITION BY cc.customer_id ORDER BY cc.data_date) AS prev_tags
			FROM customer_characteristics cc
			JOIN customers cust ON cust.id = cc.customer_id`
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE cc.org_id = $%d::uuid", len(args))
	}
	q += `)`
	if date != "" {
		args = append(args, date)
		q += fmt.Sprintf(` WHERE data_date = $%d`, len(args))
	}
	q += ` LIMIT 200`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*TagChangeItem
	for rows.Next() {
		var c TagChangeItem
		var tags, prevTags []string
		var customerID uuid.UUID
		if err := rows.Scan(&customerID, &c.CustomerName, &c.Date, &tags, &prevTags); err != nil {
			return nil, err
		}
		c.CustomerID = customerID.String()
		// 计算 added/removed
		prevSet := map[string]bool{}
		for _, t := range prevTags {
			prevSet[t] = true
		}
		curSet := map[string]bool{}
		for _, t := range tags {
			curSet[t] = true
			if !prevSet[t] {
				c.AddedTags = append(c.AddedTags, t)
			}
		}
		for _, t := range prevTags {
			if !curSet[t] {
				c.RemovedTags = append(c.RemovedTags, t)
			}
		}
		if len(c.AddedTags) > 0 || len(c.RemovedTags) > 0 {
			out = append(out, &c)
		}
	}
	return out, rows.Err()
}
