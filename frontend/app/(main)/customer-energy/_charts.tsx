'use client';

// 客户历史电量趋势柱状图。
// 单独成文件以便 next/dynamic 懒加载，剥离 recharts 出首屏。
import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { ChartContainer } from '@/components/charts/chart-container';

function fmtNum(n: number | null): string {
  if (n === null || n === undefined) return '-';
  return n.toLocaleString('zh-CN', { maximumFractionDigits: 2 });
}

export function EnergyBarChart({ data }: { data: { month: string; energy: number }[] }) {
  return (
    <ChartContainer title="月度电量趋势">
      <ResponsiveContainer width="100%" height={280}>
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} />
          <XAxis dataKey="month" tick={{ fontSize: 12 }} />
          <YAxis tick={{ fontSize: 12 }} />
          <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v: number) => fmtNum(v)} />
          <Bar dataKey="energy" name="月度电量" fill="#6366f1" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </ChartContainer>
  );
}
