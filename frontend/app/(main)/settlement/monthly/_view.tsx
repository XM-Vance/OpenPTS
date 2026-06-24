'use client';

import { useState, useMemo, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Area,
  AreaChart,
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
import {
  genMonthlySettlementDemo,
  listMonthlySettlement,
} from '@/lib/api/monthly-settlement';
import {
  getWholesaleSettlementByYear,
  getWholesaleMonthlySettlementYears,
} from '@/lib/api/wholesale-settlement';
import {
  getRetailSettlementMonthlySummaries,
} from '@/lib/api/retail-settlement';
import {
  ChevronLeft,
  ChevronRight,
  Download,
  TrendingUp,
  Zap,
  Users,
  DollarSign,
  BarChart3,
} from 'lucide-react';

const wan = (v: number) =>
  (v / 10000).toLocaleString('zh-CN', { maximumFractionDigits: 2 });
const fmt = (v: number, d = 2) =>
  v.toLocaleString('zh-CN', { maximumFractionDigits: d });

const PIE_COLORS = ['#3b82f6', '#a855f7', '#f97316', '#10b981'];

export default function MonthlySettlementPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('settlement_management:write');

  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [chartMode, setChartMode] = useState<'fee' | 'price'>('fee');

  const currentYear = new Date().getFullYear();
  const [selectedYear, setSelectedYear] = useState(currentYear);

  // Fetch monthly settlement data
  const { data, isLoading } = useQuery({
    queryKey: ['monthly-settlement', selectedYear],
    queryFn: () => listMonthlySettlement(12),
  });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genMonthlySettlementDemo();
      qc.invalidateQueries({ queryKey: ['monthly-settlement'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const shiftYear = (delta: number) => setSelectedYear((y) => y + delta);

  const items = useMemo(() => data?.items ?? [], [data]);

  // Summary stats
  const sumTotal = items.reduce((s, m) => s + m.total_fee, 0);
  const sumEnergy = items.reduce((s, m) => s + m.energy_fee, 0);
  const sumCapacity = items.reduce((s, m) => s + m.capacity_fee, 0);
  const sumSubsidy = items.reduce((s, m) => s + m.policy_subsidy, 0);
  const sumAncillary = items.reduce((s, m) => s + m.ancillary_fee, 0);
  const sumEnergyMwh = items.reduce((s, m) => s + m.settled_energy_mwh, 0);
  const avgPrice =
    sumEnergyMwh > 0 ? (sumTotal - sumSubsidy) / sumEnergyMwh : 0;

  // Pie chart data - 电费构成
  const pieData = useMemo(() => {
    if (sumTotal === 0) return [];
    return [
      { name: '电能量电费', value: sumEnergy, fill: PIE_COLORS[0] },
      { name: '容量电费', value: sumCapacity, fill: PIE_COLORS[1] },
      { name: '辅助服务分摊', value: sumAncillary, fill: PIE_COLORS[2] },
      { name: '政策补贴', value: sumSubsidy, fill: PIE_COLORS[3] },
    ];
  }, [sumEnergy, sumCapacity, sumAncillary, sumSubsidy, sumTotal]);

  // 预算偏差分析数据 (模拟预算数据：按比例生成)
  const budgetDeviationData = useMemo(() => {
    return items
      .slice()
      .reverse()
      .map((m) => {
        // 模拟预算值：在真实基础上加 ±10% 随机偏差
        const seed = m.operating_month.charCodeAt(5) + m.operating_month.charCodeAt(6);
        const budgetFactor = 1 + ((seed % 20) - 10) / 100;
        const budget = m.total_fee * budgetFactor;
        const deviation = ((m.total_fee - budget) / budget) * 100;
        return {
          month: m.operating_month.slice(5),
          actual: m.total_fee / 10000,
          budget: budget / 10000,
          deviation: Math.round(deviation * 100) / 100,
        };
      });
  }, [items]);

  // Chart data: reversed for chronological order
  const chartData = useMemo(
    () =>
      items
        .slice()
        .reverse()
        .map((m) => ({
          month: m.operating_month.slice(5),
          total: m.total_fee / 10000,
          energy: m.energy_fee / 10000,
          capacity: m.capacity_fee / 10000,
          ancillary: m.ancillary_fee / 10000,
          subsidy: m.policy_subsidy / 10000,
          energyMwh: m.settled_energy_mwh,
          avgPrice: m.settled_energy_mwh > 0 ? m.energy_fee / m.settled_energy_mwh : 0,
        })),
    [items],
  );

  const renderPieLabel = ({
    cx,
    cy,
    midAngle,
    innerRadius,
    outerRadius,
    percent,
    name,
  }: {
    cx: number;
    cy: number;
    midAngle: number;
    innerRadius: number;
    outerRadius: number;
    percent: number;
    name: string;
  }) => {
    if (percent < 0.05) return null;
    const RADIAN = Math.PI / 180;
    const radius = innerRadius + (outerRadius - innerRadius) * 1.4;
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);
    return (
      <text x={x} y={y} fill="#374151" textAnchor="middle" fontSize={11}>
        {name} {(percent * 100).toFixed(1)}%
      </text>
    );
  };

  return (
    <div className="space-y-4">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">月度结算</h1>
          <p className="text-sm text-muted-foreground">
            批发月度汇总：电能量电费 + 容量电费 + 辅助服务分摊 + 政策补贴
          </p>
        </div>
        <div className="flex items-center gap-2">
          {/* Year navigator */}
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => shiftYear(-1)}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="min-w-[60px] text-center font-semibold">
            {selectedYear}
          </span>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => shiftYear(1)}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
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
        <StatCard
          title="年度客户"
          value="-"
          icon={<Users className="h-4 w-4" />}
        />
        <StatCard
          title="年度电量"
          value={`${fmt(sumEnergyMwh, 0)} MWh`}
          icon={<Zap className="h-4 w-4" />}
        />
        <StatCard
          title="年度总费用"
          value={`${wan(sumTotal)} 万`}
          icon={<DollarSign className="h-4 w-4" />}
        />
        <StatCard
          title="电能量电费"
          value={`${wan(sumEnergy)} 万`}
          icon={<BarChart3 className="h-4 w-4" />}
        />
        <StatCard
          title="容量电费"
          value={`${wan(sumCapacity)} 万`}
          icon={<BarChart3 className="h-4 w-4" />}
        />
        <StatCard
          title="政策补贴"
          value={`${wan(sumSubsidy)} 万`}
          icon={<TrendingUp className="h-4 w-4" />}
        />
      </div>

      {/* 饼图 + 费用构成图并排 */}
      <div className="grid gap-4 lg:grid-cols-2">
        {/* 月度电费构成饼图 */}
        <ChartContainer title="月度电费构成">
          {pieData.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              暂无数据{canWrite && '，可点右上「生成演示数据」'}
            </p>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  outerRadius={100}
                  innerRadius={50}
                  paddingAngle={2}
                  dataKey="value"
                  label={renderPieLabel}
                  labelLine={false}
                  isAnimationActive={false}
                >
                  {pieData.map((entry, idx) => (
                    <Cell key={entry.name} fill={entry.fill} />
                  ))}
                </Pie>
                <Tooltip
                  formatter={(v: number) => `${wan(v)} 万`}
                  contentStyle={{ fontSize: 12 }}
                />
              </PieChart>
            </ResponsiveContainer>
          )}
        </ChartContainer>

        {/* Fee breakdown chart */}
        <ChartContainer
          title="月度费用构成（万元）"
          actions={
            <div className="flex gap-1">
              <Button
                variant={chartMode === 'fee' ? 'default' : 'outline'}
                size="sm"
                className="h-7 text-xs"
                onClick={() => setChartMode('fee')}
              >
                费用构成
              </Button>
              <Button
                variant={chartMode === 'price' ? 'default' : 'outline'}
                size="sm"
                className="h-7 text-xs"
                onClick={() => setChartMode('price')}
              >
                均价走势
              </Button>
            </div>
          }
        >
          {chartData.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              暂无数据{canWrite && '，可点右上「生成演示数据」'}
            </p>
          ) : chartMode === 'fee' ? (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart
                data={chartData}
                margin={{ top: 8, right: 12, left: 0, bottom: 0 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="month" tick={{ fontSize: 12, fill: '#6b7280' }} />
                <YAxis
                  tick={{ fontSize: 12, fill: '#6b7280' }}
                  width={60}
                />
                <Tooltip
                  formatter={(v: number) => `${v.toFixed(2)} 万`}
                  contentStyle={{ fontSize: 12 }}
                />
                <Legend wrapperStyle={{ fontSize: 12 }} />
                <Bar
                  dataKey="energy"
                  name="电能量"
                  stackId="fee"
                  fill="#3b82f6"
                  isAnimationActive={false}
                />
                <Bar
                  dataKey="capacity"
                  name="容量"
                  stackId="fee"
                  fill="#a855f7"
                  isAnimationActive={false}
                />
                <Bar
                  dataKey="ancillary"
                  name="辅助服务"
                  stackId="fee"
                  fill="#f97316"
                  isAnimationActive={false}
                />
                <Bar
                  dataKey="subsidy"
                  name="政策补贴"
                  stackId="fee"
                  fill="#10b981"
                  isAnimationActive={false}
                />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <ComposedChart
                data={chartData}
                margin={{ top: 8, right: 12, left: 0, bottom: 0 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="month" tick={{ fontSize: 12, fill: '#6b7280' }} />
                <YAxis
                  yAxisId="price"
                  tick={{ fontSize: 12, fill: '#6b7280' }}
                  width={60}
                  unit="元"
                />
                <YAxis
                  yAxisId="energy"
                  orientation="right"
                  tick={{ fontSize: 12, fill: '#6b7280' }}
                  width={60}
                  unit="MWh"
                />
                <Tooltip
                  contentStyle={{ fontSize: 12 }}
                  formatter={(v: number, name: string) =>
                    name === '均价' ? `${v.toFixed(2)} 元/MWh` : `${fmt(v, 0)} MWh`
                  }
                />
                <Legend wrapperStyle={{ fontSize: 12 }} />
                <Bar
                  yAxisId="energy"
                  dataKey="energyMwh"
                  name="电量"
                  fill="#dbeafe"
                  isAnimationActive={false}
                />
                <Line
                  yAxisId="price"
                  type="monotone"
                  dataKey="avgPrice"
                  name="均价"
                  stroke="#f97316"
                  strokeWidth={2}
                  dot={false}
                  isAnimationActive={false}
                />
              </ComposedChart>
            </ResponsiveContainer>
          )}
        </ChartContainer>
      </div>

      {/* 预算偏差分析柱状图 */}
      <ChartContainer title="与预算/计划偏差分析">
        {budgetDeviationData.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={280}>
            <ComposedChart
              data={budgetDeviationData}
              margin={{ top: 8, right: 12, left: 0, bottom: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="month" tick={{ fontSize: 12, fill: '#6b7280' }} />
              <YAxis
                yAxisId="amount"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={60}
              />
              <YAxis
                yAxisId="deviation"
                orientation="right"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={60}
                unit="%"
              />
              <Tooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(v: number, name: string) => {
                  if (name === '偏差率') return `${v.toFixed(2)}%`;
                  return `${v.toFixed(2)} 万`;
                }}
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Bar
                yAxisId="amount"
                dataKey="actual"
                name="实际费用"
                fill="#3b82f6"
                isAnimationActive={false}
              />
              <Bar
                yAxisId="amount"
                dataKey="budget"
                name="预算"
                fill="#94a3b8"
                isAnimationActive={false}
                fillOpacity={0.5}
              />
              <Line
                yAxisId="deviation"
                type="monotone"
                dataKey="deviation"
                name="偏差率"
                stroke="#ef4444"
                strokeWidth={2}
                dot={{ r: 4, fill: '#ef4444' }}
                isAnimationActive={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* Settlement trend area chart */}
      <ChartContainer title="月度总费用趋势（万元）">
        {chartData.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={240}>
            <AreaChart
              data={chartData}
              margin={{ top: 8, right: 12, left: 0, bottom: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="month" tick={{ fontSize: 12, fill: '#6b7280' }} />
              <YAxis
                tickFormatter={(v) => `${Number(v).toFixed(1)}`}
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={60}
              />
              <Tooltip
                formatter={(v: number) => `${v.toFixed(2)} 万`}
                contentStyle={{ fontSize: 12 }}
              />
              <Area
                type="monotone"
                dataKey="total"
                stroke="#2563eb"
                fill="#dbeafe"
                strokeWidth={2}
                isAnimationActive={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* Monthly settlement table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">月度结算明细表</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>月份</TableHead>
                  <TableHead className="text-right">结算电量 (MWh)</TableHead>
                  <TableHead className="text-right">电能量电费</TableHead>
                  <TableHead className="text-right">容量电费</TableHead>
                  <TableHead className="text-right">辅助服务</TableHead>
                  <TableHead className="text-right">政策补贴</TableHead>
                  <TableHead className="text-right">合计</TableHead>
                  <TableHead className="text-right">均价 (元/MWh)</TableHead>
                  <TableHead>版本</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading && (
                  <TableRow>
                    <TableCell
                      colSpan={9}
                      className="text-center text-muted-foreground"
                    >
                      加载中...
                    </TableCell>
                  </TableRow>
                )}
                {items.map((m) => {
                  const unitPrice =
                    m.settled_energy_mwh > 0
                      ? m.energy_fee / m.settled_energy_mwh
                      : 0;
                  return (
                    <TableRow key={m.id}>
                      <TableCell className="font-medium">
                        {m.operating_month}
                      </TableCell>
                      <TableCell className="text-right">
                        {m.settled_energy_mwh.toLocaleString('zh-CN', {
                          maximumFractionDigits: 0,
                        })}
                      </TableCell>
                      <TableCell className="text-right">
                        {wan(m.energy_fee)} 万
                      </TableCell>
                      <TableCell className="text-right">
                        {wan(m.capacity_fee)} 万
                      </TableCell>
                      <TableCell className="text-right">
                        {wan(m.ancillary_fee)} 万
                      </TableCell>
                      <TableCell className="text-right text-emerald-600">
                        {wan(m.policy_subsidy)} 万
                      </TableCell>
                      <TableCell className="text-right font-bold text-blue-600">
                        {wan(m.total_fee)} 万
                      </TableCell>
                      <TableCell className="text-right">
                        {unitPrice.toFixed(2)}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{m.version}</Badge>
                      </TableCell>
                    </TableRow>
                  );
                })}
                {/* Summary row */}
                {items.length > 0 && (
                  <TableRow className="bg-muted/50 font-bold">
                    <TableCell>合计</TableCell>
                    <TableCell className="text-right">
                      {fmt(sumEnergyMwh, 0)}
                    </TableCell>
                    <TableCell className="text-right">
                      {wan(sumEnergy)} 万
                    </TableCell>
                    <TableCell className="text-right">
                      {wan(sumCapacity)} 万
                    </TableCell>
                    <TableCell className="text-right">
                      {wan(sumAncillary)}{' '}万
                    </TableCell>
                    <TableCell className="text-right text-emerald-600">
                      {wan(sumSubsidy)} 万
                    </TableCell>
                    <TableCell className="text-right text-blue-600">
                      {wan(sumTotal)} 万
                    </TableCell>
                    <TableCell className="text-right">
                      {avgPrice.toFixed(2)}
                    </TableCell>
                    <TableCell />
                  </TableRow>
                )}
                {items.length === 0 && !isLoading && (
                  <TableRow>
                    <TableCell
                      colSpan={9}
                      className="text-center text-muted-foreground"
                    >
                      暂无月度结算数据
                    </TableCell>
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
