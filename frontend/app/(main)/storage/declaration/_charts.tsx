'use client';

// 储能申报图表（容量对比 Composed + 收益趋势 Bar + 充放电曲线 Bar）。
// 单独成文件以便 next/dynamic 懒加载，剥离 recharts 出首屏。
import {
  Bar,
  BarChart,
  CartesianGrid,
  ComposedChart,
  Legend,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

export function CapacityComposed({ data }: { data: any[] }) {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <ComposedChart
        data={data}
        margin={{ top: 8, right: 16, bottom: 8, left: 8 }}
      >
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
        <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
        <YAxis yAxisId="cap" tick={{ fontSize: 11 }} width={60} unit=" MWh" />
        <YAxis yAxisId="rate" orientation="right" tick={{ fontSize: 11 }} width={60} unit="%" />
        <Tooltip
          contentStyle={{ fontSize: 12 }}
          formatter={(v: number, name: string) => {
            if (name === '利用率') return `${v.toFixed(1)}%`;
            return `${v.toFixed(1)} MWh`;
          }}
        />
        <Legend wrapperStyle={{ fontSize: 12 }} />
        <Bar
          yAxisId="cap"
          dataKey="declared"
          name="申报容量"
          fill="#3b82f6"
          isAnimationActive={false}
        />
        <Bar
          yAxisId="cap"
          dataKey="actual"
          name="实际可用"
          fill="#10b981"
          isAnimationActive={false}
        />
        <Line
          yAxisId="rate"
          type="monotone"
          dataKey="utilization"
          name="利用率"
          stroke="#f97316"
          strokeWidth={2}
          dot={{ r: 3 }}
          isAnimationActive={false}
        />
      </ComposedChart>
    </ResponsiveContainer>
  );
}

export function RevenueBar({ data }: { data: any[] }) {
  return (
    <ResponsiveContainer width="100%" height={200}>
      <BarChart
        data={data}
        margin={{ top: 8, right: 16, bottom: 8, left: 8 }}
      >
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
        <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
        <YAxis tick={{ fontSize: 11 }} width={60} />
        <Tooltip
          formatter={(v: number) => `${v.toFixed(2)} 万`}
          contentStyle={{ fontSize: 12 }}
        />
        <Bar
          dataKey="expected"
          name="预测收益"
          fill="#8b5cf6"
          isAnimationActive={false}
        />
      </BarChart>
    </ResponsiveContainer>
  );
}

export function ChargeDischargeBar({ data }: { data: any[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart data={data} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
        <XAxis dataKey="period" tick={{ fontSize: 10, fill: '#6b7280' }} interval={11} />
        <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} unit=" MW" />
        <Tooltip
          formatter={(v: number) => `${Math.abs(v).toFixed(1)} MW`}
          contentStyle={{ fontSize: 12 }}
        />
        <Bar
          dataKey="charge"
          name="充电（负方向）"
          fill="#3b82f6"
          isAnimationActive={false}
        />
        <Bar
          dataKey="discharge"
          name="放电"
          fill="#f59e0b"
          isAnimationActive={false}
        />
      </BarChart>
    </ResponsiveContainer>
  );
}
