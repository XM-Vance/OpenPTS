'use client';

// 合同状态分布环形图。单独成文件以便页面用 next/dynamic 懒加载，
// 把 recharts 从 /retail/contracts 首屏 JS 中剥离。
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip as RechartsTooltip } from 'recharts';

export interface StatusPieDatum {
  name: string;
  value: number;
  color: string;
}

export function StatusPie({ data }: { data: StatusPieDatum[] }) {
  return (
    <ResponsiveContainer width="100%" height={180}>
      <PieChart>
        <Pie
          data={data}
          dataKey="value"
          nameKey="name"
          cx="50%"
          cy="50%"
          innerRadius={50}
          outerRadius={75}
          isAnimationActive={false}
        >
          {data.map((entry, index) => (
            <Cell key={index} fill={entry.color} />
          ))}
        </Pie>
        <RechartsTooltip
          formatter={(v: number, name: string) => [`${v} 个`, name]}
          contentStyle={{ fontSize: 12 }}
        />
      </PieChart>
    </ResponsiveContainer>
  );
}
