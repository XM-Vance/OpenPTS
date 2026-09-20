'use client';

import { useState, useCallback, useMemo, useEffect } from 'react';
import { apiClient } from '@/lib/api/client';
import { useQuery } from '@tanstack/react-query';
import dynamic from 'next/dynamic';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { CHART_SERIES } from '@/components/charts/palette';
import {
  getDashboardSummary,
  getSettlementSeries,
  getFreqSeries,
} from '@/lib/api/dashboard';
import {
  type DashboardConfig,
  getDefaultConfig,
  loadConfig,
  saveConfig,
} from '@/lib/api/dashboard-config';
import {
  Users,
  FileText,
  Layers,
  AlertTriangle,
  Radio,
  BatteryCharging,
  Zap,
  DollarSign,
  ArrowUpRight,
  ArrowDownRight,
  Maximize2,
  Minimize2,
  CalendarDays,
  TrendingUp,
  Bell,
  ClipboardList,
  Settings2,
} from 'lucide-react';

import AlertsPanel from './components/AlertsPanel';
import CustomizePanel from './components/CustomizePanel';
import { EmptyState } from '@/components/feedback';

// 图表懒加载：骨架屏占位，避免加载完成后页面跳动
const chartLoading = () => (
  <div className="flex min-h-[300px] flex-col justify-center gap-4 rounded-lg border bg-card p-4">
    <Skeleton className="h-4 w-1/4" />
    <Skeleton className="h-full min-h-[220px] w-full" />
  </div>
);

const SettlementPanel = dynamic(() => import('./components/SettlementPanel'), { ssr: false, loading: chartLoading });
const TradeReviewPanel = dynamic(() => import('./components/TradeReviewPanel'), { ssr: false, loading: chartLoading });
const CustomerOverviewPanel = dynamic(() => import('./components/CustomerOverviewPanel'), { ssr: false, loading: chartLoading });
const MarketPricePanel = dynamic(() => import('./components/MarketPricePanel'), { ssr: false, loading: chartLoading });

const SparkCard = dynamic(
  () => import('./components/InlineCharts').then((m) => ({ default: m.SparkCard })),
  { ssr: false, loading: () => <Skeleton className="h-[110px] rounded-lg" /> },
);
const MarketOverviewChart = dynamic(
  () => import('./components/InlineCharts').then((m) => ({ default: m.MarketOverviewChart })),
  { ssr: false, loading: chartLoading },
);

/* ── Types ── */
type TimeRange = 'today' | 'week' | 'month';

/* ── 数据占位（真实数据待市场行情/预警/待办后端接入后填充） ──
 * P0-F3：原硬编码 Mock（市场概览/预警待办/KPI 涨跌幅）已移除，
 * 改为空数据 + 空状态展示，避免假数据与真实 KPI 混排误导决策。 */
interface MarketOverviewRow {
  name: string;
  volume: number;
  avgPrice: number;
  fill: string;
}

function useMarketOverviewData(_range: TimeRange): MarketOverviewRow[] {
  // WP1.5：接全市场购电均价口径（market_avg_price + 本公司当月电量，最近 6 个月）
  const { data } = useQuery({
    queryKey: ['dashboard-market-overview'],
    queryFn: async () => {
      const res = await apiClient.get('/api/v1/dashboard/market-overview');
      return res.data.items as { month: string; market_avg: number; own_energy_mwh: number }[];
    },
    staleTime: 5 * 60 * 1000,
  });
  return (data ?? [])
    .slice()
    .reverse()
    .map((it) => ({
      name: it.month,
      volume: Number(it.own_energy_mwh),
      avgPrice: Number(it.market_avg),
      fill: CHART_SERIES[0],
    }));
}

function useTodoAlerts() {
  // 真实预警/待办待后端接入
  return [];
}

function useKpiDeltas(_range: TimeRange): Record<string, { value: number; direction: 'up' | 'down' }> {
  // KPI 环比涨跌幅待后端接入；返回空对象，KpiCard 不渲染涨跌标识
  return {};
}

/* ── KPI Card ── */
function KpiCard({
  icon: Icon,
  label,
  value,
  sub,
  color,
  delta,
  inverted,
}: {
  icon: React.ElementType;
  label: string;
  value: string | number;
  sub?: string;
  color?: string;
  delta?: { value: number; direction: 'up' | 'down' };
  inverted?: boolean;
}) {
  const isPositive = inverted
    ? delta?.direction === 'down'
    : delta?.direction === 'up';

  return (
    <div className="flex items-center gap-3 rounded-lg border bg-card px-4 py-3 shadow-sm">
      <div
        className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg"
        style={{ backgroundColor: color ? `${color}15` : undefined }}
      >
        <Icon className="h-5 w-5" style={{ color }} />
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate text-xs text-muted-foreground">{label}</p>
        <div className="flex items-center gap-1.5">
          <p className="text-lg font-bold tabular-nums leading-tight">{value}</p>
          {delta && delta.value > 0 && (
            <div
              className={`flex shrink-0 items-center gap-0.5 rounded-full px-1.5 py-0.5 text-xs font-semibold tabular-nums ${
                isPositive
                  ? 'bg-red-50 text-red-700 dark:bg-red-500/15 dark:text-red-400'
                  : 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400'
              }`}
            >
              {delta.direction === 'up' ? (
                <ArrowUpRight className="h-3 w-3" />
              ) : (
                <ArrowDownRight className="h-3 w-3" />
              )}
              {delta.value}%
            </div>
          )}
        </div>
        {sub && <p className="text-[10px] text-muted-foreground">{sub}</p>}
      </div>
    </div>
  );
}

/* ── Timeline ── */
function TimelinePanel({ items }: { items: { id: string; type: 'alert' | 'todo'; level: string; message: string; time: string }[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-1.5 text-sm font-medium">
          <ClipboardList className="h-4 w-4 text-amber-500" />
          待办事项 & 预警通知
        </CardTitle>
      </CardHeader>
      <CardContent className="max-h-72 overflow-y-auto">
        {items.length === 0 ? (
          <EmptyState compact className="py-8" title="暂无待办与预警" />
        ) : (
        <div className="relative space-y-0">
          <div className="absolute left-[11px] top-2 bottom-2 w-px bg-border" />
          {items.map((item) => {
            const isAlert = item.type === 'alert';
            const dotColor =
              item.level === 'critical'
                ? 'bg-red-500'
                : item.level === 'warning'
                  ? 'bg-amber-500'
                  : 'bg-blue-500';
            return (
              <div key={item.id} className="relative flex items-start gap-3 pb-4 pl-7">
                <div className={`absolute left-[7px] top-1.5 h-[9px] w-[9px] rounded-full ${dotColor} ring-2 ring-white`} />
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-1.5">
                    {isAlert ? (
                      <Bell className="h-3.5 w-3.5 shrink-0 text-amber-500" />
                    ) : (
                      <ClipboardList className="h-3.5 w-3.5 shrink-0 text-blue-500" />
                    )}
                    <span className="text-xs text-muted-foreground">{item.time}</span>
                  </div>
                  <p className="mt-0.5 text-sm">{item.message}</p>
                </div>
              </div>
            );
          })}
        </div>
        )}
      </CardContent>
    </Card>
  );
}

/* ── 隐藏面板占位（仅在编辑模式下显示） ── */
function HiddenPlaceholder() {
  return (
    <div className="flex min-h-[120px] items-center justify-center rounded-lg border border-dashed bg-muted/20 text-sm text-muted-foreground">
      已隐藏
    </div>
  );
}

/* ── Main ── */
export default function DashboardPage() {
  const [timeRange, setTimeRange] = useState<TimeRange>('week');
  const [fullScreen, setFullScreen] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [config, setConfig] = useState<DashboardConfig>(getDefaultConfig());
  const [configLoaded, setConfigLoaded] = useState(false);

  // 加载用户配置
  useEffect(() => {
    loadConfig().then((c) => {
      setConfig(c);
      setConfigLoaded(true);
    });
  }, []);

  const { data: summary, isLoading: summaryLoading } = useQuery({
    queryKey: ['dashboard-summary'],
    queryFn: getDashboardSummary,
  });

  const { data: settlementSeries } = useQuery({
    queryKey: ['settlement-series-14'],
    queryFn: () => getSettlementSeries(14),
  });

  const { data: freqSeries } = useQuery({
    queryKey: ['freq-series-14'],
    queryFn: () => getFreqSeries(14),
  });

  const settlementData = (settlementSeries?.items ?? []).map((p) => ({
    date: p.date.slice(5),
    value: p.value,
  }));

  const freqData = (freqSeries?.items ?? []).map((p) => ({
    date: p.date.slice(5),
    value: p.value,
  }));

  const marketData = useMarketOverviewData(timeRange);
  const todoAlerts = useTodoAlerts();
  const kpiDeltas = useKpiDeltas(timeRange);

  const enterFS = useCallback(() => {
    document.documentElement.requestFullscreen?.();
    setFullScreen(true);
  }, []);

  const exitFS = useCallback(() => {
    document.exitFullscreen?.();
    setFullScreen(false);
  }, []);

  // 配置快捷判断
  const isVisible = (id: string) => config.widgets.find((w) => w.id === id)?.visible ?? true;

  // KPI 卡片不参与隐藏/排列（作为核心数据始终显示）
  // 可配置的面板按 config.widgets 的顺序和可见性渲染

  // 获取所有可见的 panel 类型 widget（按用户排序）
  const visiblePanels = config.widgets.filter(
    (w) => w.visible && ['market_overview', 'timeline', 'settlement_panel', 'trade_review', 'customer_overview', 'market_price', 'alerts'].includes(w.id),
  );
  const visibleSparks = config.widgets.filter(
    (w) => w.visible && ['spark_settlement', 'spark_freq'].includes(w.id),
  );

  // 渲染单个 widget 内容
  const renderWidget = (id: string) => {
    switch (id) {
      case 'market_overview':
        return marketData.length > 0 ? (
          <MarketOverviewChart data={marketData} />
        ) : (
          <p className="py-8 text-center text-sm text-muted-foreground">
            市场行情汇总待接入（全市场购电均价口径开发中）
          </p>
        );
      case 'timeline':
        return <TimelinePanel items={todoAlerts} />;
      case 'settlement_panel':
        return <SettlementPanel />;
      case 'trade_review':
        return <TradeReviewPanel />;
      case 'customer_overview':
        return <CustomerOverviewPanel />;
      case 'market_price':
        return <MarketPricePanel />;
      case 'alerts':
        return <AlertsPanel />;
      default:
        return null;
    }
  };

  // 渲染 spark widget
  const renderSpark = (id: string) => {
    switch (id) {
      case 'spark_settlement':
        return (
          <SparkCard title="近 14 日结算额" data={settlementData} color="#6366f1" icon={Zap} />
        );
      case 'spark_freq':
        return (
          <SparkCard
            title="近 14 日调频收益"
            data={freqData}
            color="#f97316"
            icon={Radio}
            fmt={(v) => (v >= 10000 ? `${(v / 10000).toFixed(2)}万` : v.toLocaleString())}
          />
        );
      default:
        return null;
    }
  };

  // 将可见面板两两配对成行
  const panelRows: string[][] = [];
  for (let i = 0; i < visiblePanels.length; i += 2) {
    panelRows.push(visiblePanels.slice(i, i + 2).map((w) => w.id));
  }

  return (
    <div className="space-y-6">
      {/* Page header + controls */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">仪表盘</h1>
          <p className="text-sm text-muted-foreground">
            电力交易信息系统 — 核心业务数据概览
          </p>
        </div>
        <div className="flex items-center gap-2">
          {/* Time range selector */}
          <div className="flex rounded-lg border bg-card p-0.5 shadow-sm">
            {(['today', 'week', 'month'] as const).map((r) => (
              <Button
                key={r}
                size="sm"
                variant={timeRange === r ? 'default' : 'ghost'}
                className="h-7 px-3 text-xs"
                onClick={() => setTimeRange(r)}
              >
                {r === 'today' ? '今日' : r === 'week' ? '本周' : '本月'}
              </Button>
            ))}
          </div>
          {/* Customize toggle */}
          <Button
            size="sm"
            variant={editMode ? 'default' : 'outline'}
            className="h-7 gap-1 text-xs"
            onClick={() => setEditMode(!editMode)}
          >
            <Settings2 className="h-3.5 w-3.5" />
            {editMode ? '完成' : '自定义'}
          </Button>
          {/* Fullscreen toggle */}
          <Button
            size="sm"
            variant="outline"
            className="h-7 gap-1 text-xs"
            onClick={fullScreen ? exitFS : enterFS}
          >
            {fullScreen ? (
              <>
                <Minimize2 className="h-3.5 w-3.5" />
                退出全屏
              </>
            ) : (
              <>
                <Maximize2 className="h-3.5 w-3.5" />
                全屏
              </>
            )}
          </Button>
        </div>
      </div>

      {/* Edit mode: customization panel */}
      {editMode && (
        <CustomizePanel config={config} onChange={setConfig} />
      )}

      {/* Row 1: KPI stat cards (always visible) */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
        <KpiCard icon={Users} label="客户总数" value={summary?.customer_count ?? '-'} color="#6366f1" delta={kpiDeltas.customer_count} />
        <KpiCard icon={FileText} label="活跃合同" value={summary?.active_contracts ?? '-'} color="#3b82f6" delta={kpiDeltas.active_contracts} />
        <KpiCard icon={Layers} label="活跃套餐" value={summary?.active_packages ?? '-'} color="#10b981" delta={kpiDeltas.active_packages} />
        <KpiCard
          icon={AlertTriangle}
          label="待处理告警"
          value={summary?.pending_alerts ?? '-'}
          sub={summary?.critical_alerts ? `严重 ${summary.critical_alerts}` : undefined}
          color="#f97316"
          delta={kpiDeltas.pending_alerts}
          inverted
        />
        <KpiCard icon={Radio} label="活跃电站" value={summary?.active_stations ?? '-'} color="#8b5cf6" delta={kpiDeltas.active_stations} />
        <KpiCard
          icon={DollarSign}
          label="最新结算额"
          value={summary?.latest_settlement_fee != null ? summary.latest_settlement_fee.toLocaleString() : '-'}
          color="#ec4899"
          delta={kpiDeltas.latest_settlement_fee}
        />
      </div>

      {/* Row 2: Sparklines (configurable) */}
      {visibleSparks.length > 0 && (
        <div className={`grid grid-cols-1 gap-4 ${visibleSparks.length > 1 ? 'md:grid-cols-2' : ''}`}>
          {visibleSparks.map((w) => (
            <div key={w.id}>{renderSpark(w.id)}</div>
          ))}
        </div>
      )}

      {/* Row 3+: Configurable panels (rendered in pairs by user order) */}
      {panelRows.map((row, i) => (
        <div key={i} className={`grid grid-cols-1 gap-4 ${row.length > 1 ? 'lg:grid-cols-2' : ''}`}>
          {row.map((id) => (
            <div key={id}>{renderWidget(id)}</div>
          ))}
        </div>
      ))}

      {/* Empty state: all panels hidden */}
      {visiblePanels.length === 0 && !editMode && (
        <div className="flex min-h-[200px] items-center justify-center rounded-lg border border-dashed bg-muted/20 text-sm text-muted-foreground">
          所有面板已隐藏。点击「自定义」重新开启。
        </div>
      )}
    </div>
  );
}
