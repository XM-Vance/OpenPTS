'use client';

// 仪表盘内联图表（Sparkline + 市场概览柱状图）。
// 单独成文件以便在 page.tsx 中用 next/dynamic 懒加载——
// 把 recharts 从仪表盘（登录后落地页）的首屏 JS 中剥离，改为挂载后异步加载。
import React from 'react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { BarChart3 } from 'lucide-react';

/* ── Sparkline ── */
export function SparkCard({
  title,
  data,
  color,
  icon: Icon,
  fmt,
}: {
  title: string;
  data: { date: string; value: number }[];
  color: string;
  icon: React.ElementType;
  fmt?: (v: number) => string;
}) {
  const last = data.length > 0 ? data[data.length - 1].value : 0;
  return (
    <Card className="flex flex-col">
      <CardHeader className="flex flex-row items-center justify-between pb-1">
        <CardTitle className="flex items-center gap-1.5 text-sm font-medium">
          <Icon className="h-4 w-4" style={{ color }} />
          {title}
        </CardTitle>
        <Badge variant="secondary" className="text-xs font-bold">
          {fmt ? fmt(last) : last.toLocaleString()}
        </Badge>
      </CardHeader>
      <CardContent className="flex-1 pb-2">
        {data.length > 0 ? (
          <div className="h-16">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={data} margin={{ top: 2, right: 4, left: 4, bottom: 2 }}>
                <Line
                  type="monotone"
                  dataKey="value"
                  stroke={color}
                  strokeWidth={1.5}
                  dot={false}
                />
                <Tooltip
                  formatter={(v: number) => (fmt ? fmt(v) : v.toLocaleString())}
                  contentStyle={{ fontSize: 11 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">暂无数据</p>
        )}
      </CardContent>
    </Card>
  );
}

/* ── Market Overview Bar Chart ── */
export function MarketOverviewChart({ data }: { data: { name: string; volume: number; avgPrice: number; fill: string }[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-1.5 text-sm font-medium">
          <BarChart3 className="h-4 w-4 text-blue-500" />
          市场交易概览
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" tick={{ fontSize: 12 }} />
              <YAxis yAxisId="volume" orientation="left" tick={{ fontSize: 11 }} label={{ value: '成交量(MWh)', angle: -90, position: 'insideLeft', style: { fontSize: 11 } }} />
              <YAxis yAxisId="price" orientation="right" tick={{ fontSize: 11 }} label={{ value: '均价(元)', angle: 90, position: 'insideRight', style: { fontSize: 11 } }} />
              <Tooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(v: number, name: string) => {
                  if (name === '成交量') return [`${v.toLocaleString()} MWh`, name];
                  return [`¥${v.toFixed(1)}`, name];
                }}
              />
              <Legend />
              <Bar yAxisId="volume" dataKey="volume" name="成交量" radius={[4, 4, 0, 0]}>
                {data.map((entry, index) => (
                  <Cell key={index} fill={entry.fill} />
                ))}
              </Bar>
              <Bar yAxisId="price" dataKey="avgPrice" name="均价" fill="#f59e0b" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
