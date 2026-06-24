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
          <div className="flex h-48 items-center justify-center text-sm text-muted-foreground">
            加载中...
          </div>
        ) : chartData.length > 0 ? (
          <div className="h-52">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{ top: 4, right: 12, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="label" tick={{ fontSize: 11, fill: '#6b7280' }} />
                <YAxis
                  tick={{ fontSize: 11, fill: '#6b7280' }}
                  width={55}
                  tickFormatter={(v: number) => (v >= 10000 ? `${(v / 10000).toFixed(1)}万` : String(v))}
                />
                <Tooltip contentStyle={{ fontSize: 12 }} />
                <Legend wrapperStyle={{ fontSize: 12 }} />
                <Line type="monotone" dataKey="purchase" name="购电" stroke="#3b82f6" strokeWidth={2} dot={false} />
                <Line type="monotone" dataKey="retail" name="零售" stroke="#10b981" strokeWidth={2} dot={false} />
                <Line type="monotone" dataKey="wholesale" name="批发" stroke="#f97316" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
            暂无交易数据
          </div>
        )}
      </CardContent>
    </Card>
  );
}
