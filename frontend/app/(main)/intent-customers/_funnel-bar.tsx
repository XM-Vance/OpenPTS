'use client';

// 意向客户跟进漏斗（横向条形图）。单独成文件以便 next/dynamic 懒加载，剥离 recharts 出首屏。
import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip as RechartsTooltip, XAxis, YAxis } from 'recharts';

export interface FunnelBarDatum {
  label: string;
  count: number;
  color: string;
}

export function FunnelBar({ data }: { data: FunnelBarDatum[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart
        data={data}
        layout="vertical"
        margin={{ top: 8, right: 32, bottom: 8, left: 8 }}
      >
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" horizontal={false} />
        <XAxis type="number" tick={{ fontSize: 12, fill: '#6b7280' }} />
        <YAxis
          type="category"
          dataKey="label"
          width={80}
          tick={{ fontSize: 12, fill: '#374151' }}
        />
        <RechartsTooltip
          formatter={(v: number) => [`${v} 家`, '客户数']}
          contentStyle={{ fontSize: 12 }}
        />
        <Bar dataKey="count" radius={[0, 6, 6, 0]} isAnimationActive={false}>
          {data.map((entry, index) => (
            <Cell key={index} fill={entry.color} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}
