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
import { ChartContainer } from '@/components/charts/chart-container';
import { CustomTooltip } from '@/components/charts/custom-tooltip';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import { diagnoseIntent, genIntentDemo } from '@/lib/api/intent-customer';

function scoreVariant(s: number): 'default' | 'secondary' | 'destructive' | 'success' {
  if (s >= 80) return 'success';
  if (s >= 60) return 'default';
  return 'destructive';
}

// 跟进漏斗阶段定义
const FUNNEL_STAGES = [
  { key: 'initial', label: '初次接触', color: '#93c5fd' },
  { key: 'confirm', label: '需求确认', color: '#60a5fa' },
  { key: 'quote', label: '方案报价', color: '#3b82f6' },
  { key: 'sign', label: '合同签署', color: '#1d4ed8' },
];

// 根据综合分划分漏斗阶段
function getFunnelStage(score: number): string {
  if (score >= 80) return 'sign';
  if (score >= 60) return 'quote';
  if (score >= 40) return 'confirm';
  return 'initial';
}

// 模拟跟进时间线的假数据生成器（基于实际客户数据）
function buildGanttData(
  items: Array<{
    customer_name: string;
    overall_score: number;
    coverage_days?: number | null;
    created_at?: string;
  }>,
) {
  return items.slice(0, 8).map((item, idx) => {
    const baseDays = item.coverage_days ?? 30;
    const progress = Math.min(item.overall_score, 100);
    const stages: Array<{ stage: string; start: number; duration: number }> = [];
    let offset = 0;

    if (progress >= 10) {
      stages.push({ stage: '初次接触', start: offset, duration: Math.round(baseDays * 0.2) });
      offset += Math.round(baseDays * 0.2);
    }
    if (progress >= 30) {
      stages.push({ stage: '需求确认', start: offset, duration: Math.round(baseDays * 0.25) });
      offset += Math.round(baseDays * 0.25);
    }
    if (progress >= 55) {
      stages.push({ stage: '方案报价', start: offset, duration: Math.round(baseDays * 0.3) });
      offset += Math.round(baseDays * 0.3);
    }
    if (progress >= 80) {
      stages.push({ stage: '合同签署', start: offset, duration: Math.round(baseDays * 0.25) });
    }

    return { name: item.customer_name.length > 6 ? item.customer_name.slice(0, 6) + '…' : item.customer_name, stages, idx };
  });
}

// recharts 懒加载：从首屏 JS 剥离，挂载后异步加载（渲染内容不变）。
const FunnelBar = dynamic(
  () => import('./_funnel-bar').then((m) => ({ default: m.FunnelBar })),
  { ssr: false, loading: () => <div className="h-full w-full" /> },
);

export default function IntentCustomersPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('customer_management:write');

  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['intent-diagnose'],
    queryFn: diagnoseIntent,
  });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genIntentDemo();
      qc.invalidateQueries({ queryKey: ['intent-diagnose'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const items = useMemo(() => data?.items ?? [], [data]);
  const greenCount = items.filter((i) => i.overall_score >= 80).length;
  const yellowCount = items.filter((i) => i.overall_score >= 60 && i.overall_score < 80).length;
  const redCount = items.filter((i) => i.overall_score < 60).length;

  // ── 漏斗图数据 ──
  const funnelData = useMemo(() => {
    const counts = FUNNEL_STAGES.map((stage) => ({
      ...stage,
      count: items.filter((i) => getFunnelStage(i.overall_score) === stage.key).length,
    }));
    // 累积：后面的阶段包含前面
    const cumulative: Array<{ label: string; count: number; color: string }> = [];
    let running = 0;
    for (let i = FUNNEL_STAGES.length - 1; i >= 0; i--) {
      running += counts[i].count;
      cumulative.unshift({ label: counts[i].label, count: running, color: counts[i].color });
    }
    return cumulative;
  }, [items]);

  // ── 甘特图数据 ──
  const ganttData = useMemo(() => buildGanttData(items), [items]);

  // ── 超时提醒：覆盖天数>7 且综合分<80 的客户 ──
  const overdueItems = useMemo(
    () => items.filter((i) => (i.coverage_days ?? 0) > 7 && i.overall_score < 80),
    [items],
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">意向客户诊断</h1>
          <p className="text-sm text-muted-foreground">
            综合数据完整度 + 覆盖时长 + 负荷规模 → 转化推荐
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
            <p className="text-xs text-muted-foreground">总意向客户</p>
            <p className="mt-1 text-2xl font-bold">{items.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">推荐转化（&ge;80）</p>
            <p className="mt-1 text-2xl font-bold text-emerald-600">{greenCount}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">可试用（60-80）</p>
            <p className="mt-1 text-2xl font-bold text-blue-600">{yellowCount}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">需补数据（&lt;60）</p>
            <p className="mt-1 text-2xl font-bold text-destructive">{redCount}</p>
          </CardContent>
        </Card>
      </div>

      {/* ── 跟进漏斗图 ── */}
      <ChartContainer title="意向客户跟进漏斗" minHeight={260}>
        {funnelData.length > 0 && items.length > 0 ? (
          <FunnelBar data={funnelData} />
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            暂无数据
          </div>
        )}
      </ChartContainer>

      {/* ── 跟进状态时间线（甘特图）── */}
      <ChartContainer title="客户跟进进度时间线" minHeight={Math.max(200, ganttData.length * 44 + 40)}>
        {ganttData.length > 0 ? (
          <div className="space-y-1">
            {/* 图例 */}
            <div className="flex items-center gap-4 pb-2 text-xs">
              <span className="flex items-center gap-1"><span className="inline-block h-3 w-3 rounded-sm" style={{ backgroundColor: '#93c5fd' }} /> 初次接触</span>
              <span className="flex items-center gap-1"><span className="inline-block h-3 w-3 rounded-sm" style={{ backgroundColor: '#60a5fa' }} /> 需求确认</span>
              <span className="flex items-center gap-1"><span className="inline-block h-3 w-3 rounded-sm" style={{ backgroundColor: '#3b82f6' }} /> 方案报价</span>
              <span className="flex items-center gap-1"><span className="inline-block h-3 w-3 rounded-sm" style={{ backgroundColor: '#1d4ed8' }} /> 合同签署</span>
            </div>
            {ganttData.map((row) => {
              const totalDays = row.stages.reduce((s, st) => s + st.duration, 0) || 1;
              return (
                <div key={row.idx} className="flex items-center gap-2">
                  <div className="w-20 shrink-0 truncate text-right text-xs text-zinc-500">{row.name}</div>
                  <div className="relative h-7 flex-1 rounded bg-zinc-100">
                    {row.stages.map((stage, si) => {
                      const leftPct = (stage.start / totalDays) * 100;
                      const widthPct = (stage.duration / totalDays) * 100;
                      const colors: Record<string, string> = {
                        '初次接触': '#93c5fd',
                        '需求确认': '#60a5fa',
                        '方案报价': '#3b82f6',
                        '合同签署': '#1d4ed8',
                      };
                      return (
                        <div
                          key={si}
                          className="absolute top-0 h-full rounded-sm"
                          style={{
                            left: `${leftPct}%`,
                            width: `${widthPct}%`,
                            backgroundColor: colors[stage.stage] ?? '#94a3b8',
                          }}
                          title={`${stage.stage}: 第${stage.start}-${stage.start + stage.duration}天`}
                        />
                      );
                    })}
                  </div>
                  <span className="w-12 shrink-0 text-right text-xs text-zinc-400">{totalDays}天</span>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            暂无数据
          </div>
        )}
      </ChartContainer>

      {/* ── 超时提醒 ── */}
      {overdueItems.length > 0 && (
        <Alert className="border-orange-300 bg-orange-50 text-orange-800">
          <AlertDescription>
            <span className="font-semibold">⏰ 跟进超时提醒：</span>
            {overdueItems.length} 个客户已超过 7 天未转化 —{' '}
            {overdueItems.slice(0, 5).map((i) => i.customer_name).join('、')}
            {overdueItems.length > 5 && ` 等${overdueItems.length}家`}
          </AlertDescription>
        </Alert>
      )}

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>客户名</TableHead>
              <TableHead className="text-right">覆盖天数</TableHead>
              <TableHead className="text-right">完整度</TableHead>
              <TableHead className="text-right">日均负荷 (kW)</TableHead>
              <TableHead className="text-right">数据分</TableHead>
              <TableHead className="text-right">覆盖分</TableHead>
              <TableHead className="text-right">规模分</TableHead>
              <TableHead className="text-right">综合分</TableHead>
              <TableHead>推荐套餐</TableHead>
              <TableHead>建议</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={10} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            )}
            {items.map((i) => {
              const isOverdue = (i.coverage_days ?? 0) > 7 && i.overall_score < 80;
              return (
                <TableRow
                  key={i.id}
                  className={isOverdue ? 'bg-orange-50' : undefined}
                >
                  <TableCell className="font-medium">
                    {i.customer_name}
                    {isOverdue && (
                      <Badge variant="destructive" className="ml-2 text-[10px] px-1.5 py-0">
                        超时
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">{i.coverage_days ?? '-'}</TableCell>
                  <TableCell className="text-right">
                    {i.completeness != null ? `${i.completeness.toFixed(1)}%` : '-'}
                  </TableCell>
                  <TableCell className="text-right">
                    {i.avg_daily_load != null ? i.avg_daily_load.toFixed(0) : '-'}
                  </TableCell>
                  <TableCell className="text-right">{i.data_score.toFixed(0)}</TableCell>
                  <TableCell className="text-right">{i.coverage_score.toFixed(0)}</TableCell>
                  <TableCell className="text-right">{i.load_score.toFixed(0)}</TableCell>
                  <TableCell className="text-right">
                    <Badge variant={scoreVariant(i.overall_score)}>
                      {i.overall_score.toFixed(1)}
                    </Badge>
                  </TableCell>
                  <TableCell>{i.matched_package ?? '-'}</TableCell>
                  <TableCell className="text-muted-foreground">{i.recommendation}</TableCell>
                </TableRow>
              );
            })}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={10} className="text-center text-muted-foreground">
                  暂无意向客户{canWrite && '，可点右上「生成演示数据」'}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
