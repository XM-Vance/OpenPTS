'use client';

// 分时电价时段条形图。单独成文件以便 next/dynamic 懒加载，剥离 recharts 出首屏。
import { Bar, BarChart, Cell, ResponsiveContainer, Tooltip, XAxis } from 'recharts';

export interface TouBarDatum {
  label: string;
  value: number;
  color: string;
  tagLabel?: string;
  tag?: string;
}

export function TouBar({ data }: { data: TouBarDatum[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart data={data} margin={{ top: 4, right: 12, left: 0, bottom: 4 }}>
        <XAxis
          dataKey="label"
          tick={{ fontSize: 10, fill: '#6b7280' }}
          interval={1}
        />
        <Tooltip
          formatter={(_v: number, _name: string, props: { payload?: { tagLabel?: string; tag?: string } }) => {
            return [props.payload?.tagLabel ?? '', '时段'];
          }}
          contentStyle={{ fontSize: 12 }}
        />
        <Bar dataKey="value" isAnimationActive={false} name="时段">
          {data.map((entry, index) => (
            <Cell key={index} fill={entry.color} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}
