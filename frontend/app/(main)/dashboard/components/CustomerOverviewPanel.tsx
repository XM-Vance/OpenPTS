'use client';

import {
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useQuery } from '@tanstack/react-query';
import { getSettlementSummary } from '@/lib/api/dashboard';
import { CHART_SERIES } from '@/components/charts/palette';
import { ChartLoading, EmptyState } from '@/components/feedback';

const COLORS = [...CHART_SERIES];

export default function CustomerOverviewPanel() {
  const { data, isLoading } = useQuery({
    queryKey: ['settlement-summary'],
    queryFn: () => getSettlementSummary(),
  });

  const overview = data?.customer_overview;
  const totalCustomers = overview?.total ?? 0;
  const byType = overview?.by_type ?? {};
  const byStatus = overview?.by_status ?? {};

  const pieData = Object.entries(byType).map(([name, value]) => ({
    name,
    value: value as number,
  }));

  const statusEntries = Object.entries(byStatus);

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">客户概览</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-baseline gap-2">
          <span className="text-3xl font-bold text-indigo-600">{totalCustomers}</span>
          <span className="text-sm text-muted-foreground">个客户</span>
        </div>

        {/* By type pie */}
        {isLoading ? (
          <ChartLoading className="h-40" />
        ) : pieData.length > 0 ? (
          <div className="h-44">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={40}
                  outerRadius={65}
                  paddingAngle={2}
                  dataKey="value"
                  nameKey="name"
                >
                  {pieData.map((_, idx) => (
                    <Cell key={idx} fill={COLORS[idx % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip contentStyle={{ fontSize: 12 }} />
                <Legend wrapperStyle={{ fontSize: 11 }} />
              </PieChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <EmptyState compact className="h-24" title="暂无客户数据" />
        )}

        {/* By status badges */}
        {statusEntries.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {statusEntries.map(([status, count]) => (
              <Badge key={status} variant="outline" className="text-xs">
                {status}: {count as number}
              </Badge>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
