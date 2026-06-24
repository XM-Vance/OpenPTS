'use client';

// 代理商图表（业绩排行 TOP10 横向条形 + 区域分布饼图）。
// 单独成文件以便 next/dynamic 懒加载，剥离 recharts 出首屏。
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
} from 'recharts';

export function AgentTop10Bar({ data, colors }: { data: any[]; colors: string[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart
        data={data}
        layout="vertical"
        margin={{ top: 8, right: 24, bottom: 8, left: 8 }}
      >
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" horizontal={false} />
        <XAxis type="number" tick={{ fontSize: 11, fill: '#6b7280' }} unit="%" />
        <YAxis type="category" dataKey="name" width={72} tick={{ fontSize: 11, fill: '#374151' }} />
        <RechartsTooltip
          formatter={(v: number) => [`${v}%`, '佣金比例']}
          contentStyle={{ fontSize: 12 }}
        />
        <Bar dataKey="rate" radius={[0, 6, 6, 0]} isAnimationActive={false}>
          {data.map((_, index) => (
            <Cell key={index} fill={colors[index % colors.length]} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}

export function AgentRegionPie({ data, colors }: { data: any[]; colors: string[] }) {
  return (
    <ResponsiveContainer width="60%" height={240}>
      <PieChart>
        <Pie
          data={data}
          dataKey="value"
          nameKey="name"
          cx="50%"
          cy="50%"
          outerRadius={90}
          isAnimationActive={false}
        >
          {data.map((_, index) => (
            <Cell key={index} fill={colors[index % colors.length]} />
          ))}
        </Pie>
        <RechartsTooltip
          formatter={(v: number, name: string) => [`${v} 家`, name]}
          contentStyle={{ fontSize: 12 }}
        />
      </PieChart>
    </ResponsiveContainer>
  );
}
