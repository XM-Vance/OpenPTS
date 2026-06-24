'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import dynamic from 'next/dynamic';
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
import { DemoBadge } from '@/components/feedback';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  listGreenPowerTrades,
  genGreenPowerDemo,
  type GreenPowerTrade,
} from '@/lib/api/green-power';
import {
  Leaf,
  TrendingUp,
  BarChart3,
  Droplets,
} from 'lucide-react';

const SELECT_CLASS =
  'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';
const fmt = (v: number) => v.toLocaleString('zh-CN', { maximumFractionDigits: 2 });
const fmtInt = (v: number) => v.toLocaleString('zh-CN', { maximumFractionDigits: 0 });

function statusVariant(s: string): 'default' | 'secondary' | 'destructive' | 'success' {
  if (s === 'completed') return 'success';
  if (s === 'settling') return 'default';
  if (s === 'pending') return 'secondary';
  return 'secondary';
}

const STATUS_LABEL: Record<string, string> = {
  completed: '已完成',
  settling: '结算中',
  pending: '待处理',
};

// Demo green cert price trend
function generateGreenCertPriceTrend() {
  const months = ['01', '02', '03', '04', '05', '06', '07', '08', '09', '10', '11', '12'];
  return months.map((m, i) => ({
    month: m,
    greenCert: 30 + Math.round(Math.sin(i * 0.5) * 8 + i * 1.2 + (Math.random() - 0.5) * 4),
    carbonPrice: 55 + Math.round(Math.cos(i * 0.4) * 10 + i * 0.8 + (Math.random() - 0.5) * 3),
  }));
}

// Demo green power ratio trend
function generateGreenPowerRatioTrend() {
  const months = ['01', '02', '03', '04', '05', '06', '07', '08', '09', '10', '11', '12'];
  return months.map((m, i) => ({
    month: m,
    greenRatio: Math.min(100, Math.round(15 + i * 3.5 + (Math.random() - 0.5) * 5)),
    totalEnergy: 8000 + Math.round(Math.random() * 2000),
    greenEnergy: 0, // will be computed
  })).map(d => ({ ...d, greenEnergy: Math.round(d.totalEnergy * d.greenRatio / 100) }));
}

// recharts 懒加载：从首屏 JS 剥离，挂载后异步加载（渲染内容不变）。
const GreenCertPriceLine = dynamic(
  () => import('./_charts').then((m) => ({ default: m.GreenCertPriceLine })),
  { ssr: false, loading: () => <div className="h-[300px] w-full" /> },
);
const GreenRatioComposed = dynamic(
  () => import('./_charts').then((m) => ({ default: m.GreenRatioComposed })),
  { ssr: false, loading: () => <div className="h-[300px] w-full" /> },
);

export default function GreenPowerPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('retail_management:write');

  const [status, setStatus] = useState('');
  const [days, setDays] = useState(30);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['green-power', status, days],
    queryFn: () => listGreenPowerTrades({ status: status || undefined, days }),
  });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genGreenPowerDemo();
      qc.invalidateQueries({ queryKey: ['green-power'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const items: GreenPowerTrade[] = data?.items ?? [];

  const totalEnergy = items.reduce((s, i) => s + i.energy_mwh, 0);
  const totalAmount = items.reduce((s, i) => s + i.amount, 0);
  const totalCerts = items.reduce((s, i) => s + i.green_cert_count, 0);
  const completedCount = items.filter((i) => i.status === 'completed').length;

  // Carbon reduction estimate: 0.5 tCO2 per MWh (approximate emission factor)
  const carbonReduction = useMemo(() => Math.round(totalEnergy * 0.5), [totalEnergy]);
  // Equivalent trees planted: ~22kg CO2 per tree per year
  const equivalentTrees = useMemo(() => Math.round(carbonReduction * 1000 / 22), [carbonReduction]);

  const greenCertPriceTrend = useMemo(() => generateGreenCertPriceTrend(), []);
  const greenRatioTrend = useMemo(() => generateGreenPowerRatioTrend(), []);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">绿电交易</h1>
          <p className="text-sm text-muted-foreground">
            绿色电力交易记录、绿证跟踪与环境价值分析
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={days}
            onChange={(e) => setDays(Number(e.target.value))}
            className={SELECT_CLASS}
          >
            <option value={7}>近 7 天</option>
            <option value={30}>近 30 天</option>
            <option value={60}>近 60 天</option>
            <option value={90}>近 90 天</option>
          </select>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className={SELECT_CLASS}
          >
            <option value="">全部状态</option>
            <option value="completed">已完成</option>
            <option value="settling">结算中</option>
            <option value="pending">待处理</option>
          </select>
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

      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">交易笔数</p>
            <p className="mt-1 text-2xl font-bold">{items.length}</p>
            <p className="mt-1 text-xs text-muted-foreground">
              已完成 {completedCount}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">总电量 (MWh)</p>
            <p className="mt-1 text-2xl font-bold text-emerald-600">
              {fmtInt(totalEnergy)}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">总金额 (元)</p>
            <p className="mt-1 text-2xl font-bold text-blue-600">
              {fmtInt(totalAmount)}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">绿证数量</p>
            <p className="mt-1 text-2xl font-bold">{totalCerts}</p>
          </CardContent>
        </Card>
      </div>

      {/* ── 碳减排量可视化卡片 ── */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card className="border-emerald-200 bg-emerald-50/50">
          <CardContent className="pt-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-xs text-emerald-700">碳减排量</p>
                <p className="mt-1 text-3xl font-bold text-emerald-700">
                  {carbonReduction.toLocaleString('zh-CN')}
                </p>
                <p className="mt-1 text-xs text-emerald-600">吨 CO₂</p>
              </div>
              <div className="rounded-full bg-emerald-100 p-3">
                <Leaf className="h-6 w-6 text-emerald-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="border-blue-200 bg-blue-50/50">
          <CardContent className="pt-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-xs text-blue-700">等效植树</p>
                <p className="mt-1 text-3xl font-bold text-blue-700">
                  {equivalentTrees.toLocaleString('zh-CN')}
                </p>
                <p className="mt-1 text-xs text-blue-600">棵/年</p>
              </div>
              <div className="rounded-full bg-blue-100 p-3">
                <Droplets className="h-6 w-6 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card className="border-violet-200 bg-violet-50/50">
          <CardContent className="pt-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-xs text-violet-700">绿证均价</p>
                <p className="mt-1 text-3xl font-bold text-violet-700">
                  {totalCerts > 0 ? (totalAmount / totalCerts).toFixed(1) : '0'}
                </p>
                <p className="mt-1 text-xs text-violet-600">元/张</p>
              </div>
              <div className="rounded-full bg-violet-100 p-3">
                <TrendingUp className="h-6 w-6 text-violet-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* ── 绿证价格走势折线图 ── */}
      <ChartContainer title="绿证价格与碳价走势" actions={<DemoBadge tooltip="绿证价/碳价为随机生成的演示趋势" />}>
        <GreenCertPriceLine data={greenCertPriceTrend} />
      </ChartContainer>

      {/* ── 绿电占比趋势 AreaChart ── */}
      <ChartContainer title="绿电占比趋势" actions={<DemoBadge tooltip="绿电占比/总量为随机生成的演示趋势" />}>
        <GreenRatioComposed data={greenRatioTrend} />
      </ChartContainer>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>交易日期</TableHead>
              <TableHead>产品</TableHead>
              <TableHead>交易对手</TableHead>
              <TableHead className="text-right">电量 (MWh)</TableHead>
              <TableHead className="text-right">价格 (元/MWh)</TableHead>
              <TableHead className="text-right">金额 (元)</TableHead>
              <TableHead className="text-right">绿证数</TableHead>
              <TableHead>状态</TableHead>
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
            {items.map((t) => (
              <TableRow key={t.id}>
                <TableCell className="font-medium">
                  {t.trade_date?.slice(0, 10)}
                </TableCell>
                <TableCell>{t.product_name}</TableCell>
                <TableCell>{t.counterparty}</TableCell>
                <TableCell className="text-right">{fmt(t.energy_mwh)}</TableCell>
                <TableCell className="text-right">{fmt(t.price)}</TableCell>
                <TableCell className="text-right text-blue-600">
                  {fmtInt(t.amount)}
                </TableCell>
                <TableCell className="text-right">{t.green_cert_count}</TableCell>
                <TableCell>
                  <Badge variant={statusVariant(t.status)}>
                    {STATUS_LABEL[t.status] ?? t.status}
                  </Badge>
                </TableCell>
              </TableRow>
            ))}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground">
                  暂无绿电交易数据{canWrite && '，可点右上「生成演示数据」'}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
