'use client';

import { useState, useMemo, useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  ComposedChart,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  ReferenceArea,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { DemoBadge } from '@/components/feedback';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { StatCard } from '@/components/data-display/stat-card';
import { PageHeader } from '@/components/data-display/page-header';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import { genTotalLoadDemo, listTotalLoad, type TotalLoadDaily } from '@/lib/api/total-load';
import {
  Zap,
  TrendingUp,
  TrendingDown,
  Activity,
  BarChart3,
} from 'lucide-react';

const fmt = (v: number | null | undefined, digits = 1) =>
  v == null ? '-' : v.toLocaleString('zh-CN', { maximumFractionDigits: digits });

/* ── 峰谷平时段定义 ── */
const PERIOD_BANDS = [
  { label: '深谷', start: '00:00', end: '06:00', color: '#10b98133' },
  { label: '平段', start: '06:00', end: '08:00', color: '#3b82f622' },
  { label: '高峰', start: '08:00', end: '11:00', color: '#f59e0b33' },
  { label: '平段', start: '11:00', end: '14:00', color: '#3b82f622' },
  { label: '尖峰', start: '14:00', end: '17:00', color: '#ef444433' },
  { label: '平段', start: '17:00', end: '19:00', color: '#3b82f622' },
  { label: '高峰', start: '19:00', end: '21:00', color: '#f59e0b33' },
  { label: '平段', start: '21:00', end: '23:00', color: '#3b82f622' },
  { label: '低谷', start: '23:00', end: '24:00', color: '#10b98133' },
];

/* ── Demo: 生成"去年"曲线 ── */
function generateLastYearCurve(current: number[]): number[] {
  return current.map((v) => v * (0.88 + Math.random() * 0.08));
}

function seededRandom(seed: number) {
  const x = Math.sin(seed) * 10000;
  return x - Math.floor(x);
}

function generateDemoCurve96(seed: number): number[] {
  const base = 900 + seededRandom(seed) * 500;
  const curve: number[] = [];
  for (let i = 0; i < 96; i++) {
    const hour = i / 4;
    const shape =
      0.6 +
      0.25 * Math.sin(((hour - 6) * Math.PI) / 12) +
      0.15 * Math.sin(((hour - 14) * Math.PI) / 8);
    const noise = (seededRandom(i * 7 + seed) - 0.5) * 30;
    curve.push(Math.max(0, Math.round((base * shape + noise) * 10) / 10));
  }
  return curve;
}

export default function TotalLoadPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('load_management:write');

  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<TotalLoadDaily | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['total-load'],
    queryFn: () => listTotalLoad(14),
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
      await genTotalLoadDemo();
      qc.invalidateQueries({ queryKey: ['total-load'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  // 96 点曲线数据（含峰谷平时段着色）
  const curveData = useMemo(() => {
    const curve96 = selected?.curve_96 ?? generateDemoCurve96(42);
    return curve96.map((v, i) => ({
      period: `${String(Math.floor(i / 4)).padStart(2, '0')}:${String((i % 4) * 15 || '00').slice(0, 2)}`,
      负荷: v,
    }));
  }, [selected]);

  // 去年同期曲线
  const lastYearCurve = useMemo(() => {
    const current = selected?.curve_96 ?? generateDemoCurve96(42);
    return generateLastYearCurve(current);
  }, [selected]);

  // 对比图数据
  const comparisonData = useMemo(() => {
    const curve96 = selected?.curve_96 ?? generateDemoCurve96(42);
    return curve96.map((v, i) => ({
      period: `${String(Math.floor(i / 4)).padStart(2, '0')}:${String((i % 4) * 15 || '00').slice(0, 2)}`,
      今年: Math.round(v * 10) / 10,
      去年同期: Math.round(lastYearCurve[i] * 10) / 10,
    }));
  }, [selected, lastYearCurve]);

  // 负荷率统计
  const loadRate = useMemo(() => {
    const curve = selected?.curve_96 ?? generateDemoCurve96(42);
    const peak = Math.max(...curve);
    const valley = Math.min(...curve);
    const avg = curve.reduce((a, b) => a + b, 0) / curve.length;
    return {
      peak,
      valley,
      avg,
      loadRate: peak > 0 ? avg / peak : 0,
      peakValleyRatio: valley > 0 ? peak / valley : 0,
    };
  }, [selected]);

  const trendData = items
    .slice()
    .reverse()
    .map((d) => ({
      date: d.data_date.slice(5, 10).replace('-', '/'),
      peak: d.peak_load,
      valley: d.valley_load,
      avg: d.avg_load,
    }));

  return (
    <div className="space-y-6">
      <PageHeader
        title="系统总负荷"
        description="区域级 96 点曲线 · 峰谷平时段着色 · 历年同期对比 · 负荷率统计"
        actions={
          canWrite ? (
            <Button variant="outline" onClick={onGen} disabled={busy}>
              {busy ? '生成中...' : '生成演示数据'}
            </Button>
          ) : undefined
        }
      />

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-5">
        <StatCard
          title="峰负荷 (MW)"
          value={fmt(loadRate.peak)}
          icon={<TrendingUp className="h-4 w-4" />}
        />
        <StatCard
          title="谷负荷 (MW)"
          value={fmt(loadRate.valley)}
          icon={<TrendingDown className="h-4 w-4" />}
        />
        <StatCard
          title="平均负荷 (MW)"
          value={fmt(loadRate.avg)}
          icon={<Activity className="h-4 w-4" />}
        />
        <StatCard
          title="负荷率"
          value={`${(loadRate.loadRate * 100).toFixed(1)}%`}
          icon={<Zap className="h-4 w-4" />}
          trend={loadRate.loadRate > 0.75 ? 3 : loadRate.loadRate > 0.6 ? 0 : -5}
          trendLabel="平均/最大"
        />
        <StatCard
          title="峰谷比"
          value={loadRate.peakValleyRatio.toFixed(2)}
          icon={<BarChart3 className="h-4 w-4" />}
        />
      </div>

      {/* 96 点曲线（含峰谷平时段着色） */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">
            96 点曲线 · {selected ? selected.data_date.slice(0, 10) : '演示数据'}
          </CardTitle>
          <CardDescription className="text-xs">
            时段着色：尖峰(红)、高峰(黄)、平段(蓝)、低谷/深谷(绿)
          </CardDescription>
        </CardHeader>
        <CardContent>
          {curveData.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {canWrite ? '请点右上「生成演示数据」' : '暂无数据'}
            </p>
          ) : (
            <div
              className="relative [&_.recharts-surface:focus]:outline-none"
              style={{ width: '100%', height: 280 }}
            >
              {/* 时段着色背景带 */}
              <div className="absolute inset-0 flex pointer-events-none" style={{ top: 28, bottom: 28 }}>
                {PERIOD_BANDS.map((band, idx) => {
                  const startIdx = parseInt(band.start.split(':')[0], 10) * 4;
                  const endIdx = parseInt(band.end.split(':')[0], 10) * 4;
                  const totalPoints = 96;
                  const leftPct = (startIdx / totalPoints) * 100;
                  const widthPct = ((endIdx - startIdx) / totalPoints) * 100;
                  return (
                    <div
                      key={idx}
                      className="absolute top-0 bottom-0 flex items-start justify-center pt-0.5"
                      style={{
                        left: `${leftPct + 4}%`,
                        width: `${widthPct * 0.92}%`,
                        backgroundColor: band.color,
                      }}
                    >
                      <span className="text-[9px] opacity-60">{band.label}</span>
                    </div>
                  );
                })}
              </div>
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={curveData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="period" tick={{ fontSize: 10, fill: '#6b7280' }} interval={11} />
                  <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} unit=" MW" />
                  <Tooltip
                    formatter={(v: number) => `${fmt(v)} MW`}
                    contentStyle={{ fontSize: 12 }}
                  />
                  <Area
                    type="monotone"
                    dataKey="负荷"
                    stroke="#2563eb"
                    fill="#dbeafe"
                    strokeWidth={2}
                    isAnimationActive={false}
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 历年同期对比 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">历年同期对比
            <DemoBadge className="ml-1" tooltip="去年同期曲线含随机抖动，非真实历史数据" />
          </CardTitle>
          <CardDescription className="text-xs">
            今年 vs 去年同期负荷曲线叠加
          </CardDescription>
        </CardHeader>
        <CardContent>
          {comparisonData.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无数据</p>
          ) : (
            <div
              className="[&_.recharts-surface:focus]:outline-none"
              style={{ width: '100%', height: 260 }}
            >
              <ResponsiveContainer width="100%" height="100%">
                <ComposedChart data={comparisonData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="period" tick={{ fontSize: 10, fill: '#6b7280' }} interval={11} />
                  <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} unit=" MW" />
                  <Tooltip
                    contentStyle={{ fontSize: 12 }}
                    formatter={(v: number) => `${fmt(v)} MW`}
                  />
                  <Line
                    type="monotone"
                    dataKey="今年"
                    stroke="#2563eb"
                    strokeWidth={2}
                    dot={false}
                    isAnimationActive={false}
                    name="今年"
                  />
                  <Line
                    type="monotone"
                    dataKey="去年同期"
                    stroke="#94a3b8"
                    strokeWidth={2}
                    strokeDasharray="6 3"
                    dot={false}
                    isAnimationActive={false}
                    name="去年同期"
                  />
                </ComposedChart>
              </ResponsiveContainer>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 14 日峰/谷/均趋势 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">14 日峰/谷/均</CardTitle>
        </CardHeader>
        <CardContent>
          {trendData.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无数据</p>
          ) : (
            <div
              className="[&_.recharts-surface:focus]:outline-none"
              style={{ width: '100%', height: 200 }}
            >
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={trendData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#6b7280' }} />
                  <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} />
                  <Tooltip contentStyle={{ fontSize: 12 }} />
                  <Bar dataKey="peak" name="峰" fill="#f59e0b" isAnimationActive={false} />
                  <Bar dataKey="avg" name="均" fill="#3b82f6" isAnimationActive={false} />
                  <Bar dataKey="valley" name="谷" fill="#10b981" isAnimationActive={false} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </CardContent>
      </Card>

      {isLoading && <p className="text-sm text-muted-foreground">加载中...</p>}
    </div>
  );
}
