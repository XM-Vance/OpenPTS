'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ComposedChart,
  Legend,
  Line,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { StatCard } from '@/components/data-display/stat-card';
import { ChartContainer } from '@/components/charts/chart-container';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import { genCustProfitDemo, listCustomerProfit } from '@/lib/api/customer-profit-analysis';
import {
  customerProfitAnalysisApi,
  type ProfitViewMode,
} from '@/lib/api/customer-profit-analysis';
import {
  ChevronLeft,
  ChevronRight,
  Users,
  Zap,
  DollarSign,
  TrendingUp,
  BarChart3,
} from 'lucide-react';

const wan = (v: number) =>
  (v / 10000).toLocaleString('zh-CN', { maximumFractionDigits: 2 });
const fmt = (v: number, d = 2) =>
  v.toLocaleString('zh-CN', { maximumFractionDigits: d, minimumFractionDigits: d });

function marginVariant(
  m: number,
): 'default' | 'secondary' | 'destructive' | 'success' {
  if (m >= 8) return 'success';
  if (m >= 4) return 'default';
  if (m >= 0) return 'secondary';
  return 'destructive';
}

const PIE_COLORS = [
  '#3b82f6',
  '#10b981',
  '#f97316',
  '#8b5cf6',
  '#06b6d4',
  '#94a3b8',
];

const CURRENT_YEAR = new Date().getFullYear();
const CURRENT_MONTH = new Date().getMonth() + 1;

export default function CustomerProfitPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('analytics:write');

  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedMonth, setSelectedMonth] = useState(CURRENT_MONTH);
  const [viewMode, setViewMode] = useState<ProfitViewMode>('monthly');

  const { data, isLoading } = useQuery({
    queryKey: ['customer-profit'],
    queryFn: () => listCustomerProfit({ limit: 50 }),
  });

  const { data: dashboard } = useQuery({
    queryKey: ['profit-dashboard', selectedMonth, viewMode],
    queryFn: () =>
      customerProfitAnalysisApi.getDashboardData({
        year: CURRENT_YEAR,
        month: selectedMonth,
        view_mode: viewMode,
      }),
  });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genCustProfitDemo();
      qc.invalidateQueries({ queryKey: ['customer-profit'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const items = useMemo(() => data?.items ?? [], [data]);

  const totalProfit = items.reduce((s, i) => s + i.gross_profit, 0);
  const totalRevenue = items.reduce((s, i) => s + i.revenue, 0);
  const totalCost = items.reduce((s, i) => s + i.cost, 0);
  const avgMargin = totalRevenue === 0 ? 0 : (totalProfit / totalRevenue) * 100;

  // ── Chart 1: 利润瀑布图 ──
  const waterfallData = useMemo(() => {
    if (items.length === 0) return [];
    const totalRev = items.reduce((s, i) => s + i.revenue, 0);
    const totalCst = items.reduce((s, i) => s + i.cost, 0);
    const energyCost = totalCst * 0.65;
    const capacityCost = totalCst * 0.2;
    const otherCost = totalCst * 0.1;
    const taxFee = totalCst * 0.05;
    const netProfit = totalRev - totalCst;
    // 瀑布图: invisible bar + visible bar
    return [
      { name: '收入', invisible: 0, value: totalRev / 10000, color: '#3b82f6' },
      { name: '购电成本', invisible: (totalRev - energyCost) / 10000, value: -energyCost / 10000, color: '#ef4444' },
      { name: '容量成本', invisible: (totalRev - energyCost - capacityCost) / 10000, value: -capacityCost / 10000, color: '#f97316' },
      { name: '其他费用', invisible: (totalRev - energyCost - capacityCost - otherCost) / 10000, value: -otherCost / 10000, color: '#f59e0b' },
      { name: '税费', invisible: (totalRev - energyCost - capacityCost - otherCost - taxFee) / 10000, value: -taxFee / 10000, color: '#8b5cf6' },
      { name: '净利润', invisible: 0, value: netProfit / 10000, color: '#10b981' },
    ];
  }, [items]);

  // ── Chart 2: 客户利润排行 ──
  const byCustomer = useMemo(() => {
    const map = new Map<string, { name: string; profit: number; energy: number; revenue: number; cost: number }>();
    for (const i of items) {
      const cur = map.get(i.customer_id) ?? { name: i.customer_name, profit: 0, energy: 0, revenue: 0, cost: 0 };
      cur.profit += i.gross_profit;
      cur.energy += i.energy_mwh;
      cur.revenue += i.revenue;
      cur.cost += i.cost;
      map.set(i.customer_id, cur);
    }
    return Array.from(map.values()).sort((a, b) => b.profit - a.profit).slice(0, 10);
  }, [items]);

  const ranking = byCustomer.map((c) => ({ name: c.name, profit: c.profit / 10000 }));

  // ── Chart 3: 利润敏感度分析（电价±5%影响）──
  const sensitivityData = useMemo(() => {
    const basePrice = totalRevenue > 0 ? totalRevenue / items.reduce((s, i) => s + i.energy_mwh, 0) : 400;
    const points = [-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5];
    return points.map((pct) => {
      const priceChange = basePrice * (pct / 100);
      const newRevenue = totalRevenue + priceChange * items.reduce((s, i) => s + i.energy_mwh, 0);
      const newProfit = newRevenue - totalCost;
      return {
        电价变动: `${pct > 0 ? '+' : ''}${pct}%`,
        利润: +(newProfit / 10000).toFixed(2),
        毛利率: newRevenue > 0 ? +((newProfit / newRevenue) * 100).toFixed(1) : 0,
      };
    });
  }, [items, totalRevenue, totalCost]);

  const monthOptions = Array.from({ length: CURRENT_MONTH }, (_, i) => i + 1);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">客户利润分析</h1>
          <p className="text-sm text-muted-foreground">
            按客户聚合收入 / 成本 / 毛利 / 毛利率
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="icon" className="h-8 w-8"
            onClick={() => setSelectedMonth((m) => Math.max(1, m - 1))} disabled={selectedMonth <= 1}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <select value={selectedMonth} onChange={(e) => setSelectedMonth(Number(e.target.value))}
            className="flex h-8 rounded-md border border-input bg-transparent px-3 text-sm">
            {monthOptions.map((m) => (
              <option key={m} value={m}>{CURRENT_YEAR}-{String(m).padStart(2, '0')}</option>
            ))}
          </select>
          <Button variant="outline" size="icon" className="h-8 w-8"
            onClick={() => setSelectedMonth((m) => Math.min(CURRENT_MONTH, m + 1))} disabled={selectedMonth >= CURRENT_MONTH}>
            <ChevronRight className="h-4 w-4" />
          </Button>
          <div className="flex gap-1">
            <Button variant={viewMode === 'monthly' ? 'default' : 'outline'} size="sm" className="h-8 text-xs"
              onClick={() => setViewMode('monthly')}>月度</Button>
            <Button variant={viewMode === 'ytd' ? 'default' : 'outline'} size="sm" className="h-8 text-xs"
              onClick={() => setViewMode('ytd')}>累计</Button>
          </div>
          {canWrite && (
            <Button variant="outline" onClick={onGen} disabled={busy}>
              {busy ? '生成中...' : '生成演示数据'}
            </Button>
          )}
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* KPI stat cards */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-6">
        <StatCard title="客户数" value={`${dashboard?.kpi.customer_count ?? items.length}`} icon={<Users className="h-4 w-4" />} />
        <StatCard title="结算电量" value={`${fmt(dashboard?.kpi.total_energy_mwh ?? 0)} MWh`} icon={<Zap className="h-4 w-4" />} />
        <StatCard title="零售收入" value={`${wan(dashboard?.kpi.retail_revenue ?? totalRevenue)} 万`} icon={<DollarSign className="h-4 w-4" />} />
        <StatCard title="批发成本" value={`${wan(dashboard?.kpi.wholesale_cost ?? totalCost)} 万`} icon={<DollarSign className="h-4 w-4" />} />
        <StatCard title="毛利" value={`${wan(dashboard?.kpi.gross_profit ?? totalProfit)} 万`} icon={<TrendingUp className="h-4 w-4" />} />
        <StatCard title="平均价差" value={`${fmt(dashboard?.kpi.avg_spread ?? 0, 3)} 元/MWh`} icon={<BarChart3 className="h-4 w-4" />} />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {/* Chart 1: 利润瀑布图 */}
        <ChartContainer title="利润瀑布图（收入→成本→费用→净利润）">
          {waterfallData.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              暂无利润数据{canWrite && '，可点右上「生成演示数据」'}
            </p>
          ) : (
            <ResponsiveContainer width="100%" height={320}>
              <BarChart data={waterfallData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="name" tick={{ fontSize: 11, fill: '#6b7280' }} />
                <YAxis tick={{ fontSize: 11, fill: '#6b7280' }} width={60} unit=" 万" />
                <Tooltip formatter={(v: number, name: string) =>
                  name === 'invisible' ? null : `${v.toFixed(2)} 万元`}
                  contentStyle={{ fontSize: 12 }} />
                <Bar dataKey="invisible" stackId="a" fill="transparent" isAnimationActive={false} />
                <Bar dataKey="value" stackId="a" isAnimationActive={false}>
                  {waterfallData.map((d, idx) => (
                    <Cell key={idx} fill={d.color} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          )}
        </ChartContainer>

        {/* Chart 2: 客户利润排行 */}
        <ChartContainer title="Top-10 客户毛利排行（万元）">
          {ranking.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              暂无利润数据{canWrite && '，可点右上「生成演示数据」'}
            </p>
          ) : (
            <ResponsiveContainer width="100%" height={320}>
              <BarChart data={ranking} layout="vertical" margin={{ top: 8, right: 12, left: 80, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis type="number" tick={{ fontSize: 12, fill: '#6b7280' }} />
                <YAxis type="category" dataKey="name" tick={{ fontSize: 11, fill: '#6b7280' }} width={120} />
                <Tooltip formatter={(v: number) => `${v.toFixed(2)} 万元`} contentStyle={{ fontSize: 12 }} />
                <Bar dataKey="profit" fill="#10b981" isAnimationActive={false}>
                  {ranking.map((_, idx) => (
                    <Cell key={idx} fill={PIE_COLORS[idx % PIE_COLORS.length]} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          )}
        </ChartContainer>
      </div>

      {/* Chart 3: 利润敏感度分析 */}
      <ChartContainer title="利润敏感度分析（电价±5%影响）">
        {sensitivityData.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={280}>
            <ComposedChart data={sensitivityData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="电价变动" tick={{ fontSize: 11, fill: '#6b7280' }} />
              <YAxis yAxisId="profit" tick={{ fontSize: 11, fill: '#6b7280' }} width={60} unit=" 万" />
              <YAxis yAxisId="margin" orientation="right" tick={{ fontSize: 11, fill: '#6b7280' }} width={60} unit="%" />
              <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v: number, name: string) =>
                name === '毛利率' ? `${v}%` : `${v} 万元`} />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Line yAxisId="profit" type="monotone" dataKey="利润" stroke="#3b82f6" strokeWidth={2} dot={{ r: 3 }} isAnimationActive={false} />
              <Line yAxisId="margin" type="monotone" dataKey="毛利率" stroke="#f97316" strokeWidth={2} strokeDasharray="5 5" dot={false} isAnimationActive={false} />
            </ComposedChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* 客户利润明细 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">客户利润明细</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>月份</TableHead>
                  <TableHead>客户</TableHead>
                  <TableHead className="text-right">电量 MWh</TableHead>
                  <TableHead className="text-right">收入（万元）</TableHead>
                  <TableHead className="text-right">成本（万元）</TableHead>
                  <TableHead className="text-right">毛利（万元）</TableHead>
                  <TableHead className="text-right">毛利率</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading && (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center text-muted-foreground">加载中...</TableCell>
                  </TableRow>
                )}
                {items.map((p) => (
                  <TableRow key={p.id}>
                    <TableCell className="font-medium">{p.operating_month}</TableCell>
                    <TableCell>{p.customer_name}</TableCell>
                    <TableCell className="text-right">{fmt(p.energy_mwh, 1)}</TableCell>
                    <TableCell className="text-right">{wan(p.revenue)}</TableCell>
                    <TableCell className="text-right">{wan(p.cost)}</TableCell>
                    <TableCell className="text-right font-bold text-emerald-600">{wan(p.gross_profit)}</TableCell>
                    <TableCell className="text-right">
                      <Badge variant={marginVariant(p.gross_margin)}>{fmt(p.gross_margin, 1)}%</Badge>
                    </TableCell>
                  </TableRow>
                ))}
                {items.length === 0 && !isLoading && (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center text-muted-foreground">暂无利润数据</TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
