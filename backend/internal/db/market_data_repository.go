package db

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MarketDataRow 通用市场数据行（各表字段取并集，不存在的字段返回 null）
type MarketDataRow struct {
	TradeDate     *time.Time `json:"trade_date,omitempty"`
	StatDate      *time.Time `json:"stat_date,omitempty"`
	ObsTime       *time.Time `json:"obs_time,omitempty"`
	ObsDate       *time.Time `json:"obs_date,omitempty"`
	LocationCode  *string    `json:"location_code,omitempty"`
	LocationName  *string    `json:"location_name,omitempty"`
	OpenPrice     *float64   `json:"open_price,omitempty"`
	HighPrice     *float64   `json:"high_price,omitempty"`
	LowPrice      *float64   `json:"low_price,omitempty"`
	ClosePrice    *float64   `json:"close_price,omitempty"`
	Volume        *float64   `json:"volume,omitempty"`
	Value         *float64   `json:"value,omitempty"`
	YoyGrowth     *float64   `json:"yoy_growth,omitempty"`
	CumGrowth     *float64   `json:"cum_growth,omitempty"`
	PmiValue      *float64   `json:"pmi_value,omitempty"`
	CpiYoy        *float64   `json:"cpi_yoy,omitempty"`
	CpiMom        *float64   `json:"cpi_mom,omitempty"`
	PpiYoy        *float64   `json:"ppi_yoy,omitempty"`
	PpiMom        *float64   `json:"ppi_mom,omitempty"`
	GdpYoy        *float64   `json:"gdp_yoy,omitempty"`
	TotalUsage    *float64   `json:"total_usage,omitempty"`
	M2Yoy         *float64   `json:"m2_yoy,omitempty"`
	M2Balance     *float64   `json:"m2_balance,omitempty"`
	Lpr1y         *float64   `json:"lpr_1y,omitempty"`
	Lpr5y         *float64   `json:"lpr_5y,omitempty"`
	Overnight     *float64   `json:"overnight,omitempty"`
	Week1         *float64   `json:"week_1,omitempty"`
	Month1        *float64   `json:"month_1,omitempty"`
	Month3        *float64   `json:"month_3,omitempty"`
	Month6        *float64   `json:"month_6,omitempty"`
	Year1         *float64   `json:"year_1,omitempty"`
	Cn2y          *float64   `json:"cn_2y,omitempty"`
	Cn5y          *float64   `json:"cn_5y,omitempty"`
	Cn10y         *float64   `json:"cn_10y,omitempty"`
	Cn30y         *float64   `json:"cn_30y,omitempty"`
	Us2y          *float64   `json:"us_2y,omitempty"`
	Us5y          *float64   `json:"us_5y,omitempty"`
	Us10y         *float64   `json:"us_10y,omitempty"`
	Us30y         *float64   `json:"us_30y,omitempty"`
	BdiValue      *float64   `json:"bdi_value,omitempty"`
	GasolinePrice *float64   `json:"gasoline_price,omitempty"`
	DieselPrice   *float64   `json:"diesel_price,omitempty"`
	WindSpeed100m *float64   `json:"wind_speed_100m,omitempty"`
	WindDir100m   *int16     `json:"wind_dir_100m,omitempty"`
	Temperature2m *float64   `json:"temperature_2m,omitempty"`
	Humidity2m    *float64   `json:"humidity_2m,omitempty"`
	TempMean      *float64   `json:"temp_mean,omitempty"`
	HumidityMean  *float64   `json:"humidity_mean,omitempty"`
	Precipitation *float64   `json:"precipitation_sum,omitempty"`
	RainSum       *float64   `json:"rain_sum,omitempty"`
	Et0           *float64   `json:"et0_evapotranspiration,omitempty"`
	WindSpeed10m  *float64   `json:"wind_speed_10m_mean,omitempty"`
}

// MarketDataTableInfo 表元信息
type MarketDataTableInfo struct {
	TableName string `json:"table_name"`
	Label     string `json:"label"`
	Category  string `json:"category"`
	Scope     string `json:"scope"` // "national" | "provincial"
	RowCount  int    `json:"row_count"`
	DateRange string `json:"date_range,omitempty"`
}

// MarketDataRepository 通用市场数据查询
type MarketDataRepository struct{ pool *Pool }

func NewMarketDataRepository(pool *Pool) *MarketDataRepository {
	return &MarketDataRepository{pool: pool}
}

// GetTableMeta 返回表的元信息（scope/label 等），不存在返回 ok=false。
func (r *MarketDataRepository) GetTableMeta(tableName string) (struct {
	Label    string
	Category string
	DateCol  string
	Scope    string
}, bool) {
	meta, ok := marketDataTables[tableName]
	return meta, ok
}

// 表白名单：允许查询的表（防止SQL注入）
// Scope: "national"=全国数据（皆可见）；"provincial"=按省份过滤
var marketDataTables = map[string]struct {
	Label    string
	Category string
	DateCol  string
	Scope    string
}{
	"md_macro_gdp":               {"GDP", "macro", "stat_date", "national"},
	"md_macro_cpi":               {"CPI", "macro", "stat_date", "national"},
	"md_macro_ppi":               {"PPI", "macro", "stat_date", "national"},
	"md_macro_pmi":               {"PMI", "macro", "stat_date", "national"},
	"md_macro_electricity":       {"全社会用电量", "macro", "stat_date", "national"},
	"md_macro_m2":                {"M2货币供应", "macro", "stat_date", "national"},
	"md_macro_industrial_output": {"工业增加值", "macro", "stat_date", "national"},
	"md_fuel_wti":                {"WTI原油", "fuel", "trade_date", "national"},
	"md_fuel_natgas_hh":          {"天然气HH", "fuel", "trade_date", "national"},
	"md_fuel_ine_crude":          {"INE原油", "fuel", "trade_date", "national"},
	"md_fuel_cn_oil_price":       {"国内汽柴油", "fuel", "adjust_date", "national"},
	"md_futures_rb":              {"螺纹钢", "futures", "trade_date", "national"},
	"md_futures_i":               {"铁矿石", "futures", "trade_date", "national"},
	"md_futures_al":              {"沪铝", "futures", "trade_date", "national"},
	"md_futures_au":              {"沪金", "futures", "trade_date", "national"},
	"md_futures_cu":              {"沪铜", "futures", "trade_date", "national"},
	"md_futures_zn":              {"沪锌", "futures", "trade_date", "national"},
	"md_futures_hc":              {"热卷", "futures", "trade_date", "national"},
	"md_futures_fg":              {"玻璃", "futures", "trade_date", "national"},
	"md_futures_sa":              {"纯碱", "futures", "trade_date", "national"},
	"md_futures_zc":              {"动力煤", "futures", "trade_date", "national"},
	"md_rate_shibor":             {"Shibor", "rate", "trade_date", "national"},
	"md_rate_lpr":                {"LPR", "rate", "stat_date", "national"},
	"md_fx_usdcny":               {"USD/CNY", "fx", "trade_date", "national"},
	"md_bond_zh_us_yield":        {"中美国债收益率", "bond", "trade_date", "national"},
	"md_index_dxy":               {"美元指数", "index", "trade_date", "national"},
	"md_index_bdi":               {"BDI波罗的海指数", "index", "trade_date", "national"},
}

// mdStat 单表统计：行数 + 日期范围。
type mdStat struct {
	count     int
	dateRange string
}

// tableStats 用单条 UNION ALL 一次往返取回全部表的行数与日期范围，
// 替代原先「每表 1~2 次查询」的 N+1（30 表 → 60 次往返收敛为 1 次）。
// 表名与日期列均来自 marketDataTables 白名单，拼接安全。
func (r *MarketDataRepository) tableStats(ctx context.Context) (map[string]mdStat, error) {
	parts := make([]string, 0, len(marketDataTables))
	for tbl, meta := range marketDataTables {
		parts = append(parts, fmt.Sprintf(
			"SELECT '%s' AS t, COUNT(*) AS c, to_char(MIN(%s),'YYYY-MM-DD') AS mn, to_char(MAX(%s),'YYYY-MM-DD') AS mx FROM %s",
			tbl, meta.DateCol, meta.DateCol, tbl,
		))
	}
	rows, err := r.pool.Query(ctx, strings.Join(parts, "\nUNION ALL\n"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]mdStat, len(marketDataTables))
	for rows.Next() {
		var t string
		var c int
		var mn, mx *string
		if err := rows.Scan(&t, &c, &mn, &mx); err != nil {
			return nil, err
		}
		dr := ""
		if mn != nil && mx != nil {
			dr = *mn + " ~ " + *mx
		}
		out[t] = mdStat{count: c, dateRange: dr}
	}
	return out, rows.Err()
}

// ListTables 返回所有市场数据表概览
func (r *MarketDataRepository) ListTables(ctx context.Context) ([]*MarketDataTableInfo, error) {
	stats, err := r.tableStats(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*MarketDataTableInfo, 0, len(marketDataTables))
	for tbl, meta := range marketDataTables {
		s := stats[tbl]
		scope := meta.Scope
		if scope == "" {
			scope = "national"
		}
		list = append(list, &MarketDataTableInfo{
			TableName: tbl,
			Label:     meta.Label,
			Category:  meta.Category,
			Scope:     scope,
			RowCount:  s.count,
			DateRange: s.dateRange,
		})
	}
	return list, nil
}

// QueryTable 查询指定表的数据（白名单校验）
func (r *MarketDataRepository) QueryTable(ctx context.Context, tableName string, days int, locationCode string) ([]map[string]interface{}, error) {
	meta, ok := marketDataTables[tableName]
	if !ok {
		return nil, fmt.Errorf("未知的表: %s", tableName)
	}

	if days <= 0 || days > 3650 {
		days = 30
	}

	// 构建查询
	args := []interface{}{}
	conditions := []string{}

	// 时间过滤
	since := time.Now().AddDate(0, 0, -days)
	conditions = append(conditions, fmt.Sprintf("%s >= $%d", meta.DateCol, len(args)+1))
	args = append(args, since)

	// 风速/水文支持按站点过滤
	if (tableName == "md_weather_wind_hourly" || tableName == "md_weather_hydrology_daily") && locationCode != "" {
		conditions = append(conditions, fmt.Sprintf("location_code = $%d", len(args)+1))
		args = append(args, locationCode)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	sql := fmt.Sprintf("SELECT * FROM %s %s ORDER BY %s DESC LIMIT 5000", tableName, where, meta.DateCol)
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]map[string]interface{}, 0)
	fieldDescriptions := rows.FieldDescriptions()
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make(map[string]interface{})
		for i, fd := range fieldDescriptions {
			row[fd.Name] = values[i]
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// Overview 按分类汇总统计
func (r *MarketDataRepository) Overview(ctx context.Context) (map[string]interface{}, error) {
	type catStat struct {
		Count       int      `json:"count"`
		Tables      int      `json:"tables"`
		Tables_list []string `json:"table_list"`
	}
	stats, err := r.tableStats(ctx)
	if err != nil {
		return nil, err
	}
	categories := map[string]*catStat{}
	for tbl, meta := range marketDataTables {
		if _, ok := categories[meta.Category]; !ok {
			categories[meta.Category] = &catStat{Tables_list: []string{}}
		}
		categories[meta.Category].Tables++
		categories[meta.Category].Tables_list = append(categories[meta.Category].Tables_list, tbl)
		categories[meta.Category].Count += stats[tbl].count
	}

	result := map[string]interface{}{
		"total_tables": len(marketDataTables),
		"categories":   categories,
	}
	return result, nil
}
