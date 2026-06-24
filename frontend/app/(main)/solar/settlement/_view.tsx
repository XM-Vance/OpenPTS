'use client';

import { useEffect, useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
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
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { DemoBadge } from '@/components/feedback';
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
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  type SolarStation,
  type SolarRevenue,
  generateSolarDemoData,
  listSolarStations,
  listSolarRevenue,
} from '@/lib/api/solar';

const FIELD_CLASS =
  'flex h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm';

function fmt(v?: number | null, digits = 2): string {
  return v == null ? '-' : v.toLocaleString('zh-CN', { maximumFractionDigits: digits });
}

const STACKED_COLORS = ['#3b82f6', '#10b981', '#f59e0b'];
const PIE_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ef4444', '#06b6d4'];

export default function SolarSettlementPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('storage:write');

  const [stationId, setStationId] = useState('');
  const [generating, setGenerating] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

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

  const { data: revenueData, isLoading: revenueLoading } = useQuery({
    queryKey: ['solar-revenue', stationId],
    queryFn: () => listSolarRevenue({ station_id: stationId, limit: 12 }),
    enabled: !!stationId,
  });

  const onGenerate = async () => {
    setError(null);
    setNotice(null);
    setGenerating(true);
    try {
      const r = await generateSolarDemoData(30);
      setNotice(`已生成 ${r.stations} 个站点演示数据（含 12 个月结算）`);
      qc.invalidateQueries({ queryKey: ['solar-stations'] });
      qc.invalidateQueries({ queryKey: ['solar-revenue'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setGenerating(false);
    }
  };

  const selectedStation: SolarStation | undefined = stationsData?.items.find(
    (s) => s.id === stationId,
  );

  const items = useMemo(() => revenueData?.items ?? [], [revenueData]);

  // ── Chart 1: 收益构成堆叠柱状图（上网电价 + 补贴 + 绿证模拟）──
  const stackedData = useMemo(
    () =>
      items
        .slice()
        .reverse()
        .map((r) => ({
          月份: r.settlement_month,
          上网电价: +(r.revenue - r.subsidy).toFixed(0),
          补贴: +r.subsidy.toFixed(0),
          绿证: +((r.net_income - r.revenue) * 0.3).toFixed(0), // 模拟绿证收入
        })),
    [items],
  );

  // ── Chart 2: 各电站收益排行 ──
  // 需要所有电站的数据，暂时用当前站点的汇总
  const stationRanking = useMemo(() => {
    if (!stationsData?.items) return [];
    return stationsData.items.map((s) => ({
      name: s.station_name.length > 8 ? s.station_name.slice(0, 8) + '…' : s.station_name,
      收益: stationId === s.id
        ? items.reduce((a, r) => a + r.net_income, 0)
        : +(s.capacity_kw * (2 + Math.random() * 3) * 1000).toFixed(0), // 模拟其他站
    })).sort((a, b) => b.收益 - a.收益);
  }, [stationsData, stationId, items]);

  // ── Chart 3: 投资回收进度环形图 ──
  const paybackData = useMemo(() => {
    if (!selectedStation || items.length === 0) return null;
    const totalInvestment = selectedStation.capacity_kw * 4000; // 假设 4元/W
    const totalIncome = items.reduce((a, r) => a + r.net_income, 0);
    const pct = Math.min(100, (totalIncome / totalInvestment) * 100);
    return {
      pct: +pct.toFixed(1),
      totalInvestment,
      totalIncome,
      chartData: [
        { name: '已回收', value: +pct.toFixed(1), fill: '#10b981' },
        { name: '剩余', value: +(100 - pct).toFixed(1), fill: '#e5e7eb' },
      ],
    };
  }, [selectedStation, items]);

  // 汇总
  const totalEnergy = items.reduce((a, r) => a + r.energy_kwh, 0);
  const totalRevenue = items.reduce((a, r) => a + r.revenue, 0);
  const totalSubsidy = items.reduce((a, r) => a + r.subsidy, 0);
  const totalNet = items.reduce((a, r) => a + r.net_income, 0);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">光伏收益结算</h1>
          <p className="text-sm text-muted-foreground">
            光伏电站月度结算收入、补贴与净收益
          </p>
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

      {(stationsData?.items.length ?? 0) > 0 ? (
        <>
          <Card>
            <CardHeader>
              <CardTitle className="text-base">站点选择</CardTitle>
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
                        {s.station_name}
                      </option>
                    ))}
                  </select>
                </div>
                {selectedStation && (
                  <>
                    <StatMini
                      label="装机容量 (kW)"
                      value={fmt(selectedStation.capacity_kw, 0)}
                    />
                    <StatMini
                      label="状态"
                      value={
                        <Badge
                          variant={selectedStation.status === 'active' ? 'success' : 'secondary'}
                        >
                          {selectedStation.status === 'active' ? '运行中' : selectedStation.status}
                        </Badge>
                      }
                    />
                  </>
                )}
              </div>
            </CardContent>
          </Card>

          {/* KPI 汇总 */}
          {items.length > 0 && (
            <div className="grid gap-4 md:grid-cols-4">
              <StatCard label="累计发电量 (kWh)" value={fmt(totalEnergy, 0)} />
              <StatCard label="累计电费收入 (¥)" value={fmt(totalRevenue)} />
              <StatCard label="累计补贴 (¥)" value={fmt(totalSubsidy)} />
              <StatCard label="累计净收益 (¥)" value={fmt(totalNet)} />
            </div>
          )}

          <div className="grid gap-4 md:grid-cols-2">
            {/* Chart 1: 收益构成堆叠柱状图 */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">收益构成（上网电价 + 补贴 + 绿证）</CardTitle>
              </CardHeader>
              <CardContent>
                {stackedData.length > 0 ? (
                  <div className="h-72 w-full [&_.recharts-surface:focus]:outline-none">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart
                        data={stackedData}
                        margin={{ top: 8, right: 16, bottom: 8, left: 8 }}
                      >
                        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                        <XAxis dataKey="月份" tick={{ fontSize: 11 }} />
                        <YAxis tick={{ fontSize: 11 }} width={72} />
                        <Tooltip />
                        <Legend wrapperStyle={{ fontSize: 12 }} />
                        <Bar
                          dataKey="上网电价"
                          stackId="a"
                          fill={STACKED_COLORS[0]}
                          isAnimationActive={false}
                        />
                        <Bar
                          dataKey="补贴"
                          stackId="a"
                          fill={STACKED_COLORS[1]}
                          isAnimationActive={false}
                        />
                        <Bar
                          dataKey="绿证"
                          stackId="a"
                          fill={STACKED_COLORS[2]}
                          isAnimationActive={false}
                        />
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">暂无数据</p>
                )}
              </CardContent>
            </Card>

            {/* Chart 2: 各电站收益排行 */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">各电站收益排行 (¥)
                  <DemoBadge className="ml-1" tooltip="非当前电站的收益含随机系数生成" />
                </CardTitle>
              </CardHeader>
              <CardContent>
                {stationRanking.length > 0 ? (
                  <div className="h-72 w-full [&_.recharts-surface:focus]:outline-none">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart
                        data={stationRanking}
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
                        <Tooltip formatter={(v: number) => `¥${v.toLocaleString()}`} />
                        <Bar dataKey="收益" isAnimationActive={false}>
                          {stationRanking.map((_, idx) => (
                            <Cell key={idx} fill={PIE_COLORS[idx % PIE_COLORS.length]} />
                          ))}
                        </Bar>
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">暂无排行数据</p>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Chart 3: 投资回收进度环形图 */}
          {paybackData && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">投资回收进度</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-8">
                  <div className="h-48 w-48 [&_.recharts-surface:focus]:outline-none">
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChart>
                        <Pie
                          data={paybackData.chartData}
                          cx="50%"
                          cy="50%"
                          innerRadius={55}
                          outerRadius={80}
                          dataKey="value"
                          startAngle={90}
                          endAngle={-270}
                          isAnimationActive={false}
                        >
                          {paybackData.chartData.map((entry, idx) => (
                            <Cell key={idx} fill={entry.fill} />
                          ))}
                        </Pie>
                        <Tooltip formatter={(v: number) => `${v}%`} />
                      </PieChart>
                    </ResponsiveContainer>
                  </div>
                  <div className="space-y-3">
                    <div>
                      <p className="text-xs text-muted-foreground">总投资额</p>
                      <p className="text-lg font-semibold">¥{paybackData.totalInvestment.toLocaleString()}</p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">累计净收益</p>
                      <p className="text-lg font-semibold text-emerald-600">
                        ¥{paybackData.totalIncome.toLocaleString()}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">回收进度</p>
                      <p className="text-2xl font-bold text-primary">{paybackData.pct}%</p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* 明细表 */}
          {stationId && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">月度结算明细</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="rounded-lg border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>结算月份</TableHead>
                        <TableHead className="text-right">发电量 (kWh)</TableHead>
                        <TableHead className="text-right">电费收入 (¥)</TableHead>
                        <TableHead className="text-right">均价 (¥/kWh)</TableHead>
                        <TableHead className="text-right">补贴 (¥)</TableHead>
                        <TableHead className="text-right">净收益 (¥)</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {revenueLoading && (
                        <TableRow>
                          <TableCell colSpan={6} className="text-center text-muted-foreground">
                            加载中...
                          </TableCell>
                        </TableRow>
                      )}
                      {items.map((r) => (
                        <TableRow key={r.id}>
                          <TableCell className="font-medium">
                            {r.settlement_month}
                          </TableCell>
                          <TableCell className="text-right">
                            {fmt(r.energy_kwh, 0)}
                          </TableCell>
                          <TableCell className="text-right font-medium">
                            {fmt(r.revenue)}
                          </TableCell>
                          <TableCell className="text-right">{fmt(r.avg_price, 4)}</TableCell>
                          <TableCell className="text-right">{fmt(r.subsidy)}</TableCell>
                          <TableCell className="text-right font-semibold">
                            {fmt(r.net_income)}
                          </TableCell>
                        </TableRow>
                      ))}
                      {items.length === 0 && !revenueLoading && (
                        <TableRow>
                          <TableCell colSpan={6} className="text-center text-muted-foreground">
                            暂无结算数据
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                </div>
              </CardContent>
            </Card>
          )}
        </>
      ) : (
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground">
              暂无光伏站点{canWrite ? '，请点右上「生成演示数据」' : ''}
            </p>
          </CardContent>
        </Card>
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
