'use client';

import { useState } from 'react';
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
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useQuery } from '@tanstack/react-query';
import { getSettlementSummary } from '@/lib/api/dashboard';

type ViewMode = 'monthly' | 'yearly';

export default function SettlementPanel() {
  const [mode, setMode] = useState<ViewMode>('monthly');

  const { data, isLoading } = useQuery({
    queryKey: ['settlement-summary'],
    queryFn: () => getSettlementSummary(),
  });

  const kpi = data?.kpi;
  const monthlyChart = data?.monthly_chart ?? [];
  const yearlyChart = data?.yearly_chart ?? [];

  const chartData = mode === 'monthly' ? monthlyChart : yearlyChart;
  const profitKey = mode === 'monthly' ? 'monthly_gross_profit' : 'yearly_gross_profit';

  const fmt = (v: number | null | undefined, decimals = 2) => {
    if (v == null) return '-';
    if (Math.abs(v) >= 10000) return (v / 10000).toFixed(decimals) + ' 万';
    return v.toFixed(decimals);
  };

  const kpis = kpi
    ? [
        {
          title: '年度累计毛利',
          value: fmt(kpi.yearly_gross_profit),
          color: Number(kpi.yearly_gross_profit ?? 0) >= 0 ? 'text-emerald-600' : 'text-red-500',
        },
        {
          title: '月度毛利',
          value: fmt(kpi.monthly_gross_profit),
          color: Number(kpi.monthly_gross_profit ?? 0) >= 0 ? 'text-emerald-600' : 'text-red-500',
        },
        {
          title: '购售电价差',
          value: fmt(kpi.price_spread, 3) + ' 元/MWh',
          color: Number(kpi.price_spread ?? 0) >= 0 ? 'text-emerald-600' : 'text-red-500',
        },
        {
          title: '售电均价',
          value: fmt(kpi.retail_avg_price, 3) + ' 元/MWh',
          color: 'text-indigo-600',
        },
      ]
    : [];

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-base">结算概览</CardTitle>
        <div className="flex gap-1">
          <Button
            variant={mode === 'monthly' ? 'default' : 'outline'}
            size="sm"
            className="h-7 text-xs"
            onClick={() => setMode('monthly')}
          >
            月度
          </Button>
          <Button
            variant={mode === 'yearly' ? 'default' : 'outline'}
            size="sm"
            className="h-7 text-xs"
            onClick={() => setMode('yearly')}
          >
            年度
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* KPI cards */}
        <div className="grid grid-cols-2 gap-3">
          {kpis.map((k) => (
            <div key={k.title} className="rounded-lg border bg-muted/30 px-3 py-2">
              <div className="text-xs text-muted-foreground">{k.title}</div>
              <div className={`text-lg font-bold ${k.color}`}>{k.value}</div>
            </div>
          ))}
        </div>

        {/* Bar chart */}
        {isLoading ? (
          <div className="flex h-48 items-center justify-center text-sm text-muted-foreground">
            加载中...
          </div>
        ) : chartData.length > 0 ? (
          <div className="h-52">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData} margin={{ top: 4, right: 12, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="label" tick={{ fontSize: 11, fill: '#6b7280' }} />
                <YAxis
                  tick={{ fontSize: 11, fill: '#6b7280' }}
                  width={55}
                  tickFormatter={(v: number) => (v >= 10000 ? `${(v / 10000).toFixed(1)}万` : String(v))}
                />
                <Tooltip
                  formatter={(v: number) => `${fmt(v)}`}
                  contentStyle={{ fontSize: 12 }}
                />
                <Bar dataKey={profitKey} name="毛利" radius={[4, 4, 0, 0]}>
                  {chartData.map((entry, idx) => (
                    <Cell
                      key={idx}
                      fill={Number((entry as unknown as Record<string, unknown>)[profitKey] ?? 0) >= 0 ? '#10b981' : '#ef4444'}
                    />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
            暂无结算数据
          </div>
        )}
      </CardContent>
    </Card>
  );
}
