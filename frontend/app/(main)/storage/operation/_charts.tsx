'use client';

// 储能运行图表（SOC 曲线 + 充放电功率 + 30 日充放电与收益）。
// 单独成文件以便 next/dynamic 懒加载，剥离 recharts 出首屏。
import {
  Area,
  AreaChart,
  Bar,
  CartesianGrid,
  ComposedChart,
  Legend,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

export function SocArea({ data }: { data: any[] }) {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <AreaChart data={data} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis dataKey="hour" tick={{ fontSize: 11 }} />
        <YAxis tick={{ fontSize: 11 }} width={50} domain={[0, 100]} unit="%" />
        <Tooltip
          contentStyle={{ fontSize: 12 }}
          formatter={(v: number) => `${v.toFixed(1)}%`}
        />
        <Area
          type="monotone"
          dataKey="soc"
          name="SOC"
          stroke="#3b82f6"
          fill="#dbeafe"
          strokeWidth={2}
          isAnimationActive={false}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}

export function PowerComposed({ data }: { data: any[] }) {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <ComposedChart data={data} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis dataKey="hour" tick={{ fontSize: 11 }} />
        <YAxis tick={{ fontSize: 11 }} width={60} unit=" MW" />
        <Tooltip
          contentStyle={{ fontSize: 12 }}
          formatter={(v: number) => `${v.toFixed(1)} MW`}
        />
        <Legend wrapperStyle={{ fontSize: 12 }} />
        {/* 正区域（放电）着色 */}
        <Area
          type="monotone"
          dataKey="power"
          name="功率"
          stroke="#f97316"
          fill="#f97316"
          fillOpacity={0.3}
          strokeWidth={2}
          isAnimationActive={false}
          baseValue={0}
        />
        {/* 零线 */}
        <Line
          type="monotone"
          dataKey={() => 0}
          stroke="#6b7280"
          strokeWidth={1}
          strokeDasharray="3 3"
          dot={false}
          isAnimationActive={false}
          name=""
        />
      </ComposedChart>
    </ResponsiveContainer>
  );
}

export function DailyComposed({ data }: { data: any[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <ComposedChart
        data={data}
        margin={{ top: 8, right: 16, bottom: 8, left: 8 }}
      >
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis dataKey="date" interval={3} tick={{ fontSize: 11 }} />
        <YAxis yAxisId="left" tick={{ fontSize: 11 }} width={56} />
        <YAxis
          yAxisId="right"
          orientation="right"
          tick={{ fontSize: 11 }}
          width={70}
        />
        <Tooltip />
        <Legend wrapperStyle={{ fontSize: 12 }} />
        <Bar
          yAxisId="left"
          dataKey="充电"
          fill="#06b6d4"
          isAnimationActive={false}
        />
        <Bar
          yAxisId="left"
          dataKey="放电"
          fill="#f97316"
          isAnimationActive={false}
        />
        <Line
          yAxisId="right"
          type="monotone"
          dataKey="收益"
          stroke="#8b5cf6"
          strokeWidth={2}
          dot={false}
          isAnimationActive={false}
        />
      </ComposedChart>
    </ResponsiveContainer>
  );
}
