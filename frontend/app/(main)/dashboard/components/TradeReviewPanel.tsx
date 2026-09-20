'use client';

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
import { useQuery } from '@tanstack/react-query';
import { getSettlementSummary } from '@/lib/api/dashboard';
import { CHART_SERIES } from '@/components/charts/palette';
import { ChartLoading, EmptyState } from '@/components/feedback';

export default function TradeReviewPanel() {
  const { data, isLoading } = useQuery({
    queryKey: ['settlement-summary'],
    queryFn: () => getSettlementSummary(),
  });

  const chartData = (data?.monthly_chart ?? []).map((d: any) => ({
    label: d.label ?? '',
    purchase: Number(d.total_purchase ?? 0),
    retail: Number(d.total_retail ?? 0),
    wholesale: Number(d.total_wholesale ?? 0),
  }));

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">交易复盘 — 购售电量对比</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <ChartLoading className="h-48" />
        ) : chartData.length > 0 ? (
          <div className="h-52">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{ top: 4, right: 12, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="label" tick={{ fontSize: 11 }} />
                <YAxis
                  tick={{ fontSize: 11 }}
                  width={55}
                  tickFormatter={(v: number) => (v >= 10000 ? `${(v / 10000).toFixed(1)}万` : String(v))}
                />
                <Tooltip contentStyle={{ fontSize: 12 }} />
                <Legend wrapperStyle={{ fontSize: 12 }} />
                <Line type="monotone" dataKey="purchase" name="购电" stroke={CHART_SERIES[0]} strokeWidth={2} dot={false} />
                <Line type="monotone" dataKey="retail" name="零售" stroke={CHART_SERIES[2]} strokeWidth={2} dot={false} />
                <Line type="monotone" dataKey="wholesale" name="批发" stroke={CHART_SERIES[1]} strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <EmptyState compact className="h-32" title="暂无交易数据" />
        )}
      </CardContent>
    </Card>
  );
}
