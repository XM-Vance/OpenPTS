'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { RetailTabs } from '@/components/retail/retail-tabs';
import { ChartContainer } from '@/components/charts/chart-container';
import { DemoBadge } from '@/components/feedback';
import { PriceHeatmap } from '@/components/charts/price-heatmap';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import { listContracts } from '@/lib/api/retail';
import {
  genContractPriceDemo,
  listContractPriceDaily,
} from '@/lib/api/contract-price';

const SELECT_CLASS =
  'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';
const fmt = (v: number) => v.toLocaleString('zh-CN', { maximumFractionDigits: 1 });
const wan = (v: number) => (v / 10000).toLocaleString('zh-CN', { maximumFractionDigits: 2 });

export default function ContractPriceDailyPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('retail_management:write');

  const [contractId, setContractId] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data: contracts } = useQuery({
    queryKey: ['retail-contracts', ''],
    queryFn: () => listContracts({ keyword: '' }),
  });

  const { data, isLoading } = useQuery({
    queryKey: ['contract-price-daily', contractId],
    queryFn: () => listContractPriceDaily({ contract_id: contractId || undefined, days: 30 }),
  });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genContractPriceDemo();
      qc.invalidateQueries({ queryKey: ['contract-price-daily'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const items = useMemo(() => data?.items ?? [], [data]);

  // 折线图：选定合同时按日均价；未选时聚合所有合同日均价
  const trend = (() => {
    const byDate: Record<string, { sum: number; count: number }> = {};
    for (const p of items) {
      const d = p.price_date.slice(0, 10);
      if (!byDate[d]) byDate[d] = { sum: 0, count: 0 };
      byDate[d].sum += p.unit_price;
      byDate[d].count += 1;
    }
    return Object.entries(byDate)
      .map(([date, { sum, count }]) => ({
        date: date.slice(5).replace('-', '/'),
        price: sum / count,
      }))
      .sort((a, b) => a.date.localeCompare(b.date));
  })();

  const totalEnergy = items.reduce((s, i) => s + i.daily_energy, 0);
  const totalAmount = items.reduce((s, i) => s + i.daily_amount, 0);

  // ── 热力图数据：24h × 最近7天 ──
  const heatmapData = useMemo(() => {
    if (!items.length) return [];
    // 为每个日期生成24h模拟电价（基于当日均价浮动）
    const dateMap = new Map<string, number>();
    for (const p of items) {
      const d = p.price_date.slice(0, 10);
      if (!dateMap.has(d)) dateMap.set(d, p.unit_price);
    }
    const dates = Array.from(dateMap.keys()).sort().slice(-7);
    const cells: Array<{ date: string; hour: number; value: number }> = [];
    for (const d of dates) {
      const base = dateMap.get(d) ?? 300;
      for (let h = 0; h < 24; h++) {
        // 模拟分时：峰(8-11,17-21) 平(7,12-16) 谷(0-6,22-23)
        let factor: number;
        if (h >= 8 && h <= 11) factor = 1.3 + Math.random() * 0.1;
        else if (h >= 17 && h <= 21) factor = 1.25 + Math.random() * 0.1;
        else if (h === 7 || (h >= 12 && h <= 16)) factor = 1.0 + Math.random() * 0.05;
        else factor = 0.65 + Math.random() * 0.08;
        cells.push({ date: d, hour: h, value: +(base * factor).toFixed(2) });
      }
    }
    return cells;
  }, [items]);

  // ── 合同价 vs 现货价对比图 ──
  const comparisonData = useMemo(() => {
    if (!trend.length) return [];
    return trend.map((t, idx) => {
      // 模拟现货价格：围绕合同价波动 ±15%
      const spotBase = t.price * (0.85 + Math.sin(idx * 0.5) * 0.15);
      const spotNoise = (Math.random() - 0.5) * t.price * 0.1;
      return {
        date: t.date,
        合同价: +t.price.toFixed(2),
        现货价: +(spotBase + spotNoise).toFixed(2),
      };
    });
  }, [trend]);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">零售管理</h1>
        <p className="text-sm text-muted-foreground">合同日维度电价与累计执行</p>
      </div>
      <RetailTabs />

      <div className="flex items-center justify-between pt-2">
        <select
          value={contractId}
          onChange={(e) => setContractId(e.target.value)}
          className={SELECT_CLASS}
        >
          <option value="">全部合同（聚合）</option>
          {contracts?.map((c) => (
            <option key={c.id} value={c.id}>
              {c.customer_name} · {c.package_name_snapshot}
            </option>
          ))}
        </select>
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

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">30 日累计电量 (MWh)</p>
            <p className="mt-1 text-2xl font-bold">{fmt(totalEnergy)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">30 日累计金额（万元）</p>
            <p className="mt-1 text-2xl font-bold text-blue-600">{wan(totalAmount)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">条目数</p>
            <p className="mt-1 text-2xl font-bold">{items.length}</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        {/* ── 分时电价热力图 ── */}
        <ChartContainer title="分时电价热力图（24h × 最近7天）" minHeight={240} actions={<DemoBadge tooltip="分时电价为含随机系数的模拟数据" />}>
          {heatmapData.length > 0 ? (
            <PriceHeatmap data={heatmapData} />
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              暂无数据{canWrite && '，可点右上「生成演示数据」'}
            </div>
          )}
        </ChartContainer>

        {/* ── 合同价 vs 现货价对比图 ── */}
        <ChartContainer title="合同价 vs 现货价对比（元/MWh）" minHeight={240} actions={<DemoBadge tooltip="现货价含随机系数，非真实现货数据" />}>
          {comparisonData.length > 0 ? (
            <div
              className="[&_.recharts-surface:focus]:outline-none"
              style={{ width: '100%', height: '100%' }}
            >
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={comparisonData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#6b7280' }} />
                  <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} />
                  <RechartsTooltip
                    formatter={(v: number) => `${fmt(v)} 元/MWh`}
                    contentStyle={{ fontSize: 12 }}
                  />
                  <Legend wrapperStyle={{ fontSize: 12 }} />
                  <Line
                    type="monotone"
                    dataKey="合同价"
                    stroke="#2563eb"
                    strokeWidth={2}
                    dot={false}
                    isAnimationActive={false}
                  />
                  <Line
                    type="monotone"
                    dataKey="现货价"
                    stroke="#f59e0b"
                    strokeWidth={2}
                    dot={false}
                    strokeDasharray="5 3"
                    isAnimationActive={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              暂无数据
            </div>
          )}
        </ChartContainer>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">日均单价走势（元/MWh）</CardTitle>
        </CardHeader>
        <CardContent>
          {trend.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              暂无数据{canWrite && '，可点右上「生成演示数据」'}
            </p>
          ) : (
            <div
              className="[&_.recharts-surface:focus]:outline-none"
              style={{ width: '100%', height: 220 }}
            >
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={trend} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#6b7280' }} />
                  <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} />
                  <RechartsTooltip
                    formatter={(v: number) => `${fmt(v)} 元/MWh`}
                    contentStyle={{ fontSize: 12 }}
                  />
                  <Line
                    type="monotone"
                    dataKey="price"
                    stroke="#2563eb"
                    strokeWidth={2}
                    dot={false}
                    isAnimationActive={false}
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
              <TableHead>日期</TableHead>
              <TableHead className="text-right">单价 (元/MWh)</TableHead>
              <TableHead className="text-right">日电量 (MWh)</TableHead>
              <TableHead className="text-right">日金额（万元）</TableHead>
              <TableHead className="text-right">累计电量 (MWh)</TableHead>
              <TableHead className="text-right">累计金额（万元）</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            )}
            {items.map((p) => (
              <TableRow key={p.id}>
                <TableCell className="font-medium">{p.price_date.slice(0, 10)}</TableCell>
                <TableCell className="text-right">{fmt(p.unit_price)}</TableCell>
                <TableCell className="text-right">{fmt(p.daily_energy)}</TableCell>
                <TableCell className="text-right text-blue-600">{wan(p.daily_amount)}</TableCell>
                <TableCell className="text-right">{fmt(p.cumulative_energy)}</TableCell>
                <TableCell className="text-right">{wan(p.cumulative_amount)}</TableCell>
              </TableRow>
            ))}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  暂无合同电价数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
