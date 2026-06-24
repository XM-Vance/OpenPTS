'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  listContractProgress,
  generateContractProgressDemoData,
  type ContractProgress,
} from '@/lib/api/contract-progress';
import { StatCard } from '@/components/data-display/stat-card';
import { ChartContainer } from '@/components/charts/chart-container';
import { DataTable, type DataTableColumn, type DataRow } from '@/components/data-display/data-table';
import { PageHeader } from '@/components/data-display/page-header';
import {
  BarChart3,
  AlertCircle,
  RefreshCw,
  TrendingUp,
  TrendingDown,
  CheckCircle2,
  Clock,
  FileText,
} from 'lucide-react';

function fmt(v: number | null | undefined, digits = 2): string {
  if (v == null) return '-';
  return v.toLocaleString('zh-CN', { maximumFractionDigits: digits });
}

const STATUS_MAP: Record<string, { label: string; variant: 'default' | 'secondary' | 'outline' | 'destructive' }> = {
  on_track: { label: '正常', variant: 'secondary' },
  ahead: { label: '超前', variant: 'default' },
  behind: { label: '滞后', variant: 'destructive' },
  completed: { label: '已完成', variant: 'outline' },
};

// recharts 懒加载：从首屏 JS 剥离，挂载后异步加载（渲染内容不变）。
const PlanActualComposed = dynamic(
  () => import('./_charts').then((m) => ({ default: m.PlanActualComposed })),
  { ssr: false, loading: () => <div className="h-full w-full" /> },
);
const CompletionComposed = dynamic(
  () => import('./_charts').then((m) => ({ default: m.CompletionComposed })),
  { ssr: false, loading: () => <div className="h-full w-full" /> },
);

export default function ContractProgressPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('retail_management:write');

  const [filterMonth, setFilterMonth] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  // 合同进度列表
  const { data: listData, isLoading: listLoading } = useQuery({
    queryKey: ['contract-progress', filterMonth, filterStatus],
    queryFn: () => listContractProgress(filterMonth, filterStatus, 200),
  });

  const items = useMemo(() => (listData?.items ?? []) as ContractProgress[], [listData]);

  // 生成演示数据
  const onGenerate = async () => {
    setError(null);
    setNotice(null);
    setGenerating(true);
    try {
      const r = await generateContractProgressDemoData();
      setNotice(`${r.message}（${r.rows} 条）`);
      qc.invalidateQueries({ queryKey: ['contract-progress'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setGenerating(false);
    }
  };

  // 汇总统计
  const summary = useMemo(() => {
    if (!items.length) return null;
    const totalPlanned = items.reduce((s, i) => s + i.planned_energy_mwh, 0);
    const totalActual = items.reduce((s, i) => s + i.actual_energy_mwh, 0);
    const avgRate = items.length > 0
      ? items.reduce((s, i) => s + i.completion_rate, 0) / items.length
      : 0;
    const onTrackCount = items.filter((i) => i.status === 'on_track').length;
    const aheadCount = items.filter((i) => i.status === 'ahead').length;
    const behindCount = items.filter((i) => i.status === 'behind').length;
    const completedCount = items.filter((i) => i.status === 'completed').length;
    return {
      totalPlanned, totalActual, avgRate,
      onTrackCount, aheadCount, behindCount, completedCount,
      count: items.length,
    };
  }, [items]);

  // 按月份趋势图数据
  const trendData = useMemo(() => {
    if (!items.length) return [];
    const monthMap = new Map<string, { planned: number; actual: number; count: number }>();
    for (const item of items) {
      const m = item.operating_month;
      const prev = monthMap.get(m) ?? { planned: 0, actual: 0, count: 0 };
      prev.planned += item.planned_energy_mwh;
      prev.actual += item.actual_energy_mwh;
      prev.count += 1;
      monthMap.set(m, prev);
    }
    return [...monthMap.entries()]
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([month, d]) => ({
        month,
        计划电量: Math.round(d.planned),
        实际电量: Math.round(d.actual),
        完成率: d.planned > 0 ? Math.round((d.actual / d.planned) * 100) : 0,
      }));
  }, [items]);

  // ── 进度条可视化数据（计划 vs 实际）──
  const progressBarData = useMemo(() => {
    if (!items.length) return [];
    // 按客户聚合
    const custMap = new Map<string, { planned: number; actual: number; name: string }>();
    for (const item of items) {
      const name = item.customer_name || '未知';
      const prev = custMap.get(name) ?? { planned: 0, actual: 0, name };
      prev.planned += item.planned_energy_mwh;
      prev.actual += item.actual_energy_mwh;
      custMap.set(name, prev);
    }
    return Array.from(custMap.values())
      .sort((a, b) => b.planned - a.planned)
      .slice(0, 12)
      .map((c) => ({
        name: c.name.length > 6 ? c.name.slice(0, 6) + '…' : c.name,
        fullName: c.name,
        计划: Math.round(c.planned),
        实际: Math.round(c.actual),
        偏差: Math.round(c.actual - c.planned),
        完成率: c.planned > 0 ? Math.round((c.actual / c.planned) * 100) : 0,
      }));
  }, [items]);

  // ── 偏差预警 ──
  const deviationAlerts = useMemo(() => {
    return items
      .filter((i) => i.completion_rate < 80 && i.status !== 'completed')
      .sort((a, b) => a.completion_rate - b.completion_rate)
      .slice(0, 5)
      .map((i) => ({
        name: i.customer_name || '未知',
        month: i.operating_month,
        rate: i.completion_rate,
        gap: i.planned_energy_mwh - i.actual_energy_mwh,
      }));
  }, [items]);

  // ── 多合同执行进度汇总看板 ──
  const dashboardCards = useMemo(() => {
    if (!items.length) return [];
    const custMap = new Map<string, ContractProgress[]>();
    for (const item of items) {
      const name = item.customer_name || '未知';
      if (!custMap.has(name)) custMap.set(name, []);
      custMap.get(name)!.push(item);
    }
    return Array.from(custMap.entries())
      .map(([name, records]) => {
        const planned = records.reduce((s, r) => s + r.planned_energy_mwh, 0);
        const actual = records.reduce((s, r) => s + r.actual_energy_mwh, 0);
        const rate = planned > 0 ? (actual / planned) * 100 : 0;
        const hasBehind = records.some((r) => r.status === 'behind');
        return { name, planned, actual, rate, hasBehind, recordCount: records.length };
      })
      .sort((a, b) => a.rate - b.rate)
      .slice(0, 8);
  }, [items]);

  // 表格列
  const columns: DataTableColumn[] = [
    {
      key: 'customer_name',
      header: '客户名称',
      render: (row: DataRow) => (
        <span className="font-medium">{(row.customer_name as string) || '-'}</span>
      ),
    },
    {
      key: 'operating_month',
      header: '运行月份',
      sortable: true,
      render: (row: DataRow) => row.operating_month as string,
    },
    {
      key: 'planned_energy_mwh',
      header: '计划电量 (MWh)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => fmt(row.planned_energy_mwh as number),
    },
    {
      key: 'actual_energy_mwh',
      header: '实际电量 (MWh)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => fmt(row.actual_energy_mwh as number),
    },
    {
      key: 'completion_rate',
      header: '完成率 (%)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => {
        const v = row.completion_rate as number;
        return (
          <span className={v >= 100 ? 'text-green-600 font-medium' : v >= 80 ? 'text-amber-600' : 'text-red-600'}>
            {fmt(v, 1)}%
          </span>
        );
      },
    },
    {
      key: 'status',
      header: '状态',
      render: (row: DataRow) => {
        const s = STATUS_MAP[row.status as string] ?? { label: row.status as string, variant: 'outline' as const };
        return <Badge variant={s.variant}>{s.label}</Badge>;
      },
    },
    {
      key: 'note',
      header: '备注',
      render: (row: DataRow) => {
        const note = row.note as string | null;
        if (!note) return <span className="text-muted-foreground">-</span>;
        const display = note.length > 20 ? note.slice(0, 20) + '…' : note;
        return <span className="text-sm">{display}</span>;
      },
    },
  ];

  // 可选月份列表（从数据中提取）
  const monthOptions = useMemo(() => {
    const set = new Set(items.map((i) => i.operating_month));
    return [...set].sort().reverse();
  }, [items]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="合同进度跟踪"
        description="零售合同执行进度监控 · 月度完成率分析 · 状态跟踪"
        actions={
          canWrite ? (
            <Button variant="outline" onClick={onGenerate} disabled={generating}>
              {generating ? (
                <RefreshCw className="mr-1 h-4 w-4 animate-spin" />
              ) : (
                <BarChart3 className="mr-1 h-4 w-4" />
              )}
              {generating ? '生成中...' : '生成演示数据'}
            </Button>
          ) : undefined
        }
      />

      {notice && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{notice}</AlertDescription>
        </Alert>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 汇总卡片 */}
      {summary && (
        <div className="grid gap-4 md:grid-cols-4 lg:grid-cols-5">
          <StatCard
            title="总计划电量 (MWh)"
            value={fmt(summary.totalPlanned)}
            icon={<FileText className="h-4 w-4" />}
          />
          <StatCard
            title="总实际电量 (MWh)"
            value={fmt(summary.totalActual)}
            icon={<TrendingUp className="h-4 w-4" />}
          />
          <StatCard
            title="平均完成率 (%)"
            value={fmt(summary.avgRate, 1)}
            icon={<CheckCircle2 className="h-4 w-4" />}
          />
          <StatCard
            title="正常/超前"
            value={`${summary.onTrackCount} / ${summary.aheadCount}`}
            icon={<TrendingUp className="h-4 w-4" />}
          />
          <StatCard
            title="滞后/已完成"
            value={`${summary.behindCount} / ${summary.completedCount}`}
            icon={<TrendingDown className="h-4 w-4" />}
          />
        </div>
      )}

      {/* 偏差预警 */}
      {deviationAlerts.length > 0 && (
        <Alert className="border-red-300 bg-red-50">
          <AlertCircle className="h-4 w-4 text-red-600" />
          <AlertDescription className="text-red-700">
            <span className="font-semibold">⚠️ 偏差预警：</span>
            {deviationAlerts.map((a) => (
              <span key={`${a.name}-${a.month}`} className="inline-flex items-center gap-1 mr-3">
                <span className="font-medium">{a.name}</span>
                <span className="text-xs">({a.month})</span>
                <Badge variant="destructive" className="text-[10px] px-1 py-0">{a.rate.toFixed(0)}%</Badge>
                <span className="text-xs">差{fmt(a.gap, 0)} MWh</span>
              </span>
            ))}
          </AlertDescription>
        </Alert>
      )}

      {/* 趋势图 */}
      <div className="grid gap-6 md:grid-cols-2">
        <ChartContainer title="月度计划 vs 实际电量" minHeight={280}>
          {trendData.length > 0 ? (
            <PlanActualComposed data={trendData} />
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              暂无数据
            </div>
          )}
        </ChartContainer>

        <ChartContainer title="月度完成率趋势" minHeight={280}>
          {trendData.length > 0 ? (
            <CompletionComposed data={trendData} />
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              暂无趋势数据
            </div>
          )}
        </ChartContainer>
      </div>

      {/* ── 进度条可视化（计划 vs 实际）── */}
      <ChartContainer title="客户维度进度条（计划 vs 实际）" minHeight={Math.max(200, progressBarData.length * 44 + 60)}>
        {progressBarData.length > 0 ? (
          <div className="space-y-2">
            {progressBarData.map((row) => {
              const maxVal = Math.max(row.计划, row.实际, 1);
              const planPct = (row.计划 / maxVal) * 100;
              const actualPct = (row.实际 / maxVal) * 100;
              const isOver = row.偏差 < 0 && row.完成率 < 80;
              return (
                <div key={row.name} className="flex items-center gap-2">
                  <div className="w-20 shrink-0 truncate text-right text-xs text-zinc-500">{row.name}</div>
                  <div className="relative h-6 flex-1">
                    {/* 计划底色 */}
                    <div className="absolute inset-0 rounded bg-slate-100" style={{ width: `${planPct}%` }} />
                    {/* 实际覆盖 */}
                    <div
                      className="absolute top-0 h-full rounded"
                      style={{
                        width: `${actualPct}%`,
                        backgroundColor: isOver ? '#ef4444' : row.完成率 >= 100 ? '#10b981' : '#3b82f6',
                      }}
                    />
                  </div>
                  <div className="w-16 shrink-0 text-right">
                    <span className={`text-xs font-medium ${isOver ? 'text-red-600' : row.完成率 >= 100 ? 'text-green-600' : 'text-blue-600'}`}>
                      {row.完成率}%
                    </span>
                  </div>
                </div>
              );
            })}
            <div className="flex items-center gap-4 pt-1 text-xs text-zinc-400">
              <span className="flex items-center gap-1"><span className="inline-block h-3 w-3 rounded-sm" style={{ backgroundColor: '#3b82f6' }} /> 正常</span>
              <span className="flex items-center gap-1"><span className="inline-block h-3 w-3 rounded-sm" style={{ backgroundColor: '#10b981' }} /> 超额</span>
              <span className="flex items-center gap-1"><span className="inline-block h-3 w-3 rounded-sm" style={{ backgroundColor: '#ef4444' }} /> 偏差</span>
              <span className="flex items-center gap-1"><span className="inline-block h-3 w-3 rounded-sm" style={{ backgroundColor: '#e2e8f0' }} /> 计划</span>
            </div>
          </div>
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            暂无数据
          </div>
        )}
      </ChartContainer>

      {/* ── 多合同执行进度汇总看板 ── */}
      {dashboardCards.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <FileText className="h-4 w-4 text-primary" />
              多合同执行进度汇总看板
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {dashboardCards.map((card) => {
                const barColor = card.hasBehind
                  ? 'bg-red-500'
                  : card.rate >= 100
                    ? 'bg-emerald-500'
                    : card.rate >= 80
                      ? 'bg-blue-500'
                      : 'bg-amber-500';
                const borderColor = card.hasBehind ? 'border-red-200' : 'border-slate-200';
                return (
                  <div key={card.name} className={`rounded-lg border ${borderColor} p-3 space-y-2`}>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium truncate">{card.name}</span>
                      {card.hasBehind && (
                        <Badge variant="destructive" className="text-[10px] px-1 py-0">滞后</Badge>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="flex-1 h-2 rounded-full bg-zinc-100">
                        <div
                          className={`h-full rounded-full ${barColor}`}
                          style={{ width: `${Math.min(card.rate, 100)}%` }}
                        />
                      </div>
                      <span className="text-xs font-medium w-12 text-right">{card.rate.toFixed(0)}%</span>
                    </div>
                    <div className="flex justify-between text-[10px] text-zinc-400">
                      <span>计划: {fmt(card.planned, 0)} MWh</span>
                      <span>实际: {fmt(card.actual, 0)} MWh</span>
                    </div>
                    <div className="text-[10px] text-zinc-400">{card.recordCount} 个月度记录</div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* 进度列表 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Clock className="h-4 w-4 text-primary" />
            合同进度明细
            <span className="text-xs font-normal text-muted-foreground ml-2">
              {items.length} 条记录
            </span>
          </CardTitle>
          {/* 筛选区 */}
          <div className="flex items-center gap-3 pt-2">
            <label className="text-sm text-muted-foreground whitespace-nowrap">月份筛选</label>
            <select
              className="h-8 rounded-md border border-input bg-background px-2 text-sm"
              value={filterMonth}
              onChange={(e) => setFilterMonth(e.target.value)}
            >
              <option value="">全部月份</option>
              {monthOptions.map((m) => (
                <option key={m} value={m}>{m}</option>
              ))}
            </select>

            <label className="text-sm text-muted-foreground whitespace-nowrap">状态筛选</label>
            <select
              className="h-8 rounded-md border border-input bg-background px-2 text-sm"
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
            >
              <option value="">全部状态</option>
              <option value="on_track">正常</option>
              <option value="ahead">超前</option>
              <option value="behind">滞后</option>
              <option value="completed">已完成</option>
            </select>

            {(filterMonth || filterStatus) && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setFilterMonth('');
                  setFilterStatus('');
                }}
              >
                清除筛选
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          <DataTable
            columns={columns}
            data={items as unknown as DataRow[]}
            rowKey="id"
            pageSize={15}
            showPagination
            loading={listLoading}
          />
        </CardContent>
      </Card>
    </div>
  );
}
