'use client';

import { useState, useMemo, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  getCustomerLoadSummary,
  getCustomerLoadCurve,
} from '@/lib/api/customer-load';

const fmt = (v: number) => v.toLocaleString('zh-CN', { maximumFractionDigits: 2 });

const LINE_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4'];

// 负荷模板分类
const TEMPLATES = [
  { name: '单峰型', desc: '白天单峰，夜间低谷', color: '#3b82f6' },
  { name: '双峰型', desc: '上午+下午双峰', color: '#10b981' },
  { name: '平缓型', desc: '全天负荷平稳', color: '#f59e0b' },
  { name: '夜峰型', desc: '夜间负荷偏高', color: '#ef4444' },
  { name: '尖峰型', desc: '短时高峰突出', color: '#8b5cf6' },
];

function classifyTemplate(cv: number, peakValley: number): { name: string; color: string } {
  if (peakValley > 3) return { name: '尖峰型', color: '#8b5cf6' };
  if (cv > 0.3) return { name: '双峰型', color: '#10b981' };
  if (cv < 0.1) return { name: '平缓型', color: '#f59e0b' };
  if (peakValley < 1.5) return { name: '夜峰型', color: '#ef4444' };
  return { name: '单峰型', color: '#3b82f6' };
}

export default function CustomerLoadAnalysisPage() {
  const [selected, setSelected] = useState<string | null>(null);
  const [compareIds, setCompareIds] = useState<string[]>([]);

  const { data: summary, isLoading } = useQuery({
    queryKey: ['cust-load-summary'],
    queryFn: () => getCustomerLoadSummary(14),
  });
  const { data: curve } = useQuery({
    queryKey: ['cust-load-curve', selected],
    queryFn: () => getCustomerLoadCurve(selected!),
    enabled: !!selected,
  });

  // 多客户曲线数据
  const { data: compareCurves } = useQuery({
    queryKey: ['cust-load-compare', compareIds],
    queryFn: async () => {
      const results = await Promise.all(
        compareIds.map(async (id) => {
          const c = await getCustomerLoadCurve(id);
          return { id, name: c.customer_name, curve: c.curve_96 };
        }),
      );
      return results;
    },
    enabled: compareIds.length > 0,
  });

  const items = useMemo(() => summary?.items ?? [], [summary]);
  // 自动选中首位：放 useEffect 里（渲染期内 setTimeout setState 是反模式）
  useEffect(() => {
    if (!selected && items.length > 0) {
      setSelected(items[0].customer_id);
      if (compareIds.length === 0) {
        setCompareIds(items.slice(0, 3).map((c) => c.customer_id));
      }
    }
  }, [selected, items, compareIds]);

  const chartData = (curve?.curve_96 ?? []).map((v, i) => ({
    period: `${String(Math.floor(i / 4)).padStart(2, '0')}:${(i % 4) * 15}0`.slice(0, 5),
    load: v,
  }));

  // ── Chart 1: 多客户负荷曲线叠加对比 ──
  const multiCurveData = useMemo(() => {
    if (!compareCurves || compareCurves.length === 0) return [];
    const periods = Array.from({ length: 96 }, (_, i) =>
      `${String(Math.floor(i / 4)).padStart(2, '0')}:${(i % 4) * 15}0`.slice(0, 5),
    );
    return periods.map((p, i) => {
      const row: Record<string, string | number> = { period: p };
      compareCurves.forEach((c) => {
        row[c.name] = c.curve[i] ?? 0;
      });
      return row;
    });
  }, [compareCurves]);

  // ── Chart 2: 负荷模板匹配 ──
  const templateMatchData = useMemo(
    () =>
      items.map((c) => {
        const tpl = classifyTemplate(c.cv, c.peak_valley_ratio);
        return {
          customer_id: c.customer_id,
          customer_name: c.customer_name,
          template: tpl.name,
          color: tpl.color,
          cv: c.cv,
          peakValley: c.peak_valley_ratio,
          avgDaily: c.avg_daily,
        };
      }),
    [items],
  );

  const templateCounts = useMemo(() => {
    const map = new Map<string, number>();
    templateMatchData.forEach((d) => {
      map.set(d.template, (map.get(d.template) || 0) + 1);
    });
    return Array.from(map.entries()).map(([name, count]) => {
      const tpl = TEMPLATES.find((t) => t.name === name);
      return { name, count, color: tpl?.color ?? '#6b7280' };
    });
  }, [templateMatchData]);

  // ── Chart 3: 最大需量统计卡片 ──
  const maxDemandStats = useMemo(() => {
    if (items.length === 0) return null;
    const peakLoads = items.map((c) => c.peak_load);
    return {
      max: Math.max(...peakLoads),
      avg: peakLoads.reduce((a, b) => a + b, 0) / peakLoads.length,
      min: Math.min(...peakLoads),
      customers: items.length,
    };
  }, [items]);

  const toggleCompare = (id: string) => {
    setCompareIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : prev.length < 6 ? [...prev, id] : prev,
    );
  };

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">客户负荷分析</h1>
        <p className="text-sm text-muted-foreground">
          多客户负荷对比 · 模板匹配 · 最大需量统计
        </p>
      </div>

      {/* Chart 3: 最大需量统计卡片 */}
      {maxDemandStats && (
        <div className="grid gap-4 md:grid-cols-4">
          <StatMini label="客户总数" value={`${maxDemandStats.customers}`} />
          <StatMini label="最大需量 (kW)" value={fmt(maxDemandStats.max)} />
          <StatMini label="平均最大需量 (kW)" value={fmt(maxDemandStats.avg)} />
          <StatMini label="最小需量 (kW)" value={fmt(maxDemandStats.min)} />
        </div>
      )}

      <div className="grid gap-4 lg:grid-cols-5">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">客户列表（点击对比）</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <div className="rounded-lg border-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>客户</TableHead>
                    <TableHead className="text-right">日均 MWh</TableHead>
                    <TableHead className="text-right">峰/谷比</TableHead>
                    <TableHead className="text-right">CV</TableHead>
                    <TableHead>模板</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {isLoading && (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center text-muted-foreground">加载中...</TableCell>
                    </TableRow>
                  )}
                  {items.map((c) => {
                    const tpl = classifyTemplate(c.cv, c.peak_valley_ratio);
                    const isCompared = compareIds.includes(c.customer_id);
                    return (
                      <TableRow
                        key={c.customer_id}
                        className={
                          selected === c.customer_id
                            ? 'cursor-pointer bg-amber-50 ring-1 ring-amber-300'
                            : 'cursor-pointer hover:bg-muted/50'
                        }
                        onClick={() => {
                          setSelected(c.customer_id);
                          toggleCompare(c.customer_id);
                        }}
                      >
                        <TableCell className="font-medium">
                          <div className="flex items-center gap-1">
                            {isCompared && <div className="h-2 w-2 rounded-full bg-blue-500" />}
                            {c.customer_name}
                          </div>
                        </TableCell>
                        <TableCell className="text-right">{fmt(c.avg_daily)}</TableCell>
                        <TableCell className="text-right">
                          <Badge variant={c.peak_valley_ratio > 3 ? 'destructive' : 'outline'}>
                            {fmt(c.peak_valley_ratio)}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">{fmt(c.cv)}</TableCell>
                        <TableCell>
                          <span className="text-xs" style={{ color: tpl.color }}>{tpl.name}</span>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                  {items.length === 0 && !isLoading && (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center text-muted-foreground">暂无负荷数据</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>

        <div className="lg:col-span-3 space-y-4">
          {/* Chart 1: 多客户负荷曲线叠加对比 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">多客户负荷曲线叠加对比</CardTitle>
            </CardHeader>
            <CardContent>
              {multiCurveData.length > 0 ? (
                <div className="[&_.recharts-surface:focus]:outline-none" style={{ width: '100%', height: 320 }}>
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={multiCurveData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                      <XAxis dataKey="period" tick={{ fontSize: 10, fill: '#6b7280' }} interval={11} />
                      <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} unit=" kW" />
                      <Tooltip formatter={(v: number) => `${fmt(v)} kW`} contentStyle={{ fontSize: 12 }} />
                      <Legend wrapperStyle={{ fontSize: 11 }} />
                      {compareCurves?.map((c, idx) => (
                        <Line
                          key={c.id}
                          type="monotone"
                          dataKey={c.name}
                          stroke={LINE_COLORS[idx % LINE_COLORS.length]}
                          strokeWidth={2}
                          dot={false}
                          isAnimationActive={false}
                        />
                      ))}
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              ) : chartData.length > 0 ? (
                <div className="[&_.recharts-surface:focus]:outline-none" style={{ width: '100%', height: 320 }}>
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={chartData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                      <XAxis dataKey="period" tick={{ fontSize: 10, fill: '#6b7280' }} interval={11} />
                      <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} unit=" kW" />
                      <Tooltip formatter={(v: number) => `${fmt(v)} kW`} contentStyle={{ fontSize: 12 }} />
                      <Line type="monotone" dataKey="load" stroke="#2563eb" strokeWidth={2} dot={false} isAnimationActive={false} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">请从左侧选择客户</p>
              )}
            </CardContent>
          </Card>

          {/* Chart 2: 负荷模板匹配统计 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">负荷模板匹配</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-3 md:grid-cols-3">
                {templateCounts.map((t) => (
                  <div key={t.name} className="rounded-md border p-3">
                    <div className="flex items-center gap-2">
                      <div className="h-3 w-3 rounded-full" style={{ backgroundColor: t.color }} />
                      <span className="text-sm font-medium">{t.name}</span>
                    </div>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {TEMPLATES.find((tp) => tp.name === t.name)?.desc}
                    </p>
                    <p className="mt-2 text-lg font-bold" style={{ color: t.color }}>
                      {t.count} 户
                    </p>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function StatMini({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border bg-muted/30 p-4">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 text-xl font-semibold">{value}</p>
    </div>
  );
}
