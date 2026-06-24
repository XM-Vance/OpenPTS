import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
} from 'recharts';
import type { DisplayTrendItem } from '@/lib/api/display-screen';
import { C } from './constants';
import { ChartTooltip } from './chart-tooltip';

interface BottomBarProps {
  trend: DisplayTrendItem[];
}

export function BottomBar({ trend }: BottomBarProps) {
  const energyData = trend.map((d) => ({
    date: d.date.slice(5),
    energy: d.energy_mwh,
    revenue: d.revenue,
  }));

  return (
    <div
      style={{
        gridColumn: '1 / 4',
        background: C.cardBg,
        border: `1px solid ${C.cardBorder}`,
        borderRadius: 8,
        padding: '10px 16px 4px',
        backdropFilter: 'blur(8px)',
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 4,
        }}
      >
        <span
          style={{
            fontSize: 12,
            color: C.blue,
            fontWeight: 600,
            letterSpacing: 1,
          }}
        >
          营收日报
        </span>
        <span style={{ fontSize: 10, color: C.textDim }}>
          单位：万元
        </span>
      </div>
      <ResponsiveContainer width="100%" height="82%">
        <BarChart data={energyData}>
          <XAxis
            dataKey="date"
            stroke={C.textDim}
            tick={{ fill: C.textDim, fontSize: 10 }}
            axisLine={{ stroke: C.divider }}
            tickLine={false}
          />
          <YAxis
            stroke={C.textDim}
            tick={{ fill: C.textDim, fontSize: 9 }}
            axisLine={false}
            tickLine={false}
            width={40}
            tickFormatter={(v: number) => `${(v / 10000).toFixed(0)}`}
          />
          <Tooltip
            content={
              <ChartTooltip
                formatter={(v) => `${(v / 10000).toFixed(2)} 万元`}
              />
            }
          />
          <Bar
            dataKey="revenue"
            fill={C.blue}
            radius={[3, 3, 0, 0]}
            maxBarSize={30}
            name="营收"
            opacity={0.8}
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
