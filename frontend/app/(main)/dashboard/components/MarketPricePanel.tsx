'use client';

import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useQuery } from '@tanstack/react-query';
import { getSettlementSeries } from '@/lib/api/dashboard';

export default function MarketPricePanel() {
  const { data, isLoading } = useQuery({
    queryKey: ['settlement-series-14'],
    queryFn: () => getSettlementSeries(14),
  });

  const chartData = (data?.items ?? []).map((p) => ({
    date: p.date.slice(5),
    value: p.value,
  }));

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">结算趋势（近 14 天）</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex h-48 items-center justify-center text-sm text-muted-foreground">
            加载中...
          </div>
        ) : chartData.length > 0 ? (
          <div className="h-48">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{ top: 4, right: 12, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#6b7280' }} />
                <YAxis
                  tick={{ fontSize: 11, fill: '#6b7280' }}
                  width={55}
                  tickFormatter={(v: number) =>
                    v >= 10000 ? `${(v / 10000).toFixed(1)}万` : String(v)
                  }
                />
                <Tooltip
                  formatter={(v: number) => v.toLocaleString()}
                  contentStyle={{ fontSize: 12 }}
                />
                <Line
                  type="monotone"
                  dataKey="value"
                  name="结算额"
                  stroke="#6366f1"
                  strokeWidth={2}
                  dot={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
            暂无数据
          </div>
        )}
      </CardContent>
    </Card>
  );
}
