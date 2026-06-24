'use client';

import { useState, useMemo, useCallback } from 'react';
import { format, subDays } from 'date-fns';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ComposedChart,
  Line,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
  Legend,
  ReferenceArea,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  listDeviations,
  getDeviationSummary,
  generateDeviationDemoData,
} from '@/lib/api/deviation-settlement';
import { StatCard } from '@/components/data-display/stat-card';
import { ChartContainer } from '@/components/charts/chart-container';
import { CustomTooltip } from '@/components/charts/custom-tooltip';
import { DataTable, type DataTableColumn, type DataRow } from '@/components/data-display/data-table';
import { PageHeader } from '@/components/data-display/page-header';
import {
  AlertCircle,
  RefreshCw,
  TrendingUp,
  TrendingDown,
  Zap,
  DollarSign,
  BarChart3,
  PieChart as PieChartIcon,
  Activity,
  Search,
  ShieldAlert,
  CalendarDays,
} from 'lucide-react';

// ─── 常量 ───

/** 免赔区间阈值 (MWh) */
const DEDUCTIBLE_ZONE = 2;
/** 偏差率预警阈值 (%) */
const WARNING_RATE_THRESHOLD = 5;

// ─── 辅助函数 ───

function fmtMoney(v?: number | null): string {
  if (v == null) return '-';
  return v.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function fmt(v: number | null | undefined, digits = 2): string {
  if (v == null) return '-';
  return v.toLocaleString('zh-CN', { maximumFractionDigits: digits });
}

/** 类别 -> 中文名 */
const CATEGORY_LABELS: Record<string, string> = {
  day_ahead: '日前偏差',
  real_time: '实时偏差',
  intraday: '日内偏差',
};

/** 类别 -> 颜色 */
const CATEGORY_COLORS: Record<string, string> = {
  day_ahead: '#2563eb',
  real_time: '#f59e0b',
  intraday: '#10b981',
};

function catLabel(cat: string): string {
  return CATEGORY_LABELS[cat] || cat;
}

function catColor(cat: string): string {
  return CATEGORY_COLORS[cat] || '#6b7280';
}

// ─── 页面组件 ───

export default function DeviationManagementPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('settlement_management:write');

  // 筛选条件
  const [days, setDays] = useState(30);
  const [categoryFilter, setCategoryFilter] = useState('');
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  // ── 数据查询 ──

  // 偏差列表
  const {
    data: listData,
    isLoading: listLoading,
  } = useQuery({
    queryKey: ['deviation-list', days, categoryFilter],
    queryFn: () => listDeviations(days, categoryFilter),
  });

  // 偏差汇总
  const {
    data: summaryData,
    isLoading: summaryLoading,
  } = useQuery({
    queryKey: ['deviation-summary', days],
    queryFn: () => getDeviationSummary(days),
  });

  // ── 操作 ──

  const onGenerate = useCallback(async () => {
    setError(null);
    setNotice(null);
    setGenerating(true);
    try {
      const r = await generateDeviationDemoData();
      setNotice(`已生成 ${r.rows} 条偏差结算演示数据`);
      qc.invalidateQueries({ queryKey: ['deviation-list'] });
      qc.invalidateQueries({ queryKey: ['deviation-summary'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setGenerating(false);
    }
  }, [qc]);

  // ── 计算数据 ──

  const items = useMemo(() => listData?.items ?? [], [listData]);

  // 汇总卡片数据
  const overall = useMemo(() => {
    if (items.length === 0) return null;
    const totalDeviation = items.reduce(
      (s, i) => s + Math.abs(i.deviation_energy_mwh),
      0,
    );
    const positiveMwh = items
      .filter((i) => i.deviation_energy_mwh > 0)
      .reduce((s, i) => s + i.deviation_energy_mwh, 0);
    const negativeMwh = items
      .filter((i) => i.deviation_energy_mwh < 0)
      .reduce((s, i) => s + Math.abs(i.deviation_energy_mwh), 0);
    const totalCost = items.reduce(
      (s, i) => s + i.total_settlement,
      0,
    );
    const totalPenalty = items.reduce((s, i) => s + i.penalty_cost, 0);
    const avgRate =
      items.length > 0
        ? items.reduce((s, i) => s + Math.abs(i.deviation_rate), 0) /
          items.length
        : 0;
    return {
      totalDeviation,
      positiveMwh,
      negativeMwh,
      totalCost,
      totalPenalty,
      avgRate,
      count: items.length,
    };
  }, [items]);

  // ── Sprint 2 新增：偏差费用统计卡片数据 ──

  const cumulativePenaltyCost = useMemo(() => {
    return items.reduce((s, i) => s + i.penalty_cost, 0);
  }, [items]);

  const dailyAvgDeviation = useMemo(() => {
    if (items.length === 0) return 0;
    const dates = new Set(items.map((i) => i.operating_date?.slice(0, 10)).filter(Boolean));
    const totalAbs = items.reduce((s, i) => s + Math.abs(i.deviation_energy_mwh), 0);
    return dates.size > 0 ? totalAbs / dates.size : 0;
  }, [items]);

  const maxDeviationDay = useMemo(() => {
    if (items.length === 0) return null;
    const byDate = new Map<string, number>();
    for (const item of items) {
      const key = item.operating_date?.slice(0, 10);
      if (!key) continue;
      byDate.set(key, (byDate.get(key) || 0) + Math.abs(item.deviation_energy_mwh));
    }
    let maxDate = '';
    let maxVal = 0;
    for (const [date, val] of byDate) {
      if (val > maxVal) {
        maxVal = val;
        maxDate = date;
      }
    }
    return maxDate ? { date: maxDate, value: maxVal } : null;
  }, [items]);

  // 超标记录数
  const warningCount = useMemo(() => {
    return items.filter((i) => Math.abs(i.deviation_rate) > WARNING_RATE_THRESHOLD).length;
  }, [items]);

  // 偏差趋势数据（按日期聚合）—— 增加净偏差
  const trendData = useMemo(() => {
    if (items.length === 0) return [];
    const map = new Map<
      string,
      { date: string; positive: number; negative: number; cost: number; net: number }
    >();
    for (const item of items) {
      const key = item.operating_date?.slice(0, 10);
      if (!key) continue;
      const prev = map.get(key) || {
        date: key.slice(5),
        positive: 0,
        negative: 0,
        cost: 0,
        net: 0,
      };
      prev.net += item.deviation_energy_mwh;
      if (item.deviation_energy_mwh > 0) {
        prev.positive += item.deviation_energy_mwh;
      } else {
        prev.negative += Math.abs(item.deviation_energy_mwh);
      }
      prev.cost += item.total_settlement;
      map.set(key, prev);
    }
    return Array.from(map.values()).sort((a, b) => a.date.localeCompare(b.date));
  }, [items]);

  // 按类别统计分布
  const categoryDistData = useMemo(() => {
    if (items.length === 0) return [];
    const map = new Map<string, { name: string; 偏差电量: number; 偏差费用: number; 数量: number }>();
    for (const item of items) {
      const label = catLabel(item.category);
      const prev = map.get(item.category) || {
        name: label,
        偏差电量: 0,
        偏差费用: 0,
        数量: 0,
      };
      prev.偏差电量 += Math.abs(item.deviation_energy_mwh);
      prev.偏差费用 += item.total_settlement;
      prev.数量 += 1;
      map.set(item.category, prev);
    }
    return Array.from(map.values());
  }, [items]);

  // ── Sprint 2 新增：偏差归因分析数据 ──
  const attributionData = useMemo(() => {
    if (items.length === 0) return [];
    const totalAbs = items.reduce((s, i) => s + Math.abs(i.deviation_energy_mwh), 0);
    if (totalAbs === 0) return [];

    const forecastError = items
      .filter((i) => Math.abs(i.deviation_rate) > 10)
      .reduce((s, i) => s + Math.abs(i.deviation_energy_mwh), 0);
    const contractDecomp = items
      .filter((i) => Math.abs(i.deviation_rate) > 5 && Math.abs(i.deviation_rate) <= 10)
      .reduce((s, i) => s + Math.abs(i.deviation_energy_mwh), 0);
    const actualChange = items
      .filter((i) => Math.abs(i.deviation_rate) <= 5)
      .reduce((s, i) => s + Math.abs(i.deviation_energy_mwh), 0);

    const result: { name: string; value: number; color: string }[] = [];
    if (forecastError > 0) result.push({ name: '预测误差', value: forecastError, color: '#ef4444' });
    if (contractDecomp > 0) result.push({ name: '合同分解', value: contractDecomp, color: '#f59e0b' });
    if (actualChange > 0) result.push({ name: '实际变化', value: actualChange, color: '#10b981' });
    return result;
  }, [items]);

  // 汇总数据（来自 summary API）
  const summaryStats = useMemo(() => {
    if (!summaryData?.items?.length) return null;
    const all = summaryData.items;
    return {
      totalDeviationMwh: all.reduce((s, i) => s + i.total_deviation_energy_mwh, 0),
      totalCost: all.reduce((s, i) => s + i.total_cost, 0),
      avgRate:
        all.length > 0
          ? all.reduce((s, i) => s + i.avg_deviation_rate, 0) / all.length
          : 0,
      totalRecords: all.reduce((s, i) => s + i.count, 0),
      categories: all,
    };
  }, [summaryData]);

  // ── 表格列定义 ──

  const columns: DataTableColumn[] = [
    {
      key: 'operating_date',
      header: '运行日期',
      sortable: true,
      render: (row: DataRow) =>
        (row.operating_date as string)?.slice(0, 10) ?? '-',
    },
    {
      key: 'category',
      header: '类别',
      render: (row: DataRow) => {
        const cat = row.category as string;
        return (
          <Badge
            variant="outline"
            style={{ borderColor: catColor(cat), color: catColor(cat) }}
          >
            {catLabel(cat)}
          </Badge>
        );
      },
    },
    {
      key: 'declared_energy_mwh',
      header: '申报电量 (MWh)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => fmt(row.declared_energy_mwh as number, 2),
    },
    {
      key: 'actual_energy_mwh',
      header: '实际电量 (MWh)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => fmt(row.actual_energy_mwh as number, 2),
    },
    {
      key: 'deviation_energy_mwh',
      header: '偏差电量 (MWh)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => {
        const v = row.deviation_energy_mwh as number;
        const isPositive = v >= 0;
        const isWarning = Math.abs(v) > DEDUCTIBLE_ZONE;
        return (
          <span className={isWarning ? 'font-bold' : ''} style={isWarning ? { color: isPositive ? '#16a34a' : '#dc2626' } : { color: isPositive ? '#ef4444' : '#10b981' }}>
            {isPositive ? '+' : ''}
            {fmt(v, 2)}
          </span>
        );
      },
    },
    {
      key: 'deviation_rate',
      header: '偏差率',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => {
        const v = row.deviation_rate as number;
        const isWarning = Math.abs(v) > WARNING_RATE_THRESHOLD;
        return (
          <span
            className={`font-semibold ${isWarning ? 'bg-red-100 text-red-700 px-1.5 py-0.5 rounded' : ''}`}
            style={isWarning ? {} : { color: v >= 0 ? '#ef4444' : '#10b981' }}
          >
            {v >= 0 ? '+' : ''}
            {fmt(v, 1)}%
            {isWarning && ' ⚠'}
          </span>
        );
      },
    },
    {
      key: 'deviation_cost',
      header: '偏差费用 (¥)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => {
        const v = row.deviation_cost as number;
        const isWarning = Math.abs(v) > 5000;
        return (
          <span className={isWarning ? 'font-bold text-red-600 bg-red-50 px-1.5 py-0.5 rounded' : ''}>
            {fmtMoney(v)}
          </span>
        );
      },
    },
    {
      key: 'penalty_cost',
      header: '考核费用 (¥)',
      align: 'right',
      render: (row: DataRow) => {
        const v = row.penalty_cost as number;
        const isWarning = v > 3000;
        return (
          <span className={isWarning ? 'font-bold text-red-600 bg-red-50 px-1.5 py-0.5 rounded' : ''}>
            {fmtMoney(v)}
          </span>
        );
      },
    },
    {
      key: 'total_settlement',
      header: '总结算 (¥)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => (
        <span className="font-semibold">
          {fmtMoney(row.total_settlement as number)}
        </span>
      ),
    },
  ];

  // ── 渲染 ──

  return (
    <div className="space-y-6">
      <PageHeader
        title="偏差管理"
        description="偏差结算数据管理 · 偏差电量分析 · 偏差预警"
        actions={
          canWrite ? (
            <Button variant="outline" onClick={onGenerate} disabled={generating}>
              {generating ? (
                <RefreshCw className="mr-1 h-4 w-4 animate-spin" />
              ) : (
                <BarChart3 className="mr-1 h-4 w-4" />
              )}
              {generating ? '生成中...' : '生成演示数据'}
            </Button>
          ) : undefined
        }
      />

      {/* 提示 & 错误 */}
      {notice && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{notice}</AlertDescription>
        </Alert>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 筛选条件 */}
      <Card>
        <CardContent className="flex flex-wrap items-end gap-4 pt-6">
          <div className="space-y-1">
            <Label>天数范围</Label>
            <Input
              type="number"
              className="w-24"
              min={1}
              max={90}
              value={days}
              onChange={(e) =>
                setDays(Math.min(90, Math.max(1, Number(e.target.value) || 1)))
              }
            />
          </div>
          <div className="space-y-1">
            <Label>偏差类别</Label>
            <select
              className="flex h-9 w-40 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              value={categoryFilter}
              onChange={(e) => setCategoryFilter(e.target.value)}
            >
              <option value="">全部类别</option>
              <option value="day_ahead">日前偏差</option>
              <option value="real_time">实时偏差</option>
              <option value="intraday">日内偏差</option>
            </select>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => {
                setDays(30);
                setCategoryFilter('');
              }}
            >
              <Search className="mr-1 h-3.5 w-3.5" />
              重置
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* ── Sprint 2：偏差费用统计卡片 ── */}
      {overall && (
        <div className="grid gap-4 md:grid-cols-4 lg:grid-cols-7">
          <StatCard
            title="总偏差电量 (MWh)"
            value={fmt(overall.totalDeviation, 2)}
            icon={<Zap className="h-4 w-4" />}
          />
          <StatCard
            title="正偏差 (MWh)"
            value={`+${fmt(overall.positiveMwh, 2)}`}
            icon={<TrendingUp className="h-4 w-4 text-red-500" />}
          />
          <StatCard
            title="负偏差 (MWh)"
            value={`-${fmt(overall.negativeMwh, 2)}`}
            icon={<TrendingDown className="h-4 w-4 text-emerald-500" />}
          />
          <StatCard
            title="总偏差费用 (¥)"
            value={fmtMoney(overall.totalCost)}
            icon={<DollarSign className="h-4 w-4" />}
          />
          <StatCard
            title="累计考核费用 (¥)"
            value={fmtMoney(cumulativePenaltyCost)}
            icon={<ShieldAlert className="h-4 w-4 text-red-500" />}
            className={cumulativePenaltyCost > 50000 ? 'border-red-300 bg-red-50' : ''}
          />
          <StatCard
            title="日均偏差电量 (MWh)"
            value={fmt(dailyAvgDeviation, 2)}
            icon={<Activity className="h-4 w-4" />}
          />
          <StatCard
            title="最大偏差日"
            value={maxDeviationDay ? maxDeviationDay.date.slice(5) : '-'}
            icon={<CalendarDays className="h-4 w-4" />}
            trendLabel={maxDeviationDay ? `${fmt(maxDeviationDay.value, 1)} MWh` : undefined}
          />
        </div>
      )}

      {/* 预警统计条 */}
      {warningCount > 0 && (
        <Alert className="border-orange-300 bg-orange-50">
          <ShieldAlert className="h-4 w-4 text-orange-600" />
          <AlertDescription className="text-orange-800">
            <span className="font-semibold">{warningCount}</span> 条记录偏差率超过 {WARNING_RATE_THRESHOLD}% 预警阈值，请关注偏差控制
          </AlertDescription>
        </Alert>
      )}

      {/* ── Sprint 2：偏差电量趋势图（日度柱状图 + 免赔区间） ── */}
      <ChartContainer title="偏差电量趋势（日度）" minHeight={300}>
        {trendData.length > 0 ? (
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={trendData}
              margin={{ top: 8, right: 16, bottom: 8, left: 8 }}
            >
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis dataKey="date" interval="preserveStartEnd" tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 11 }} width={56} />
              <RechartsTooltip
                content={({ active, payload, label }) => {
                  if (!active || !payload?.length) return null;
                  const d = payload[0]?.payload as typeof trendData[number];
                  return (
                    <div className="rounded-md border bg-white px-3 py-2 text-xs shadow-lg">
                      <p className="mb-1 font-semibold">日期: {label}</p>
                      <p>净偏差: <span style={{ color: d?.net >= 0 ? '#10b981' : '#ef4444' }}>{d?.net >= 0 ? '+' : ''}{fmt(d?.net, 2)} MWh</span></p>
                      <p style={{ color: '#10b981' }}>正偏差: +{fmt(d?.positive, 2)} MWh</p>
                      <p style={{ color: '#ef4444' }}>负偏差: -{fmt(d?.negative, 2)} MWh</p>
                      <p className="text-muted-foreground">偏差费用: ¥{fmtMoney(d?.cost)}</p>
                      <p className="text-muted-foreground mt-1 text-[10px]">免赔区间: ±{DEDUCTIBLE_ZONE} MWh</p>
                    </div>
                  );
                }}
              />
              <Legend />
              {/* ±免赔区间标注 */}
              <ReferenceArea
                y1={-DEDUCTIBLE_ZONE}
                y2={DEDUCTIBLE_ZONE}
                fill="#10b981"
                fillOpacity={0.06}
                stroke="#10b981"
                strokeOpacity={0.2}
                strokeDasharray="3 3"
              />
              <Bar
                dataKey="net"
                isAnimationActive={false}
                name="净偏差电量 (MWh)"
              >
                {trendData.map((entry, idx) => (
                  <Cell
                    key={idx}
                    fill={entry.net >= 0 ? '#10b981' : '#ef4444'}
                  />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            暂无趋势数据
          </div>
        )}
      </ChartContainer>

      {/* ── Sprint 2：偏差归因分析 + 类别分布 ── */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* 偏差归因分析饼图 */}
        <ChartContainer title="偏差归因分析" minHeight={280}>
          {attributionData.length > 0 ? (
            <div className="flex h-full items-center justify-center">
              <ResponsiveContainer width="100%" height={260}>
                <PieChart>
                  <Pie
                    data={attributionData}
                    dataKey="value"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    outerRadius={90}
                    innerRadius={50}
                    isAnimationActive={false}
                    label={({ name, percent }) =>
                      `${name} ${(percent * 100).toFixed(0)}%`
                    }
                  >
                    {attributionData.map((entry, idx) => (
                      <Cell key={idx} fill={entry.color} />
                    ))}
                  </Pie>
                  <RechartsTooltip
                    content={({ active, payload }) => {
                      if (!active || !payload?.length) return null;
                      const d = payload[0].payload as (typeof attributionData)[number];
                      return (
                        <div className="rounded-md border bg-white px-3 py-2 text-xs shadow-lg">
                          <p className="font-semibold">{d.name}</p>
                          <p>偏差电量: {fmt(d.value, 2)} MWh</p>
                          <p>占比: {((d.value / attributionData.reduce((s, i) => s + i.value, 0)) * 100).toFixed(1)}%</p>
                        </div>
                      );
                    }}
                  />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            </div>
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              暂无归因数据
            </div>
          )}
        </ChartContainer>

        {/* 类别分布（原有） */}
        <ChartContainer title="偏差类别分布" minHeight={280}>
          {categoryDistData.length > 0 ? (
            <div className="flex h-full items-center justify-center">
              <ResponsiveContainer width="100%" height={240}>
                <PieChart>
                  <Pie
                    data={categoryDistData}
                    dataKey="偏差电量"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    outerRadius={90}
                    innerRadius={50}
                    isAnimationActive={false}
                    label={({ name, percent }) =>
                      `${name} ${(percent * 100).toFixed(0)}%`
                    }
                  >
                    {categoryDistData.map((entry, idx) => (
                      <Cell
                        key={idx}
                        fill={
                          Object.values(CATEGORY_COLORS)[idx] || '#6b7280'
                        }
                      />
                    ))}
                  </Pie>
                  <RechartsTooltip
                    content={({ active, payload }) => {
                      if (!active || !payload?.length) return null;
                      const d = payload[0].payload as any;
                      return (
                        <div className="rounded-md border bg-white px-3 py-2 text-xs shadow-lg">
                          <p className="font-semibold">{d.name}</p>
                          <p>偏差电量: {fmt(d.偏差电量, 2)} MWh</p>
                          <p>偏差费用: {fmtMoney(d.偏差费用)}</p>
                          <p>记录数: {d.数量}</p>
                        </div>
                      );
                    }}
                  />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            </div>
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              暂无分类数据
            </div>
          )}
        </ChartContainer>
      </div>

      {/* 汇总统计卡片（后端 summary API） */}
      {summaryStats && (
        <div className="grid gap-4 md:grid-cols-4">
          {summaryStats.categories.map((cat) => {
            const isWarning = cat.avg_deviation_rate > WARNING_RATE_THRESHOLD;
            return (
              <Card key={cat.category} className={isWarning ? 'border-red-300 bg-red-50/50' : ''}>
                <CardHeader className="pb-2">
                  <CardTitle className="flex items-center gap-2 text-sm font-medium">
                    <span
                      className="inline-block h-3 w-3 rounded-full"
                      style={{ backgroundColor: catColor(cat.category) }}
                    />
                    {catLabel(cat.category)}
                    {isWarning && (
                      <Badge variant="destructive" className="ml-auto text-[10px] px-1.5 py-0">超标</Badge>
                    )}
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-1 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">偏差电量</span>
                    <span className="font-medium">
                      {fmt(cat.total_deviation_energy_mwh, 2)} MWh
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">偏差费用</span>
                    <span className={`font-medium ${isWarning ? 'text-red-600' : ''}`}>
                      {fmtMoney(cat.total_cost)}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">平均偏差率</span>
                    <span className={`font-medium ${isWarning ? 'text-red-600 font-bold' : ''}`}>
                      {fmt(cat.avg_deviation_rate, 1)}%
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">记录数</span>
                    <span className="font-medium">{cat.count}</span>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* 偏差明细表格 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <BarChart3 className="h-4 w-4 text-primary" />
            偏差明细
            <span className="ml-2 text-xs font-normal text-muted-foreground">
              {items.length} 条记录
            </span>
            {warningCount > 0 && (
              <Badge variant="destructive" className="ml-2 text-[10px]">
                {warningCount} 条超标
              </Badge>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <DataTable
            columns={columns}
            data={items as DataRow[]}
            rowKey="id"
            pageSize={15}
            showPagination
            loading={listLoading}
          />
        </CardContent>
      </Card>
    </div>
  );
}
