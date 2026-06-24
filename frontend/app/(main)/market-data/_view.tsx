'use client';

import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  ComposedChart,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  getMarketDataOverview,
  getMarketDataTables,
  queryMarketData,
  type MarketDataTableInfo,
} from '@/lib/api/market-data';
import {
  TrendingUp,
  Fuel,
  Landmark,
  BarChart3,
  Activity,
  ArrowUpDown,
} from 'lucide-react';

/* ── 分类配置 ── */
const CATEGORY_META: Record<
  string,
  { label: string; icon: typeof TrendingUp; color: string }
> = {
  macro: { label: '宏观经济', icon: Landmark, color: 'text-blue-600' },
  fuel: { label: '能源燃料', icon: Fuel, color: 'text-orange-500' },
  futures: { label: '商品期货', icon: BarChart3, color: 'text-purple-600' },
  rate: { label: '利率', icon: Activity, color: 'text-green-600' },
  fx: { label: '汇率', icon: ArrowUpDown, color: 'text-cyan-600' },
  bond: { label: '债券', icon: Landmark, color: 'text-yellow-600' },
  index: { label: '指数', icon: TrendingUp, color: 'text-pink-600' },
};

const CATEGORY_ORDER = [
  'macro',
  'fuel',
  'futures',
  'rate',
  'fx',
  'bond',
  'index',
];

/* ── 日期格式 ── */
function fmtDate(v: unknown): string {
  if (!v) return '-';
  const s = String(v);
  // "2024-01-15T00:00:00Z" or "2024-01-15T00:00:00+08:00"
  return s.slice(0, 10);
}

function fmtNum(v: unknown, decimals = 2): string {
  if (v == null) return '-';
  const n = Number(v);
  if (isNaN(n)) return '-';
  return n.toLocaleString('zh-CN', {
    minimumFractionDigits: 0,
    maximumFractionDigits: decimals,
  });
}

/* ── 图表展示的字段映射 ── */
interface ChartField {
  key: string;
  label: string;
  color: string;
}

function getChartFields(
  table: string,
  sample: Record<string, unknown>,
): ChartField[] {
  // 通用 OHLC 表 → 收盘价
  if (
    ['close_price', 'close'].some(
      (k) => sample[k] != null,
    )
  ) {
    return [{ key: 'close_price', label: '收盘价', color: '#3b82f6' }];
  }
  // 宏观类：找 value / yoy 字段
  const valueKeys = [
    'value',
    'yoy_growth',
    'gdp_yoy',
    'cpi_yoy',
    'ppi_yoy',
    'pmi_value',
    'total_usage',
    'm2_yoy',
    'm2_balance',
  ];
  for (const k of valueKeys) {
    if (sample[k] != null) {
      return [{ key: k, label: k, color: '#3b82f6' }];
    }
  }
  // Shibor → overnight + week_1 + month_1
  if (sample['overnight'] != null) {
    return [
      { key: 'overnight', label: '隔夜', color: '#3b82f6' },
      { key: 'week_1', label: '1周', color: '#f97316' },
      { key: 'month_1', label: '1月', color: '#22c55e' },
    ];
  }
  // LPR
  if (sample['lpr_1y'] != null) {
    return [
      { key: 'lpr_1y', label: '1年期', color: '#3b82f6' },
      { key: 'lpr_5y', label: '5年期', color: '#f97316' },
    ];
  }
  // 国债收益率
  if (sample['cn_10y'] != null) {
    return [
      { key: 'cn_10y', label: '中国10Y', color: '#ef4444' },
      { key: 'us_10y', label: '美国10Y', color: '#3b82f6' },
    ];
  }
  // 汽柴油
  if (sample['gasoline_price'] != null) {
    return [
      { key: 'gasoline_price', label: '汽油', color: '#f97316' },
      { key: 'diesel_price', label: '柴油', color: '#64748b' },
    ];
  }
  // 风速
  if (sample['wind_speed_100m'] != null) {
    return [{ key: 'wind_speed_100m', label: '100m风速(m/s)', color: '#3b82f6' }];
  }
  // 水文
  if (sample['precipitation_sum'] != null) {
    return [
      { key: 'precipitation_sum', label: '降水(mm)', color: '#3b82f6' },
      { key: 'et0_evapotranspiration', label: '蒸发(mm)', color: '#f97316' },
    ];
  }
  // fallback
  const numKeys = Object.entries(sample)
    .filter(([, v]) => typeof v === 'number')
    .map(([k]) => k)
    .filter(
      (k) =>
        !['id', 'fetched_at'].includes(k) && !k.includes('date') && !k.includes('time'),
    );
  if (numKeys.length > 0) {
    return [{ key: numKeys[0], label: numKeys[0], color: '#3b82f6' }];
  }
  return [];
}

function getDateKey(sample: Record<string, unknown>): string {
  for (const k of [
    'trade_date',
    'stat_date',
    'obs_time',
    'obs_date',
    'adjust_date',
  ]) {
    if (sample[k] != null) return k;
  }
  return Object.keys(sample).find((k) => k.includes('date') || k.includes('time')) || '';
}

/* ── 表格列 ── */
function getTableColumns(
  table: string,
  sample: Record<string, unknown>,
): { key: string; label: string; fmt: (v: unknown) => string }[] {
  const cols: { key: string; label: string; fmt: (v: unknown) => string }[] = [];
  const dateKey = getDateKey(sample);

  // 日期列
  if (dateKey) {
    cols.push({ key: dateKey, label: '日期', fmt: fmtDate });
  }

  // 站点名
  if (sample['location_name'] != null) {
    cols.push({ key: 'location_name', label: '站点', fmt: (v) => (v ? String(v) : '-') });
  }

  // OHLCV
  const priceFields = [
    { key: 'open_price', label: '开盘' },
    { key: 'high_price', label: '最高' },
    { key: 'low_price', label: '最低' },
    { key: 'close_price', label: '收盘' },
    { key: 'volume', label: '成交量' },
  ];
  for (const f of priceFields) {
    if (sample[f.key] != null) {
      cols.push({ key: f.key, label: f.label, fmt: fmtNum });
    }
  }

  // 宏观 value / yoy / cum
  const macroFields = [
    { key: 'value', label: '数值' },
    { key: 'yoy_growth', label: '同比(%)' },
    { key: 'cum_growth', label: '累计(%)' },
    { key: 'gdp_yoy', label: 'GDP同比(%)' },
    { key: 'cpi_yoy', label: 'CPI同比(%)' },
    { key: 'cpi_mom', label: 'CPI环比(%)' },
    { key: 'ppi_yoy', label: 'PPI同比(%)' },
    { key: 'ppi_mom', label: 'PPI环比(%)' },
    { key: 'pmi_value', label: 'PMI' },
    { key: 'total_usage', label: '用电量(亿kWh)' },
    { key: 'm2_yoy', label: 'M2同比(%)' },
    { key: 'm2_balance', label: 'M2余额(万亿)' },
  ];
  for (const f of macroFields) {
    if (sample[f.key] != null) {
      cols.push({ key: f.key, label: f.label, fmt: fmtNum });
    }
  }

  // 利率
  const rateFields = [
    { key: 'overnight', label: '隔夜(%)' },
    { key: 'week_1', label: '1周(%)' },
    { key: 'month_1', label: '1月(%)' },
    { key: 'month_3', label: '3月(%)' },
    { key: 'month_6', label: '6月(%)' },
    { key: 'year_1', label: '1年(%)' },
    { key: 'lpr_1y', label: 'LPR 1Y(%)' },
    { key: 'lpr_5y', label: 'LPR 5Y(%)' },
  ];
  for (const f of rateFields) {
    if (sample[f.key] != null) {
      cols.push({ key: f.key, label: f.label, fmt: fmtNum });
    }
  }

  // 国债
  const bondFields = [
    { key: 'cn_2y', label: '中2Y(%)' },
    { key: 'cn_5y', label: '中5Y(%)' },
    { key: 'cn_10y', label: '中10Y(%)' },
    { key: 'cn_30y', label: '中30Y(%)' },
    { key: 'us_2y', label: '美2Y(%)' },
    { key: 'us_5y', label: '美5Y(%)' },
    { key: 'us_10y', label: '美10Y(%)' },
    { key: 'us_30y', label: '美30Y(%)' },
  ];
  for (const f of bondFields) {
    if (sample[f.key] != null) {
      cols.push({ key: f.key, label: f.label, fmt: fmtNum });
    }
  }

  // 汽柴油
  if (sample['gasoline_price'] != null) {
    cols.push({ key: 'gasoline_price', label: '汽油(元/吨)', fmt: fmtNum });
    cols.push({ key: 'diesel_price', label: '柴油(元/吨)', fmt: fmtNum });
  }

  // 风速
  const windFields = [
    { key: 'wind_speed_100m', label: '100m风速(m/s)' },
    { key: 'wind_dir_100m', label: '风向(°)' },
    { key: 'temperature_2m', label: '气温(°C)' },
    { key: 'humidity_2m', label: '湿度(%)' },
  ];
  for (const f of windFields) {
    if (sample[f.key] != null) {
      cols.push({ key: f.key, label: f.label, fmt: fmtNum });
    }
  }

  // 水文
  const hydroFields = [
    { key: 'temp_mean', label: '均温(°C)' },
    { key: 'precipitation_sum', label: '降水(mm)' },
    { key: 'et0_evapotranspiration', label: '蒸发(mm)' },
    { key: 'wind_speed_10m_mean', label: '风速(m/s)' },
  ];
  for (const f of hydroFields) {
    if (sample[f.key] != null) {
      cols.push({ key: f.key, label: f.label, fmt: fmtNum });
    }
  }

  return cols;
}

/* ── 时间范围选择 ── */
const DAY_OPTIONS = [
  { value: 7, label: '7天' },
  { value: 30, label: '30天' },
  { value: 90, label: '90天' },
  { value: 180, label: '半年' },
  { value: 365, label: '1年' },
  { value: 3650, label: '全部' },
];

export default function MarketDataPage() {
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [selectedTable, setSelectedTable] = useState<string | null>(null);
  const [days, setDays] = useState(30);

  // ── 概览 ──
  const { data: overview } = useQuery({
    queryKey: ['market-data', 'overview'],
    queryFn: async () => {
      const r = await getMarketDataOverview();
      return r.data;
    },
  });

  // ── 表列表 ──
  const { data: tablesResp } = useQuery({
    queryKey: ['market-data', 'tables'],
    queryFn: async () => {
      const r = await getMarketDataTables();
      return r.data;
    },
  });

  const tables = useMemo(() => tablesResp?.tables ?? [], [tablesResp]);

  const filteredTables = useMemo(() => {
    if (!selectedCategory) return tables;
    return tables.filter((t) => t.category === selectedCategory);
  }, [tables, selectedCategory]);

  // ── 选中表数据 ──
  const { data: queryResult, isLoading: queryLoading } = useQuery({
    queryKey: ['market-data', 'query', selectedTable, days],
    queryFn: async () => {
      if (!selectedTable) return null;
      const r = await queryMarketData(selectedTable, { days });
      return r.data;
    },
    enabled: !!selectedTable,
  });

  const selectedTableMeta = useMemo(
    () => tables.find((t) => t.table_name === selectedTable),
    [tables, selectedTable],
  );

  // ── 图表数据 ──
  const chartData = useMemo(() => {
    if (!queryResult?.data?.length) return [];
    const dateKey = getDateKey(queryResult.data[0]);
    const fields = getChartFields(selectedTable!, queryResult.data[0]);
    return queryResult.data
      .map((row) => {
        const point: Record<string, unknown> = {
          date: fmtDate(row[dateKey]),
        };
        for (const f of fields) {
          point[f.label] = row[f.key] != null ? Number(row[f.key]) : null;
        }
        return point;
      })
      .reverse(); // 时间正序
  }, [queryResult, selectedTable]);

  const chartFields = useMemo(
    () =>
      queryResult?.data?.length
        ? getChartFields(selectedTable!, queryResult.data[0])
        : [],
    [queryResult, selectedTable],
  );

  const tableColumns = useMemo(
    () =>
      queryResult?.data?.length
        ? getTableColumns(selectedTable!, queryResult.data[0])
        : [],
    [queryResult, selectedTable],
  );

  return (
    <div className="space-y-6 p-6">
      {/* 标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">市场行情</h1>
          <p className="text-sm text-muted-foreground">
            涵盖宏观经济、能源燃料、商品期货、利率汇率、债券指数等 {overview?.total_tables ?? 30} 类数据
          </p>
        </div>
      </div>

      {/* 分类卡片 */}
      <div className="grid grid-cols-3 gap-3 sm:grid-cols-5 lg:grid-cols-9">
        <button
          key="__all"
          onClick={() => {
            setSelectedCategory(null);
            setSelectedTable(null);
          }}
          className={`rounded-lg border p-3 text-center transition-colors ${
            !selectedCategory
              ? 'border-blue-400 bg-blue-50 shadow-sm'
              : 'border-border hover:bg-muted/50'
          }`}
        >
          <div className="text-lg font-semibold">{overview?.total_tables ?? 30}</div>
          <div className="text-xs text-muted-foreground">全部</div>
        </button>
        {CATEGORY_ORDER.map((cat) => {
          const meta = CATEGORY_META[cat];
          const catInfo = overview?.categories?.[cat];
          const Icon = meta.icon;
          const active = selectedCategory === cat;
          return (
            <button
              key={cat}
              onClick={() => {
                setSelectedCategory(active ? null : cat);
                setSelectedTable(null);
              }}
              className={`rounded-lg border p-3 text-center transition-colors ${
                active
                  ? 'border-blue-400 bg-blue-50 shadow-sm'
                  : 'border-border hover:bg-muted/50'
              }`}
            >
              <Icon className={`mx-auto h-4 w-4 ${meta.color}`} />
              <div className="mt-1 text-sm font-medium">{meta.label}</div>
              <div className="text-xs text-muted-foreground">
                {catInfo ? `${catInfo.tables}表 · ${(catInfo.count / 1000).toFixed(1)}k` : '-'}
              </div>
            </button>
          );
        })}
      </div>

      {/* 表列表 */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">
            {selectedCategory
              ? `${CATEGORY_META[selectedCategory]?.label ?? selectedCategory} · 数据表`
              : '全部数据表'}
            <span className="ml-2 text-sm font-normal text-muted-foreground">
              ({filteredTables.length})
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
            {filteredTables.map((t) => {
              const active = selectedTable === t.table_name;
              return (
                <button
                  key={t.table_name}
                  onClick={() => setSelectedTable(active ? null : t.table_name)}
                  className={`flex flex-col items-start rounded-lg border p-3 text-left transition-colors ${
                    active
                      ? 'border-blue-400 bg-blue-50'
                      : 'border-border hover:bg-muted/50'
                  }`}
                >
                  <div className="flex w-full items-center justify-between">
                    <span className="text-sm font-medium">{t.label}</span>
                    {(t.scope === 'national' || !t.scope) && (
                      <span className="text-[10px] text-blue-500" title="全国数据">🇨🇳</span>
                    )}
                    {t.scope === 'provincial' && (
                      <span className="text-[10px] text-orange-500" title="分省数据">📍</span>
                    )}
                  </div>
                  <span className="text-xs text-muted-foreground">
                    {t.row_count.toLocaleString()}条
                  </span>
                  <span className="text-xs text-muted-foreground">{t.date_range}</span>
                </button>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* 选中表详情 */}
      {selectedTable && selectedTableMeta && (
        <div className="space-y-4">
          {/* 标题 + 时间选择 */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <h2 className="text-lg font-semibold">{selectedTableMeta.label}</h2>
              <Badge variant="outline" className="text-xs">
                {selectedTableMeta.row_count.toLocaleString()}条
              </Badge>
              {(queryResult?.scope_label || selectedTableMeta.scope === 'national' || !selectedTableMeta.scope) ? (
                <Badge variant="info" className="text-xs">
                  {queryResult?.scope_label || '全国数据'}
                </Badge>
              ) : selectedTableMeta.scope === 'provincial' ? (
                <Badge variant="warning" className="text-xs">
                  {queryResult?.scope_label || '分省数据'}
                </Badge>
              ) : null}
            </div>
            <div className="flex gap-1">
              {DAY_OPTIONS.map((opt) => (
                <button
                  key={opt.value}
                  onClick={() => setDays(opt.value)}
                  className={`rounded-md px-2.5 py-1 text-xs transition-colors ${
                    days === opt.value
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-muted-foreground hover:bg-muted'
                  }`}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </div>

          {queryLoading ? (
            <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">
              加载中…
            </div>
          ) : (
            <>
              {/* 图表 */}
              {chartData.length > 0 && chartFields.length > 0 && (
                <Card>
                  <CardContent className="pt-6">
                    <ResponsiveContainer width="100%" height={320}>
                      {chartFields.length === 1 ? (
                        <AreaChart data={chartData}>
                          <defs>
                            <linearGradient
                              id="grad"
                              x1="0"
                              y1="0"
                              x2="0"
                              y2="1"
                            >
                              <stop offset="5%" stopColor={chartFields[0].color} stopOpacity={0.15} />
                              <stop offset="95%" stopColor={chartFields[0].color} stopOpacity={0} />
                            </linearGradient>
                          </defs>
                          <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
                          <XAxis
                            dataKey="date"
                            tick={{ fontSize: 11 }}
                            interval="preserveStartEnd"
                          />
                          <YAxis tick={{ fontSize: 11 }} width={60} />
                          <Tooltip
                            contentStyle={{ fontSize: 12 }}
                            formatter={(val: number) => fmtNum(val)}
                          />
                          <Area
                            type="monotone"
                            dataKey={chartFields[0].label}
                            stroke={chartFields[0].color}
                            fill="url(#grad)"
                            strokeWidth={1.5}
                            dot={false}
                          />
                        </AreaChart>
                      ) : (
                        <LineChart data={chartData}>
                          <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
                          <XAxis
                            dataKey="date"
                            tick={{ fontSize: 11 }}
                            interval="preserveStartEnd"
                          />
                          <YAxis tick={{ fontSize: 11 }} width={60} />
                          <Tooltip
                            contentStyle={{ fontSize: 12 }}
                            formatter={(val: number) => fmtNum(val)}
                          />
                          <Legend wrapperStyle={{ fontSize: 12 }} />
                          {chartFields.map((f) => (
                            <Line
                              key={f.key}
                              type="monotone"
                              dataKey={f.label}
                              stroke={f.color}
                              strokeWidth={1.5}
                              dot={false}
                            />
                          ))}
                        </LineChart>
                      )}
                    </ResponsiveContainer>
                  </CardContent>
                </Card>
              )}

              {/* 数据表格 */}
              {queryResult && queryResult.data.length > 0 && (
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm">
                      数据明细
                      <span className="ml-1 font-normal text-muted-foreground">
                        (最近 {queryResult.data.length} 条)
                      </span>
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="max-h-[420px] overflow-auto">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            {tableColumns.map((col) => (
                              <TableHead key={col.key} className="whitespace-nowrap text-xs">
                                {col.label}
                              </TableHead>
                            ))}
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {queryResult.data.slice(0, 100).map((row, i) => (
                            <TableRow key={i}>
                              {tableColumns.map((col) => (
                                <TableCell key={col.key} className="whitespace-nowrap text-xs">
                                  {col.fmt(row[col.key])}
                                </TableCell>
                              ))}
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                  </CardContent>
                </Card>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
}
