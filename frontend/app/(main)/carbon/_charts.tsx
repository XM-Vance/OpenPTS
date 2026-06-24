'use client';

// 碳价走势图（CEA/CCER/EUA 收盘价对比）。
// 单独成文件以便 next/dynamic 懒加载，把 recharts 剥离出首屏。
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

export interface CarbonTrendPoint {
  date: string;
  CEA: number | null;
  CCER: number | null;
  EUA: number | null;
}

export function CarbonTrendLine({ data }: { data: CarbonTrendPoint[] }) {
  return (
    <ResponsiveContainer width="100%" height={320}>
      <LineChart data={data} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
        <XAxis dataKey="date" tick={{ fontSize: 11 }} minTickGap={32} />
        <YAxis tick={{ fontSize: 11 }} width={52} />
        <Tooltip
          contentStyle={{ fontSize: 12 }}
          formatter={(v: number, name: string) => [v?.toLocaleString('zh-CN'), name]}
        />
        <Legend wrapperStyle={{ fontSize: 12 }} />
        <Line
          type="monotone"
          dataKey="CEA"
          stroke="#10b981"
          strokeWidth={2}
          dot={false}
          name="CEA 全国碳配额(元)"
          connectNulls
          isAnimationActive={false}
        />
        <Line
          type="monotone"
          dataKey="CCER"
          stroke="#3b82f6"
          strokeWidth={2}
          dot={false}
          name="CCER 自愿减排(元)"
          connectNulls
          isAnimationActive={false}
        />
        <Line
          type="monotone"
          dataKey="EUA"
          stroke="#f59e0b"
          strokeWidth={2}
          dot={false}
          name="EUA 欧盟配额(欧元)"
          connectNulls
          isAnimationActive={false}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
