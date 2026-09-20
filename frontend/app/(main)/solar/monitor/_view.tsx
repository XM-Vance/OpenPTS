'use client';

import { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { EmptyState, DemoBadge } from '@/components/feedback';
import { Label } from '@/components/ui/label';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { MobileCardList, type MobileCardField, type MobileRow } from '@/components/data-display/mobile-card-list';
import { useIsMobile } from '@/hooks/use-mobile';
import { listSolarStations, type SolarStation } from '@/lib/api/solar';

const FIELD_CLASS =
  'flex h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm';

function fmt(v?: number | null, digits = 2): string {
  return v == null ? '-' : v.toLocaleString('zh-CN', { maximumFractionDigits: digits });
}

const PIE_COLORS = ['#f59e0b', '#10b981', '#3b82f6', '#8b5cf6', '#ef4444', '#06b6d4'];

// 模拟站点实时监控数据
interface MonitorRow {
  station: SolarStation;
  currentPowerKw: number;
  dailyKwh: number;
  efficiency: number;
  irradiance: number;
  panelTemp: number;
  dropAlert: boolean;
}

// 按站点 id + 字段盐值派生稳定伪随机数（0~1）：演示数据不随重渲染/重新挂载跳变。
function seededRand(seed: string, salt: number): number {
  let h = 2166136261 ^ salt;
  for (let i = 0; i < seed.length; i++) {
    h ^= seed.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return ((h >>> 0) % 10000) / 10000;
}

function genMockMonitor(stations: SolarStation[]): MonitorRow[] {
  return stations.map((s) => {
    const hour = new Date().getHours();
    const isDay = hour >= 6 && hour <= 18;
    const cap = s.capacity_kw;
    const current = isDay ? +(cap * (0.3 + 0.5 * seededRand(s.id, 1))).toFixed(1) : 0;
    // 模拟出力骤降告警
    const dropAlert = seededRand(s.id, 2) < 0.15 && isDay;
    const actualPower = dropAlert ? +(current * 0.3).toFixed(1) : current;
    return {
      station: s,
      currentPowerKw: actualPower,
      dailyKwh: +(cap * (1 + 4 * seededRand(s.id, 3))).toFixed(0),
      efficiency: +(85 + 10 * seededRand(s.id, 4)).toFixed(1),
      irradiance: isDay ? +(400 + 500 * seededRand(s.id, 5)).toFixed(0) : 0,
      panelTemp: +(25 + 20 * seededRand(s.id, 6)).toFixed(1),
      dropAlert,
    };
  });
}

export default function SolarMonitorPage() {
  const [stationId, setStationId] = useState('');

  const { data: stationsData } = useQuery({
    queryKey: ['solar-stations'],
    queryFn: listSolarStations,
  });

  useEffect(() => {
    const items = stationsData?.items;
    if (!stationId && items && items.length > 0) {
      setStationId(items[0].id);
    }
  }, [stationsData, stationId]);

  const stations = useMemo(() => stationsData?.items ?? [], [stationsData]);
  // 演示数据按站点列表 memo：避免 hover/setState 等重渲染时整表重新随机。
  const monitorData = useMemo(() => genMockMonitor(stations), [stations]);
  const selected = monitorData.find((m) => m.station.id === stationId);

  // 状态统计
  const statusCounts = {
    active: stations.filter((s) => s.status === 'active').length,
    maintenance: stations.filter((s) => s.status === 'maintenance').length,
    offline: stations.filter((s) => s.status === 'offline').length,
  };
  const totalCapacity = stations.reduce((a, s) => a + s.capacity_kw, 0);
  const totalCurrentPower = monitorData.reduce((a, m) => a + m.currentPowerKw, 0);

  // ── Chart 1: 电站地区分布 PieChart（代替地图）──
  const regionData = (() => {
    const regionMap = new Map<string, number>();
    for (const s of stations) {
      const region = s.location?.split(/[·\-–]/)[0]?.trim() || '未知';
      regionMap.set(region, (regionMap.get(region) || 0) + s.capacity_kw);
    }
    return Array.from(regionMap.entries()).map(([name, value]) => ({ name, value: +value.toFixed(0) }));
  })();

  // ── Chart 2: 各电站实时出力排行 BarChart ──
  const rankingData = [...monitorData]
    .sort((a, b) => b.currentPowerKw - a.currentPowerKw)
    .slice(0, 15)
    .map((m) => ({
      name: m.station.station_name.length > 8
        ? m.station.station_name.slice(0, 8) + '…'
        : m.station.station_name,
      出力: m.currentPowerKw,
      装机: m.station.capacity_kw,
      alert: m.dropAlert,
    }));

  // ── Chart 3: 异常电站告警列表 ──
  const alertList = monitorData.filter((m) => m.dropAlert);
  const isMobile = useIsMobile();
  const alertMobileFields: MobileCardField<MobileRow>[] = [
    { primary: true, render: (m) => <span className="text-red-700">{(m as any).station.station_name}</span> },
    { label: '出力比', emphasize: true, render: (m) => <span className="text-red-700">{((m as any).currentPowerKw / (m as any).station.capacity_kw * 100).toFixed(1)}%</span> },
    { label: '当前出力', emphasize: true, render: (m) => <span className="text-red-700 font-bold">{fmt((m as any).currentPowerKw)} kW</span> },
    { label: '装机', render: (m) => `${fmt((m as any).station.capacity_kw, 0)} kW` },
    { label: '所在地', render: (m) => <span className="text-muted-foreground">{(m as any).station.location}</span> },
    { label: '告警', render: () => <Badge variant="destructive">出力骤降</Badge> },
  ];

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">发电监控</h1>
        <p className="text-sm text-muted-foreground">
          电站实时运行状态与关键指标
        </p>
      </div>

      {(stations.length ?? 0) === 0 ? (
        <Card>
          <CardContent className="pt-6">
            <EmptyState compact title="暂无电站数据" />
          </CardContent>
        </Card>
      ) : (
        <>
          {/* KPI 概览 */}
          <div className="grid gap-4 md:grid-cols-4">
            <StatCard label="总装机容量" value={`${fmt(totalCapacity, 0)} kW`} />
            <StatCard label="当前总出力" value={`${fmt(totalCurrentPower, 0)} kW`} />
            <StatCard
              label="运行 / 维护 / 离线"
              value={`${statusCounts.active} / ${statusCounts.maintenance} / ${statusCounts.offline}`}
            />
            <StatCard
              label="平均效率"
              value={
                monitorData.length > 0
                  ? `${fmt(monitorData.reduce((a, m) => a + m.efficiency, 0) / monitorData.length, 1)}%`
                  : '-'
              }
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            {/* Chart 1: 电站地区分布 PieChart */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">电站分布（按地区装机容量）</CardTitle>
              </CardHeader>
              <CardContent>
                {regionData.length > 0 ? (
                  <div className="h-64 w-full [&_.recharts-surface:focus]:outline-none">
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChart>
                        <Pie
                          data={regionData}
                          cx="50%"
                          cy="50%"
                          innerRadius={50}
                          outerRadius={90}
                          dataKey="value"
                          label={({ name, percent }) =>
                            `${name} ${(percent * 100).toFixed(0)}%`
                          }
                          isAnimationActive={false}
                        >
                          {regionData.map((_, idx) => (
                            <Cell key={idx} fill={PIE_COLORS[idx % PIE_COLORS.length]} />
                          ))}
                        </Pie>
                        <Tooltip formatter={(v: number) => `${v} kW`} />
                      </PieChart>
                    </ResponsiveContainer>
                  </div>
                ) : (
                  <EmptyState compact title="暂无分布数据" />
                )}
              </CardContent>
            </Card>

            {/* Chart 2: 各电站实时出力排行 */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">各电站实时出力排行 (kW)
                  <DemoBadge className="ml-1" tooltip="实时出力为随机生成的演示数据，非实时采集" />
                </CardTitle>
              </CardHeader>
              <CardContent>
                {rankingData.length > 0 ? (
                  <div className="h-64 w-full [&_.recharts-surface:focus]:outline-none">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart
                        data={rankingData}
                        layout="vertical"
                        margin={{ top: 8, right: 16, bottom: 8, left: 80 }}
                      >
                        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                        <XAxis type="number" tick={{ fontSize: 11 }} />
                        <YAxis
                          type="category"
                          dataKey="name"
                          tick={{ fontSize: 11 }}
                          width={80}
                        />
                        <Tooltip />
                        <Legend wrapperStyle={{ fontSize: 12 }} />
                        <Bar
                          dataKey="出力"
                          name="实时出力"
                          isAnimationActive={false}
                        >
                          {rankingData.map((d, idx) => (
                            <Cell
                              key={idx}
                              fill={d.alert ? '#ef4444' : '#10b981'}
                            />
                          ))}
                        </Bar>
                        <Bar
                          dataKey="装机"
                          name="装机容量"
                          fill="#94a3b8"
                          isAnimationActive={false}
                          opacity={0.4}
                        />
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                ) : (
                  <EmptyState compact title="暂无出力数据" />
                )}
              </CardContent>
            </Card>
          </div>

          {/* Chart 3: 异常电站告警列表 */}
          {alertList.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base text-red-700">
                  ⚠ 异常电站告警（出力骤降）
                  <DemoBadge className="ml-1" tooltip="出力骤降告警为随机触发的演示数据" />
                </CardTitle>
              </CardHeader>
              <CardContent>
                {isMobile ? (
                  <MobileCardList items={alertList as unknown as MobileRow[]} itemKey={(m) => String((m as any).station.id)} fields={alertMobileFields} />
                ) : (
                <div className="rounded-lg border border-red-200 dark:border-red-500/30 bg-red-50/50 dark:bg-red-500/10">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>站点名称</TableHead>
                        <TableHead>所在地</TableHead>
                        <TableHead className="text-right">当前出力 (kW)</TableHead>
                        <TableHead className="text-right">装机容量 (kW)</TableHead>
                        <TableHead className="text-right">出力比</TableHead>
                        <TableHead>告警</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {alertList.map((m) => (
                        <TableRow key={m.station.id} className="bg-red-50/30">
                          <TableCell className="font-medium text-red-700">
                            {m.station.station_name}
                          </TableCell>
                          <TableCell>{m.station.location}</TableCell>
                          <TableCell className="text-right font-bold text-red-700">
                            {fmt(m.currentPowerKw)}
                          </TableCell>
                          <TableCell className="text-right">
                            {fmt(m.station.capacity_kw, 0)}
                          </TableCell>
                          <TableCell className="text-right text-red-700">
                            {((m.currentPowerKw / m.station.capacity_kw) * 100).toFixed(1)}%
                          </TableCell>
                          <TableCell>
                            <Badge variant="destructive">出力骤降</Badge>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* 站点选择 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">站点选择</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <Label>当前站点</Label>
                <select
                  value={stationId}
                  onChange={(e) => setStationId(e.target.value)}
                  className={FIELD_CLASS}
                >
                  {stations.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.station_name}
                    </option>
                  ))}
                </select>
              </div>
            </CardContent>
          </Card>

          {/* 选中站点详情 */}
          {selected && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">
                  {selected.station.station_name} — 运行详情
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 md:grid-cols-3">
                  <StatMini label="装机容量 (kW)" value={fmt(selected.station.capacity_kw, 0)} />
                  <StatMini label="当前出力 (kW)" value={fmt(selected.currentPowerKw)} />
                  <StatMini
                    label="状态"
                    value={
                      <Badge
                        variant={selected.station.status === 'active' ? 'success' : 'secondary'}
                      >
                        {selected.station.status === 'active'
                          ? '运行中'
                          : selected.station.status === 'maintenance'
                            ? '维护中'
                            : '离线'}
                      </Badge>
                    }
                  />
                </div>
                <div className="mt-4 grid gap-4 md:grid-cols-3">
                  <StatMini label="今日发电 (kWh)" value={fmt(selected.dailyKwh, 0)} />
                  <StatMini label="转换效率 (%)" value={fmt(selected.efficiency)} />
                  <StatMini label="辐照度 (W/m²)" value={fmt(selected.irradiance, 0)} />
                </div>
                <div className="mt-4 grid gap-4 md:grid-cols-3">
                  <StatMini label="组件温度 (°C)" value={fmt(selected.panelTemp)} />
                  <StatMini label="所在地" value={selected.station.location || '-'} />
                  <StatMini
                    label="投产日期"
                    value={selected.station.installed_date?.slice(0, 10) || '-'}
                  />
                </div>
              </CardContent>
            </Card>
          )}
        </>
      )}
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border bg-muted/30 p-4">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-semibold">{value}</p>
    </div>
  );
}

function StatMini({
  label,
  value,
}: {
  label: string;
  value: string | React.ReactNode;
}) {
  return (
    <div className="rounded-md border bg-muted/30 p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 text-xl font-semibold">{value}</p>
    </div>
  );
}
