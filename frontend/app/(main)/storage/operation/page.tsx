'use client';

import { useEffect, useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Label } from '@/components/ui/label';
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
  generateStorageDemoData,
  listStorageOperations,
  listStorageStations,
} from '@/lib/api/storage';
import { Battery, TrendingUp, DollarSign, Zap } from 'lucide-react';

const FIELD_CLASS =
  'flex h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm';

function fmt(v?: number | null, digits = 2): string {
  return v == null ? '-' : v.toLocaleString('zh-CN', { maximumFractionDigits: digits });
}

const wan = (v: number) => (v / 10000).toLocaleString('zh-CN', { maximumFractionDigits: 2 });

// 生成模拟 SOC 24h 曲线数据
function generateSOCCurve() {
  const data = [];
  for (let h = 0; h < 24; h++) {
    // 模拟：夜间充电 SOC 上升，白天放电 SOC 下降
    let soc: number;
    if (h >= 0 && h < 6) soc = 20 + h * 10; // 充电：20->80
    else if (h >= 6 && h < 10) soc = 80 - (h - 6) * 5; // 缓慢放电
    else if (h >= 10 && h < 14) soc = 60 - (h - 10) * 10; // 快速放电
    else if (h >= 14 && h < 18) soc = 20 + (h - 14) * 15; // 再次充电
    else soc = 80 - (h - 18) * 10; // 晚间放电
    data.push({
      hour: `${String(h).padStart(2, '0')}:00`,
      soc: Math.max(5, Math.min(95, soc + (Math.random() - 0.5) * 5)),
    });
  }
  return data;
}

// 生成模拟充放电功率曲线（正负区域着色）
function generatePowerCurve() {
  const data = [];
  for (let h = 0; h < 24; h++) {
    let power: number;
    if (h >= 0 && h < 6) power = -5 - Math.random() * 5; // 充电（负值）
    else if (h >= 6 && h < 10) power = 2 + Math.random() * 3; // 放电（正值）
    else if (h >= 10 && h < 14) power = 8 + Math.random() * 4; // 峰时大功率放电
    else if (h >= 14 && h < 18) power = -6 - Math.random() * 4; // 充电
    else power = 4 + Math.random() * 3; // 晚高峰放电
    data.push({
      hour: `${String(h).padStart(2, '0')}:00`,
      power: Math.round(power * 100) / 100,
    });
  }
  return data;
}

// recharts 懒加载：从首屏 JS 剥离，挂载后异步加载（渲染内容不变）。
const SocArea = dynamic(
  () => import('./_charts').then((m) => ({ default: m.SocArea })),
  { ssr: false, loading: () => <div className="h-[280px] w-full" /> },
);
const PowerComposed = dynamic(
  () => import('./_charts').then((m) => ({ default: m.PowerComposed })),
  { ssr: false, loading: () => <div className="h-[280px] w-full" /> },
);
const DailyComposed = dynamic(
  () => import('./_charts').then((m) => ({ default: m.DailyComposed })),
  { ssr: false, loading: () => <div className="h-full w-full" /> },
);

export default function StorageOperationPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('storage:write');

  const [stationId, setStationId] = useState('');
  const [generating, setGenerating] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const { data: stationsData } = useQuery({
    queryKey: ['storage-stations'],
    queryFn: listStorageStations,
  });

  // 默认选第一个站点
  useEffect(() => {
    const items = stationsData?.items;
    if (!stationId && items && items.length > 0) {
      setStationId(items[0].id);
    }
  }, [stationsData, stationId]);

  const { data: opsData, isLoading: opsLoading } = useQuery({
    queryKey: ['storage-ops', stationId],
    queryFn: () => listStorageOperations(stationId, 30),
    enabled: !!stationId,
  });

  const onGenerate = async () => {
    setError(null);
    setNotice(null);
    setGenerating(true);
    try {
      const r = await generateStorageDemoData(30);
      setNotice(`已生成 ${r.stations} 个站点 × ${r.days} 天演示运营数据`);
      qc.invalidateQueries({ queryKey: ['storage-stations'] });
      qc.invalidateQueries({ queryKey: ['storage-ops'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setGenerating(false);
    }
  };

  const selectedStation = stationsData?.items.find((s) => s.id === stationId);
  const opsItems = opsData?.items ?? [];

  // 汇总统计
  const totalCharge = opsItems.reduce((s, o) => s + o.charge_mwh, 0);
  const totalDischarge = opsItems.reduce((s, o) => s + o.discharge_mwh, 0);
  const totalRevenue = opsItems.reduce((s, o) => s + (o.revenue ?? 0), 0);
  const avgSoc = opsItems.length > 0
    ? opsItems.reduce((s, o) => s + (o.avg_soc ?? 0), 0) / opsItems.length
    : 0;

  // 峰谷套利收益计算
  const peakValleyRevenue = useMemo(() => {
    // 简化计算：假设峰谷价差 0.5 元/kWh = 500 元/MWh，转换效率 90%
    const efficiency = 0.9;
    const priceDiff = 500;
    return totalDischarge * priceDiff * efficiency;
  }, [totalDischarge]);

  const chartData = opsItems
    .slice()
    .reverse()
    .map((o) => ({
      date: o.operation_date.slice(5, 10),
      充电: o.charge_mwh,
      放电: o.discharge_mwh,
      收益: o.revenue ?? 0,
    }));

  // SOC 实时监控曲线（演示数据：挂载时生成一次）
  const socCurve = useMemo(() => generateSOCCurve(), []);

  // 充放电功率曲线（演示数据：挂载时生成一次）
  const powerCurve = useMemo(() => generatePowerCurve(), []);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">储能管理</h1>
          <p className="text-sm text-muted-foreground">储能电站运营信息（充放电、收益、SOC）</p>
        </div>
        {canWrite && (
          <Button variant="outline" onClick={onGenerate} disabled={generating}>
            {generating ? '生成中...' : '生成演示数据'}
          </Button>
        )}
      </div>

      {notice && (
        <Alert>
          <AlertDescription>{notice}</AlertDescription>
        </Alert>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 储能收益计算卡片（峰谷套利） */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
        <StatCard
          title="总充电量 (MWh)"
          value={fmt(totalCharge, 1)}
          icon={<Zap className="h-4 w-4" />}
        />
        <StatCard
          title="总放电量 (MWh)"
          value={fmt(totalDischarge, 1)}
          icon={<TrendingUp className="h-4 w-4" />}
        />
        <StatCard
          title="实际总收益 (¥)"
          value={fmt(totalRevenue)}
          icon={<DollarSign className="h-4 w-4" />}
        />
        <StatCard
          title="峰谷套利收益 (¥)"
          value={fmt(peakValleyRevenue)}
          icon={<Battery className="h-4 w-4" />}
          trend={totalRevenue > 0 && peakValleyRevenue > 0 ? ((totalRevenue / peakValleyRevenue - 1) * 100) : undefined}
          trendLabel="vs 理论值"
        />
      </div>

      {/* 站点选择 */}
      {(stationsData?.items.length ?? 0) > 0 ? (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">站点</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-3">
              <div className="space-y-2">
                <Label>当前站点</Label>
                <select
                  value={stationId}
                  onChange={(e) => setStationId(e.target.value)}
                  className={FIELD_CLASS}
                >
                  {stationsData?.items.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </select>
              </div>
              {selectedStation && (
                <>
                  <StatMini
                    label="额定容量 (MWh)"
                    value={fmt(selectedStation.capacity_mwh, 0)}
                  />
                  <StatMini
                    label="最大功率 (MW)"
                    value={fmt(selectedStation.max_power_mw, 0)}
                  />
                </>
              )}
            </div>
            {selectedStation?.location && (
              <p className="mt-3 text-sm text-muted-foreground">
                所在地：{selectedStation.location} · 状态：
                <Badge
                  variant={selectedStation.status === 'active' ? 'success' : 'secondary'}
                  className="ml-1"
                >
                  {selectedStation.status}
                </Badge>
              </p>
            )}
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground">
              暂无储能站点{canWrite ? '，请点右上「生成演示数据」' : ''}
            </p>
          </CardContent>
        </Card>
      )}

      {/* SOC 曲线实时监控折线图 */}
      <ChartContainer title="SOC 实时监控曲线（%）" actions={<DemoBadge tooltip="SOC 曲线为随机生成的演示数据，非实时采集" />}>
        <SocArea data={socCurve} />
      </ChartContainer>

      {/* 充放电功率曲线（正负区域着色） */}
      <ChartContainer title="充放电功率曲线（正=放电，负=充电）" actions={<DemoBadge tooltip="充放电功率曲线为随机生成的演示数据" />}>
        <PowerComposed data={powerCurve} />
        <p className="mt-1 text-xs text-muted-foreground px-2">
          🟠 正值区域：放电功率 | 🔵 负值区域：充电功率
        </p>
      </ChartContainer>

      {/* 最近 30 日充放电与收益 */}
      {chartData.length > 0 && (
        <ChartContainer title="最近 30 日充放电与收益">
          <div className="h-80 w-full [&_.recharts-surface:focus]:outline-none">
            <DailyComposed data={chartData} />
          </div>
          <p className="mt-2 text-xs text-muted-foreground">
            左轴：充放电量 (MWh)；右轴：日收益 (¥)
          </p>
        </ChartContainer>
      )}

      {stationId && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">最近 30 日运营明细</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>日期</TableHead>
                    <TableHead className="text-right">充电 (MWh)</TableHead>
                    <TableHead className="text-right">放电 (MWh)</TableHead>
                    <TableHead className="text-right">日收益 (¥)</TableHead>
                    <TableHead className="text-right">平均 SOC (%)</TableHead>
                    <TableHead className="text-right">循环次数</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {opsLoading && (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center text-muted-foreground">
                        加载中...
                      </TableCell>
                    </TableRow>
                  )}
                  {opsItems.map((o) => (
                    <TableRow key={o.id}>
                      <TableCell className="font-medium">
                        {o.operation_date.slice(0, 10)}
                      </TableCell>
                      <TableCell className="text-right">{fmt(o.charge_mwh)}</TableCell>
                      <TableCell className="text-right">{fmt(o.discharge_mwh)}</TableCell>
                      <TableCell className="text-right font-medium">
                        {fmt(o.revenue)}
                      </TableCell>
                      <TableCell className="text-right">{fmt(o.avg_soc, 1)}</TableCell>
                      <TableCell className="text-right">{fmt(o.cycles)}</TableCell>
                    </TableRow>
                  ))}
                  {opsItems.length === 0 && !opsLoading && (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center text-muted-foreground">
                        该站点暂无运营数据
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function StatMini({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border bg-muted/30 p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 text-xl font-semibold">{value}</p>
    </div>
  );
}
