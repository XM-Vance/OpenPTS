'use client';

import { useState, useMemo, useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
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
import { ChartContainer } from '@/components/charts/chart-container';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import { genPreSettleDemo, listPreSettle, type PreSettleDaily } from '@/lib/api/pre-settle';

const fmt = (v: number) => v.toLocaleString('zh-CN', { maximumFractionDigits: 1 });
const wan = (v: number) => (v / 10000).toLocaleString('zh-CN', { maximumFractionDigits: 2 });
const DEVIATION_THRESHOLD = 0.05; // 5% 阈值

export default function PreSettlementPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('settlement_management:write');

  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<PreSettleDaily | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['pre-settle'],
    queryFn: () => listPreSettle(14),
  });

  const items = useMemo(() => data?.items ?? [], [data]);
  // 自动选中首项：放 useEffect 里（渲染期内 setTimeout setState 是反模式）
  useEffect(() => {
    if (!selected && items.length > 0) setSelected(items[0]);
  }, [selected, items]);

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genPreSettleDemo();
      qc.invalidateQueries({ queryKey: ['pre-settle'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  // 96 点曲线对比数据
  const chartData = selected
    ? selected.declared_curve_96.map((d, i) => ({
        period: `${String(Math.floor(i / 4)).padStart(2, '0')}:${String((i % 4) * 15 || 0).padStart(2, '0')}`.slice(0, 5),
        declared: d,
        cleared: selected.cleared_curve_96[i] ?? 0,
        price: selected.spot_price_96[i] ?? 0,
      }))
    : [];

  // 预结算 vs 最终结算差异对比数据（日维度）
  const comparisonData = useMemo(
    () =>
      items
        .slice()
        .reverse()
        .map((it) => {
          const preSettlement = it.energy_revenue;
          const finalSettlement = it.final_amount;
          const diff = finalSettlement - preSettlement;
          const diffRate = preSettlement !== 0 ? (diff / preSettlement) * 100 : 0;
          return {
            date: it.operating_date.slice(5, 10),
            pre: preSettlement / 10000,
            final: finalSettlement / 10000,
            diffRate: Math.round(diffRate * 100) / 100,
            isOverThreshold: Math.abs(it.deviation_ratio) > DEVIATION_THRESHOLD,
          };
        }),
    [items],
  );

  // 预结算准确率趋势数据
  const accuracyData = useMemo(
    () =>
      items
        .slice()
        .reverse()
        .map((it) => {
          const accuracy =
            it.total_declared > 0
              ? (1 - Math.abs(it.total_declared - it.total_cleared) / it.total_declared) * 100
              : 100;
          return {
            date: it.operating_date.slice(5, 10),
            accuracy: Math.round(accuracy * 100) / 100,
            deviation: Math.round(Math.abs(it.deviation_ratio) * 10000) / 100,
          };
        }),
    [items],
  );

  // 偏差预警条目
  const warningItems = items.filter((it) => Math.abs(it.deviation_ratio) > DEVIATION_THRESHOLD);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">预结算明细</h1>
          <p className="text-sm text-muted-foreground">
            日维度 96 点对账：申报 vs 出清 + 现货价格 + 偏差/收益
          </p>
        </div>
        {canWrite && (
          <Button variant="outline" onClick={onGen} disabled={busy}>
            {busy ? '生成中...' : '生成演示数据'}
          </Button>
        )}
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">申报电量 (MWh)</p>
            <p className="mt-1 text-2xl font-bold">{selected ? fmt(selected.total_declared) : '-'}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">出清电量 (MWh)</p>
            <p className="mt-1 text-2xl font-bold text-blue-600">
              {selected ? fmt(selected.total_cleared) : '-'}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">偏差率</p>
            <p
              className={`mt-1 text-2xl font-bold ${
                selected && Math.abs(selected.deviation_ratio) > DEVIATION_THRESHOLD
                  ? 'text-destructive'
                  : 'text-emerald-600'
              }`}
            >
              {selected ? (selected.deviation_ratio * 100).toFixed(2) + '%' : '-'}
              {selected && Math.abs(selected.deviation_ratio) > DEVIATION_THRESHOLD && (
                <Badge variant="destructive" className="ml-2 text-xs">超阈值</Badge>
              )}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">最终结算（万元）</p>
            <p className="mt-1 text-2xl font-bold text-amber-600">
              {selected ? wan(selected.final_amount) : '-'}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 偏差预警面板 */}
      {warningItems.length > 0 && (
        <Alert variant="destructive">
          <AlertDescription>
            <span className="font-semibold">偏差预警：</span>
            以下日期偏差率超过 {DEVIATION_THRESHOLD * 100}% 阈值：{' '}
            {warningItems.map((it) => (
              <Badge key={it.id} variant="destructive" className="mr-1 text-xs">
                {it.operating_date.slice(0, 10)} ({(it.deviation_ratio * 100).toFixed(1)}%)
              </Badge>
            ))}
          </AlertDescription>
        </Alert>
      )}

      {/* 预结算 vs 最终结算差异对比双线图 */}
      <ChartContainer title="预结算 vs 最终结算差异对比">
        {comparisonData.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            {canWrite ? '请点右上「生成演示数据」' : '暂无数据'}
          </p>
        ) : (
          <ResponsiveContainer width="100%" height={300}>
            <ComposedChart
              data={comparisonData}
              margin={{ top: 8, right: 30, left: 0, bottom: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
              <YAxis
                yAxisId="amount"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={70}
              />
              <YAxis
                yAxisId="rate"
                orientation="right"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={60}
                unit="%"
              />
              <Tooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(v: number, name: string) => {
                  if (name === '差异率') return `${v.toFixed(2)}%`;
                  return `${v.toFixed(2)} 万`;
                }}
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Line
                yAxisId="amount"
                type="monotone"
                dataKey="pre"
                name="预结算"
                stroke="#3b82f6"
                strokeWidth={2}
                dot={{ r: 3 }}
                isAnimationActive={false}
              />
              <Line
                yAxisId="amount"
                type="monotone"
                dataKey="final"
                name="最终结算"
                stroke="#f59e0b"
                strokeWidth={2}
                dot={(props: Record<string, unknown>) => {
                  const { cx, cy, payload } = props as { cx: number; cy: number; payload: { isOverThreshold: boolean } };
                  return (
                    <circle
                      key={`dot-${cx}-${cy}`}
                      cx={cx}
                      cy={cy}
                      r={payload.isOverThreshold ? 5 : 3}
                      fill={payload.isOverThreshold ? '#ef4444' : '#f59e0b'}
                      stroke={payload.isOverThreshold ? '#ef4444' : '#f59e0b'}
                    />
                  );
                }}
                isAnimationActive={false}
              />
              <Line
                yAxisId="rate"
                type="monotone"
                dataKey="diffRate"
                name="差异率"
                stroke="#94a3b8"
                strokeWidth={1.5}
                strokeDasharray="5 5"
                dot={false}
                isAnimationActive={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* 预结算准确率趋势折线图 */}
      <ChartContainer title="预结算准确率趋势">
        {accuracyData.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={260}>
            <ComposedChart
              data={accuracyData}
              margin={{ top: 8, right: 30, left: 0, bottom: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
              <YAxis
                yAxisId="acc"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={60}
                unit="%"
                domain={[80, 100]}
              />
              <YAxis
                yAxisId="dev"
                orientation="right"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={60}
                unit="%"
              />
              <Tooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(v: number, name: string) => `${v.toFixed(2)}%`}
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Line
                yAxisId="acc"
                type="monotone"
                dataKey="accuracy"
                name="准确率"
                stroke="#10b981"
                strokeWidth={2}
                dot={{ r: 3 }}
                isAnimationActive={false}
              />
              <Line
                yAxisId="dev"
                type="monotone"
                dataKey="deviation"
                name="偏差率"
                stroke="#ef4444"
                strokeWidth={2}
                dot={false}
                strokeDasharray="5 5"
                isAnimationActive={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">
            96 点对比 · {selected ? selected.operating_date.slice(0, 10) : '请选择日期'}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {chartData.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {canWrite ? '请点右上「生成演示数据」' : '暂无数据'}
            </p>
          ) : (
            <div
              className="[&_.recharts-surface:focus]:outline-none"
              style={{ width: '100%', height: 300 }}
            >
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData} margin={{ top: 8, right: 30, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="period" tick={{ fontSize: 10, fill: '#6b7280' }} interval={11} />
                  <YAxis
                    yAxisId="left"
                    tick={{ fontSize: 12, fill: '#6b7280' }}
                    width={60}
                    unit=" MW"
                  />
                  <YAxis
                    yAxisId="right"
                    orientation="right"
                    tick={{ fontSize: 12, fill: '#6b7280' }}
                    width={60}
                    unit=" 元"
                  />
                  <Tooltip contentStyle={{ fontSize: 12 }} />
                  <Line
                    yAxisId="left"
                    type="monotone"
                    dataKey="declared"
                    name="申报 MW"
                    stroke="#94a3b8"
                    strokeWidth={2}
                    dot={false}
                    isAnimationActive={false}
                  />
                  <Line
                    yAxisId="left"
                    type="monotone"
                    dataKey="cleared"
                    name="出清 MW"
                    stroke="#2563eb"
                    strokeWidth={2}
                    dot={false}
                    isAnimationActive={false}
                  />
                  <Line
                    yAxisId="right"
                    type="monotone"
                    dataKey="price"
                    name="现货价 元/MWh"
                    stroke="#f59e0b"
                    strokeWidth={2}
                    dot={false}
                    isAnimationActive={false}
                    strokeDasharray="5 5"
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </CardContent>
      </Card>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>运行日</TableHead>
              <TableHead className="text-right">申报 (MWh)</TableHead>
              <TableHead className="text-right">出清 (MWh)</TableHead>
              <TableHead className="text-right">偏差</TableHead>
              <TableHead className="text-right">偏差率</TableHead>
              <TableHead className="text-right">电费收益（万元）</TableHead>
              <TableHead className="text-right">偏差罚款（万元）</TableHead>
              <TableHead className="text-right">最终（万元）</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            )}
            {items.map((it) => (
              <TableRow
                key={it.id}
                className={
                  selected?.id === it.id
                    ? 'cursor-pointer bg-amber-50 ring-1 ring-amber-300'
                    : 'cursor-pointer hover:bg-muted/50'
                }
                onClick={() => setSelected(it)}
              >
                <TableCell className="font-medium">{it.operating_date.slice(0, 10)}</TableCell>
                <TableCell className="text-right">{fmt(it.total_declared)}</TableCell>
                <TableCell className="text-right">{fmt(it.total_cleared)}</TableCell>
                <TableCell className="text-right">{fmt(it.total_deviation)}</TableCell>
                <TableCell
                  className={`text-right ${
                    Math.abs(it.deviation_ratio) > DEVIATION_THRESHOLD ? 'text-destructive font-bold' : ''
                  }`}
                >
                  {(it.deviation_ratio * 100).toFixed(2)}%
                  {Math.abs(it.deviation_ratio) > DEVIATION_THRESHOLD && (
                    <Badge variant="destructive" className="ml-1 text-[10px] px-1">!</Badge>
                  )}
                </TableCell>
                <TableCell className="text-right">{wan(it.energy_revenue)}</TableCell>
                <TableCell className="text-right text-destructive">
                  {wan(it.deviation_penalty)}
                </TableCell>
                <TableCell className="text-right font-bold text-amber-600">
                  {wan(it.final_amount)}
                </TableCell>
              </TableRow>
            ))}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground">
                  暂无预结算数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
