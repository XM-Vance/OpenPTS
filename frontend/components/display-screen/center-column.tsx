import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  LineChart,
  Line,
} from 'recharts';
import type { DisplayTrendItem } from '@/lib/api/display-screen';
import { C } from './constants';
import { ChartTooltip } from './chart-tooltip';
import { FujianMap } from './fujian-map';

interface CenterColumnProps {
  trend: DisplayTrendItem[];
}

export function CenterColumn({ trend }: CenterColumnProps) {
  const energyData = trend.map((d) => ({
    date: d.date.slice(5),
    energy: d.energy_mwh,
    revenue: d.revenue,
  }));
  const priceData = trend.map((d) => ({
    date: d.date.slice(5),
    price: d.avg_price,
  }));

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
      {/* 福建省电力地图 */}
      <div
        style={{
          background: C.cardBg,
          border: `1px solid ${C.cardBorder}`,
          borderRadius: 8,
          padding: '10px 14px 4px',
          backdropFilter: 'blur(8px)',
          flex: 0.55,
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            fontSize: 12,
            color: C.blue,
            fontWeight: 600,
            letterSpacing: 1,
            marginBottom: 4,
          }}
        >
          福建省电力流向图
        </div>
        <div style={{ width: '100%', height: 'calc(100% - 22px)' }}>
          <FujianMap />
        </div>
      </div>

      {/* 用量与营收趋势 */}
      <div
        style={{
          background: C.cardBg,
          border: `1px solid ${C.cardBorder}`,
          borderRadius: 8,
          padding: '10px 14px 4px',
          backdropFilter: 'blur(8px)',
          flex: 0.65,
        }}
      >
        <div
          style={{
            fontSize: 12,
            color: C.blue,
            fontWeight: 600,
            letterSpacing: 1,
            marginBottom: 4,
          }}
        >
          用电量与营收趋势
        </div>
        <ResponsiveContainer width="100%" height="88%">
          <AreaChart data={energyData}>
            <defs>
              <linearGradient id="energyGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={C.cyan} stopOpacity={0.3} />
                <stop offset="100%" stopColor={C.cyan} stopOpacity={0} />
              </linearGradient>
              <linearGradient id="revenueGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={C.gold} stopOpacity={0.25} />
                <stop offset="100%" stopColor={C.gold} stopOpacity={0} />
              </linearGradient>
            </defs>
            <XAxis
              dataKey="date"
              stroke={C.textDim}
              tick={{ fill: C.textDim, fontSize: 10 }}
              axisLine={{ stroke: C.divider }}
              tickLine={false}
            />
            <YAxis
              yAxisId="left"
              stroke={C.textDim}
              tick={{ fill: C.textDim, fontSize: 9 }}
              axisLine={false}
              tickLine={false}
              width={40}
              tickFormatter={(v: number) => `${(v / 1000).toFixed(0)}k`}
            />
            <YAxis
              yAxisId="right"
              orientation="right"
              stroke={C.textDim}
              tick={{ fill: C.textDim, fontSize: 9 }}
              axisLine={false}
              tickLine={false}
              width={50}
              tickFormatter={(v: number) => `${(v / 1000000).toFixed(1)}m`}
            />
            <Tooltip
              content={
                <ChartTooltip
                  formatter={(v) => v.toLocaleString()}
                />
              }
            />
            <Area
              yAxisId="left"
              type="monotone"
              dataKey="energy"
              stroke={C.cyan}
              strokeWidth={2}
              fill="url(#energyGrad)"
              dot={false}
              name="用电量 (MWh)"
            />
            <Area
              yAxisId="right"
              type="monotone"
              dataKey="revenue"
              stroke={C.gold}
              strokeWidth={2}
              fill="url(#revenueGrad)"
              dot={false}
              name="营收 (元)"
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      {/* 电价趋势 */}
      <div
        style={{
          background: C.cardBg,
          border: `1px solid ${C.cardBorder}`,
          borderRadius: 8,
          padding: '10px 14px 4px',
          backdropFilter: 'blur(8px)',
          flex: 0.35,
        }}
      >
        <div
          style={{
            fontSize: 12,
            color: C.blue,
            fontWeight: 600,
            letterSpacing: 1,
            marginBottom: 2,
          }}
        >
          均价走势
        </div>
        <ResponsiveContainer width="100%" height="85%">
          <LineChart data={priceData}>
            <XAxis
              dataKey="date"
              stroke={C.textDim}
              tick={{ fill: C.textDim, fontSize: 9 }}
              axisLine={{ stroke: C.divider }}
              tickLine={false}
            />
            <YAxis
              stroke={C.textDim}
              tick={{ fill: C.textDim, fontSize: 9 }}
              axisLine={false}
              tickLine={false}
              width={35}
              domain={['auto', 'auto']}
            />
            <Tooltip
              content={<ChartTooltip formatter={(v) => `${v.toFixed(2)} 元/MWh`} />}
            />
            <Line
              type="monotone"
              dataKey="price"
              stroke={C.orange}
              strokeWidth={2.5}
              dot={false}
              activeDot={{ fill: C.orange, r: 4, stroke: C.text, strokeWidth: 2 }}
              name="均价"
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
