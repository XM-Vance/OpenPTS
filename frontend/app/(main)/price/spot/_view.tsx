'use client';

import { useState, useMemo } from 'react';
import { format, subDays } from 'date-fns';
import { useQuery } from '@tanstack/react-query';
import {
  CartesianGrid,
  ComposedChart,
  Line,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { usePermission } from '@/lib/auth/use-permission';
import {
  getSpotMarketStatistics,
  getSpotMarketPriceCurve,
  getSpotMarketList,
} from '@/lib/api/spot-market';
import { StatCard } from '@/components/data-display/stat-card';
import { ChartContainer } from '@/components/charts/chart-container';
import { CustomTooltip } from '@/components/charts/custom-tooltip';
import { DataTable, type DataTableColumn, type DataRow } from '@/components/data-display/data-table';
import { PageHeader } from '@/components/data-display/page-header';
import {
  DollarSign,
  TrendingUp,
  TrendingDown,
  BarChart3,
  Activity,
  Eye,
  EyeOff,
} from 'lucide-react';

function pointTime(i: number, points = 48): string {
  const mins = (i * 24 * 60) / points;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
}

function fmt(v: number | null | undefined, digits = 2): string {
  if (v == null) return '-';
  return v.toLocaleString('zh-CN', { maximumFractionDigits: digits });
}

const FIELD_CLASS =
  'flex h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm';

export default function SpotPricePage() {
  const { has } = usePermission();

  const [dateRange, setDateRange] = useState({
    start: format(subDays(new Date(), 7), 'yyyy-MM-dd'),
    end: format(new Date(), 'yyyy-MM-dd'),
  });
  const [selectedDate, setSelectedDate] = useState(format(new Date(), 'yyyy-MM-dd'));
  const [showDA, setShowDA] = useState(true);
  const [showRT, setShowRT] = useState(true);
  const [showForecast, setShowForecast] = useState(true);

  // 统计
  const { data: statsRes } = useQuery({
    queryKey: ['spot-market', 'statistics', dateRange.start, dateRange.end],
    queryFn: () => getSpotMarketStatistics({ start_date: dateRange.start, end_date: dateRange.end }),
    enabled: !!dateRange.start && !!dateRange.end,
  });
  const statistics = (statsRes as any)?.data;

  // 价格曲线
  const { data: curveRes, isLoading: curveLoading } = useQuery({
    queryKey: ['spot-market', 'curve', selectedDate],
    queryFn: () => getSpotMarketPriceCurve({ target_date: selectedDate }),
    enabled: !!selectedDate,
  });
  const priceCurve = (curveRes as any)?.data;

  // 列表
  const { data: listRes, isLoading: listLoading } = useQuery({
    queryKey: ['spot-market', 'list', dateRange.start, dateRange.end],
    queryFn: () => getSpotMarketList({ start_date: dateRange.start, end_date: dateRange.end, limit: 100 }),
    enabled: !!dateRange.start && !!dateRange.end,
  });
  const spotList = (listRes as any)?.data;

  // 统计卡片
  const statsCards = useMemo(() => {
    if (!statistics) return null;
    return {
      maxPrice: statistics.max_price ?? statistics.high ?? null,
      minPrice: statistics.min_price ?? statistics.low ?? null,
      avgPrice: statistics.avg_price ?? statistics.average ?? null,
      volatility: statistics.volatility ?? null,
      daAvgPrice: statistics.da_avg_price ?? null,
      rtAvgPrice: statistics.rt_avg_price ?? null,
    };
  }, [statistics]);

  // 曲线数据
  const curveData = useMemo(() => {
    if (!priceCurve) return [];
    const daPrices: number[] = priceCurve.day_ahead_prices ?? priceCurve.da_prices ?? [];
    const rtPrices: number[] = priceCurve.realtime_prices ?? priceCurve.rt_prices ?? [];
    const forecastPrices: number[] = priceCurve.forecast_prices ?? priceCurve.da_forecast ?? [];
    const points = Math.max(daPrices.length, rtPrices.length, forecastPrices.length, 1);
    return Array.from({ length: points }, (_, i) => ({
      t: pointTime(i, points),
      日前价格: daPrices[i] ?? null,
      实时价格: rtPrices[i] ?? null,
      预测价格: forecastPrices[i] ?? null,
    }));
  }, [priceCurve]);

  // 列表列
  const columns: DataTableColumn[] = [
    { key: 'date', header: '日期', sortable: true },
    {
      key: 'da_avg_price',
      header: '日前均价 (¥/MWh)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => fmt(row.da_avg_price ?? row.day_ahead_price, 2),
    },
    {
      key: 'rt_avg_price',
      header: '实时均价 (¥/MWh)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => fmt(row.rt_avg_price ?? row.realtime_price, 2),
    },
    {
      key: 'max_price',
      header: '最高价',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => (
        <span className="text-destructive font-medium">{fmt(row.max_price ?? row.high, 2)}</span>
      ),
    },
    {
      key: 'min_price',
      header: '最低价',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => (
        <span className="text-emerald-600 font-medium">{fmt(row.min_price ?? row.low, 2)}</span>
      ),
    },
    {
      key: 'volatility',
      header: '波动率 (%)',
      align: 'right',
      render: (row: DataRow) => {
        const v = row.volatility as number;
        return v != null ? `${(v * 100).toFixed(1)}%` : '-';
      },
    },
  ];

  const listData = useMemo(() => {
    if (!spotList) return [];
    return Array.isArray(spotList) ? spotList : spotList?.items ?? [];
  }, [spotList]);

  // 多日价格走势数据
  const multiDayData = useMemo(() => {
    return (listData as any[]).map((row: any) => ({
      date: (row.date ?? row.trading_date ?? '').slice(5, 10),
      日前均价: row.da_avg_price ?? row.day_ahead_price ?? 0,
      实时均价: row.rt_avg_price ?? row.realtime_price ?? 0,
    }));
  }, [listData]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="现货价格"
        description="现货市场价格走势 · 日前 / 实时价格对比 · 统计指标"
      />

      {/* 日期筛选 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Activity className="h-4 w-4 text-primary" />
            日期范围
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-4">
            <div className="space-y-2">
              <Label>起始日期</Label>
              <input
                type="date"
                value={dateRange.start}
                onChange={(e) => setDateRange((prev) => ({ ...prev, start: e.target.value }))}
                className={FIELD_CLASS}
              />
            </div>
            <div className="space-y-2">
              <Label>结束日期</Label>
              <input
                type="date"
                value={dateRange.end}
                onChange={(e) => setDateRange((prev) => ({ ...prev, end: e.target.value }))}
                className={FIELD_CLASS}
              />
            </div>
            <div className="space-y-2">
              <Label>曲线日期</Label>
              <input
                type="date"
                value={selectedDate}
                onChange={(e) => setSelectedDate(e.target.value)}
                className={FIELD_CLASS}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-5">
        <StatCard title="期间最高价 (¥/MWh)" value={statsCards?.maxPrice != null ? fmt(statsCards.maxPrice, 2) : '-'} icon={<TrendingUp className="h-4 w-4" />} />
        <StatCard title="期间最低价 (¥/MWh)" value={statsCards?.minPrice != null ? fmt(statsCards.minPrice, 2) : '-'} icon={<TrendingDown className="h-4 w-4" />} />
        <StatCard title="期间均价 (¥/MWh)" value={statsCards?.avgPrice != null ? fmt(statsCards.avgPrice, 2) : '-'} icon={<DollarSign className="h-4 w-4" />} />
        <StatCard title="日前均价" value={statsCards?.daAvgPrice != null ? fmt(statsCards.daAvgPrice, 2) : '-'} icon={<BarChart3 className="h-4 w-4" />} />
        <StatCard title="波动率" value={statsCards?.volatility != null ? `${(statsCards.volatility * 100).toFixed(1)}%` : '-'} icon={<Activity className="h-4 w-4" />} />
      </div>

      {/* 价格走势曲线 */}
      <ChartContainer
        title={`${selectedDate} 现货价格曲线`}
        minHeight={360}
        actions={
          <>
            <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={() => setShowDA(!showDA)}>
              {showDA ? <EyeOff className="mr-1 h-3 w-3" /> : <Eye className="mr-1 h-3 w-3" />} 日前
            </Button>
            <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={() => setShowRT(!showRT)}>
              {showRT ? <EyeOff className="mr-1 h-3 w-3" /> : <Eye className="mr-1 h-3 w-3" />} 实时
            </Button>
            <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={() => setShowForecast(!showForecast)}>
              {showForecast ? <EyeOff className="mr-1 h-3 w-3" /> : <Eye className="mr-1 h-3 w-3" />} 预测
            </Button>
          </>
        }
      >
        {curveLoading ? (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">加载价格曲线...</div>
        ) : curveData.length > 0 ? (
          <ResponsiveContainer width="100%" height="100%">
            <ComposedChart data={curveData} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis dataKey="t" interval={11} tick={{ fontSize: 12 }} />
              <YAxis tick={{ fontSize: 12 }} width={64} />
              <RechartsTooltip content={<CustomTooltip unit="¥/MWh" />} />
              {showDA && <Line dataKey="日前价格" stroke="#2563eb" strokeWidth={2} dot={false} isAnimationActive={false} name="日前价格" />}
              {showRT && <Line dataKey="实时价格" stroke="#f59e0b" strokeWidth={2} strokeDasharray="6 3" dot={false} isAnimationActive={false} name="实时价格" />}
              {showForecast && <Line dataKey="预测价格" stroke="#10b981" strokeWidth={1.5} strokeDasharray="3 3" dot={false} isAnimationActive={false} name="预测价格" />}
            </ComposedChart>
          </ResponsiveContainer>
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">暂无曲线数据</div>
        )}
      </ChartContainer>

      {/* 多日走势 */}
      {multiDayData.length > 0 && (
        <ChartContainer title="多日价格走势" minHeight={260}>
          <ResponsiveContainer width="100%" height="100%">
            <ComposedChart data={multiDayData} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis dataKey="date" tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 11 }} width={64} />
              <RechartsTooltip content={<CustomTooltip unit="¥/MWh" />} />
              <Line type="monotone" dataKey="日前均价" stroke="#2563eb" strokeWidth={2} dot={false} isAnimationActive={false} name="日前均价" />
              <Line type="monotone" dataKey="实时均价" stroke="#f59e0b" strokeWidth={2} strokeDasharray="6 3" dot={false} isAnimationActive={false} name="实时均价" />
            </ComposedChart>
          </ResponsiveContainer>
        </ChartContainer>
      )}

      {/* 价格列表 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <BarChart3 className="h-4 w-4 text-primary" />
            现货价格列表
            <span className="text-xs font-normal text-muted-foreground ml-2">{listData.length} 条记录</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <DataTable
            columns={columns}
            data={listData as DataRow[]}
            rowKey={(row: DataRow) => (row as any).id ?? (row as any).date ?? Math.random().toString()}
            pageSize={10}
            showPagination
            loading={listLoading}
          />
        </CardContent>
      </Card>
    </div>
  );
}
