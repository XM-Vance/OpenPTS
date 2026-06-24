'use client';

import { useState, useEffect, useMemo } from 'react';
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
  genStorageDeclDemo,
  listStorageDeclarations,
} from '@/lib/api/storage-declaration';
import { Clock, TrendingUp, DollarSign, Battery } from 'lucide-react';

const wan = (v: number) => (v / 10000).toLocaleString('zh-CN', { maximumFractionDigits: 2 });
const fmt = (v: number, d = 2) => v.toLocaleString('zh-CN', { maximumFractionDigits: d });

// 申报截止倒计时组件
function DeclarationCountdown() {
  const [timeLeft, setTimeLeft] = useState({ days: 0, hours: 0, minutes: 0, seconds: 0 });

  useEffect(() => {
    // 假设截止日期为当月最后一天 17:00
    const getDeadline = () => {
      const now = new Date();
      const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0);
      return new Date(lastDay.getFullYear(), lastDay.getMonth(), lastDay.getDate(), 17, 0, 0);
    };

    const update = () => {
      const now = new Date();
      const deadline = getDeadline();
      const diff = Math.max(0, deadline.getTime() - now.getTime());

      if (diff === 0) {
        setTimeLeft({ days: 0, hours: 0, minutes: 0, seconds: 0 });
        return;
      }

      const days = Math.floor(diff / (1000 * 60 * 60 * 24));
      const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
      const seconds = Math.floor((diff % (1000 * 60)) / 1000);

      setTimeLeft({ days, hours, minutes, seconds });
    };

    update();
    const timer = setInterval(update, 1000);
    return () => clearInterval(timer);
  }, []);

  const isUrgent = timeLeft.days < 3;

  return (
    <Card className={isUrgent ? 'border-destructive' : ''}>
      <CardContent className="pt-6">
        <div className="flex items-center gap-2">
          <Clock className={`h-5 w-5 ${isUrgent ? 'text-destructive' : 'text-muted-foreground'}`} />
          <p className="text-xs text-muted-foreground">申报截止倒计时</p>
        </div>
        <div className="mt-2 flex items-baseline gap-1">
          {timeLeft.days > 0 && (
            <span className={`text-3xl font-bold ${isUrgent ? 'text-destructive' : ''}`}>
              {timeLeft.days}
              <span className="text-sm font-normal text-muted-foreground ml-0.5">天</span>
            </span>
          )}
          <span className={`text-3xl font-bold font-mono ${isUrgent ? 'text-destructive' : ''}`}>
            {String(timeLeft.hours).padStart(2, '0')}:{String(timeLeft.minutes).padStart(2, '0')}:{String(timeLeft.seconds).padStart(2, '0')}
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          截止：本月末 17:00
        </p>
        {isUrgent && (
          <Badge variant="destructive" className="mt-2">即将截止</Badge>
        )}
      </CardContent>
    </Card>
  );
}

// recharts 懒加载：从首屏 JS 剥离，挂载后异步加载（渲染内容不变）。
const CapacityComposed = dynamic(
  () => import('./_charts').then((m) => ({ default: m.CapacityComposed })),
  { ssr: false, loading: () => <div className="h-[280px] w-full" /> },
);
const RevenueBar = dynamic(
  () => import('./_charts').then((m) => ({ default: m.RevenueBar })),
  { ssr: false, loading: () => <div className="h-[200px] w-full" /> },
);
const ChargeDischargeBar = dynamic(
  () => import('./_charts').then((m) => ({ default: m.ChargeDischargeBar })),
  { ssr: false, loading: () => <div className="h-full w-full" /> },
);

export default function StorageDeclarationPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('storage:write');

  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['storage-declarations'],
    queryFn: () => listStorageDeclarations({ days: 7 }),
  });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genStorageDeclDemo();
      qc.invalidateQueries({ queryKey: ['storage-declarations'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const items = useMemo(() => data?.items ?? [], [data]);
  // 自动选中首项：放 useEffect 里（渲染期内 setTimeout setState 是反模式，会反复排队 + 卸载后仍触发）
  useEffect(() => {
    if (!selected && items.length > 0) setSelected(items[0].id);
  }, [selected, items]);

  const sel = items.find((i) => i.id === selected);

  // 96 时段充放电曲线
  const chartData = sel
    ? sel.charge_mw.map((c, i) => ({
        period: `${String(Math.floor(i / 4)).padStart(2, '0')}:${String((i % 4) * 15 || 0).padStart(2, '0')}`.slice(0, 5),
        charge: -c,
        discharge: sel.discharge_mw[i] ?? 0,
      }))
    : [];

  const totalRev = items.reduce((s, i) => s + i.expected_revenue, 0);

  // 申报 vs 实际可用容量对比数据
  const capacityCompareData = useMemo(
    () =>
      items.map((i) => {
        // 模拟实际可用容量 = 申报容量 × (85%-105%)
        const declaredCapacity = i.discharge_mw.reduce((s, v) => s + v, 0) / 4;
        const actualFactor = 0.85 + Math.random() * 0.2;
        const actualCapacity = declaredCapacity * actualFactor;
        return {
          date: i.declared_date.slice(5, 10),
          station: i.station_name,
          declared: Math.round(declaredCapacity * 10) / 10,
          actual: Math.round(actualCapacity * 10) / 10,
          utilization: Math.round(actualFactor * 10000) / 100,
        };
      }),
    [items],
  );

  // 日收益趋势
  const revenueTrend = useMemo(
    () =>
      items
        .slice()
        .reverse()
        .map((i) => ({
          date: i.declared_date.slice(5, 10),
          expected: i.expected_revenue / 10000,
        })),
    [items],
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">储能申报策略</h1>
          <p className="text-sm text-muted-foreground">
            96 时段充放电策略 + 预测收益（选择条目查看曲线）
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

      {/* 申报收益预估卡片 + 倒计时 */}
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard
          title="7 日累计申报"
          value={`${items.length} 条`}
          icon={<Battery className="h-4 w-4" />}
        />
        <StatCard
          title="累计预测收益（万元）"
          value={wan(totalRev)}
          icon={<DollarSign className="h-4 w-4" />}
          trend={items.length >= 2 && items[0].expected_revenue > 0 && items[1].expected_revenue > 0
            ? ((items[0].expected_revenue / items[1].expected_revenue - 1) * 100)
            : undefined
          }
          trendLabel="日环比"
        />
        <StatCard
          title="日均收益（万元）"
          value={items.length === 0 ? '0' : wan(totalRev / items.length)}
          icon={<TrendingUp className="h-4 w-4" />}
        />
        <DeclarationCountdown />
      </div>

      {/* 申报 vs 实际可用容量对比图 */}
      <ChartContainer title="申报 vs 实际可用容量对比 (MWh)" actions={<DemoBadge tooltip="实际可用容量含随机系数，非真实采集" />}>
        {capacityCompareData.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            暂无数据{canWrite && '，可点右上「生成演示数据」'}
          </p>
        ) : (
          <CapacityComposed data={capacityCompareData} />
        )}
      </ChartContainer>

      {/* 收益预估趋势 */}
      <ChartContainer title="申报收益预估趋势（万元）">
        {revenueTrend.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无数据</p>
        ) : (
          <RevenueBar data={revenueTrend} />
        )}
      </ChartContainer>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">
            {sel ? `${sel.station_name} · ${sel.declared_date.slice(0, 10)} 充放电曲线` : '充放电曲线'}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {!sel ? (
            <p className="text-sm text-muted-foreground">请从下方表格选择一条申报</p>
          ) : (
            <div
              className="[&_.recharts-surface:focus]:outline-none"
              style={{ width: '100%', height: 280 }}
            >
              <ChargeDischargeBar data={chartData} />
            </div>
          )}
        </CardContent>
      </Card>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>申报日</TableHead>
              <TableHead>站点</TableHead>
              <TableHead className="text-right">预测收益（万元）</TableHead>
              <TableHead>策略</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            )}
            {items.map((i) => (
              <TableRow
                key={i.id}
                className={
                  selected === i.id
                    ? 'cursor-pointer bg-amber-50 ring-1 ring-amber-300'
                    : 'cursor-pointer hover:bg-muted/50'
                }
                onClick={() => setSelected(i.id)}
              >
                <TableCell className="font-medium">{i.declared_date.slice(0, 10)}</TableCell>
                <TableCell>{i.station_name}</TableCell>
                <TableCell className="text-right text-emerald-600">
                  {wan(i.expected_revenue)}
                </TableCell>
                <TableCell>
                  <Badge variant="outline">{i.strategy_note ?? '-'}</Badge>
                </TableCell>
              </TableRow>
            ))}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-muted-foreground">
                  暂无申报数据{canWrite && '，可点右上「生成演示数据」'}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
